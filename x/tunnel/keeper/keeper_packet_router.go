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
	payload, err := types.EncodingPacketABI(packet)
	if err != nil {
		return nil, err
	}

	// get router ibc channel and integration contract
	params := k.GetParams(ctx)
	routerIBCChannel := params.RouterIBCChannel
	routerIntegrationContract := params.RouterIntegrationContract

	// mint coin to the fee payer
	err = k.MintIBCHookCoinToAccount(ctx, packet.TunnelID, feePayer)
	if err != nil {
		return nil, err
	}

	// create memo string for ibc transfer
	memoStr := types.NewRouterMemo(
		routerIntegrationContract,
		route.DestinationChainID,
		route.DestinationContractAddress,
		route.DestinationGasLimit,
		base64.StdEncoding.EncodeToString(payload),
	).JSONString()

	msg := ibctransfertypes.NewMsgTransfer(
		ibctransfertypes.PortID,
		routerIBCChannel,
		sdk.NewInt64Coin(types.FormatHookDenomIdentifier(packet.TunnelID), types.HookTransferAmount),
		feePayer.String(),
		routerIntegrationContract,
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
