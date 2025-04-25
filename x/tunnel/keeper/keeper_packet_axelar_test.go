package keeper_test

import (
	"time"

	"go.uber.org/mock/gomock"

	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	feedstypes "github.com/bandprotocol/chain/v3/x/feeds/types"
	"github.com/bandprotocol/chain/v3/x/tunnel/types"
)

func (s *KeeperTestSuite) TestSendAxelarPacket() {
	ctx, k := s.ctx, s.keeper

	axelarFee := sdk.NewInt64Coin("uband", 100)

	tunnelID := uint64(1)
	route := &types.AxelarRoute{
		DestinationChainID:         "mock-chain",
		DestinationContractAddress: "0x1234567890",
		Fee:                        axelarFee,
	}
	packet := types.Packet{
		TunnelID:  tunnelID,
		Sequence:  1,
		Prices:    []feedstypes.Price{},
		CreatedAt: time.Now().Unix(),
	}
	interval := uint64(60)
	feePayer := sdk.AccAddress([]byte("feePayer"))

	expectedPacketReceipt := types.AxelarPacketReceipt{
		Sequence: 1,
	}

	s.transferKeeper.EXPECT().Transfer(ctx, gomock.Any()).Return(&ibctransfertypes.MsgTransferResponse{
		Sequence: 1,
	}, nil)

	content, err := k.SendAxelarPacket(
		ctx,
		route,
		packet,
		feePayer,
		interval,
	)
	s.Require().NoError(err)

	receipt, ok := content.(*types.AxelarPacketReceipt)
	s.Require().True(ok)
	s.Require().Equal(expectedPacketReceipt, *receipt)
}
