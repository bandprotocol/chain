package keeper_test

import (
	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/x/tunnel/types"
)

func (s *KeeperTestSuite) TestGRPCQueryTunnels() {
	ctx, k, q := s.ctx, s.keeper, s.queryServer

	tunnel1 := types.Tunnel{
		ID: 1,
	}
	tunnel2 := types.Tunnel{
		ID: 2,
	}
	k.SetTunnel(ctx, tunnel1)
	k.SetTunnel(ctx, tunnel2)

	resp, err := q.Tunnels(ctx, &types.QueryTunnelsRequest{})
	s.Require().NoError(err)
	s.Require().NotNil(resp)
	s.Require().Len(resp.Tunnels, 2)
	s.Require().Equal(tunnel1, *resp.Tunnels[0])
	s.Require().Equal(tunnel2, *resp.Tunnels[1])
}

func (s *KeeperTestSuite) TestGRPCQueryTunnel() {
	ctx, k, q := s.ctx, s.keeper, s.queryServer

	tunnel := types.Tunnel{
		ID: 1,
	}
	k.SetTunnel(ctx, tunnel)

	resp, err := q.Tunnel(ctx, &types.QueryTunnelRequest{
		TunnelId: 1,
	})
	s.Require().NoError(err)
	s.Require().NotNil(resp)
	s.Require().Equal(tunnel, resp.Tunnel)
}

func (s *KeeperTestSuite) TestGRPCQueryDeposits() {
	ctx, k, q := s.ctx, s.keeper, s.queryServer

	tunnel := types.Tunnel{
		ID:       1,
		Sequence: 2,
	}
	r := types.TSSRoute{
		DestinationChainID:         "1",
		DestinationContractAddress: "0x123",
	}
	err := tunnel.SetRoute(&r)
	s.Require().NoError(err)
	k.SetTunnel(ctx, tunnel)

	deposit1 := types.Deposit{
		TunnelID:  1,
		Depositor: sdk.AccAddress([]byte("depositor1")).String(),
		Amount:    sdk.NewCoins(sdk.NewCoin("band", sdkmath.NewInt(100))),
	}
	deposit2 := types.Deposit{
		TunnelID:  1,
		Depositor: sdk.AccAddress([]byte("depositor2")).String(),
		Amount:    sdk.NewCoins(sdk.NewCoin("band", sdkmath.NewInt(110))),
	}
	k.SetDeposit(ctx, deposit1)
	k.SetDeposit(ctx, deposit2)

	resp, err := q.Deposits(ctx, &types.QueryDepositsRequest{
		TunnelId: 1,
	})
	s.Require().NoError(err)
	s.Require().NotNil(resp)
	s.Require().Len(resp.Deposits, 2)
	s.Require().Equal(deposit1, *resp.Deposits[0])
	s.Require().Equal(deposit2, *resp.Deposits[1])
}

func (s *KeeperTestSuite) TestGRPCQueryDeposit() {
	ctx, k, q := s.ctx, s.keeper, s.queryServer

	tunnel := types.Tunnel{
		ID:       1,
		Sequence: 2,
	}
	r := types.TSSRoute{
		DestinationChainID:         "1",
		DestinationContractAddress: "0x123",
	}
	err := tunnel.SetRoute(&r)
	s.Require().NoError(err)
	k.SetTunnel(ctx, tunnel)

	deposit1 := types.Deposit{
		TunnelID:  1,
		Depositor: sdk.AccAddress([]byte("depositor")).String(),
		Amount:    sdk.NewCoins(sdk.NewCoin("band", sdkmath.NewInt(100))),
	}
	k.SetDeposit(ctx, deposit1)

	deposit2 := types.Deposit{
		TunnelID:  1,
		Depositor: sdk.AccAddress([]byte("depositor")).String(),
		Amount:    sdk.NewCoins(sdk.NewCoin("band", sdkmath.NewInt(100))),
	}
	k.SetDeposit(ctx, deposit2)

	resp, err := q.Deposit(ctx, &types.QueryDepositRequest{
		TunnelId:  1,
		Depositor: deposit1.Depositor,
	})
	s.Require().NoError(err)
	s.Require().NotNil(resp)
	s.Require().Equal(deposit1, resp.Deposit)
}

func (s *KeeperTestSuite) TestGRPCQueryPackets() {
	ctx, k, q := s.ctx, s.keeper, s.queryServer

	tunnel := types.Tunnel{
		ID:       1,
		Sequence: 2,
	}
	r := types.TSSRoute{
		DestinationChainID:         "1",
		DestinationContractAddress: "0x123",
	}
	err := tunnel.SetRoute(&r)
	s.Require().NoError(err)

	k.SetTunnel(ctx, tunnel)

	packet1 := types.Packet{
		TunnelID: 1,
		Sequence: 1,
	}
	packet2 := types.Packet{
		TunnelID: 1,
		Sequence: 2,
	}
	packet3 := types.Packet{
		TunnelID: 2,
		Sequence: 1,
	}
	err = packet1.SetReceipt(&types.TSSPacketReceipt{
		SigningID: 1,
	})
	s.Require().NoError(err)
	err = packet2.SetReceipt(&types.TSSPacketReceipt{
		SigningID: 2,
	})
	s.Require().NoError(err)
	err = packet3.SetReceipt(&types.TSSPacketReceipt{
		SigningID: 3,
	})
	s.Require().NoError(err)

	k.SetPacket(ctx, packet1)
	k.SetPacket(ctx, packet2)
	k.SetPacket(ctx, packet3)

	resp, err := q.Packets(ctx, &types.QueryPacketsRequest{
		TunnelId: 1,
	})
	s.Require().NoError(err)
	s.Require().NotNil(resp)
	s.Require().Len(resp.Packets, 2)
	s.Require().Equal(packet1, *resp.Packets[0])
	s.Require().Equal(packet2, *resp.Packets[1])
}

func (s *KeeperTestSuite) TestGRPCQueryPacket() {
	ctx, k, q := s.ctx, s.keeper, s.queryServer

	// set tunnel
	tunnel := types.Tunnel{
		ID:       1,
		Sequence: 2,
	}
	r := types.TSSRoute{
		DestinationChainID:         "1",
		DestinationContractAddress: "0x123",
	}
	err := tunnel.SetRoute(&r)
	s.Require().NoError(err)
	k.SetTunnel(ctx, tunnel)

	packet1 := types.Packet{
		TunnelID: 1,
		Sequence: 1,
	}
	err = packet1.SetReceipt(&types.TSSPacketReceipt{
		SigningID: 1,
	})
	s.Require().NoError(err)
	k.SetPacket(ctx, packet1)

	packet2 := types.Packet{
		TunnelID: 1,
		Sequence: 2,
	}
	err = packet2.SetReceipt(&types.TSSPacketReceipt{})
	s.Require().NoError(err)
	k.SetPacket(ctx, packet2)

	res, err := q.Packet(ctx, &types.QueryPacketRequest{
		TunnelId: 1,
		Sequence: 1,
	})
	s.Require().NoError(err)
	s.Require().NotNil(res)
	s.Require().Equal(packet1, *res.Packet)
}

func (s *KeeperTestSuite) TestGRPCQueryTotalFees() {
	ctx, k, q := s.ctx, s.keeper, s.queryServer

	// Set total fees
	totalFees := types.TotalFees{
		TotalBasePacketFee: sdk.NewCoins(sdk.NewCoin("band", sdkmath.NewInt(100))),
	}
	k.SetTotalFees(ctx, totalFees)

	// Query total fees
	res, err := q.TotalFees(ctx, &types.QueryTotalFeesRequest{})
	s.Require().NoError(err)
	s.Require().Equal(totalFees, res.TotalFees)
}

func (s *KeeperTestSuite) TestGRPCQueryParams() {
	ctx, k, q := s.ctx, s.keeper, s.queryServer

	// Set params
	err := k.SetParams(ctx, types.DefaultParams())
	s.Require().NoError(err)

	// Query params
	res, err := q.Params(ctx, &types.QueryParamsRequest{})
	s.Require().NoError(err)
	s.Require().Equal(types.DefaultParams(), res.Params)
}
