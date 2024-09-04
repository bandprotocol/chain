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
		"band1mdnfc2ehu7vkkg5nttc8tuvwpa9f3dxskf75yxfr7zwhevvcj62q2yggu0",
		addr.String(),
		"expected generated address to match",
	)
}
