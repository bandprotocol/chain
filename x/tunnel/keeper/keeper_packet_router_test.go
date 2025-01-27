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

	tunnelID := uint64(1)
	route := &types.RouterRoute{
		DestinationChainID:         "17000",
		DestinationContractAddress: "0xDFCfEbF22e85193eDc37b8b136d4F3394987d1AE",
		DestinationGasLimit:        300000,
		DestinationGasPrice:        10000000,
	}
	packet := types.Packet{
		TunnelID:  tunnelID,
		Sequence:  1,
		Prices:    []feedstypes.Price{},
		CreatedAt: time.Now().Unix(),
	}
	interval := uint64(60)
	feePayer := sdk.AccAddress([]byte("feePayer"))
	hookCoins := sdk.NewCoins(
		sdk.NewInt64Coin(types.FormatHookDenomIdentifier(tunnelID), types.HookTransferAmount),
	)

	s.transferKeeper.EXPECT().
		Transfer(ctx, gomock.Any()).
		Return(&ibctransfertypes.MsgTransferResponse{Sequence: 1}, nil)
	s.bankKeeper.EXPECT().MintCoins(ctx, types.ModuleName, hookCoins).Return(nil)
	s.bankKeeper.EXPECT().
		SendCoinsFromModuleToAccount(ctx, types.ModuleName, feePayer, hookCoins).
		Return(nil)

	receipt, err := k.SendRouterPacket(
		ctx,
		route,
		packet,
		feePayer,
		interval,
	)
	s.Require().NoError(err)

	packetReceipt, ok := receipt.(*types.RouterPacketReceipt)
	s.Require().True(ok)
	s.Require().Equal(uint64(1), packetReceipt.Sequence)
}
