package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bandprotocol/chain/v2/x/tunnel/testutil"
	"github.com/bandprotocol/chain/v2/x/tunnel/types"
)

func TestGetSetTSSPacket(t *testing.T) {
	s := testutil.NewTestSuite(t)
	ctx, k := s.Ctx, s.Keeper
	packet := types.TSSPacket{
		ID: 1,
	}
	k.SetTSSPacket(ctx, packet)

	storedPacket, err := k.GetTSSPacket(ctx, 1)
	require.NoError(s.T(), err)
	require.Equal(s.T(), packet, storedPacket)
}

func TestGetNextTSSPacketID(t *testing.T) {
	s := testutil.NewTestSuite(t)
	ctx, k := s.Ctx, s.Keeper

	firstID := k.GetNextTSSPacketID(ctx)
	require.Equal(s.T(), uint64(1), firstID, "expected first tss packet ID to be 1")
	secondID := k.GetNextTSSPacketID(ctx)
	require.Equal(s.T(), uint64(2), secondID, "expected next tss packet ID to be 2")
}
