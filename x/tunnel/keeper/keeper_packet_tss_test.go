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
		Encoder:                    feedstypes.ENCODER_FIXED_POINT_ABI,
	}
	packet := types.NewPacket(
		1,                    // tunnelID
		1,                    // sequence
		[]feedstypes.Price{}, // priceInfos[]
		time.Now().Unix(),
	)

	s.bandtssKeeper.EXPECT().GetSigningFee(ctx).Return(sdk.NewCoins(sdk.NewCoin("uband", sdkmath.NewInt(20))), nil)

	// Mock the TSS keeper and set the state for checking later
	s.bandtssKeeper.EXPECT().CreateTunnelSigningRequest(
		gomock.Any(),
		uint64(1),
		"chain-1",
		"0x1234567890abcdef",
		gomock.Any(),
		bandtesting.Alice.Address,
		sdk.NewCoins(sdk.NewCoin("uband", sdkmath.NewInt(20))),
	).Return(bandtsstypes.SigningID(1), nil)

	// Send the TSS packet
	content, err := k.SendTSSPacket(
		ctx,
		&route,
		packet,
		bandtesting.Alice.Address,
	)
	s.Require().NoError(err)

	packetReceipt, ok := content.(*types.TSSPacketReceipt)
	s.Require().True(ok)
	s.Require().Equal(bandtsstypes.SigningID(1), packetReceipt.SigningID)
}
