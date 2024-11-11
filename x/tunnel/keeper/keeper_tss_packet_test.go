package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	bandtsstypes "github.com/bandprotocol/chain/v3/x/bandtss/types"
	"github.com/bandprotocol/chain/v3/x/tunnel/types"
)

func (s *KeeperTestSuite) TestSendTSSPacket() {
	ctx, k := s.ctx, s.keeper

	route := types.TSSRoute{
		DestinationChainID:         "chain-1",
		DestinationContractAddress: "0x1234567890abcdef",
	}
	packet := types.NewPacket(
		1,                     // tunnelID
		1,                     // sequence
		[]types.SignalPrice{}, // signalPriceInfos[]
		sdk.NewCoins(),        // baseFee
		sdk.NewCoins(),        // routeFee
		time.Now().Unix(),
	)

	content, fee, err := k.SendTSSPacket(ctx, &route, packet)
	s.Require().NoError(err)

	packetContent, ok := content.(*types.TSSPacketContent)
	s.Require().True(ok)
	s.Require().Equal("chain-1", packetContent.DestinationChainID)
	s.Require().Equal("0x1234567890abcdef", packetContent.DestinationContractAddress)
	s.Require().Equal(sdk.NewCoins(), fee)
	s.Require().Equal(bandtsstypes.SigningID(1), packetContent.SigningID)
}
