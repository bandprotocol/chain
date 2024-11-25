package keeper

import (
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/x/tunnel/types"
)

// SendIBCHookPacket sends a packet to the destination chain using IBC Hook
func (k Keeper) SendIBCHookPacket(
	ctx sdk.Context,
	route *types.IBCHookRoute,
	packet types.Packet,
	feePayer sdk.AccAddress,
) (types.PacketReceiptI, error) {
	// create memo string for ibc transfer
	memoStr, err := types.NewIBCHookMemo(
		route.DestinationContractAddress,
		packet.TunnelID,
		packet.Sequence,
		packet.Prices,
		packet.CreatedAt,
	).String()
	if err != nil {
		return nil, err
	}

	// create ibc transfer message
	msg := ibctransfertypes.NewMsgTransfer(
		ibctransfertypes.PortID,
		route.ChannelID,
		// TODO: align the token to send with msg transfer
		sdk.NewInt64Coin("uband", 1),
		feePayer.String(),
		route.DestinationContractAddress,
		clienttypes.ZeroHeight(),
		uint64(ctx.BlockTime().UnixNano()+packetExpireTime),
		memoStr,
	)

	// send packet
	res, err := k.transferKeeper.Transfer(ctx, msg)
	if err != nil {
		return nil, err
	}

	return types.NewIBCHookPacketReceipt(res.Sequence), nil
}
