package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/tunnel/types"
)

func (s *KeeperTestSuite) TestAddDeposit() {
	ctx, k := s.ctx, s.keeper

	tunnelID := uint64(1)
	depositorAddr := sdk.AccAddress([]byte("depositor"))
	depositAmount := sdk.NewCoins(sdk.NewCoin("uband", sdk.NewInt(100)))

	s.bankKeeper.EXPECT().
		SendCoinsFromAccountToModule(ctx, depositorAddr, types.ModuleName, depositAmount).
		Return(nil).Times(1)

	// Add a tunnel
	tunnel := types.Tunnel{ID: tunnelID, TotalDeposit: sdk.NewCoins()}
	k.SetTunnel(ctx, tunnel)

	// Add deposit
	err := k.AddDeposit(ctx, tunnelID, depositorAddr, depositAmount)
	s.Require().NoError(err)

	// Check deposit
	deposit, found := k.GetDeposit(ctx, tunnelID, depositorAddr)
	s.Require().True(found)
	s.Require().Equal(depositAmount, deposit.Amount)

	// Check tunnel's total deposit
	tunnel, err = k.GetTunnel(ctx, tunnelID)
	s.Require().NoError(err)
	s.Require().Equal(depositAmount, tunnel.TotalDeposit)
}

func (s *KeeperTestSuite) TestGetSetDeposit() {
	ctx, k := s.ctx, s.keeper

	tunnelID := uint64(1)
	depositorAddr := sdk.AccAddress([]byte("depositor"))
	depositAmount := sdk.NewCoins(sdk.NewCoin("band", sdk.NewInt(100)))

	// Set deposit
	deposit := types.Deposit{TunnelID: tunnelID, Depositor: depositorAddr.String(), Amount: depositAmount}
	k.SetDeposit(ctx, deposit)

	// Get deposit
	retrievedDeposit, found := k.GetDeposit(ctx, tunnelID, depositorAddr)
	s.Require().True(found)
	s.Require().Equal(deposit, retrievedDeposit)
}

func (s *KeeperTestSuite) TestGetDeposits() {
	ctx, k := s.ctx, s.keeper

	tunnelID := uint64(1)
	depositorAddr := sdk.AccAddress([]byte("depositor"))
	depositAmount := sdk.NewCoins(sdk.NewCoin("band", sdk.NewInt(100)))

	// Set deposit
	deposit := types.Deposit{TunnelID: tunnelID, Depositor: depositorAddr.String(), Amount: depositAmount}
	k.SetDeposit(ctx, deposit)

	// Get deposits
	deposits := k.GetDeposits(ctx, tunnelID)
	s.Require().Len(deposits, 1)
	s.Require().Equal(deposit, deposits[0])
}

func (s *KeeperTestSuite) TestDeleteDeposit() {
	ctx, k := s.ctx, s.keeper

	tunnelID := uint64(1)
	depositorAddr := sdk.AccAddress([]byte("depositor"))
	depositAmount := sdk.NewCoins(sdk.NewCoin("band", sdk.NewInt(100)))

	// Set deposit
	deposit := types.Deposit{TunnelID: tunnelID, Depositor: depositorAddr.String(), Amount: depositAmount}
	k.SetDeposit(ctx, deposit)

	// Delete deposit
	k.DeleteDeposit(ctx, tunnelID, depositorAddr)

	// Check deposit
	_, found := k.GetDeposit(ctx, tunnelID, depositorAddr)
	s.Require().False(found)
}

func (s *KeeperTestSuite) TestWithdrawDeposit() {
	ctx, k := s.ctx, s.keeper

	tunnelID := uint64(1)
	depositorAddr := sdk.AccAddress([]byte("depositor"))
	depositAmount := sdk.NewCoins(sdk.NewCoin("band", sdk.NewInt(1000)))

	s.bankKeeper.EXPECT().
		SendCoinsFromModuleToAccount(ctx, types.ModuleName, depositorAddr, depositAmount).
		Return(nil).Times(1)

	// Set a tunnel
	tunnel := types.Tunnel{ID: tunnelID, TotalDeposit: depositAmount, IsActive: true}
	k.SetTunnel(ctx, tunnel)

	// Set deposit
	deposit := types.Deposit{TunnelID: tunnelID, Depositor: depositorAddr.String(), Amount: depositAmount}
	k.SetDeposit(ctx, deposit)

	// Withdraw deposit
	err := k.WithdrawDeposit(ctx, tunnelID, depositAmount, depositorAddr)
	s.Require().NoError(err)

	// Check tunnel's total deposit
	tunnel, err = k.GetTunnel(ctx, tunnelID)
	s.Require().NoError(err)
	s.Require().Equal(sdk.Coins(nil), tunnel.TotalDeposit)

	// Check is active
	s.Require().False(tunnel.IsActive)
}
