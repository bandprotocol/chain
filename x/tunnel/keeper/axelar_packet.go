package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/tunnel/types"
)

// SetAxelarPacket sets a Axelar packet in the store
func (k Keeper) SetAxelarPacket(ctx sdk.Context, packet types.AxelarPacket) {
	ctx.KVStore(k.storeKey).
		Set(types.TunnelPacketStoreKey(packet.TunnelID, packet.Nonce), k.cdc.MustMarshal(&packet))
}

// AddAxelarPacket adds a Axelar packet to the store
func (k Keeper) AddAxelarPacket(ctx sdk.Context, packet types.AxelarPacket) {
	packet.CreatedAt = ctx.BlockTime().Unix()
	k.SetAxelarPacket(ctx, packet)
}

// GetAxelarPacket retrieves a Axelar packet by its tunnel ID and packet ID
func (k Keeper) GetAxelarPacket(ctx sdk.Context, tunnelID uint64, nonce uint64) (types.AxelarPacket, error) {
	bz := ctx.KVStore(k.storeKey).Get(types.TunnelPacketStoreKey(tunnelID, nonce))
	if bz == nil {
		return types.AxelarPacket{}, types.ErrPacketNotFound.Wrapf("tunnelID: %d, nonce: %d", tunnelID, nonce)
	}

	var packet types.AxelarPacket
	k.cdc.MustUnmarshal(bz, &packet)
	return packet, nil
}

func (k Keeper) AxelarPacketHandler(ctx sdk.Context, packet types.AxelarPacket) {}
