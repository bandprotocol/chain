package keeper_test

import (
	"github.com/stretchr/testify/require"

	"github.com/bandprotocol/chain/v2/x/tunnel/types"
)

func (s *KeeperTestSuite) TestGetSetTSSPacket() {
	ctx, k := s.ctx, s.keeper

	packet := types.TSSPacket{
		ID: 1,
	}
	k.SetTSSPacket(ctx, packet)

	storedPacket, err := k.GetTSSPacket(ctx, 1)
	require.NoError(s.T(), err)
	require.Equal(s.T(), packet, storedPacket)
}

func (s *KeeperTestSuite) TestGetNextTSSPacketID() {
	ctx, k := s.ctx, s.keeper

	firstID := k.GetNextTSSPacketID(ctx)
	require.Equal(s.T(), uint64(1), firstID, "expected first tss packet ID to be 1")
	secondID := k.GetNextTSSPacketID(ctx)
	require.Equal(s.T(), uint64(2), secondID, "expected next tss packet ID to be 2")
}
