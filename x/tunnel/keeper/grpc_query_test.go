package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bandprotocol/chain/v2/x/tunnel/testutil"
	"github.com/bandprotocol/chain/v2/x/tunnel/types"
)

func TestGRPCQueryPackets(t *testing.T) {
	s := testutil.NewTestSuite(t)
	q := s.QueryServer

	// set tunnel
	tunnel := types.Tunnel{
		ID:         1,
		NonceCount: 2,
	}
	r := types.TSSRoute{
		DestinationChainID:         "1",
		DestinationContractAddress: "0x123",
	}
	err := tunnel.SetRoute(&r)
	require.NoError(t, err)
	s.Keeper.SetTunnel(s.Ctx, tunnel)

	// set tss packets
	tssPackets := []types.TSSPacket{
		{
			TunnelID: 1,
			Nonce:    1,
		},
		{
			TunnelID: 1,
			Nonce:    2,
		},
	}
	for _, packet := range tssPackets {
		s.Keeper.SetTSSPacket(s.Ctx, packet)
	}

	// query packets
	res, err := q.Packets(s.Ctx, &types.QueryPacketsRequest{
		TunnelId: 1,
	})
	require.NoError(t, err)
	for i, packet := range res.Packets {
		tssPacket, ok := packet.GetCachedValue().(*types.TSSPacket)
		require.True(t, ok)
		require.Equal(t, tssPackets[i], *tssPacket)
	}
}

func TestGRPCQueryPacket(t *testing.T) {
	s := testutil.NewTestSuite(t)
	q := s.QueryServer

	// set tunnel
	tunnel := types.Tunnel{
		ID:         1,
		NonceCount: 2,
	}
	r := types.TSSRoute{
		DestinationChainID:         "1",
		DestinationContractAddress: "0x123",
	}
	err := tunnel.SetRoute(&r)
	require.NoError(t, err)
	s.Keeper.SetTunnel(s.Ctx, tunnel)

	// set tss packets
	tssPackets := []types.TSSPacket{
		{
			TunnelID: 1,
			Nonce:    1,
		},
		{
			TunnelID: 1,
			Nonce:    2,
		},
	}
	for _, packet := range tssPackets {
		s.Keeper.SetTSSPacket(s.Ctx, packet)
	}

	// query packet
	res, err := q.Packet(s.Ctx, &types.QueryPacketRequest{
		TunnelId: 1,
		Nonce:    2,
	})
	require.NoError(t, err)

	tssPacket, ok := res.Packet.GetCachedValue().(*types.TSSPacket)
	require.True(t, ok)
	require.Equal(t, tssPackets[1], *tssPacket)
}
