package keeper_test

// import (
// 	"time"

// 	"github.com/bandprotocol/chain/v3/x/tunnel/types"
// 	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
// 	"go.uber.org/mock/gomock"
// )

// // func (s *KeeperTestSuite) TestSendIBCPacket() {
// // 	ctx, k := s.ctx, s.keeper

// // 	route := &types.IBCRoute{
// // 		ChannelID: "channel-0",
// // 	}
// // 	packet := types.Packet{
// // 		TunnelID:     1,
// // 		Sequence:     1,
// // 		SignalPrices: []types.SignalPrice{},
// // 		CreatedAt:    time.Now().Unix(),
// // 	}

// // 	s.scopedKeeper.EXPECT().GetCapability(ctx, gomock.Any()).Return(&capabilitytypes.Capability{}, true)
// // 	s.channelKeeper.EXPECT().
// // 		SendPacket(ctx, gomock.Any(), types.PortID, route.ChannelID, gomock.Any(), gomock.Any(), gomock.Any()).
// // 		Return(uint64(0), nil)

// // 	content, err := k.SendIBCPacket(ctx, route, packet)
// // 	s.Require().NoError(err)

// // 	packetContent, ok := content.(*types.IBCPacketContent)
// // 	s.Require().True(ok)
// // 	s.Require().Equal("channel-0", packetContent.ChannelID)
// // }