package keeper

import (
	"encoding/base64"
	"time"

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
	feePayer sdk.AccAddress,
	interval uint64,
) (types.PacketReceiptI, error) {
	relayPacket, err := types.EncodingRouter(packet)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	routerIBCChannel := k.GetParams(ctx).RouterIBCChannel

	msg := ibctransfertypes.NewMsgTransfer(
		ibctransfertypes.PortID,
		routerIBCChannel,
		route.Fund,
		feePayer.String(),
		route.BridgeContractAddress,
		clienttypes.ZeroHeight(),
		uint64(ctx.BlockTime().UnixNano())+interval*uint64(time.Second)*2,
		memoStr,
	)

	res, err := k.transferKeeper.Transfer(ctx, msg)
	if err != nil {
		return nil, err
	}

	return types.NewRouterPacketReceipt(res.Sequence), nil
}
