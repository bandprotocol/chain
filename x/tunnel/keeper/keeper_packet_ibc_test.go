package keeper_test

import (
	"time"

	"go.uber.org/mock/gomock"

	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"

	feedstypes "github.com/bandprotocol/chain/v3/x/feeds/types"
	"github.com/bandprotocol/chain/v3/x/tunnel/types"
)

func (s *KeeperTestSuite) TestSendIBCPacket() {
	ctx, k := s.ctx, s.keeper

	route := &types.IBCRoute{
		ChannelID: "channel-0",
	}
	packet := types.Packet{
		TunnelID:  1,
		Sequence:  1,
		Prices:    []feedstypes.Price{},
		CreatedAt: time.Now().Unix(),
	}
	interval := uint64(60)

	s.scopedKeeper.EXPECT().GetCapability(ctx, gomock.Any()).Return(&capabilitytypes.Capability{}, true)
	s.icsWrapper.EXPECT().
		SendPacket(ctx, gomock.Any(), "tunnel.1", route.ChannelID, clienttypes.NewHeight(0, 0), uint64(ctx.BlockTime().UnixNano())+interval*uint64(time.Second), gomock.Any()).
		Return(uint64(1), nil)

	content, err := k.SendIBCPacket(ctx, route, packet, interval)
	s.Require().NoError(err)

	packetReceipt, ok := content.(*types.IBCPacketReceipt)
	s.Require().True(ok)
	s.Require().Equal(uint64(1), packetReceipt.Sequence)
}
