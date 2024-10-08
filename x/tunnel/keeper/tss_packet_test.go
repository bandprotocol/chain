package keeper_test

import (
	"math"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"go.uber.org/mock/gomock"

	bandtesting "github.com/bandprotocol/chain/v2/testing"
	bandtsstypes "github.com/bandprotocol/chain/v2/x/bandtss/types"
	"github.com/bandprotocol/chain/v2/x/tunnel/types"
)

func (s *KeeperTestSuite) TestSendTSSPacket() {
	ctx, k := s.ctx, s.keeper

	route := types.TSSRoute{
		DestinationChainID:         "chain-1",
		DestinationContractAddress: "0x1234567890abcdef",
	}
	packet, err := types.NewPacket(
		1,                     // tunnelID
		1,                     // sequence
		[]types.SignalPrice{}, // signalPriceInfos[]
		time.Now().Unix(),
	)
	s.Require().NoError(err)

	s.bandtssKeeper.EXPECT().GetParams(gomock.Any()).Return(bandtsstypes.Params{
		Fee: sdk.NewCoins(sdk.NewCoin("uband", sdk.NewInt(10))),
	})
	s.bandtssKeeper.EXPECT().CreateTunnelSigningRequest(
		gomock.Any(),
		uint64(1),
		"0x1234567890abcdef",
		"chain-1",
		gomock.Any(),
		bandtesting.Alice.Address,
		sdk.NewCoins(sdk.NewCoin("uband", sdk.NewInt(math.MaxInt))),
	).Return(bandtsstypes.SigningID(1), nil)

	k.SetTunnel(ctx, types.Tunnel{
		ID:       1,
		Encoder:  types.ENCODER_FIXED_POINT_ABI,
		FeePayer: bandtesting.Alice.Address.String(),
	})

	// Send the TSS packet
	content, err := k.SendTSSPacket(ctx, &route, packet)
	s.Require().NoError(err)

	packetContent, ok := content.(*types.TSSPacketContent)
	s.Require().True(ok)
	s.Require().Equal("chain-1", packetContent.DestinationChainID)
	s.Require().Equal("0x1234567890abcdef", packetContent.DestinationContractAddress)
	s.Require().Equal(bandtsstypes.SigningID(1), packetContent.SigningID)
}
