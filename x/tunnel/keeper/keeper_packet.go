package keeper

import (
	"fmt"

	sdkerrors "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	feedstypes "github.com/bandprotocol/chain/v3/x/feeds/types"
	"github.com/bandprotocol/chain/v3/x/tunnel/types"
)

// DeductBasePacketFee deducts the base packet fee from the fee payer of the packet
func (k Keeper) DeductBasePacketFee(ctx sdk.Context, feePayer sdk.AccAddress) error {
	basePacketFee := k.GetParams(ctx).BasePacketFee
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, feePayer, types.ModuleName, basePacketFee); err != nil {
		return err
	}

	// update total fees
	totalFees := k.GetTotalFees(ctx)
	totalFees.TotalBasePacketFee = totalFees.TotalBasePacketFee.Add(basePacketFee...)
	k.SetTotalFees(ctx, totalFees)
	return nil
}

// SetPacket sets a packet in the store
func (k Keeper) SetPacket(ctx sdk.Context, packet types.Packet) {
	ctx.KVStore(k.storeKey).
		Set(types.TunnelPacketStoreKey(packet.TunnelID, packet.Sequence), k.cdc.MustMarshal(&packet))
}

// GetPacket retrieves a packet by its tunnel ID and packet ID
func (k Keeper) GetPacket(ctx sdk.Context, tunnelID uint64, sequence uint64) (types.Packet, error) {
	bz := ctx.KVStore(k.storeKey).Get(types.TunnelPacketStoreKey(tunnelID, sequence))
	if bz == nil {
		return types.Packet{}, types.ErrPacketNotFound.Wrapf("tunnelID: %d, sequence: %d", tunnelID, sequence)
	}

	var packet types.Packet
	k.cdc.MustUnmarshal(bz, &packet)
	return packet, nil
}

// ProduceActiveTunnelPackets generates packets and sends packets to the destination route for all active tunnels
func (k Keeper) ProduceActiveTunnelPackets(ctx sdk.Context) error {
	// get active tunnel IDs
	ids := k.GetActiveTunnelIDs(ctx)

	prices := k.feedsKeeper.GetAllPrices(ctx)
	pricesMap := CreatePricesMap(prices)

	// create new packet. If failed to produce packet, emit an event.
	for _, id := range ids {
		if err := k.ProduceActiveTunnelPacket(ctx, id, pricesMap); err != nil {
			ctx.EventManager().EmitEvent(sdk.NewEvent(
				types.EventTypeProducePacketFail,
				sdk.NewAttribute(types.AttributeKeyTunnelID, fmt.Sprintf("%d", id)),
				sdk.NewAttribute(types.AttributeKeyReason, err.Error()),
			))
		}
	}

	return nil
}

// ProduceActiveTunnelPacket generates a packet and sends it to the destination route for the given tunnel ID.
// If not enough fund, deactivate the tunnel.
func (k Keeper) ProduceActiveTunnelPacket(
	ctx sdk.Context,
	tunnelID uint64,
	pricesMap map[string]feedstypes.Price,
) (err error) {
	// Check if the tunnel has enough fund to create a packet and deactivate the tunnel if not
	// enough fund. Error should not happen here since the tunnel is already validated.
	ok, err := k.HasEnoughFundToCreatePacket(ctx, tunnelID)
	if err != nil {
		return err
	}
	if !ok {
		return k.DeactivateTunnel(ctx, tunnelID)
	}

	// Produce a packet. If produce packet successfully, update the context state.
	cacheCtx, writeFn := ctx.CacheContext()
	if err := k.ProducePacket(cacheCtx, tunnelID, pricesMap); err != nil {
		return err
	}
	writeFn()

	return nil
}

// ProducePacket generates a packet and sends it to the destination route
func (k Keeper) ProducePacket(
	ctx sdk.Context,
	tunnelID uint64,
	feedsPricesMap map[string]feedstypes.Price,
) error {
	unixNow := ctx.BlockTime().Unix()

	// get tunnel and prices info
	tunnel, err := k.GetTunnel(ctx, tunnelID)
	if err != nil {
		return err
	}

	latestPrices, err := k.GetLatestPrices(ctx, tunnelID)
	if err != nil {
		return err
	}
	latestPricesMap := CreatePricesMap(latestPrices.Prices)

	// check if the interval has passed
	sendAll := unixNow >= int64(tunnel.Interval)+latestPrices.LastInterval

	// generate newPrices; if no newPrices, stop the process.
	newPrices := GenerateNewPrices(
		tunnel.SignalDeviations,
		latestPricesMap,
		feedsPricesMap,
		ctx.BlockTime().Unix(),
		sendAll,
	)
	if len(newPrices) == 0 {
		return nil
	}

	// create a new packet
	packet, err := k.CreatePacket(ctx, tunnel.ID, newPrices)
	if err != nil {
		return err
	}

	// send packet
	if err := k.SendPacket(ctx, packet); err != nil {
		return sdkerrors.Wrapf(err, "failed to send packet for tunnel %d", tunnel.ID)
	}

	// update latest price info.
	latestPrices.UpdatePrices(newPrices)
	if sendAll {
		latestPrices.LastInterval = unixNow
	}
	k.SetLatestPrices(ctx, latestPrices)

	// emit an event
	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeProducePacketSuccess,
		sdk.NewAttribute(types.AttributeKeyTunnelID, fmt.Sprintf("%d", tunnel.ID)),
		sdk.NewAttribute(types.AttributeKeySequence, fmt.Sprintf("%d", packet.Sequence)),
	))

	return nil
}

// CreatePacket creates a new packet of the the given tunnel. Creating a packet charges
// the base packet fee to the tunnel's fee payer.
func (k Keeper) CreatePacket(
	ctx sdk.Context,
	tunnelID uint64,
	prices []feedstypes.Price,
) (types.Packet, error) {
	// get tunnel and prices info
	params := k.GetParams(ctx)

	tunnel, err := k.GetTunnel(ctx, tunnelID)
	if err != nil {
		return types.Packet{}, err
	}

	// deduct base packet fee from the fee payer,
	feePayer := sdk.MustAccAddressFromBech32(tunnel.FeePayer)
	if err := k.DeductBasePacketFee(ctx, feePayer); err != nil {
		return types.Packet{}, sdkerrors.Wrapf(err, "failed to deduct base packet fee for tunnel %d", tunnel.ID)
	}

	// get the route
	route, err := tunnel.GetRouteValue()
	if err != nil {
		return types.Packet{}, err
	}

	// get the route fee
	routeFee, err := k.GetRouteFee(ctx, route)
	if err != nil {
		return types.Packet{}, err
	}

	tunnel.Sequence++
	packet := types.NewPacket(
		tunnelID,
		tunnel.Sequence,
		prices,
		params.BasePacketFee,
		routeFee,
		ctx.BlockTime().Unix(),
	)

	// update information in the store
	k.SetTunnel(ctx, tunnel)
	k.SetPacket(ctx, packet)

	return packet, nil
}

// SendPacket sends a packet to the destination route
func (k Keeper) SendPacket(ctx sdk.Context, packet types.Packet) (err error) {
	defer func() {
		if r := recover(); r != nil {
			ctx.Logger().Error(fmt.Sprintf("Panic recovered: %v", r))
			err = types.ErrSendPacketPanic
			return
		}
	}()

	tunnel, err := k.GetTunnel(ctx, packet.TunnelID)
	if err != nil {
		return err
	}

	// get the route
	route, err := tunnel.GetRouteValue()
	if err != nil {
		return err
	}

	// send packet to the destination route and get the route result
	var receipt types.PacketReceiptI
	switch r := route.(type) {
	case *types.TSSRoute:
		receipt, err = k.SendTSSPacket(
			ctx,
			r,
			packet,
			sdk.MustAccAddressFromBech32(tunnel.FeePayer),
		)
	case *types.IBCRoute:
		receipt, err = k.SendIBCPacket(ctx, r, packet, tunnel.Interval)
	case *types.RouterRoute:
		receipt, err = k.SendRouterPacket(ctx, r, packet, sdk.MustAccAddressFromBech32(tunnel.FeePayer), tunnel.Interval)
	default:
		return types.ErrInvalidRoute.Wrapf("no route found for tunnel ID: %d", tunnel.ID)
	}
	// return error if failed to send packet
	if err != nil {
		return err
	}

	// set the receipt to the packet
	if err := packet.SetReceipt(receipt); err != nil {
		return sdkerrors.Wrapf(
			err,
			"failed to set packet receipt for tunnel %d, sequence %d",
			packet.TunnelID,
			packet.Sequence,
		)
	}

	k.SetPacket(ctx, packet)

	return nil
}
