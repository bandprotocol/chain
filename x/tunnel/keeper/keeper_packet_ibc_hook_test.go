package keeper_test

import (
	"go.uber.org/mock/gomock"

	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	feedstypes "github.com/bandprotocol/chain/v3/x/feeds/types"
	"github.com/bandprotocol/chain/v3/x/tunnel/types"
)

func (s *KeeperTestSuite) TestSendIBCHookPacket() {
	ctx, k := s.ctx, s.keeper

	tunnelID := uint64(1)
	route := &types.IBCHookRoute{
		ChannelID:                  "channel-0",
		DestinationContractAddress: "wasm1vjq0k3fj47s8wns4a7zw5c4lsjd8l6r2kzzlpk",
	}
	packet := types.Packet{
		TunnelID:  tunnelID,
		Sequence:  1,
		Prices:    []feedstypes.Price{},
		CreatedAt: 1730358471,
	}
	interval := uint64(60)
	feePayer := sdk.AccAddress([]byte("feePayer"))
	hookCoins := sdk.NewCoins(
		sdk.NewInt64Coin(types.FormatHookDenomIdentifier(tunnelID), types.HookTransferAmount),
	)

	expectedPacketReceipt := types.IBCHookPacketReceipt{
		Sequence: 1,
	}

	s.transferKeeper.EXPECT().Transfer(ctx, gomock.Any()).Return(&ibctransfertypes.MsgTransferResponse{
		Sequence: 1,
	}, nil)
	s.bankKeeper.EXPECT().MintCoins(ctx, types.ModuleName, hookCoins).Return(nil)
	s.bankKeeper.EXPECT().
		SendCoinsFromModuleToAccount(ctx, types.ModuleName, feePayer, hookCoins).
		Return(nil)

	content, err := k.SendIBCHookPacket(
		ctx,
		route,
		packet,
		feePayer,
		interval,
	)
	s.Require().NoError(err)

	receipt, ok := content.(*types.IBCHookPacketReceipt)
	s.Require().True(ok)
	s.Require().Equal(expectedPacketReceipt, *receipt)
}

func (s *KeeperTestSuite) TestMintIBCHookCoinToAccount() {
	ctx, k := s.ctx, s.keeper

	tunnelID := uint64(1)
	account := sdk.AccAddress([]byte("test_account"))
	hookCoins := sdk.NewCoins(
		sdk.NewInt64Coin(types.FormatHookDenomIdentifier(tunnelID), types.HookTransferAmount),
	)

	s.bankKeeper.EXPECT().MintCoins(ctx, types.ModuleName, hookCoins).Return(nil)
	s.bankKeeper.EXPECT().
		SendCoinsFromModuleToAccount(ctx, types.ModuleName, account, hookCoins).
		Return(nil)

	// Mint coins to the account
	err := k.MintIBCHookCoinToAccount(ctx, tunnelID, account)
	s.Require().NoError(err)
}
