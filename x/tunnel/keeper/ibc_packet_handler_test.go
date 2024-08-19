package keeper_test

import (
	"testing"
	"time"

	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/bandprotocol/chain/v2/x/tunnel/testutil"
	"github.com/bandprotocol/chain/v2/x/tunnel/types"
)

func TestIBCPacketHandler(t *testing.T) {
	s := testutil.NewTestSuite(t)
	ctx, k := s.Ctx, s.Keeper

	// Define a mock route
	route := &types.IBCRoute{
		ChannelID: "channel-0",
	}

	// Define a mock packet
	packet := types.Packet{
		TunnelID:         1,
		Nonce:            1,
		FeedType:         1,
		SignalPriceInfos: []types.SignalPriceInfo{},
		CreatedAt:        time.Now().Unix(),
	}

	// Mock the scoped keeper and channel keeper
	s.MockScopedKeeper.EXPECT().GetCapability(ctx, gomock.Any()).Return(&capabilitytypes.Capability{}, true)
	s.MockChannelKeeper.EXPECT().
		SendPacket(ctx, gomock.Any(), types.PortID, route.ChannelID, gomock.Any(), gomock.Any(), gomock.Any()).
		Return(uint64(0), nil)

	// Run the IBCPacketHandler function
	k.IBCPacketHandler(ctx, route, packet)

	packet, err := k.GetPacket(ctx, packet.TunnelID, packet.Nonce)
	require.NoError(t, err)

	// Check if the packet was added
	require.NotNil(t, packet)
}
