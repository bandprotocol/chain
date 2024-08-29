package keeper

import (
	"fmt"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	feedsTypes "github.com/bandprotocol/chain/v2/x/feeds/types"
	"github.com/bandprotocol/chain/v2/x/tunnel/types"
)

// CreateTunnel creates a new tunnel
func (k Keeper) CreateTunnel(
	ctx sdk.Context,
	route *codectypes.Any,
	feedType feedsTypes.FeedType,
	signalInfos []types.SignalInfo,
	interval uint64,
	creator string,
) (types.Tunnel, error) {
	id := k.GetTunnelCount(ctx)
	newID := id + 1

	// Generate a new tunnel account
	acc, err := k.GenerateAccount(ctx, fmt.Sprintf("%d", newID))
	if err != nil {
		return types.Tunnel{}, err
	}

	// Set the new tunnel count
	k.SetTunnelCount(ctx, newID)

	// Set the signal prices info
	var signalPrices []types.SignalPrice
	for _, sp := range signalInfos {
		signalPrices = append(signalPrices, types.NewSignalPrice(sp.SignalID, 0, 0))
	}
	k.SetSignalPricesInfo(ctx, types.NewSignalPricesInfo(newID, signalPrices, 0))

	return types.NewTunnel(
		id,
		0,
		route,
		feedType,
		acc.String(),
		signalInfos,
		interval,
		false,
		ctx.BlockTime().Unix(),
		creator,
	), nil
}

// SetTunnelCount sets the tunnel count in the store
func (k Keeper) SetTunnelCount(ctx sdk.Context, count uint64) {
	ctx.KVStore(k.storeKey).Set(types.TunnelCountStoreKey, sdk.Uint64ToBigEndian(count))
}

// GetTunnelCount returns the current number of all tunnels ever existed
func (k Keeper) GetTunnelCount(ctx sdk.Context) uint64 {
	return sdk.BigEndianToUint64(ctx.KVStore(k.storeKey).Get(types.TunnelCountStoreKey))
}

// SetTunnel sets a tunnel in the store
func (k Keeper) SetTunnel(ctx sdk.Context, tunnel types.Tunnel) {
	ctx.KVStore(k.storeKey).Set(types.TunnelStoreKey(tunnel.ID), k.cdc.MustMarshal(&tunnel))
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

// GetTunnelsByActiveStatus returns all tunnels by their active status
func (k Keeper) GetTunnelsByActiveStatus(ctx sdk.Context, isActive bool) []types.Tunnel {
	var tunnels []types.Tunnel
	iterator := sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.TunnelStoreKeyPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var tunnel types.Tunnel
		k.cdc.MustUnmarshal(iterator.Value(), &tunnel)

		if tunnel.IsActive == isActive {
			tunnels = append(tunnels, tunnel)
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
