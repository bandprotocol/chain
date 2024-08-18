package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bandprotocol/chain/v2/x/tunnel/testutil"
	"github.com/bandprotocol/chain/v2/x/tunnel/types"
)

func TestGetSetPacket(t *testing.T) {
	s := testutil.NewTestSuite(t)
	ctx := s.Ctx
	k := s.Keeper

	packet := types.Packet{
		TunnelID: 1,
		Nonce:    1,
	}

	k.SetPacket(ctx, packet)

	storedPacket, err := k.GetPacket(ctx, packet.TunnelID, packet.Nonce)
	require.NoError(t, err)
	require.Equal(t, packet, storedPacket)
}

func TestAddPacket(t *testing.T) {
	s := testutil.NewTestSuite(t)
	ctx := s.Ctx
	k := s.Keeper

	packet := types.Packet{
		TunnelID: 1,
		Nonce:    1,
	}

	k.AddPacket(ctx, packet)

	storedPacket, err := k.GetPacket(ctx, packet.TunnelID, packet.Nonce)
	require.NoError(t, err)
	require.Equal(t, packet.TunnelID, storedPacket.TunnelID)
	require.Equal(t, packet.Nonce, storedPacket.Nonce)
	require.NotZero(t, storedPacket.CreatedAt)
}

func TestMustGetPacket(t *testing.T) {
	s := testutil.NewTestSuite(t)
	ctx := s.Ctx
	k := s.Keeper

	packet := types.Packet{
		TunnelID: 1,
		Nonce:    1,
	}

	k.SetPacket(ctx, packet)

	storedPacket := k.MustGetPacket(ctx, packet.TunnelID, packet.Nonce)
	require.Equal(t, packet, storedPacket)

	require.Panics(t, func() {
		k.MustGetPacket(ctx, packet.TunnelID, 999)
	})
}
