package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

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
		sdk.NewCoins(),       // routeFee
		time.Now().Unix(),
	)

	content, err := k.SendTSSPacket(ctx, &route, packet)
	s.Require().NoError(err)

	packetReceipt, ok := content.(*types.TSSPacketReceipt)
	s.Require().True(ok)
	s.Require().Equal(bandtsstypes.SigningID(1), packetReceipt.SigningID)
}
