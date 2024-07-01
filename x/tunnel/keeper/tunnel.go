package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

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
func (k Keeper) AddTunnel(ctx sdk.Context, tunnel types.Tunnel) uint64 {
	tunnel.ID = k.GetNextTunnelID(ctx)
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
	switch tunnel.Route.GetCachedValue().(type) {
	case *types.TSSRoute:
		// TODO: Implement TSS packet generation
		k.TSSPacketHandler(ctx, types.TSSPacket{})
	case *types.AxelarRoute:
		// TODO: Implement Axelar packet generation
		k.AxelarPacketHandler(ctx, types.AxelarPacket{})

	default:
		return fmt.Errorf("unknown route type: %s", tunnel.Route.String())
	}

	return nil
}
