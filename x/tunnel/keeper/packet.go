package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/tunnel/types"
)

// SetPacket sets a packet in the store
func (k Keeper) SetPacket(ctx sdk.Context, packet types.Packet) {
	ctx.KVStore(k.storeKey).
		Set(types.TunnelPacketStoreKey(packet.TunnelID, packet.Nonce), k.cdc.MustMarshal(&packet))
}

// AddPacket adds a packet to the store
func (k Keeper) AddPacket(ctx sdk.Context, packet types.Packet) {
	// Set the creation time
	packet.CreatedAt = ctx.BlockTime().Unix()
	k.SetPacket(ctx, packet)
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
