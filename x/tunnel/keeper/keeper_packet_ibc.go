package keeper

import (
	"time"

	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	host "github.com/cosmos/ibc-go/v8/modules/core/24-host"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/x/tunnel/types"
)

// SendIBCPacket sends IBC packet
func (k Keeper) SendIBCPacket(
	ctx sdk.Context,
	route *types.IBCRoute,
	packet types.Packet,
	interval uint64,
) (types.PacketReceiptI, error) {
	portID := PortIDForTunnel(packet.TunnelID)
	// retrieve the dynamic capability for this channel
	channelCap, ok := k.scopedKeeper.GetCapability(
		ctx,
		host.ChannelCapabilityPath(portID, route.ChannelID),
	)
	if !ok {
		return nil, types.ErrChannelCapabilityNotFound
	}

	// create the tunnel prices packet data bytes
	packetBytes := types.NewTunnelPricesPacketData(
		packet.TunnelID,
		packet.Sequence,
		packet.Prices,
		packet.CreatedAt,
	).GetBytes()

	// send packet to IBC, authenticating with channelCap
	sequence, err := k.ics4Wrapper.SendPacket(
		ctx,
		channelCap,
		portID,
		route.ChannelID,
		clienttypes.NewHeight(0, 0),
		uint64(ctx.BlockTime().UnixNano())+interval*uint64(time.Second),
		packetBytes,
	)
	if err != nil {
		return nil, err
	}

	return types.NewIBCPacketReceipt(sequence), nil
}
