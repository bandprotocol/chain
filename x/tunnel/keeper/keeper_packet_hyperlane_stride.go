package keeper

import (
	"encoding/hex"
	"time"

	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/x/tunnel/types"
)

// SendHyperlaneStridePacket sends HyperlaneStride packet
func (k Keeper) SendHyperlaneStridePacket(
	ctx sdk.Context,
	route *types.HyperlaneStrideRoute,
	packet types.Packet,
	feePayer sdk.AccAddress,
	interval uint64,
) (types.PacketReceiptI, error) {
	relayPacket, err := types.EncodingHyperlaneStride(packet)
	if err != nil {
		return nil, err
	}

	hyperlaneStrideIBCChannel := k.GetParams(ctx).HyperlaneStrideIBCChannel
	hyperlaneStrideIntegrationContract := k.GetParams(ctx).HyperlaneStrideIntegrationContract

	// create memo string for ibc transfer
	memoStr, err := types.NewHyperlaneStrideMemo(
		hyperlaneStrideIntegrationContract,
		route.DispatchDestDomain,
		route.DispatchRecipientAddr,
		hex.EncodeToString(relayPacket),
	).String()
	if err != nil {
		return nil, err
	}

	msg := ibctransfertypes.NewMsgTransfer(
		ibctransfertypes.PortID,
		hyperlaneStrideIBCChannel,
		route.Fund,
		feePayer.String(),
		route.DispatchRecipientAddr,
		clienttypes.ZeroHeight(),
		uint64(ctx.BlockTime().UnixNano())+interval*uint64(time.Second)*2,
		memoStr,
	)

	res, err := k.transferKeeper.Transfer(ctx, msg)
	if err != nil {
		return nil, err
	}

	return types.NewHyperlaneStridePacketReceipt(res.Sequence), nil
}
