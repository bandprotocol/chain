package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/tunnel/types"
)

// SetAxelarPacketCount sets the total number of Axelar packets that have been sent
func (k Keeper) SetAxelarPacketCount(ctx sdk.Context, count uint64) {
	ctx.KVStore(k.storeKey).Set(types.AxelarPacketCountStoreKey, sdk.Uint64ToBigEndian(count))
}

// GetAxelarPacketCount returns the current number of all Axelar packets ever sent
func (k Keeper) GetAxelarPacketCount(ctx sdk.Context) uint64 {
	return sdk.BigEndianToUint64(ctx.KVStore(k.storeKey).Get(types.AxelarPacketCountStoreKey))
}

// GetNextAxelarPacketID increments the Axelar packet count and returns the current number of Axelar packets
func (k Keeper) GetNextAxelarPacketID(ctx sdk.Context) uint64 {
	packetNumber := k.GetAxelarPacketCount(ctx) + 1
	k.SetAxelarPacketCount(ctx, packetNumber)
	return packetNumber
}

// SetAxelarPacket sets a Axelar packet in the store
func (k Keeper) SetAxelarPacket(ctx sdk.Context, packet types.AxelarPacket) {
	ctx.KVStore(k.storeKey).Set(types.AxelarPacketStoreKey(packet.ID), k.cdc.MustMarshal(&packet))
}

// AddAxelarPacket adds a Axelar packet to the store and returns the new packet ID
func (k Keeper) AddAxelarPacket(ctx sdk.Context, packet types.AxelarPacket) uint64 {
	packet.ID = k.GetNextAxelarPacketID(ctx)
	k.SetAxelarPacket(ctx, packet)
	return packet.ID
}

// GetAxelarPacket retrieves a Axelar packet by its ID
func (k Keeper) GetAxelarPacket(ctx sdk.Context, id uint64) (types.AxelarPacket, error) {
	bz := ctx.KVStore(k.storeKey).Get(types.AxelarPacketStoreKey(id))
	if bz == nil {
		return types.AxelarPacket{}, types.ErrAxelarPacketNotFound.Wrapf("packetID: %d", id)
	}

	var packet types.AxelarPacket
	k.cdc.MustUnmarshal(bz, &packet)
	return packet, nil
}

func (k Keeper) AxelarPacketHandler(ctx sdk.Context, packet types.AxelarPacket) {}
