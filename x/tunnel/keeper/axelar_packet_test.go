package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bandprotocol/chain/v2/x/tunnel/testutil"
	"github.com/bandprotocol/chain/v2/x/tunnel/types"
)

func TestGetSetAxelarPacket(t *testing.T) {
	s := testutil.NewTestSuite(t)
	ctx, k := s.Ctx, s.Keeper

	packet := types.AxelarPacket{
		ID: 1,
	}
	k.SetAxelarPacket(ctx, packet)

	storedPacket, err := k.GetAxelarPacket(ctx, 1)
	require.NoError(s.T(), err)
	require.Equal(s.T(), packet, storedPacket)
}

func TestGetNextAxelarPacketID(t *testing.T) {
	s := testutil.NewTestSuite(t)
	ctx, k := s.Ctx, s.Keeper

	firstID := k.GetNextAxelarPacketID(ctx)
	require.Equal(s.T(), uint64(1), firstID, "expected first axelar packet ID to be 1")
	secondID := k.GetNextAxelarPacketID(ctx)
	require.Equal(s.T(), uint64(2), secondID, "expected next axelar packet ID to be 2")
}
