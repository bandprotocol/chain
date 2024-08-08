package keeper

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	host "github.com/cosmos/ibc-go/v7/modules/core/24-host"

	"github.com/bandprotocol/chain/v2/x/tunnel/types"
)

const (
	packetExpireTime = int64(10 * time.Minute)
)

// GetIBCPacket retrieves a IBC packet by its tunnel ID and packet ID
func (k Keeper) GetIBCPacket(ctx sdk.Context, tunnelID uint64, nonce uint64) (types.IBCPacket, error) {
	bz := ctx.KVStore(k.storeKey).Get(types.TunnelPacketStoreKey(tunnelID, nonce))
	if bz == nil {
		return types.IBCPacket{}, types.ErrPacketNotFound.Wrapf("tunnelID: %d, nonce: %d", tunnelID, nonce)
	}

	var packet types.IBCPacket
	k.cdc.MustUnmarshal(bz, &packet)
	return packet, nil
}

// IBCPacketHandler func
func (k Keeper) IBCPacketHandler(ctx sdk.Context, packet types.IBCPacket) {
	// retrieve the dynamic capability for this channel
	channelCap, ok := k.scopedKeeper.GetCapability(ctx, host.ChannelCapabilityPath(types.PortID, packet.ChannelID))
	if !ok {
		ctx.EventManager().EmitEvent(sdk.NewEvent(
			types.EventTypeSendPacketFail,
			sdk.NewAttribute(types.AttributeKeyReason, "Module does not own channel capability"),
		))

		return
	}

	// Send packet to IBC, authenticating with channelCap
	if _, err := k.channelKeeper.SendPacket(
		ctx,
		channelCap,
		types.PortID,
		packet.ChannelID,
		clienttypes.NewHeight(0, 0),
		uint64(ctx.BlockTime().UnixNano()+packetExpireTime),
		packet.GetBytes(),
	); err != nil {
		ctx.EventManager().EmitEvent(sdk.NewEvent(
			types.EventTypeSendPacketFail,
			sdk.NewAttribute(types.AttributeKeyReason, fmt.Sprintf("Unable to send packet: %s", err)),
		))
	}
}
