package keeper_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/bandprotocol/chain/v2/x/tunnel/testutil"
)

func TestGenerateAccount(t *testing.T) {
	s := testutil.NewTestSuite(t)
	ctx, k := s.Ctx, s.Keeper

	tunnelID := uint64(1)
	s.MockAccountKeeper.EXPECT().
		GetAccount(ctx, gomock.Any()).
		Return(nil).Times(1)
	s.MockAccountKeeper.EXPECT().NewAccount(ctx, gomock.Any()).Times(1)
	s.MockAccountKeeper.EXPECT().SetAccount(ctx, gomock.Any()).Times(1)

	addr, err := k.GenerateAccount(ctx, fmt.Sprintf("%d", tunnelID))
	require.NoError(s.T(), err, "expected no error generating account")
	require.NotNil(s.T(), addr, "expected generated address to be non-nil")
	require.Equal(
		s.T(),
		"cosmos1w66ct9dvwauhu68t7vt2y7gz3z73qc5kap98mzg5t0y06r3txc8spuqw0g",
		addr.String(),
		"expected generated address to match",
	)
}
