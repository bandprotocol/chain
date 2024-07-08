package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/tunnel/types"
)

// SetTSSPacketCount sets the total number of TSS packets that have been sent
func (k Keeper) SetTSSPacketCount(ctx sdk.Context, count uint64) {
	ctx.KVStore(k.storeKey).Set(types.TSSPacketCountStoreKey, sdk.Uint64ToBigEndian(count))
}

// GetTSSPacketCount returns the current number of all TSS packets ever sent
func (k Keeper) GetTSSPacketCount(ctx sdk.Context) uint64 {
	return sdk.BigEndianToUint64(ctx.KVStore(k.storeKey).Get(types.TSSPacketCountStoreKey))
}

// GetNextTSSPacketID increments the TSS packet count and returns the current number of TSS packets
func (k Keeper) GetNextTSSPacketID(ctx sdk.Context) uint64 {
	packetNumber := k.GetTSSPacketCount(ctx) + 1
	k.SetTSSPacketCount(ctx, packetNumber)
	return packetNumber
}

// SetTSSPacket sets a TSS packet in the store
func (k Keeper) SetTSSPacket(ctx sdk.Context, packet types.TSSPacket) {
	ctx.KVStore(k.storeKey).Set(types.TSSPacketStoreKey(packet.ID), k.cdc.MustMarshal(&packet))
}

// AddTSSPacket adds a TSS packet to the store and returns the new packet ID
func (k Keeper) AddTSSPacket(ctx sdk.Context, packet types.TSSPacket) uint64 {
	packet.ID = k.GetNextTSSPacketID(ctx)

	// Set the creation time
	packet.CreatedAt = ctx.BlockTime()

	k.SetTSSPacket(ctx, packet)
	return packet.ID
}

// GetTSSPacket retrieves a TSS packet by its ID
func (k Keeper) GetTSSPacket(ctx sdk.Context, id uint64) (types.TSSPacket, error) {
	bz := ctx.KVStore(k.storeKey).Get(types.TSSPacketStoreKey(id))
	if bz == nil {
		return types.TSSPacket{}, types.ErrTSSPacketNotFound.Wrapf("packetID: %d", id)
	}

	var packet types.TSSPacket
	k.cdc.MustUnmarshal(bz, &packet)
	return packet, nil
}

// MustGetTSSPacket retrieves a TSS packet by its ID and panics if the packet does not exist
func (k Keeper) MustGetTSSPacket(ctx sdk.Context, id uint64) types.TSSPacket {
	packet, err := k.GetTSSPacket(ctx, id)
	if err != nil {
		panic(err)
	}
	return packet
}

func (k Keeper) TSSPacketHandler(ctx sdk.Context, packet types.TSSPacket) uint64 {
	// TODO: Implement TSS packet handler logic
	// Sign TSS packet
	packet.SigningID = 1

	// Save the signed TSS packet
	return k.AddTSSPacket(ctx, packet)
}
