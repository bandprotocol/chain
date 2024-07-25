package keeper_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bandprotocol/chain/v2/x/tunnel/testutil"
	"github.com/bandprotocol/chain/v2/x/tunnel/types"
)

func TestGRPCQueryPackets(t *testing.T) {
	s := testutil.NewTestSuite(t)
	// q := s.QueryServer

	tunnel := types.Tunnel{
		ID:         1,
		NonceCount: 2,
	}
	r := types.TSSRoute{
		DestinationChainID:         "1",
		DestinationContractAddress: "0x123",
	}
	err := tunnel.SetRoute(&r)
	a := tunnel.Route.GetCachedValue()
	fmt.Printf("route: %+v\n", a)
	require.NoError(t, err)

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

	s.Keeper.SetTunnel(s.Ctx, tunnel)
	for _, packet := range tssPackets {
		s.Keeper.SetTSSPacket(s.Ctx, packet)
	}

	tu, err := s.Keeper.GetTunnel(s.Ctx, 1)
	require.NoError(t, err)

	ss, err := tu.UnpackRoute()
	require.NoError(t, err)

	fmt.Printf("route2 %+v\n", ss)

	// ts := s.Keeper.GetTunnels(s.Ctx)
	// for _, tunnel := range ts {
	// 	a := tunnel.Route.GetCachedValue()
	// 	fmt.Printf("route: %+v\n", a)
	// }

	// res, err := q.Packets(s.Ctx, &types.QueryPacketsRequest{
	// 	TunnelId: 1,
	// })
	// require.NoError(t, err)
	// require.Len(t, res, len(res.Packets))
}
