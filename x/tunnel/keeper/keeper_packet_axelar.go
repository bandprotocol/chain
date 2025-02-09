package keeper

import (
	"time"

	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/x/tunnel/types"
)

// SendAxelarPacket sends a packet to the destination chain using Axelar
func (k Keeper) SendAxelarPacket(
	ctx sdk.Context,
	route *types.AxelarRoute,
	packet types.Packet,
	feePayer sdk.AccAddress,
	interval uint64,
) (types.PacketReceiptI, error) {
	// get axelar params
	params := k.GetParams(ctx)
	ibcChannel := params.AxelarIBCChannel
	feeRecipient := params.AxelarFeeRecipient
	gmpAccount := params.AxelarGmpAccount

	// encode packet to abi format
	payload, err := types.EncodingPacketABI(packet)
	if err != nil {
		return nil, err
	}

	// create axelar fee
	feePayerStr := feePayer.String()
	axelarFee := types.NewAxelarFee(route.Fee.String(), feeRecipient, &feePayerStr)

	// create memo string for ibc transfer
	memoStr, err := types.NewAxelarMemo(
		route.DestinationChainID,
		route.DestinationContractAddress,
		payload,
		types.AxelarMessageTypeGeneralMessage,
		&axelarFee,
	).String()
	if err != nil {
		return nil, err
	}

	// create ibc transfer message with the memo string
	msg := ibctransfertypes.NewMsgTransfer(
		ibctransfertypes.PortID,
		ibcChannel,
		route.Fee,
		feePayer.String(),
		gmpAccount,
		clienttypes.ZeroHeight(),
		uint64(ctx.BlockTime().UnixNano())+interval*uint64(time.Second)*2,
		memoStr,
	)

	// send packet
	res, err := k.transferKeeper.Transfer(ctx, msg)
	if err != nil {
		return nil, err
	}

	return types.NewAxelarPacketReceipt(res.Sequence), nil
}
