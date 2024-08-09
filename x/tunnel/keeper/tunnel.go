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
	tunnel.CreatedAt = ctx.BlockTime()

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

// GetRequiredProcessTunnels returns all tunnels that require processing
func (k Keeper) GetRequiredProcessTunnels(
	ctx sdk.Context,
) []types.Tunnel {
	var tunnels []types.Tunnel
	activeTunnels := k.GetActiveTunnels(ctx)
	// TODO: TBD Get price may got some unavailable price
	latestPrices := k.feedsKeeper.GetPrices(ctx)
	latestPricesMap := make(map[string]feedsTypes.Price, len(latestPrices))

	// Populate the map with the latest prices
	for _, price := range latestPrices {
		latestPricesMap[price.SignalID] = price
	}

	now := ctx.BlockTime()
	unixNow := ctx.BlockTime().Unix()

	// Evaluate which tunnels require processing based on the price signals
	for i, at := range activeTunnels {
		var trigger bool
		for j, sp := range at.SignalPriceInfos {
			latestPrice, exists := latestPricesMap[sp.SignalID]
			if !exists {
				continue
			}

			deviation := math.Abs(float64(latestPrice.Price)-float64(sp.Price)) / float64(sp.Price)
			deviationInBPS := uint64(deviation * 10000)

			if deviationInBPS > sp.DeviationBPS || unixNow >= int64(sp.LastTimestamp+sp.Interval) {
				// Update the price directly
				activeTunnels[i].SignalPriceInfos[j].Price = latestPrice.Price
				activeTunnels[i].SignalPriceInfos[j].LastTimestamp = uint64(now.Unix())
				trigger = true
			}
		}

		if trigger {
			tunnels = append(tunnels, at)
		}
	}

	// add pending trigger tunnels
	pendingTriggerTunnels := k.GetPendingTriggerTunnels(ctx)
	for _, id := range pendingTriggerTunnels {
		if !types.IsTunnelInList(id, tunnels) {
			tunnel := k.MustGetTunnel(ctx, id)
			for i, sp := range tunnel.SignalPriceInfos {
				latestPrice, exists := latestPricesMap[sp.SignalID]
				if !exists {
					ctx.EventManager().EmitEvent(sdk.NewEvent(
						types.EventTypeSignalIDNotFound,
						sdk.NewAttribute(types.AttributeKeyTunnelID, fmt.Sprintf("%d", tunnel.ID)),
						sdk.NewAttribute(types.AttributeSignalID, sp.SignalID),
					))
					continue
				}

				tunnel.SignalPriceInfos[i].Price = latestPrice.Price
				tunnel.SignalPriceInfos[i].LastTimestamp = uint64(now.Unix())
				tunnels = append(tunnels, tunnel)
			}
		}
	}
	return tunnels
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

// SetParams sets the tunnel module parameters
func (k Keeper) ProcessTunnel(ctx sdk.Context, tunnel types.Tunnel) {
	// Increment the nonce
	tunnel.NonceCount += 1

	switch r := tunnel.Route.GetCachedValue().(type) {
	case *types.TSSRoute:
		k.TSSPacketHandler(ctx, types.TSSPacket{
			TunnelID:                   tunnel.ID,
			SignalPriceInfos:           tunnel.SignalPriceInfos,
			DestinationChainID:         r.DestinationChainID,
			DestinationContractAddress: r.DestinationContractAddress,
		})
	case *types.AxelarRoute:
		k.AxelarPacketHandler(ctx, types.AxelarPacket{})
	case *types.IBCRoute:
		fmt.Printf("Generating IBC packets for tunnel %d, route %s\n", tunnel.ID, r.String())
		k.IBCPacketHandler(ctx, types.IBCPacket{
			TunnelID:         tunnel.ID,
			Nonce:            tunnel.NonceCount,
			FeedType:         tunnel.FeedType,
			SignalPriceInfos: tunnel.SignalPriceInfos,
			ChannelID:        r.ChannelID,
		})
	}

	// Update the tunnel
	k.SetTunnel(ctx, tunnel)
}
