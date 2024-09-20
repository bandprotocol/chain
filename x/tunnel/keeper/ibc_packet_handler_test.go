package keeper_test

import (
	"time"

	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	"go.uber.org/mock/gomock"

	"github.com/bandprotocol/chain/v2/x/tunnel/types"
)

func (s *KeeperTestSuite) TestIBCPacketHandler() {
	ctx, k := s.ctx, s.keeper

	route := &types.IBCRoute{
		ChannelID: "channel-0",
	}
	packet := types.Packet{
		TunnelID:     1,
		Nonce:        1,
		SignalPrices: []types.SignalPrice{},
		CreatedAt:    time.Now().Unix(),
	}

	s.scopedKeeper.EXPECT().GetCapability(ctx, gomock.Any()).Return(&capabilitytypes.Capability{}, true)
	s.channelKeeper.EXPECT().
		SendPacket(ctx, gomock.Any(), types.PortID, route.ChannelID, gomock.Any(), gomock.Any(), gomock.Any()).
		Return(uint64(0), nil)

	k.IBCPacketHandler(ctx, route, packet)

	packet, err := k.GetPacket(ctx, packet.TunnelID, packet.Nonce)
	s.Require().NoError(err)
	s.Require().NotNil(packet)
}
