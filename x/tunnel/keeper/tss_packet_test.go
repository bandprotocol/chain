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
		TunnelID: 1,
		Nonce:    1,
	}
	k.SetTSSPacket(ctx, packet)

	storedPacket, err := k.GetTSSPacket(ctx, 1, 1)
	require.NoError(s.T(), err)
	require.Equal(s.T(), packet, storedPacket)
}
