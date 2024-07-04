package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/bandprotocol/chain/v2/x/tunnel/types"
)

// GenerateTunnelAccount generates a new tunnel account
func (k Keeper) GenerateTunnelAccount(ctx sdk.Context, tunnelID uint64) sdk.AccAddress {
	tacc := authtypes.NewEmptyModuleAccount(fmt.Sprintf("%s-%d", types.ModuleName, tunnelID))
	taccI := k.authKeeper.NewAccount(ctx, tacc)
	k.authKeeper.SetAccount(ctx, taccI)
	return taccI.GetAddress()
}

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
func (k Keeper) AddTunnel(ctx sdk.Context, tunnel types.Tunnel) uint64 {
	tunnel.ID = k.GetNextTunnelID(ctx)

	// Generate a new tunnel account
	acc := k.GenerateTunnelAccount(ctx, tunnel.ID)
	tunnel.FeePayer = acc.String()

	// Set the creation time
	tunnel.CreatedAt = ctx.BlockTime()

	k.SetTunnel(ctx, tunnel)
	return tunnel.ID
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

func (k Keeper) GeneratePackets(ctx sdk.Context, tunnel types.Tunnel) error {
	var route types.Route
	route, ok := tunnel.Route.GetCachedValue().(*types.TSSRoute)
	if ok {
		fmt.Printf("TSSRoute: %v\n", route)
		return nil
	}
	route, ok = tunnel.Route.GetCachedValue().(*types.AxelarRoute)
	if ok {
		fmt.Printf("AxelarRoute: %v\n", route)
		return nil
	}
	return fmt.Errorf("unknown route type: %s", tunnel.Route.String())
}
