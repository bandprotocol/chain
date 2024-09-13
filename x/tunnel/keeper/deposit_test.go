package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/bandprotocol/chain/v2/x/tunnel/testutil"
	"github.com/bandprotocol/chain/v2/x/tunnel/types"
)

func TestAddDeposit(t *testing.T) {
	s := testutil.NewTestSuite(t)
	ctx, k := s.Ctx, s.Keeper

	tunnelID := uint64(1)
	depositorAddr := sdk.AccAddress([]byte("depositor"))
	depositAmount := sdk.NewCoins(sdk.NewCoin("band", sdk.NewInt(100)))

	s.MockBankKeeper.EXPECT().
		SendCoinsFromAccountToModule(ctx, depositorAddr, types.ModuleName, depositAmount).
		Return(nil).Times(1)

	// Add a tunnel
	tunnel := types.Tunnel{ID: tunnelID, TotalDeposit: sdk.NewCoins()}
	k.SetTunnel(ctx, tunnel)

	// Add deposit
	err := k.AddDeposit(ctx, tunnelID, depositorAddr, depositAmount)
	require.NoError(t, err)

	// Check deposit
	deposit, found := k.GetDeposit(ctx, tunnelID, depositorAddr)
	require.True(t, found)
	require.Equal(t, depositAmount, deposit.Amount)

	// Check tunnel's total deposit
	tunnel, err = k.GetTunnel(ctx, tunnelID)
	require.NoError(t, err)
	require.Equal(t, depositAmount, tunnel.TotalDeposit)
}

func TestGetSetDeposit(t *testing.T) {
	s := testutil.NewTestSuite(t)
	ctx, k := s.Ctx, s.Keeper

	tunnelID := uint64(1)
	depositorAddr := sdk.AccAddress([]byte("depositor"))
	depositAmount := sdk.NewCoins(sdk.NewCoin("band", sdk.NewInt(100)))

	// Set deposit
	deposit := types.Deposit{TunnelID: tunnelID, Depositor: depositorAddr.String(), Amount: depositAmount}
	k.SetDeposit(ctx, deposit)

	// Get deposit
	retrievedDeposit, found := k.GetDeposit(ctx, tunnelID, depositorAddr)
	require.True(t, found)
	require.Equal(t, deposit, retrievedDeposit)
}

func TestGetDeposits(t *testing.T) {
	s := testutil.NewTestSuite(t)
	ctx, k := s.Ctx, s.Keeper

	tunnelID := uint64(1)
	depositorAddr := sdk.AccAddress([]byte("depositor"))
	depositAmount := sdk.NewCoins(sdk.NewCoin("band", sdk.NewInt(100)))

	// Set deposit
	deposit := types.Deposit{TunnelID: tunnelID, Depositor: depositorAddr.String(), Amount: depositAmount}
	k.SetDeposit(ctx, deposit)

	// Get deposits
	deposits := k.GetDeposits(ctx, tunnelID)
	require.Len(t, deposits, 1)
	require.Equal(t, deposit, deposits[0])
}

func TestDeleteDeposit(t *testing.T) {
	s := testutil.NewTestSuite(t)
	ctx, k := s.Ctx, s.Keeper

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
	require.False(t, found)
}

func TestWithdrawDeposit(t *testing.T) {
	s := testutil.NewTestSuite(t)
	ctx, k := s.Ctx, s.Keeper

	tunnelID := uint64(1)
	depositorAddr := sdk.AccAddress([]byte("depositor"))
	depositAmount := sdk.NewCoins(sdk.NewCoin("band", sdk.NewInt(1000)))

	s.MockBankKeeper.EXPECT().
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
	require.NoError(t, err)

	// Check tunnel's total deposit
	tunnel, err = k.GetTunnel(ctx, tunnelID)
	require.NoError(t, err)
	require.Equal(t, sdk.Coins(nil), tunnel.TotalDeposit)

	// Check is active
	require.False(t, tunnel.IsActive)
}
