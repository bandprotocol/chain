package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bandprotocol/chain/x/oracle/testapp"
)

func TestDepositRequestPool(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true)
	require.False(t, k.HasRequest(ctx, 42))
	err := k.DepositRequestPool(ctx, "beeb", "port-1", "channel-1", testapp.Coins1000000uband, testapp.Alice.Address)
	require.Nil(t, err)
	poolBalance := k.GetRequetPoolBalance(ctx, "beeb", "port-1", "channel-1")
	require.Equal(t, testapp.Coins1000000uband[0].Amount, poolBalance[0].Amount)
}
