package keeper_test

import (
	"time"

	"github.com/stretchr/testify/require"

	feedtypes "github.com/bandprotocol/chain/v2/x/feeds/types"
	"github.com/bandprotocol/chain/v2/x/tunnel/types"
)

func (s *KeeperTestSuite) TestGenerateTunnelAccount() {
	ctx, k := s.ctx, s.keeper

	tunnelID := uint64(1)
	addr := k.GenerateTunnelAccount(ctx, tunnelID)

	// Assert
	require.NotNil(s.T(), addr, "expected generated address to be non-nil")
	require.Equal(
		s.T(),
		"band1mfkys3fdex2pvylxdutwk3ng26ys8pxtmjstgp",
		addr.String(),
		"expected generated address to match",
	)
}

func (s *KeeperTestSuite) TestAddTunnel() {
	ctx, k := s.ctx, s.keeper

	// Create a new tunnel instance
	tunnel := types.Tunnel{ID: 1}

	// Add the tunnel to the keeper
	k.AddTunnel(s.ctx, tunnel)

	// Attempt to retrieve the tunnel by its ID
	retrievedTunnel, err := k.GetTunnel(ctx, tunnel.ID)
	require.NoError(s.T(), err, "retrieving tunnel should not produce an error")

	expected := types.Tunnel{
		ID:                       1,
		Route:                    nil,
		FeedType:                 feedtypes.FEED_TYPE_UNSPECIFIED,
		FeePayer:                 "band1mfkys3fdex2pvylxdutwk3ng26ys8pxtmjstgp",
		SignalPriceInfos:         nil,
		LastTriggeredBlockHeight: 0,
		IsActive:                 false,
		CreatedAt:                time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC),
		Creator:                  "",
	}

	// Assert the retrieved tunnel matches the one we set
	require.Equal(s.T(), expected, retrievedTunnel, "the retrieved tunnel should match the original")
}

func (s *KeeperTestSuite) TestGetSetTunnel() {
	ctx, k := s.ctx, s.keeper

	// Create a new tunnel instance
	tunnel := types.Tunnel{ID: 1}

	// Set the tunnel in the keeper
	k.SetTunnel(s.ctx, tunnel)

	// Attempt to retrieve the tunnel by its ID
	retrievedTunnel, err := k.GetTunnel(ctx, tunnel.ID)

	// Assert no error occurred during retrieval
	require.NoError(s.T(), err, "retrieving tunnel should not produce an error")

	// Assert the retrieved tunnel matches the one we set
	require.Equal(s.T(), tunnel, retrievedTunnel, "the retrieved tunnel should match the original")
}

func (s *KeeperTestSuite) TestGetNextTunnelID() {
	ctx, k := s.ctx, s.keeper

	firstID := k.GetNextTunnelID(ctx)
	require.Equal(s.T(), uint64(1), firstID, "expected first tunnel ID to be 1")

	secondID := k.GetNextTunnelID(ctx)
	require.Equal(s.T(), uint64(2), secondID, "expected next tunnel ID to be 2")
}
