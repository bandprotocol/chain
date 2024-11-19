package keeper

import (
	"fmt"
	"math"

	sdkerrors "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/pkg/ctxcache"
	feedstypes "github.com/bandprotocol/chain/v3/x/feeds/types"
	"github.com/bandprotocol/chain/v3/x/tunnel/types"
)

// DeductBaseFee deducts the base fee from fee payer's account.
func (k Keeper) DeductBasePacketFee(ctx sdk.Context, feePayer sdk.AccAddress) error {
	basePacketFee := k.GetParams(ctx).BasePacketFee
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, feePayer, types.ModuleName, basePacketFee); err != nil {
		return err
	}

	// update total fees
	totalFees := k.GetTotalFees(ctx)
	totalFees.TotalPacketFee = totalFees.TotalPacketFee.Add(basePacketFee...)
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

// MustGetPacket retrieves a packet by its tunnel ID and packet ID and panics if the packet does not exist
func (k Keeper) MustGetPacket(ctx sdk.Context, tunnelID uint64, sequence uint64) types.Packet {
	packet, err := k.GetPacket(ctx, tunnelID, sequence)
	if err != nil {
		panic(err)
	}
	return packet
}

// ProduceActiveTunnelPackets generates packets and sends packets to the destination route for all active tunnels
func (k Keeper) ProduceActiveTunnelPackets(ctx sdk.Context) {
	// get active tunnel IDs
	ids := k.GetActiveTunnelIDs(ctx)

	prices := k.feedsKeeper.GetAllPrices(ctx)
	pricesMap := CreatePricesMap(prices)

	// create new packet if possible for active tunnels. If not enough fund, deactivate the tunnel.
	for _, id := range ids {
		ok, err := k.HasEnoughFundToCreatePacket(ctx, id)
		if err != nil {
			continue
		}
		if !ok {
			k.MustDeactivateTunnel(ctx, id)
			continue
		}

		producePacketFunc := func(ctx sdk.Context) error {
			return k.ProducePacket(ctx, id, pricesMap)
		}

		// Produce a packet. If error, emits an event.
		if err := ctxcache.ApplyFuncIfNoError(ctx, producePacketFunc); err != nil {
			ctx.EventManager().EmitEvent(sdk.NewEvent(
				types.EventTypeProducePacketFail,
				sdk.NewAttribute(types.AttributeKeyTunnelID, fmt.Sprintf("%d", id)),
				sdk.NewAttribute(types.AttributeKeyReason, err.Error()),
			))
		}
	}
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
	newPrices, err := GenerateNewPrices(tunnel.SignalDeviations, latestPricesMap, feedsPricesMap, sendAll)
	if err != nil {
		return err
	}
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
		return sdkerrors.Wrapf(err, "failed to create packet for tunnel %d", tunnel.ID)
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

	// get the route fee
	route, ok := tunnel.Route.GetCachedValue().(types.RouteI)
	if !ok {
		return types.Packet{}, types.ErrInvalidRoute
	}

	routeFee, err := k.GetRouterFee(ctx, route)
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
func (k Keeper) SendPacket(ctx sdk.Context, packet types.Packet) error {
	tunnel, err := k.GetTunnel(ctx, packet.TunnelID)
	if err != nil {
		return err
	}

	// get the packet content, which is the information receiving after
	// sending packet to the destination route
	var content types.PacketContentI
	switch r := tunnel.Route.GetCachedValue().(type) {
	case *types.TSSRoute:
		content, err = k.SendTSSPacket(
			ctx,
			r,
			packet,
			sdk.MustAccAddressFromBech32(tunnel.FeePayer),
		)
	default:
		return types.ErrInvalidRoute.Wrapf("no route found for tunnel ID: %d", tunnel.ID)
	}
	// return error if failed to send packet
	if err != nil {
		return err
	}

	// set the packet content
	if err := packet.SetPacketContent(content); err != nil {
		return sdkerrors.Wrapf(err, "failed to set packet content for tunnel ID: %d", tunnel.ID)
	}

	k.SetPacket(ctx, packet)

	return nil
}

// calculateDeviationBPS calculates the deviation between the old price and
// the new price in basis points, i.e., |(newPrice - oldPrice)| * 10000 / oldPrice
func calculateDeviationBPS(oldPrice, newPrice sdkmath.Int) sdkmath.Int {
	if newPrice.Equal(oldPrice) {
		return sdkmath.ZeroInt()
	}

	if oldPrice.IsZero() {
		return sdkmath.NewInt(math.MaxInt64)
	}

	return newPrice.Sub(oldPrice).Abs().MulRaw(10000).Quo(oldPrice)
}

// CreatePricesMap creates a map of prices with signal ID as the key
func CreatePricesMap(prices []feedstypes.Price) map[string]feedstypes.Price {
	pricesMap := make(map[string]feedstypes.Price, len(prices))
	for _, p := range prices {
		pricesMap[p.SignalID] = p
	}
	return pricesMap
}
