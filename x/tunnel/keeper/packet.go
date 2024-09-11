package keeper

import (
	"fmt"

	sdkerrors "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/pkg/ctxcache"
	feedstypes "github.com/bandprotocol/chain/v2/x/feeds/types"
	"github.com/bandprotocol/chain/v2/x/tunnel/types"
)

// DeductBaseFee deducts the base fee from fee payer's account.
func (k Keeper) DeductBasePacketFee(ctx sdk.Context, feePayer sdk.AccAddress) error {
	basePacketFee := k.GetParams(ctx).BasePacketFee
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, feePayer, types.ModuleName, basePacketFee); err != nil {
		return err
	}

	// update total fee
	totalFee := k.GetTotalFee(ctx)
	totalFee.TotalPacketFee.Add(basePacketFee...)
	k.SetTotalFee(ctx, totalFee)
	return nil
}

// SetPacket sets a packet in the store
func (k Keeper) SetPacket(ctx sdk.Context, packet types.Packet) {
	ctx.KVStore(k.storeKey).
		Set(types.TunnelPacketStoreKey(packet.TunnelID, packet.Nonce), k.cdc.MustMarshal(&packet))
}

// GetPacket retrieves a packet by its tunnel ID and packet ID
func (k Keeper) GetPacket(ctx sdk.Context, tunnelID uint64, nonce uint64) (types.Packet, error) {
	bz := ctx.KVStore(k.storeKey).Get(types.TunnelPacketStoreKey(tunnelID, nonce))
	if bz == nil {
		return types.Packet{}, types.ErrPacketNotFound.Wrapf("tunnelID: %d, nonce: %d", tunnelID, nonce)
	}

	var packet types.Packet
	k.cdc.MustUnmarshal(bz, &packet)
	return packet, nil
}

// MustGetPacket retrieves a packet by its tunnel ID and packet ID and panics if the packet does not exist
func (k Keeper) MustGetPacket(ctx sdk.Context, tunnelID uint64, nonce uint64) types.Packet {
	packet, err := k.GetPacket(ctx, tunnelID, nonce)
	if err != nil {
		panic(err)
	}
	return packet
}

// ProduceActiveTunnelPackets generates packets and sends packets to the destination route for all active tunnels
func (k Keeper) ProduceActiveTunnelPackets(ctx sdk.Context) {
	// get active tunnel IDs
	ids := k.GetActiveTunnelIDs(ctx)

	currentPrices := k.feedsKeeper.GetCurrentPrices(ctx)
	currentPricesMap := createCurrentPricesMap(currentPrices)

	// check for active tunnels
	for _, id := range ids {
		isErrNoPacketCreated := false
		producePacketFunc := func(ctx sdk.Context) error {
			tunnel, err := k.GetTunnel(ctx, id)
			if err != nil {
				return err
			}

			// deduct base packet fee from the fee payer; deactivate tunnel if failed.
			feePayerAddr := sdk.MustAccAddressFromBech32(tunnel.FeePayer)
			if err := k.DeductBasePacketFee(ctx, feePayerAddr); err != nil {
				return k.DeactivateTunnel(ctx, id)
			}

			// Produce and send a packet, if no packet is created, return error so that
			// fee is reverted.
			isCreated, err := k.ProducePacket(ctx, id, currentPricesMap, false)
			if err != nil {
				return err
			}
			if !isCreated {
				isErrNoPacketCreated = true
				return fmt.Errorf("no packet is created for tunnel %d", id)
			}

			return nil
		}

		// produce a packet. If error, emits an event.
		if err := ctxcache.ApplyFuncIfNoError(ctx, producePacketFunc); err != nil && !isErrNoPacketCreated {
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
	triggerAll bool,
) (isCreated bool, err error) {
	unixNow := ctx.BlockTime().Unix()

	// get tunnel and signal prices info
	tunnel := k.MustGetTunnel(ctx, tunnelID)
	signalPricesInfo := k.MustGetSignalPricesInfo(ctx, tunnelID)

	// check if the interval has passed
	intervalTrigger := ctx.BlockTime().Unix() >= int64(tunnel.Interval)+signalPricesInfo.LastIntervalTimestamp

	// generate new signal prices
	nsps := GenerateSignalPrices(
		ctx,
		tunnel.ID,
		currentPricesMap,
		tunnel.GetSignalInfoMap(),
		signalPricesInfo.SignalPrices,
		triggerAll || intervalTrigger,
	)

	// return if no new signal prices
	if len(nsps) == 0 {
		return false, nil
	}

	newPacket := types.NewPacket(tunnel.ID, tunnel.NonceCount+1, nsps, nil, unixNow)
	if err := k.SendPacket(ctx, tunnel, newPacket); err != nil {
		return false, sdkerrors.Wrapf(err, "route %s failed to send packet", tunnel.Route.TypeUrl)
	}

	// update signal prices info
	signalPricesInfo.UpdateSignalPrices(nsps)
	if triggerAll || intervalTrigger {
		signalPricesInfo.LastIntervalTimestamp = unixNow
	}
	k.SetSignalPricesInfo(ctx, signalPricesInfo)

	// update tunnel nonce count
	tunnel.NonceCount++
	k.SetTunnel(ctx, tunnel)

	return true, nil
}

// SendPacket sends a packet to the destination route
func (k Keeper) SendPacket(
	ctx sdk.Context,
	tunnel types.Tunnel,
	packet types.Packet,
) error {
	var content types.PacketContentI
	var err error

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
	err = packet.SetPacketContent(content)
	if err != nil {
		panic(fmt.Sprintf("failed to set packet content: %s", err))
	}

	// set the packet in the store
	k.SetPacket(ctx, packet)
	return nil
}

// GenerateSignalPrices generates signal prices based on the current prices and signal info
func GenerateSignalPrices(
	ctx sdk.Context,
	tunnelID uint64,
	currentPricesMap map[string]feedstypes.Price,
	signalInfoMap map[string]types.SignalInfo,
	signalPrices []types.SignalPrice,
	triggerAll bool,
) []types.SignalPrice {
	var sps []types.SignalPrice
	for _, sp := range signalPrices {
		currentPrice, exists := currentPricesMap[sp.SignalID]
		if !exists || currentPrice.PriceStatus != feedstypes.PriceStatusAvailable {
			sps = append(sps, types.NewSignalPrice(sp.SignalID, 0))
			continue
		}

		// get signal info from signalInfoMap
		signalInfo, exists := signalInfoMap[sp.SignalID]
		if !exists {
			// panic if signal info not found for signal ID in the tunnel that should not happen
			panic(fmt.Sprintf("signal info not found for signal ID: %s", sp.SignalID))
		}

		// if triggerAll is true or the deviation exceeds the threshold, add the signal price info to the list
		if triggerAll ||
			deviationExceedsThreshold(
				sdk.NewIntFromUint64(sp.Price),
				sdk.NewIntFromUint64(currentPrice.Price),
				sdk.NewIntFromUint64(signalInfo.HardDeviationBPS),
			) {
			sps = append(
				sps,
				types.NewSignalPrice(
					sp.SignalID,
					currentPrice.Price,
				),
			)
		}
	}
	return sps
}

// deviationExceedsThreshold checks if the deviation between the old price and the new price exceeds the threshold
func deviationExceedsThreshold(oldPrice, newPrice, thresholdBPS sdkmath.Int) bool {
	// if the deviation is greater than the hard deviation, add the signal price info to the list
	// soft deviation is the feature to be implemented in the future
	deviation := newPrice.Sub(oldPrice).Abs().Quo(oldPrice)

	deviationInBPS := deviation.MulRaw(10000)
	return deviationInBPS.GTE(thresholdBPS)
}

// createCurrentPricesMap creates a map of current prices with signal ID as the key
func createCurrentPricesMap(latestPrices []feedstypes.Price) map[string]feedstypes.Price {
	latestPricesMap := make(map[string]feedstypes.Price, len(latestPrices))
	for _, price := range latestPrices {
		latestPricesMap[price.SignalID] = price
	}
	return latestPricesMap
}
