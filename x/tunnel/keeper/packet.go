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

	currentPrices := k.feedsKeeper.GetAllCurrentPrices(ctx)
	currentPricesMap := createPricesMap(currentPrices)

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
			return k.ProducePacket(ctx, id, currentPricesMap, false)
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
	currentPricesMap map[string]feedstypes.Price,
	sendAll bool,
) error {
	unixNow := ctx.BlockTime().Unix()

	// get tunnel and signal prices info
	tunnel, err := k.GetTunnel(ctx, tunnelID)
	if err != nil {
		return err
	}
	latestSignalPrices, err := k.GetLatestSignalPrices(ctx, tunnelID)
	if err != nil {
		return err
	}

	// check if the interval has passed
	isIntervalReached := unixNow >= int64(tunnel.Interval)+latestSignalPrices.LastIntervalTimestamp
	sendAll = sendAll || isIntervalReached

	// generate new signal prices; if no new signal prices, stop the process.
	newSignalPrices, err := k.GenerateNewSignalPrices(ctx, tunnelID, currentPricesMap, sendAll)
	if err != nil {
		return err
	}
	if len(newSignalPrices) == 0 {
		return nil
	}

	// create a new packet
	packet, err := k.CreatePacket(ctx, tunnel.ID, newSignalPrices)
	if err != nil {
		return err
	}

	// send packet
	if err := k.SendPacket(ctx, packet); err != nil {
		return sdkerrors.Wrapf(err, "failed to create packet for tunnel %d", tunnel.ID)
	}

	// update latest price info.
	if err := k.UpdatePriceTunnel(ctx, tunnel.ID, newSignalPrices); err != nil {
		return err
	}
	if sendAll {
		if err := k.UpdateLastInterval(ctx, tunnel.ID, unixNow); err != nil {
			return err
		}
	}

	return nil
}

// CreatePacket creates a new packet of the the given tunnel. Creating a packet charges
// the base packet fee to the tunnel's fee payer.
func (k Keeper) CreatePacket(
	ctx sdk.Context,
	tunnelID uint64,
	signalPrices []types.SignalPrice,
) (types.Packet, error) {
	tunnel, err := k.GetTunnel(ctx, tunnelID)
	if err != nil {
		return types.Packet{}, err
	}

	// deduct base packet fee from the fee payer,
	feePayer := sdk.MustAccAddressFromBech32(tunnel.FeePayer)
	if err := k.DeductBasePacketFee(ctx, feePayer); err != nil {
		return types.Packet{}, sdkerrors.Wrapf(err, "failed to deduct base packet fee for tunnel %d", tunnel.ID)
	}

	tunnel.Sequence++
	packet := types.NewPacket(tunnelID, tunnel.Sequence, signalPrices, ctx.BlockTime().Unix())

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
		content, err = k.SendTSSPacket(ctx, r, packet)
	case *types.AxelarRoute:
		content, err = k.SendAxelarPacket(ctx, r, packet)
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

// GenerateNewSignalPrices generates new signal prices based on the current prices
// and signal deviations.
func (k Keeper) GenerateNewSignalPrices(
	ctx sdk.Context,
	tunnelID uint64,
	currentFeedsPricesMap map[string]feedstypes.Price,
	sendAll bool,
) ([]types.SignalPrice, error) {
	// get deviation info from the tunnel
	tunnel, err := k.GetTunnel(ctx, tunnelID)
	if err != nil {
		return nil, err
	}
	signalDeviations := tunnel.GetSignalDeviationMap()

	// get latest signal prices
	latestSignalPrices, err := k.GetLatestSignalPrices(ctx, tunnelID)
	if err != nil {
		return nil, err
	}

	shouldSend := false
	newSignalPrices := make([]types.SignalPrice, 0)
	for _, sp := range latestSignalPrices.SignalPrices {
		oldPrice := sdkmath.NewIntFromUint64(sp.Price)

		// get current price from the feed, if not found, set price to 0
		price := uint64(0)
		feedPrice, ok := currentFeedsPricesMap[sp.SignalID]
		if ok && feedPrice.PriceStatus == feedstypes.PriceStatusAvailable {
			price = feedPrice.Price
		}
		newPrice := sdkmath.NewIntFromUint64(price)

		// get hard/soft deviation, panic if not found; should not happen.
		sd, ok := signalDeviations[sp.SignalID]
		if !ok {
			panic(fmt.Sprintf("deviation not found for signal ID: %s", sp.SignalID))
		}
		hardDeviation := sdkmath.NewIntFromUint64(sd.HardDeviationBPS)
		softDeviation := sdkmath.NewIntFromUint64(sd.SoftDeviationBPS)

		// calculate deviation between old price and new price and compare with the threshold.
		// shouldSend is set to true if sendAll is true or there is a signal whose deviation
		// is over the hard threshold.
		deviation := calculateDeviationBPS(oldPrice, newPrice)
		if sendAll || deviation.GTE(hardDeviation) {
			newSignalPrices = append(newSignalPrices, types.NewSignalPrice(sp.SignalID, price))
			shouldSend = true
		} else if deviation.GTE(softDeviation) {
			newSignalPrices = append(newSignalPrices, types.NewSignalPrice(sp.SignalID, price))
		}
	}

	if shouldSend {
		return newSignalPrices, nil
	} else {
		return []types.SignalPrice{}, nil
	}
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

// createPricesMap creates a map of prices with signal ID as the key
func createPricesMap(prices []feedstypes.Price) map[string]feedstypes.Price {
	pricesMap := make(map[string]feedstypes.Price, len(prices))
	for _, p := range prices {
		pricesMap[p.SignalID] = p
	}
	return pricesMap
}
