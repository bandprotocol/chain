package keeper_test

import (
	"time"

	"go.uber.org/mock/gomock"

	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	feedstypes "github.com/bandprotocol/chain/v3/x/feeds/types"
	"github.com/bandprotocol/chain/v3/x/tunnel/types"
)

func (s *KeeperTestSuite) TestSendRouterPacket() {
	ctx, k := s.ctx, s.keeper

	route := &types.IBCHookRoute{
		ChannelID:                  "channel-0",
		DestinationContractAddress: "wasm1vjq0k3fj47s8wns4a7zw5c4lsjd8l6r2kzzlpk",
	}
	packet := types.Packet{
		TunnelID:  1,
		Sequence:  1,
		Prices:    []feedstypes.Price{},
		CreatedAt: time.Now().Unix(),
	}
	interval := uint64(60)
	feePayer := sdk.AccAddress([]byte("feePayer"))

	expectedPacketReceipt := types.IBCHookPacketReceipt{
		Sequence: 1,
	}

	s.transferKeeper.EXPECT().Transfer(ctx, gomock.Any()).Return(&ibctransfertypes.MsgTransferResponse{
		Sequence: 1,
	}, nil)
	s.bankKeeper.EXPECT().MintCoins(ctx, types.ModuleName, sdk.NewCoins(types.TransferAmount)).Return(nil)
	s.bankKeeper.EXPECT().
		SendCoinsFromModuleToAccount(ctx, types.ModuleName, feePayer, sdk.NewCoins(types.TransferAmount)).
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

func (s *KeeperTestSuite) TestMintCoinsToAccount() {
	ctx, k := s.ctx, s.keeper

	account := sdk.AccAddress([]byte("test_account"))

	s.bankKeeper.EXPECT().MintCoins(ctx, types.ModuleName, sdk.NewCoins(types.TransferAmount)).Return(nil)
	s.bankKeeper.EXPECT().
		SendCoinsFromModuleToAccount(ctx, types.ModuleName, account, sdk.NewCoins(types.TransferAmount)).
		Return(nil)

	// Mint coins to the account
	err := k.MintCoinsToAccount(ctx, account)
	s.Require().NoError(err)
}
