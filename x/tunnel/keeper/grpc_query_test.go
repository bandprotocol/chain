package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/bandprotocol/chain/v2/x/tunnel/testutil"
	"github.com/bandprotocol/chain/v2/x/tunnel/types"
)

func TestGRPCQueryPackets(t *testing.T) {
	s := testutil.NewTestSuite(t)
	q := s.QueryServer

	// Set tunnel
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

	// Set packets
	packet1 := types.Packet{
		TunnelID: 1,
		Nonce:    1,
	}
	packet2 := types.Packet{
		TunnelID: 1,
		Nonce:    2,
	}
	err = packet1.SetPacketContent(&types.TSSPacketContent{
		SigningID:                  1,
		DestinationChainID:         r.DestinationChainID,
		DestinationContractAddress: r.DestinationContractAddress,
	})
	require.NoError(t, err)
	err = packet2.SetPacketContent(&types.TSSPacketContent{
		SigningID:                  2,
		DestinationChainID:         r.DestinationChainID,
		DestinationContractAddress: r.DestinationContractAddress,
	})
	require.NoError(t, err)
	s.Keeper.SetPacket(s.Ctx, packet1)
	s.Keeper.SetPacket(s.Ctx, packet2)

	// Query packets
	resp, err := q.Packets(s.Ctx, &types.QueryPacketsRequest{
		TunnelId: 1,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Packets, 2)
	require.Equal(t, packet1, *resp.Packets[0])
	require.Equal(t, packet2, *resp.Packets[1])
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

	packet1 := types.Packet{
		TunnelID: 1,
		Nonce:    1,
	}
	err = packet1.SetPacketContent(&types.TSSPacketContent{
		SigningID:                  1,
		DestinationChainID:         r.DestinationChainID,
		DestinationContractAddress: r.DestinationContractAddress,
	})
	require.NoError(t, err)
	s.Keeper.SetPacket(s.Ctx, packet1)

	res, err := q.Packet(s.Ctx, &types.QueryPacketRequest{
		TunnelId: 1,
		Nonce:    1,
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, packet1, *res.Packet)
}

func TestGRPCQueryDeposit(t *testing.T) {
	s := testutil.NewTestSuite(t)
	q := s.QueryServer

	// Set tunnel
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

	// Set deposit
	deposit := types.Deposit{
		TunnelID:  1,
		Depositor: sdk.AccAddress([]byte("depositor")).String(),
		Amount:    sdk.NewCoins(sdk.NewCoin("band", sdk.NewInt(100))),
	}
	s.Keeper.SetDeposit(s.Ctx, deposit)

	// Query deposit
	resp, err := q.Deposit(s.Ctx, &types.QueryDepositRequest{
		TunnelId:  1,
		Depositor: deposit.Depositor,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, deposit, resp.Deposit)
}

func TestGRPCQueryDeposits(t *testing.T) {
	s := testutil.NewTestSuite(t)
	q := s.QueryServer

	// Set tunnel
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

	// Set deposits
	deposit1 := types.Deposit{
		TunnelID:  1,
		Depositor: sdk.AccAddress([]byte("depositor1")).String(),
		Amount:    sdk.NewCoins(sdk.NewCoin("band", sdk.NewInt(100))),
	}
	deposit2 := types.Deposit{
		TunnelID:  1,
		Depositor: sdk.AccAddress([]byte("depositor2")).String(),
		Amount:    sdk.NewCoins(sdk.NewCoin("band", sdk.NewInt(110))),
	}
	s.Keeper.SetDeposit(s.Ctx, deposit1)
	s.Keeper.SetDeposit(s.Ctx, deposit2)

	// Query deposits
	resp, err := q.Deposits(s.Ctx, &types.QueryDepositsRequest{
		TunnelId: 1,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Deposits, 2)
	require.Equal(t, deposit1, *resp.Deposits[0])
	require.Equal(t, deposit2, *resp.Deposits[1])
}
