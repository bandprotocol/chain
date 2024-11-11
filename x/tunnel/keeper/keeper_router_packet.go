package keeper

import (
	"encoding/base64"

	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/x/tunnel/types"
)

// SendRouterPacket sends Router packet
func (k Keeper) SendRouterPacket(
	ctx sdk.Context,
	route *types.RouterRoute,
	packet types.Packet,
	encoder types.Encoder,
	feePayer sdk.AccAddress,
) (types.PacketContentI, sdk.Coins, error) {
	// create encoding packet
	encodingpacket, err := types.NewEncodingPacket(packet, encoder)
	if err != nil {
		return nil, nil, err
	}

	// encode relay packet ABI
	relayPacket, err := encodingpacket.EncodeRelayPacketABI()
	if err != nil {
		return nil, nil, err
	}

	// create memo string for ibc transfer
	memoStr, err := types.NewRouterMemo(
		route.BridgeContractAddress,
		route.DestChainID,
		route.DestContractAddress,
		route.DestGasLimit,
		route.DestGasPrice,
		base64.StdEncoding.EncodeToString(relayPacket),
		0,
		"",
	).String()
	if err != nil {
		return nil, nil, err
	}

	msg := ibctransfertypes.NewMsgTransfer(
		ibctransfertypes.PortID,
		route.ChannelID,
		route.Fund,
		feePayer.String(),
		route.BridgeContractAddress,
		clienttypes.ZeroHeight(),
		uint64(ctx.BlockTime().UnixNano()+packetExpireTime),
		memoStr,
	)

	if _, err := k.transferKeeper.Transfer(ctx, msg); err != nil {
		return nil, nil, err
	}

	packetContent := types.RouterPacketContent{
		ChannelID:             route.ChannelID,
		Fund:                  route.Fund,
		BridgeContractAddress: route.BridgeContractAddress,
		DestChainID:           route.DestChainID,
		DestContractAddress:   route.DestContractAddress,
		DestGasLimit:          route.DestGasLimit,
		DestGasPrice:          route.DestGasPrice,
	}

	fee, err := route.Fee()
	if err != nil {
		return nil, nil, err
	}

	return &packetContent, fee, nil
}
