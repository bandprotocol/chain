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
	pricePacket := types.NewTunnelPricesPacketData(packet.TunnelID, packet.Sequence, packet.Prices, packet.CreatedAt)
	memoStr := types.NewIBCHookMemo(route.DestinationContractAddress, pricePacket).JSONString()

	// mint coin to the fee payer
	err := k.MintIBCHookCoinToAccount(ctx, packet.TunnelID, feePayer)
	if err != nil {
		return nil, err
	}

	// create ibc transfer message with the memo string
	msg := ibctransfertypes.NewMsgTransfer(
		ibctransfertypes.PortID,
		route.ChannelID,
		sdk.NewInt64Coin(types.FormatHookDenomIdentifier(packet.TunnelID), types.HookTransferAmount),
		feePayer.String(),
		route.DestinationContractAddress,
		clienttypes.ZeroHeight(),
		uint64(ctx.BlockTime().UnixNano())+interval*uint64(time.Second)*2,
		memoStr,
	)

	// send packet
	res, err := k.transferKeeper.Transfer(ctx, msg)
	if err != nil {
		return nil, err
	}

	return types.NewIBCHookPacketReceipt(res.Sequence), nil
}

// MintIBCHookCoinToAccount mints hook coin to the account
func (k Keeper) MintIBCHookCoinToAccount(ctx sdk.Context, tunnelID uint64, account sdk.AccAddress) error {
	// create hook coins
	hookCoins := sdk.NewCoins(
		sdk.NewInt64Coin(types.FormatHookDenomIdentifier(tunnelID), types.HookTransferAmount),
	)

	// mint coins to the module account
	err := k.bankKeeper.MintCoins(ctx, types.ModuleName, hookCoins)
	if err != nil {
		return err
	}

	// send coins to the account
	return k.bankKeeper.SendCoinsFromModuleToAccount(
		ctx,
		types.ModuleName,
		account,
		hookCoins,
	)
}
