package keeper

import (
	"time"

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
	interval uint64,
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

	// mint coins to the fee payer
	err = k.MintCoinsToAccount(ctx, feePayer)
	if err != nil {
		return nil, err
	}

	// create ibc transfer message
	msg := ibctransfertypes.NewMsgTransfer(
		ibctransfertypes.PortID,
		route.ChannelID,
		types.TransferAmount,
		feePayer.String(),
		route.DestinationContractAddress,
		clienttypes.ZeroHeight(),
		uint64(ctx.BlockTime().UnixNano())+interval*uint64(time.Second),
		memoStr,
	)

	// send packet
	res, err := k.transferKeeper.Transfer(ctx, msg)
	if err != nil {
		return nil, err
	}

	return types.NewIBCHookPacketReceipt(res.Sequence), nil
}

// MintCoinsToAccount mints uhook coins to the account
func (k Keeper) MintCoinsToAccount(ctx sdk.Context, account sdk.AccAddress) error {
	// mint coins to the account
	err := k.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(types.TransferAmount))
	if err != nil {
		return err
	}

	// send coins to the account
	return k.bankKeeper.SendCoinsFromModuleToAccount(
		ctx,
		types.ModuleName,
		account,
		sdk.NewCoins(types.TransferAmount),
	)
}
