package keeper

import (
	"fmt"
	"math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	feedstypes "github.com/bandprotocol/chain/v2/x/feeds/types"
	"github.com/bandprotocol/chain/v2/x/tunnel/types"
)

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
	// get all active tunnels
	activeTunnels := k.GetTunnelsByActiveStatus(ctx, true)

	// check for active tunnels
	for _, at := range activeTunnels {
		signalPricesInfo, err := k.GetSignalPricesInfo(ctx, at.ID)
		if err != nil {
			// emit get signal prices info fail event
			ctx.EventManager().EmitEvent(sdk.NewEvent(
				types.EventTypeGetSignalPricesInfoFail,
				sdk.NewAttribute(types.AttributeKeyTunnelID, fmt.Sprintf("%d", at.ID)),
				sdk.NewAttribute(types.AttributeKeyReason, err.Error()),
			))
			continue
		}

		// check if the interval has passed
		intervalTrigger := ctx.BlockTime().Unix() >= int64(at.Interval)+signalPricesInfo.LastIntervalTimestamp

		// produce packet
		err = k.ProducePacket(ctx, at, signalPricesInfo, intervalTrigger)
		if err != nil {
			// emit send packet fail event
			ctx.EventManager().EmitEvent(sdk.NewEvent(
				types.EventTypeProducePacketFail,
				sdk.NewAttribute(types.AttributeKeyTunnelID, fmt.Sprintf("%d", at.ID)),
				sdk.NewAttribute(types.AttributeKeyRoute, fmt.Sprintf("%v", at.Route)),
				sdk.NewAttribute(types.AttributeKeyReason, err.Error()),
			))
			continue
		}
	}
}

// ProducePacket generates a packet and sends it to the destination route
func (k Keeper) ProducePacket(
	ctx sdk.Context,
	tunnel types.Tunnel,
	signalPricesInfo types.SignalPricesInfo,
	triggerAll bool,
) error {
	unixNow := ctx.BlockTime().Unix()

	// TODO: feeds module needs to be implemented get prices that can use
	latestPrices := k.feedsKeeper.GetPrices(ctx)
	latestPricesMap := createLatestPricesMap(latestPrices)

	// generate new signal prices
	nsps := GenerateSignalPrices(
		ctx,
		tunnel.ID,
		latestPricesMap,
		tunnel.GetSignalInfoMap(),
		signalPricesInfo.SignalPrices,
		triggerAll,
	)
	if len(nsps) > 0 {
		err := k.SendPacket(ctx, tunnel, types.NewPacket(tunnel.ID, tunnel.NonceCount+1, nsps, nil, unixNow))
		if err != nil {
			return err
		}

		// update signal prices info
		signalPricesInfo.UpdateSignalPrices(nsps)
		if triggerAll {
			signalPricesInfo.LastIntervalTimestamp = unixNow
		}
		k.SetSignalPricesInfo(ctx, signalPricesInfo)

		// update tunnel nonce count
		tunnel.NonceCount++
		k.SetTunnel(ctx, tunnel)
	}

	return nil
}

// SendPacket sends a packet to the destination route
func (k Keeper) SendPacket(
	ctx sdk.Context,
	tunnel types.Tunnel,
	packet types.Packet,
) error {
	// Process the tunnel based on the route type
	switch r := tunnel.Route.GetCachedValue().(type) {
	case *types.TSSRoute:
		err := k.TSSPacketHandle(ctx, r, packet)
		if err != nil {
			return err
		}
	case *types.AxelarRoute:
		err := k.AxelarPacketHandle(ctx, r, packet)
		if err != nil {
			return err
		}
	default:
		panic(fmt.Sprintf("unknown route type: %T", r))
	}
	return nil
}

// GenerateSignalPrices generates signal prices based on the latest prices and signal info
func GenerateSignalPrices(
	ctx sdk.Context,
	tunnelID uint64,
	latestPricesMap map[string]feedstypes.Price,
	signalInfoMap map[string]types.SignalInfo,
	signalPrices []types.SignalPrice,
	triggerAll bool,
) []types.SignalPrice {
	var sps []types.SignalPrice
	for _, sp := range signalPrices {
		latestPrice, exists := latestPricesMap[sp.SignalID]
		// TODO: remove check PriceStatusAvailable when feeds module is implemented
		if !exists || latestPrice.PriceStatus != feedstypes.PriceStatusAvailable {
			sps = append(sps, types.NewSignalPrice(sp.SignalID, 0, 0))
			ctx.EventManager().EmitEvent(sdk.NewEvent(
				types.EventTypeSignalIDNotFound,
				sdk.NewAttribute(types.AttributeKeyTunnelID, fmt.Sprintf("%d", tunnelID)),
				sdk.NewAttribute(types.AttributeKeySignalID, sp.SignalID),
			))
			continue
		}

		// get signal info from signalInfoMap
		signalInfo, exists := signalInfoMap[sp.SignalID]
		if !exists {
			ctx.EventManager().EmitEvent(sdk.NewEvent(
				types.EventTypeSignalInfoNotFound,
				sdk.NewAttribute(types.AttributeKeyTunnelID, fmt.Sprintf("%d", tunnelID)),
				sdk.NewAttribute(types.AttributeKeySignalID, sp.SignalID),
			))
			continue
		}

		// if triggerAll is true or the deviation exceeds the threshold, add the signal price info to the list
		if triggerAll || deviationExceedsThreshold(sp.Price, latestPrice.Price, signalInfo.HardDeviationBPS) {
			sps = append(
				sps,
				types.NewSignalPrice(
					sp.SignalID,
					latestPrice.Price,
					latestPrice.Timestamp,
				),
			)
		}
	}
	return sps
}

// deviationExceedsThreshold checks if the deviation between the old price and the new price exceeds the threshold
func deviationExceedsThreshold(oldPrice, newPrice uint64, thresholdBPS uint64) bool {
	// if the deviation is greater than the hard deviation, add the signal price info to the list
	// soft deviation is the feature to be implemented in the future
	deviation := math.Abs(float64(newPrice-oldPrice)) / float64(oldPrice)
	deviationInBPS := uint64(deviation * 10000)
	return deviationInBPS >= thresholdBPS
}

// createLatestPricesMap creates a map of latest prices with signal ID as the key
func createLatestPricesMap(latestPrices []feedstypes.Price) map[string]feedstypes.Price {
	latestPricesMap := make(map[string]feedstypes.Price, len(latestPrices))
	for _, price := range latestPrices {
		latestPricesMap[price.SignalID] = price
	}
	return latestPricesMap
}
