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

// IBCPacketHandler func
func (k Keeper) IBCPacketHandler(ctx sdk.Context, route *types.IBCRoute, packet types.Packet) {
	// retrieve the dynamic capability for this channel
	channelCap, ok := k.scopedKeeper.GetCapability(ctx, host.ChannelCapabilityPath(types.PortID, route.ChannelID))
	if !ok {
		ctx.EventManager().EmitEvent(sdk.NewEvent(
			types.EventTypeSendPacketFail,
			sdk.NewAttribute(types.AttributeKeyReason, "Module does not own channel capability"),
		))

		return
	}

	// create the IBC packet result bytes
	resultBytes := types.NewIBCPacketResult(
		packet.TunnelID,
		packet.Nonce,
		packet.SignalPriceInfos,
	).GetBytes()

	// Send packet to IBC, authenticating with channelCap
	if _, err := k.channelKeeper.SendPacket(
		ctx,
		channelCap,
		types.PortID,
		route.ChannelID,
		clienttypes.NewHeight(0, 0),
		uint64(ctx.BlockTime().UnixNano()+packetExpireTime),
		resultBytes,
	); err != nil {
		ctx.EventManager().EmitEvent(sdk.NewEvent(
			types.EventTypeSendPacketFail,
			sdk.NewAttribute(types.AttributeKeyReason, fmt.Sprintf("Unable to send packet: %s", err)),
		))
	}

	// Set the packet content
	packetContent := types.IBCPacketContent{
		ChannelID: route.ChannelID,
	}
	err := packet.SetPacketContent(&packetContent)
	if err != nil {
		panic(fmt.Errorf("failed to set packet content: %w", err))
	}

	k.AddPacket(ctx, packet)
}
