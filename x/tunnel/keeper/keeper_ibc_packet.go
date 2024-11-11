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
) (types.PacketContentI, sdk.Coins, error) {
	// retrieve the dynamic capability for this channel
	channelCap, ok := k.scopedKeeper.GetCapability(ctx, host.ChannelCapabilityPath(types.PortID, route.ChannelID))
	if !ok {
		return nil, nil, types.ErrChannelCapabilityNotFound
	}

	// create the IBC packet result bytes
	resultBytes := types.NewIBCPacketResult(
		packet.TunnelID,
		packet.Sequence,
		packet.SignalPrices,
		packet.CreatedAt,
	).GetBytes()

	// send packet to IBC, authenticating with channelCap
	if _, err := k.ics4Wrapper.SendPacket(
		ctx,
		channelCap,
		types.PortID,
		route.ChannelID,
		clienttypes.NewHeight(0, 0),
		uint64(ctx.BlockTime().UnixNano()+packetExpireTime),
		resultBytes,
	); err != nil {
		return nil, nil, err
	}

	fee, err := route.Fee()
	if err != nil {
		return nil, nil, err
	}

	return types.NewIBCPacketContent(route.ChannelID), fee, nil
}
