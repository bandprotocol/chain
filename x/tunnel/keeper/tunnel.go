package keeper

import (
	"fmt"
	"math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	feedsTypes "github.com/bandprotocol/chain/v2/x/feeds/types"
	"github.com/bandprotocol/chain/v2/x/tunnel/types"
)

// SetTunnelCount sets the tunnel count in the store
func (k Keeper) SetTunnelCount(ctx sdk.Context, count uint64) {
	ctx.KVStore(k.storeKey).Set(types.TunnelCountStoreKey, sdk.Uint64ToBigEndian(count))
}

// GetTunnelCount returns the current number of all tunnels ever existed
func (k Keeper) GetTunnelCount(ctx sdk.Context) uint64 {
	return sdk.BigEndianToUint64(ctx.KVStore(k.storeKey).Get(types.TunnelCountStoreKey))
}

// GetNextTunnelID increments the tunnel count and returns the current number of tunnels
func (k Keeper) GetNextTunnelID(ctx sdk.Context) uint64 {
	tunnelNumber := k.GetTunnelCount(ctx) + 1
	k.SetTunnelCount(ctx, tunnelNumber)
	return tunnelNumber
}

// SetTunnel sets a tunnel in the store
func (k Keeper) SetTunnel(ctx sdk.Context, tunnel types.Tunnel) {
	ctx.KVStore(k.storeKey).Set(types.TunnelStoreKey(tunnel.ID), k.cdc.MustMarshal(&tunnel))
}

// AddTunnel adds a tunnel to the store and returns the new tunnel ID
func (k Keeper) AddTunnel(ctx sdk.Context, tunnel types.Tunnel) (uint64, error) {
	tunnel.ID = k.GetNextTunnelID(ctx)

	// Generate a new tunnel account
	acc, err := k.GenerateAccount(ctx, fmt.Sprintf("%d", tunnel.ID))
	if err != nil {
		return 0, err
	}

	tunnel.FeePayer = acc.String()

	// Set the creation time
	tunnel.CreatedAt = ctx.BlockTime().Unix()

	k.SetTunnel(ctx, tunnel)
	return tunnel.ID, nil
}

// GetTunnel retrieves a tunnel by its ID
func (k Keeper) GetTunnel(ctx sdk.Context, id uint64) (types.Tunnel, error) {
	bz := ctx.KVStore(k.storeKey).Get(types.TunnelStoreKey(id))
	if bz == nil {
		return types.Tunnel{}, types.ErrTunnelNotFound.Wrapf("tunnelID: %d", id)
	}

	var tunnel types.Tunnel
	k.cdc.MustUnmarshal(bz, &tunnel)
	return tunnel, nil
}

// MustGetTunnel retrieves a tunnel by its ID. Panics if the tunnel does not exist.
func (k Keeper) MustGetTunnel(ctx sdk.Context, id uint64) types.Tunnel {
	tunnel, err := k.GetTunnel(ctx, id)
	if err != nil {
		panic(err)
	}
	return tunnel
}

// GetTunnels returns all tunnels
func (k Keeper) GetTunnels(ctx sdk.Context) []types.Tunnel {
	var tunnels []types.Tunnel
	iterator := sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.TunnelStoreKeyPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var tunnel types.Tunnel
		k.cdc.MustUnmarshal(iterator.Value(), &tunnel)
		tunnels = append(tunnels, tunnel)
	}
	return tunnels
}

// GetActiveTunnels returns all active tunnels
func (k Keeper) GetActiveTunnels(ctx sdk.Context) []types.Tunnel {
	var tunnels []types.Tunnel
	iterator := sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.TunnelStoreKeyPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var tunnel types.Tunnel
		k.cdc.MustUnmarshal(iterator.Value(), &tunnel)

		if tunnel.IsActive {
			tunnels = append(tunnels, tunnel)
		}
	}
	return tunnels
}

// AddPendingTriggerTunnel adds the tunnel ID to the list of pending trigger tunnels
func (k Keeper) AddPendingTriggerTunnel(ctx sdk.Context, id uint64) {
	pendingList := k.GetPendingTriggerTunnels(ctx)
	pendingList = append(pendingList, id)
	k.SetPendingTriggerTunnels(ctx, pendingList)
}

// SetPendingTriggerTunnels saves the list of pending trigger tunnels that will be executed at the end of the block.
func (k Keeper) SetPendingTriggerTunnels(ctx sdk.Context, ids []uint64) {
	bz := k.cdc.MustMarshal(&types.PendingTriggerTunnels{IDs: ids})
	if bz == nil {
		bz = []byte{}
	}
	ctx.KVStore(k.storeKey).Set(types.PendingTriggerTunnelsStoreKey, bz)
}

// GetPendingTriggerTunnels returns the list of pending trigger tunnels to be executed during EndBlock.
func (k Keeper) GetPendingTriggerTunnels(ctx sdk.Context) (ids []uint64) {
	bz := ctx.KVStore(k.storeKey).Get(types.PendingTriggerTunnelsStoreKey)
	if len(bz) == 0 { // Return an empty list if the key does not exist in the store.
		return []uint64{}
	}
	pendingTriggerTunnels := types.PendingTriggerTunnels{}
	k.cdc.MustUnmarshal(bz, &pendingTriggerTunnels)
	return pendingTriggerTunnels.IDs
}

// ActivateTunnel activates a tunnel
func (k Keeper) ActivateTunnel(ctx sdk.Context, id uint64, creator string) error {
	tunnel, err := k.GetTunnel(ctx, id)
	if err != nil {
		return err
	}

	if tunnel.Creator != creator {
		return fmt.Errorf("creator %s is not the creator of tunnel %d", creator, id)
	}
	tunnel.IsActive = true

	k.SetTunnel(ctx, tunnel)
	return nil
}

// GeneratePackets generates packets for all tunnels that require triggering
func (k Keeper) GeneratePackets(ctx sdk.Context) []types.Packet {
	packets := []types.Packet{}

	activeTunnels := k.GetActiveTunnels(ctx)
	latestPrices := k.feedsKeeper.GetPrices(ctx)
	latestPricesMap := CreateLatestPricesMap(latestPrices)
	unixNow := ctx.BlockTime().Unix()

	// check for active tunnels
	for _, at := range activeTunnels {
		if unixNow >= int64(at.Interval)+at.Timestamp {
			sps := GenerateSignalPriceInfos(ctx, at.SignalPriceInfos, latestPricesMap, at.ID)
			if len(sps) > 0 {
				packets = append(packets, types.NewPacket(at.ID, at.NonceCount+1, sps, unixNow))
			}
		} else {
			sps := GenerateSignalPriceInfosBasedOnDeviation(ctx, at.SignalPriceInfos, latestPricesMap, at.ID)
			if len(sps) > 0 {
				packets = append(packets, types.NewPacket(at.ID, at.NonceCount+1, sps, unixNow))
			}
		}
	}

	// check for pending trigger tunnels
	pendingTriggerTunnels := k.GetPendingTriggerTunnels(ctx)
	for _, id := range pendingTriggerTunnels {
		tunnel := k.MustGetTunnel(ctx, id)
		// skip if the tunnel is already in trigger list
		if unixNow >= int64(tunnel.Interval)+tunnel.Timestamp && tunnel.IsActive {
			continue
		}
		sps := GenerateSignalPriceInfos(ctx, tunnel.SignalPriceInfos, latestPricesMap, tunnel.ID)
		if len(sps) > 0 {
			packets = append(packets, types.NewPacket(tunnel.ID, tunnel.NonceCount+1, sps, unixNow))
		}
	}

	return packets
}

// HandlePacket sends a packet to destination route, stores the packet in the store, and updates the tunnel data
func (k Keeper) HandlePacket(ctx sdk.Context, packet types.Packet) {
	// get tunnel from tunnelID
	tunnel := k.MustGetTunnel(ctx, packet.TunnelID)

	// Process the tunnel based on the route type
	switch r := tunnel.Route.GetCachedValue().(type) {
	case *types.TSSRoute:
		err := k.TSSPacketHandle(ctx, r, packet)
		if err != nil {
			// Emit an event if the packet processing fails
			emitPacketFailEvent(ctx, packet.TunnelID, r, err)
			return
		}
	case *types.AxelarRoute:
		err := k.AxelarPacketHandle(ctx, r, packet)
		if err != nil {
			// Emit an event if the packet processing fails
			emitPacketFailEvent(ctx, packet.TunnelID, r, err)
			return
		}
	default:
		panic(fmt.Sprintf("unknown route type: %T", r))
	}

	// update tunnel data
	tunnel.NonceCount = packet.Nonce
	tunnel.SignalPriceInfos = packet.SignalPriceInfos
	tunnel.Timestamp = packet.CreatedAt
	k.SetTunnel(ctx, tunnel)
}

// emitPacketFailEvent emits an event when a packet fails to be sent
func emitPacketFailEvent(ctx sdk.Context, tunnelID uint64, route interface{}, err error) {
	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeSendPacketFail,
		sdk.NewAttribute(types.AttributeKeyTunnelID, fmt.Sprintf("%d", tunnelID)),
		sdk.NewAttribute(types.AttributeKeyRoute, fmt.Sprintf("%v", route)),
		sdk.NewAttribute(types.AttributeKeyReason, err.Error()),
	))
}

// GenerateSignalPriceInfos generates signal price infos based on the latest prices
func GenerateSignalPriceInfos(
	ctx sdk.Context,
	signalPriceInfos []types.SignalPriceInfo,
	latestPricesMap map[string]feedsTypes.Price,
	tunnelID uint64,
) []types.SignalPriceInfo {
	var nsps []types.SignalPriceInfo
	for _, sp := range signalPriceInfos {
		latestPrice, exists := latestPricesMap[sp.SignalID]
		if !exists || latestPrice.PriceStatus != feedsTypes.PriceStatusAvailable {
			nsps = append(nsps, types.NewSignalPriceInfo(sp.SignalID, sp.SoftDeviationBPS, sp.HardDeviationBPS, 0, 0))
			ctx.EventManager().EmitEvent(sdk.NewEvent(
				types.EventTypeSignalIDNotFound,
				sdk.NewAttribute(types.AttributeKeyTunnelID, fmt.Sprintf("%d", tunnelID)),
				sdk.NewAttribute(types.AttributeKeySignalID, sp.SignalID),
			))
			continue
		}
		nsps = append(
			nsps,
			types.NewSignalPriceInfo(
				sp.SignalID,
				sp.SoftDeviationBPS,
				sp.HardDeviationBPS,
				latestPrice.Price,
				latestPrice.Timestamp,
			),
		)
	}
	return nsps
}

// GenerateSignalPriceInfosBasedOnDeviation generates signal price infos based on the deviation of the latest prices
func GenerateSignalPriceInfosBasedOnDeviation(
	ctx sdk.Context,
	signalPriceInfos []types.SignalPriceInfo,
	latestPricesMap map[string]feedsTypes.Price,
	tunnelID uint64,
) []types.SignalPriceInfo {
	var nsps []types.SignalPriceInfo
	for _, sp := range signalPriceInfos {
		latestPrice, exists := latestPricesMap[sp.SignalID]
		if !exists || latestPrice.PriceStatus != feedsTypes.PriceStatusAvailable {
			nsps = append(nsps, types.NewSignalPriceInfo(sp.SignalID, sp.SoftDeviationBPS, sp.HardDeviationBPS, 0, 0))
			ctx.EventManager().EmitEvent(sdk.NewEvent(
				types.EventTypeSignalIDNotFound,
				sdk.NewAttribute(types.AttributeKeyTunnelID, fmt.Sprintf("%d", tunnelID)),
				sdk.NewAttribute(types.AttributeKeySignalID, sp.SignalID),
			))
			continue
		}
		deviation := math.Abs(float64(latestPrice.Price)-float64(sp.Price)) / float64(sp.Price)
		deviationInBPS := uint64(deviation * 10000)
		if deviationInBPS >= sp.HardDeviationBPS {
			nsps = append(
				nsps,
				types.NewSignalPriceInfo(
					sp.SignalID,
					sp.SoftDeviationBPS,
					sp.HardDeviationBPS,
					latestPrice.Price,
					latestPrice.Timestamp,
				),
			)
		}
	}
	return nsps
}

// CreateLatestPricesMap creates a map of latest prices with signal ID as the key
func CreateLatestPricesMap(latestPrices []feedsTypes.Price) map[string]feedsTypes.Price {
	latestPricesMap := make(map[string]feedsTypes.Price, len(latestPrices))
	for _, price := range latestPrices {
		latestPricesMap[price.SignalID] = price
	}
	return latestPricesMap
}
