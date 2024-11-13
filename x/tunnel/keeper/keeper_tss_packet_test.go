package keeper_test

import (
	"time"

	"go.uber.org/mock/gomock"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	bandtesting "github.com/bandprotocol/chain/v3/testing"
	bandtsstypes "github.com/bandprotocol/chain/v3/x/bandtss/types"
	feedstypes "github.com/bandprotocol/chain/v3/x/feeds/types"
	"github.com/bandprotocol/chain/v3/x/tunnel/types"
)

func (s *KeeperTestSuite) TestSendTSSPacket() {
	ctx, k := s.ctx, s.keeper

	route := types.TSSRoute{
		DestinationChainID:         "chain-1",
		DestinationContractAddress: "0x1234567890abcdef",
	}
	packet := types.NewPacket(
		1,                    // tunnelID
		1,                    // sequence
		[]feedstypes.Price{}, // priceInfos[]
		sdk.NewCoins(),       // baseFee
		sdk.NewCoins(sdk.NewCoin("uband", sdkmath.NewInt(20))), // routeFee
		time.Now().Unix(),
	)

	s.bandtssKeeper.EXPECT().CreateTunnelSigningRequest(
		gomock.Any(),
		uint64(1),
		"0x1234567890abcdef",
		"chain-1",
		gomock.Any(),
		bandtesting.Alice.Address,
		sdk.NewCoins(sdk.NewCoin("uband", sdkmath.NewInt(20))),
	).Return(bandtsstypes.SigningID(1), nil)

	k.SetTunnel(ctx, types.Tunnel{
		ID:       1,
		Encoder:  feedstypes.ENCODER_FIXED_POINT_ABI,
		FeePayer: bandtesting.Alice.Address.String(),
	})

	// Send the TSS packet
	content, fee, err := k.SendTSSPacket(ctx, &route, packet)
	s.Require().NoError(err)

	packetContent, ok := content.(*types.TSSPacketContent)
	s.Require().True(ok)
	s.Require().Equal("chain-1", packetContent.DestinationChainID)
	s.Require().Equal("0x1234567890abcdef", packetContent.DestinationContractAddress)
	s.Require().Equal(sdk.NewCoins(sdk.NewCoin("uband", sdkmath.NewInt(20))), fee)
	s.Require().Equal(bandtsstypes.SigningID(1), packetContent.SigningID)
}
