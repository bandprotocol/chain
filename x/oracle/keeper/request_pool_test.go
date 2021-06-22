package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/bandprotocol/chain/v2/testing/testapp"
)

func TestDepositRequestPool(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true)

	err := k.DepositRequestPool(ctx, "beeb", testapp.Port1, testapp.Channel1, testapp.Coins1uband, testapp.Alice.Address)
	require.Nil(t, err)
	poolBalance := k.GetRequestPoolBalances(ctx, "beeb", testapp.Port1, testapp.Channel1)
	require.Equal(t, testapp.Coins1uband[0].Amount, poolBalance[0].Amount)

	err = k.DepositRequestPool(ctx, "beeb2", testapp.Port2, testapp.Channel2, testapp.Coins10uband, testapp.Alice.Address)
	require.Nil(t, err)
	poolBalance = k.GetRequestPoolBalances(ctx, "beeb2", testapp.Port2, testapp.Channel2)
	require.Equal(t, testapp.Coins10uband[0].Amount, poolBalance[0].Amount)

	err = k.DepositRequestPool(ctx, "beeb", testapp.Port1, testapp.Channel1, testapp.Coins10uband, testapp.Alice.Address)
	require.Nil(t, err)
	poolBalance = k.GetRequestPoolBalances(ctx, "beeb", testapp.Port1, testapp.Channel1)
	require.Equal(t, testapp.Coins11uband[0].Amount, poolBalance[0].Amount)
}

func TestGetFromNonRequestPool(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true)
	poolBalance := k.GetRequestPoolBalances(ctx, "beeb", testapp.Port1, testapp.Channel1)
	require.Equal(t, sdk.Coins{}, poolBalance)
}
