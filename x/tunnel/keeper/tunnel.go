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

// GetRequiredProcessTunnels returns all tunnels that require processing
func (k Keeper) GetRequiredProcessTunnels(
	ctx sdk.Context,
) []types.Tunnel {
	// TODO: Remove mock test
	return k.GetTunnels(ctx)

	var tunnels []types.Tunnel

	activeTunnels := k.GetActiveTunnels(ctx)
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
			if exists {
				difference := math.Abs(float64(latestPrice.Price)-float64(sp.Price)) / float64(sp.Price)
				differenceInBPS := uint64(difference * 10000)

				activeTunnels[i].SignalPriceInfos[j].Price = latestPrice.Price
				activeTunnels[i].SignalPriceInfos[j].LastTimestamp = &now

				if differenceInBPS > sp.DeviationBPS || unixNow >= sp.LastTimestamp.Unix()+int64(sp.Interval) {
					trigger = true
				}
			}
		}

		if trigger {
			tunnels = append(tunnels, at)
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

func (k Keeper) GetPackets(ctx sdk.Context, tunnelID uint64) ([]any, error) {
	var packets []any
	iterator := sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.TunnelPacketsStoreKey(tunnelID))
	defer iterator.Close()

	tunnel, err := k.GetTunnel(ctx, tunnelID)
	if err != nil {
		return nil, err
	}

	switch tunnel.Route.GetCachedValue().(type) {
	case *types.TSSRoute:
		for ; iterator.Valid(); iterator.Next() {
			var packet types.TSSPacket
			k.cdc.MustUnmarshal(iterator.Value(), &packet)
			packets = append(packets, packet)
		}
	case *types.AxelarRoute:
		for ; iterator.Valid(); iterator.Next() {
			var packet types.AxelarPacket
			k.cdc.MustUnmarshal(iterator.Value(), &packet)
			packets = append(packets, packet)
		}
	default:
		return nil, fmt.Errorf("unknown route type")
	}

	return packets, nil
}

// SetParams sets the tunnel module parameters
func (k Keeper) ProcessTunnel(ctx sdk.Context, tunnel types.Tunnel) {
	// Increment the packet count
	tunnel.NonceCount += 1

	switch r := tunnel.Route.GetCachedValue().(type) {
	case *types.TSSRoute:
		fmt.Printf("Generating TSS packets for tunnel %d, route %s\n", tunnel.ID, r.String())
		k.TSSPacketHandler(ctx, types.TSSPacket{
			TunnelID:                   tunnel.ID,
			Nonce:                      tunnel.NonceCount,
			SignalPriceInfos:           tunnel.SignalPriceInfos,
			DestinationChainID:         r.DestinationChainID,
			DestinationContractAddress: r.DestinationContractAddress,
		})
	case *types.AxelarRoute:
		fmt.Printf("Generating Axelar packets for tunnel %d, route %s\n", tunnel.ID, r.String())
		k.AxelarPacketHandler(ctx, types.AxelarPacket{})
	default:
		panic("unknown route type")
	}

	// Set the last SignalPriceInfos
	k.SetTunnel(ctx, tunnel)
}
