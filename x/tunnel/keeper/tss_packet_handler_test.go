package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	bandtsstypes "github.com/bandprotocol/chain/v2/x/bandtss/types"
	feedstypes "github.com/bandprotocol/chain/v2/x/feeds/types"
	"github.com/bandprotocol/chain/v2/x/tunnel/testutil"
	"github.com/bandprotocol/chain/v2/x/tunnel/types"
)

func TestTSSPacketHandler(t *testing.T) {
	s := testutil.NewTestSuite(t)
	ctx, k := s.Ctx, s.Keeper

	// Create a sample TSSRoute
	route := types.TSSRoute{
		DestinationChainID:         "chain-1",
		DestinationContractAddress: "0x1234567890abcdef",
	}

	// Create a sample Packet
	packet := types.NewPacket(
		1, // tunnelID
		1, // nonce
		feedstypes.FEED_TYPE_DEFAULT,
		nil, // SignalPriceInfos
		time.Now().Unix(),
	)

	// Call the TSSPacketHandler
	k.TSSPacketHandler(ctx, &route, packet)

	// Retrieve the packet from the keeper
	savedPacket, err := k.GetPacket(ctx, 1, 1)
	require.NoError(t, err)

	// Assert the packet content
	packetContent, ok := savedPacket.PacketContent.GetCachedValue().(*types.TSSPacketContent)
	require.True(t, ok)
	require.Equal(t, "chain-1", packetContent.DestinationChainID)
	require.Equal(t, "0x1234567890abcdef", packetContent.DestinationContractAddress)
	require.Equal(t, bandtsstypes.SigningID(1), packetContent.SigningID)
}
