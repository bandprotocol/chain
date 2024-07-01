package keeper_test

import (
	"github.com/stretchr/testify/require"

	"github.com/bandprotocol/chain/v2/x/tunnel/types"
)

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
