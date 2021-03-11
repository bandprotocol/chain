package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/bandprotocol/chain/x/oracle/testapp"
)

func TestDepositRequestPool(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true)

	err := k.DepositRequestPool(ctx, "beeb", testapp.Port1, testapp.Channel1, testapp.Coins1unband, testapp.Alice.Address)
	require.Nil(t, err)
	poolBalance := k.GetRequetPoolBalances(ctx, "beeb", testapp.Port1, testapp.Channel1)
	require.Equal(t, testapp.Coins1unband[0].Amount, poolBalance[0].Amount)

	err = k.DepositRequestPool(ctx, "beeb2", testapp.Port2, testapp.Channel2, testapp.Coins10unband, testapp.Alice.Address)
	require.Nil(t, err)
	poolBalance = k.GetRequetPoolBalances(ctx, "beeb2", testapp.Port2, testapp.Channel2)
	require.Equal(t, testapp.Coins10unband[0].Amount, poolBalance[0].Amount)

	err = k.DepositRequestPool(ctx, "beeb", testapp.Port1, testapp.Channel1, testapp.Coins10unband, testapp.Alice.Address)
	require.Nil(t, err)
	poolBalance = k.GetRequetPoolBalances(ctx, "beeb", testapp.Port1, testapp.Channel1)
	require.Equal(t, testapp.Coins11unband[0].Amount, poolBalance[0].Amount)
}

func TestGetFromNonRequestPool(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true)
	poolBalance := k.GetRequetPoolBalances(ctx, "beeb", testapp.Port1, testapp.Channel1)
	require.Equal(t, sdk.Coins{}, poolBalance)
}
