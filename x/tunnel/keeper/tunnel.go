package keeper

import (
	"fmt"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	feedsTypes "github.com/bandprotocol/chain/v2/x/feeds/types"
	"github.com/bandprotocol/chain/v2/x/tunnel/types"
)

// AddTunnel adds a new tunnel
func (k Keeper) AddTunnel(
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

	// Set the signal prices info
	var signalPrices []types.SignalPrice
	for _, si := range signalInfos {
		signalPrices = append(signalPrices, types.NewSignalPrice(si.SignalID, 0, 0))
	}
	k.SetSignalPricesInfo(ctx, types.NewSignalPricesInfo(newID, signalPrices, 0))

	// Create a new tunnel
	tunnel := types.NewTunnel(
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
	)
	k.SetTunnel(ctx, tunnel)
	k.SetTunnelCount(ctx, tunnel.ID)

	return tunnel, nil
}

// EditTunnel edits a tunnel
func (k Keeper) EditTunnel(
	ctx sdk.Context,
	tunnelID uint64,
	signalInfos []types.SignalInfo,
	interval uint64,
) error {
	tunnel, err := k.GetTunnel(ctx, tunnelID)
	if err != nil {
		return err
	}

	// Edit the signal infos and interval
	tunnel.SignalInfos = signalInfos
	tunnel.Interval = interval
	k.SetTunnel(ctx, tunnel)

	// Edit the signal prices info
	var signalPrices []types.SignalPrice
	for _, sp := range signalInfos {
		signalPrices = append(signalPrices, types.NewSignalPrice(sp.SignalID, 0, 0))
	}
	k.SetSignalPricesInfo(ctx, types.NewSignalPricesInfo(tunnelID, signalPrices, 0))

	return nil
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
func (k Keeper) GetTunnel(ctx sdk.Context, tunnelID uint64) (types.Tunnel, error) {
	bz := ctx.KVStore(k.storeKey).Get(types.TunnelStoreKey(tunnelID))
	if bz == nil {
		return types.Tunnel{}, types.ErrTunnelNotFound.Wrapf("tunnelID: %d", tunnelID)
	}

	var tunnel types.Tunnel
	k.cdc.MustUnmarshal(bz, &tunnel)
	return tunnel, nil
}

// MustGetTunnel retrieves a tunnel by its ID. Panics if the tunnel does not exist.
func (k Keeper) MustGetTunnel(ctx sdk.Context, tunnelID uint64) types.Tunnel {
	tunnel, err := k.GetTunnel(ctx, tunnelID)
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

// SetActiveTunnelIDs sets the active tunnel IDs in the store
func (k Keeper) SetActiveTunnelIDs(ctx sdk.Context, ids []uint64) {
	ctx.KVStore(k.storeKey).Set(types.ActiveTunnelIDsStoreKey, k.cdc.MustMarshal(&types.ActiveTunnelIDs{IDs: ids}))
}

// GetActiveTunnelIDs retrieves the active tunnel IDs from the store
func (k Keeper) GetActiveTunnelIDs(ctx sdk.Context) ([]uint64, error) {
	bz := ctx.KVStore(k.storeKey).Get(types.ActiveTunnelIDsStoreKey)
	if bz == nil {
		return []uint64{}, types.ErrActiveTunnelIDsNotFound
	}
	var activeTunnelIDs types.ActiveTunnelIDs
	k.cdc.MustUnmarshal(bz, &activeTunnelIDs)
	return activeTunnelIDs.IDs, nil
}

// MustGetActiveTunnelIDs retrieves the active tunnel IDs from the store and panics if the IDs do not exist
func (k Keeper) MustGetActiveTunnelIDs(ctx sdk.Context) []uint64 {
	ids, err := k.GetActiveTunnelIDs(ctx)
	if err != nil {
		panic(err)
	}
	return ids
}

// ActivateTunnel activates a tunnel
func (k Keeper) ActivateTunnel(ctx sdk.Context, tunnelID uint64) error {
	tunnel, err := k.GetTunnel(ctx, tunnelID)
	if err != nil {
		return err
	}

	// Check if the tunnel is already active
	if tunnel.IsActive {
		return types.ErrTunnelAlreadyActive
	}

	// Activate the tunnel
	tunnel.IsActive = true

	// Add the tunnel ID to the active tunnel IDs
	activeTunnelIDs, err := k.GetActiveTunnelIDs(ctx)
	if err != nil {
		return err
	}

	activeTunnelIDs = append(activeTunnelIDs, tunnelID)
	k.SetActiveTunnelIDs(ctx, activeTunnelIDs)

	// Set the last interval timestamp to the current block time
	k.SetTunnel(ctx, tunnel)
	return nil
}

// DeactivateTunnel deactivates a tunnel
func (k Keeper) DeactivateTunnel(ctx sdk.Context, tunnelID uint64) error {
	tunnel, err := k.GetTunnel(ctx, tunnelID)
	if err != nil {
		return err
	}

	// Check if the tunnel is already inactive
	if !tunnel.IsActive {
		return types.ErrTunnelNotActive
	}

	tunnel.IsActive = false

	// Remove the tunnel ID from the active tunnel IDs
	activeTunnelIDs, err := k.GetActiveTunnelIDs(ctx)
	if err != nil {
		return err
	}

	// Remove the tunnel ID from the active tunnel IDs
	for i, activeID := range activeTunnelIDs {
		if activeID == tunnelID {
			activeTunnelIDs = append(activeTunnelIDs[:i], activeTunnelIDs[i+1:]...)
			break
		}
	}
	k.SetActiveTunnelIDs(ctx, activeTunnelIDs)

	// Set the last interval timestamp to the current block time
	k.SetTunnel(ctx, tunnel)
	return nil
}
