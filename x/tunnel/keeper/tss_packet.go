package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/tunnel/types"
)

// SetTSSPacket sets a TSS packet in the store
func (k Keeper) SetTSSPacket(ctx sdk.Context, packet types.TSSPacket) {
	ctx.KVStore(k.storeKey).
		Set(types.TunnelPacketStoreKey(packet.TunnelID, packet.Nonce), k.cdc.MustMarshal(&packet))
}

// AddTSSPacket adds a TSS packet to the store
func (k Keeper) AddTSSPacket(ctx sdk.Context, packet types.TSSPacket) {
	// Set the creation time
	packet.CreatedAt = ctx.BlockTime()
	k.SetTSSPacket(ctx, packet)
}

// GetTSSPacket retrieves a TSS packet by its tunnel ID and packet ID
func (k Keeper) GetTSSPacket(ctx sdk.Context, tunnelID uint64, nonce uint64) (types.TSSPacket, error) {
	bz := ctx.KVStore(k.storeKey).Get(types.TunnelPacketStoreKey(tunnelID, nonce))
	if bz == nil {
		return types.TSSPacket{}, types.ErrPacketNotFound.Wrapf("tunnelID: %d, nonce: %d", tunnelID, nonce)
	}

	var packet types.TSSPacket
	k.cdc.MustUnmarshal(bz, &packet)
	return packet, nil
}

// MustGetTSSPacket retrieves a TSS packet by its ID and panics if the packet does not exist
func (k Keeper) MustGetTSSPacket(ctx sdk.Context, tunnelID uint64, nonce uint64) types.TSSPacket {
	packet, err := k.GetTSSPacket(ctx, tunnelID, nonce)
	if err != nil {
		panic(err)
	}
	return packet
}

// TSSPacketHandler handles incoming TSS packets
func (k Keeper) TSSPacketHandler(ctx sdk.Context, packet types.TSSPacket) {
	// TODO: Implement TSS packet handler logic
	// Sign TSS packet
	packet.SigningID = 1

	// Save the signed TSS packet
	k.AddTSSPacket(ctx, packet)
}
