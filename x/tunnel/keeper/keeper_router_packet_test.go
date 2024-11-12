package keeper_test

import (
	"time"

	"go.uber.org/mock/gomock"

	sdk "github.com/cosmos/cosmos-sdk/types"

	feedstypes "github.com/bandprotocol/chain/v3/x/feeds/types"
	"github.com/bandprotocol/chain/v3/x/tunnel/types"
)

func (s *KeeperTestSuite) TestSendRouterPacket() {
	ctx, k := s.ctx, s.keeper

	route := &types.RouterRoute{
		ChannelID:             "channel-0",
		Fund:                  sdk.NewInt64Coin("uband", 50000),
		BridgeContractAddress: "router17c2txg2px6vna8a6v4ql4eh4ruvprerhytxvwt2ugp4qr473pajsyj9pgm",
		DestChainID:           "17000",
		DestContractAddress:   "0xDFCfEbF22e85193eDc37b8b136d4F3394987d1AE",
		DestGasLimit:          300000,
		DestGasPrice:          10000000,
	}

	packet := types.Packet{
		TunnelID:  1,
		Sequence:  1,
		Prices:    []feedstypes.Price{},
		CreatedAt: time.Now().Unix(),
	}

	expectedPacketContent := types.RouterPacketContent{
		ChannelID:             "channel-0",
		Fund:                  sdk.NewInt64Coin("uband", 50000),
		BridgeContractAddress: "router17c2txg2px6vna8a6v4ql4eh4ruvprerhytxvwt2ugp4qr473pajsyj9pgm",
		DestChainID:           "17000",
		DestContractAddress:   "0xDFCfEbF22e85193eDc37b8b136d4F3394987d1AE",
		DestGasLimit:          300000,
		DestGasPrice:          10000000,
	}

	s.transferKeeper.EXPECT().Transfer(ctx, gomock.Any()).Return(nil, nil)

	content, err := k.SendRouterPacket(
		ctx,
		route,
		packet,
		feedstypes.ENCODER_FIXED_POINT_ABI,
		sdk.AccAddress("feePayer"),
	)
	s.Require().NoError(err)

	packetContent, ok := content.(*types.RouterPacketContent)
	s.Require().True(ok)
	s.Require().Equal(expectedPacketContent, *packetContent)
}
