package keeper_test

import (
	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/x/tunnel/types"
)

func (s *KeeperTestSuite) TestAddDeposit() {
	ctx, k := s.ctx, s.keeper

	tunnelID := uint64(1)
	depositorAddr := sdk.AccAddress([]byte("depositor"))
	depositAmount := sdk.NewCoins(sdk.NewCoin("uband", sdkmath.NewInt(100)))

	s.bankKeeper.EXPECT().
		SendCoinsFromAccountToModule(ctx, depositorAddr, types.ModuleName, depositAmount).
		Return(nil).Times(1)

	tunnel := types.Tunnel{ID: tunnelID, TotalDeposit: sdk.NewCoins()}
	k.SetTunnel(ctx, tunnel)

	err := k.AddDeposit(ctx, tunnelID, depositorAddr, depositAmount)
	s.Require().NoError(err)

	// check deposit
	deposit, found := k.GetDeposit(ctx, tunnelID, depositorAddr)
	s.Require().True(found)
	s.Require().Equal(depositAmount, deposit.Amount)

	// check tunnel's total deposit
	tunnel, err = k.GetTunnel(ctx, tunnelID)
	s.Require().NoError(err)
	s.Require().Equal(depositAmount, tunnel.TotalDeposit)
}

func (s *KeeperTestSuite) TestGetSetDeposit() {
	ctx, k := s.ctx, s.keeper

	tunnelID := uint64(1)
	depositorAddr := sdk.AccAddress([]byte("depositor"))
	depositAmount := sdk.NewCoins(sdk.NewCoin("band", sdkmath.NewInt(100)))

	deposit := types.Deposit{TunnelID: tunnelID, Depositor: depositorAddr.String(), Amount: depositAmount}
	k.SetDeposit(ctx, deposit)

	retrievedDeposit, found := k.GetDeposit(ctx, tunnelID, depositorAddr)
	s.Require().True(found)
	s.Require().Equal(deposit, retrievedDeposit)
}

func (s *KeeperTestSuite) TestGetDeposits() {
	ctx, k := s.ctx, s.keeper

	tunnelID1 := uint64(1)

	depositorAddr1 := sdk.AccAddress([]byte("depositor1"))
	depositorAddr2 := sdk.AccAddress([]byte("depositor2"))

	depositAmount := sdk.NewCoins(sdk.NewCoin("band", sdkmath.NewInt(100)))

	// add a deposit
	deposit1 := types.Deposit{TunnelID: tunnelID1, Depositor: depositorAddr1.String(), Amount: depositAmount}
	k.SetDeposit(ctx, deposit1)

	// add another deposit
	deposit2 := types.Deposit{TunnelID: tunnelID1, Depositor: depositorAddr2.String(), Amount: depositAmount}
	k.SetDeposit(ctx, deposit2)

	deposits := k.GetDeposits(ctx, tunnelID1)
	s.Require().Len(deposits, 2)
	s.Require().Equal(deposit1, deposits[0])
	s.Require().Equal(deposit2, deposits[1])
}

func (s *KeeperTestSuite) TestGetAllDeposits() {
	ctx, k := s.ctx, s.keeper

	tunnelID1 := uint64(1)
	depositorAddr1 := sdk.AccAddress([]byte("depositor"))
	depositAmount1 := sdk.NewCoins(sdk.NewCoin("band", sdkmath.NewInt(100)))

	deposit1 := types.Deposit{TunnelID: tunnelID1, Depositor: depositorAddr1.String(), Amount: depositAmount1}
	k.SetDeposit(ctx, deposit1)

	tunnelID2 := uint64(2)
	depositorAddr2 := sdk.AccAddress([]byte("depositor2"))
	depositAmount2 := sdk.NewCoins(sdk.NewCoin("band", sdkmath.NewInt(200)))

	deposit2 := types.Deposit{TunnelID: tunnelID2, Depositor: depositorAddr2.String(), Amount: depositAmount2}
	k.SetDeposit(ctx, deposit2)

	deposits := k.GetAllDeposits(ctx)
	s.Require().Len(deposits, 2)
	s.Require().Equal(deposit1, deposits[0])
	s.Require().Equal(deposit2, deposits[1])
}

func (s *KeeperTestSuite) TestDeleteDeposit() {
	ctx, k := s.ctx, s.keeper

	tunnelID := uint64(1)
	depositorAddr := sdk.AccAddress([]byte("depositor"))
	depositAmount := sdk.NewCoins(sdk.NewCoin("band", sdkmath.NewInt(100)))

	deposit := types.Deposit{TunnelID: tunnelID, Depositor: depositorAddr.String(), Amount: depositAmount}
	k.SetDeposit(ctx, deposit)

	k.DeleteDeposit(ctx, tunnelID, depositorAddr)

	_, found := k.GetDeposit(ctx, tunnelID, depositorAddr)
	s.Require().False(found)
}

func (s *KeeperTestSuite) TestWithdrawDeposit() {
	ctx, k := s.ctx, s.keeper

	tunnelID := uint64(1)
	depositorAddr := sdk.AccAddress([]byte("depositor"))
	depositAmount := sdk.NewCoins(sdk.NewCoin("band", sdkmath.NewInt(1000)))
	firstWithdraw := sdk.NewCoins(sdk.NewCoin("band", sdkmath.NewInt(500)))
	secondWithdraw := sdk.NewCoins(sdk.NewCoin("band", sdkmath.NewInt(500)))

	s.bankKeeper.EXPECT().
		SendCoinsFromModuleToAccount(ctx, types.ModuleName, depositorAddr, firstWithdraw).
		Return(nil).Times(1)
	s.bankKeeper.EXPECT().
		SendCoinsFromModuleToAccount(ctx, types.ModuleName, depositorAddr, secondWithdraw).
		Return(nil).Times(1)

	tunnel := types.Tunnel{ID: tunnelID, TotalDeposit: depositAmount, IsActive: true}
	k.SetTunnel(ctx, tunnel)

	deposit := types.Deposit{TunnelID: tunnelID, Depositor: depositorAddr.String(), Amount: depositAmount}
	k.SetDeposit(ctx, deposit)

	// partial withdraw
	err := k.WithdrawDeposit(ctx, tunnelID, firstWithdraw, depositorAddr)
	s.Require().NoError(err)

	// check tunnel's total deposit
	tunnel, err = k.GetTunnel(ctx, tunnelID)
	s.Require().NoError(err)
	s.Require().Equal(deposit.Amount.Sub(firstWithdraw...), tunnel.TotalDeposit)

	// withdraw all
	err = k.WithdrawDeposit(ctx, tunnelID, secondWithdraw, depositorAddr)
	s.Require().NoError(err)

	// check tunnel's total deposit
	tunnel, err = k.GetTunnel(ctx, tunnelID)
	s.Require().NoError(err)
	s.Require().Equal(sdk.Coins(nil), tunnel.TotalDeposit)

	// check is active
	s.Require().False(tunnel.IsActive)
}
