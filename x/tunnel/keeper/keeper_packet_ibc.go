package keeper

import (
	"time"

	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	host "github.com/cosmos/ibc-go/v8/modules/core/24-host"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/x/tunnel/types"
)

const (
	packetExpireTime = int64(10 * time.Minute)
)

// SendIBCPacket sends IBC packet
func (k Keeper) SendIBCPacket(
	ctx sdk.Context,
	route *types.IBCRoute,
	packet types.Packet,
) (types.PacketReceiptI, error) {
	// retrieve the dynamic capability for this channel
	channelCap, ok := k.scopedKeeper.GetCapability(ctx, host.ChannelCapabilityPath(types.PortID, route.ChannelID))
	if !ok {
		return nil, types.ErrChannelCapabilityNotFound
	}

	// create the IBC packet bytes
	packetBytes := types.NewIBCPacket(
		packet.TunnelID,
		packet.Sequence,
		packet.Prices,
		packet.CreatedAt,
	).GetBytes()

	// send packet to IBC, authenticating with channelCap
	sequence, err := k.ics4Wrapper.SendPacket(
		ctx,
		channelCap,
		types.PortID,
		route.ChannelID,
		clienttypes.NewHeight(0, 0),
		uint64(ctx.BlockTime().UnixNano()+packetExpireTime),
		packetBytes,
	)
	if err != nil {
		return nil, err
	}

	return types.NewIBCPacketReceipt(route.ChannelID, sequence), nil
}
