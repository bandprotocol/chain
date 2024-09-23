package keeper_test

import (
	"fmt"

	"go.uber.org/mock/gomock"
)

func (s *KeeperTestSuite) TestGenerateAccount() {
	ctx, k := s.ctx, s.keeper

	tunnelID := uint64(1)
	s.accountKeeper.EXPECT().
		GetAccount(ctx, gomock.Any()).
		Return(nil).Times(1)
	s.accountKeeper.EXPECT().NewAccount(ctx, gomock.Any()).Times(1)
	s.accountKeeper.EXPECT().SetAccount(ctx, gomock.Any()).Times(1)

	addr, err := k.GenerateAccount(ctx, fmt.Sprintf("%d", tunnelID))
	s.Require().NoError(err, "expected no error generating account")
	s.Require().NotNil(addr, "expected generated address to be non-nil")
	s.Require().Equal(
		"band1mdnfc2ehu7vkkg5nttc8tuvwpa9f3dxskf75yxfr7zwhevvcj62q2yggu0",
		addr.String(),
		"expected generated address to match",
	)
}
