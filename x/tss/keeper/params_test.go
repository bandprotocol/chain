package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bandprotocol/chain/v2/x/tss/types"
)

func TestGetSetParams(t *testing.T) {
	s := NewKeeperTestSuite(t)
	ctx, k := s.Ctx, s.Keeper
	params := types.DefaultParams()

	err := k.SetParams(ctx, params)
	require.NoError(t, err)

	require.Equal(t, params, k.GetParams(ctx))
}
