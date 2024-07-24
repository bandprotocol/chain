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
		TunnelID: 1,
		PacketID: 1,
	}
	k.SetAxelarPacket(ctx, packet)

	storedPacket, err := k.GetAxelarPacket(ctx, 1, 1)
	require.NoError(s.T(), err)
	require.Equal(s.T(), packet, storedPacket)
}
