package keeper_test

import (
	"time"

	bandtsstypes "github.com/bandprotocol/chain/v2/x/bandtss/types"
	"github.com/bandprotocol/chain/v2/x/tunnel/types"
)

func (s *KeeperTestSuite) TestSendTSSPacket() {
	ctx, k := s.ctx, s.keeper

	// Create a sample TSSRoute
	route := types.TSSRoute{
		DestinationChainID:         "chain-1",
		DestinationContractAddress: "0x1234567890abcdef",
	}

	// Create a sample Packet
	packet := types.NewPacket(
		1,                     // tunnelID
		1,                     // nonce
		[]types.SignalPrice{}, // SignalPriceInfos
		nil,
		time.Now().Unix(),
	)

	// Send the TSS packet
	content, err := k.SendTSSPacket(ctx, &route, packet)
	s.Require().NoError(err)

	// Assert the packet content
	packetContent, ok := content.(*types.TSSPacketContent)
	s.Require().True(ok)
	s.Require().Equal("chain-1", packetContent.DestinationChainID)
	s.Require().Equal("0x1234567890abcdef", packetContent.DestinationContractAddress)
	s.Require().Equal(bandtsstypes.SigningID(1), packetContent.SigningID)
}
