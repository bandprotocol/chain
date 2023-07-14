package keeper_test

import (
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

func (s *KeeperTestSuite) TestSetInActive() {
	ctx, k := s.ctx, s.app.TSSKeeper
	s.SetupGroup(types.GROUP_STATUS_ACTIVE)

	// Success case
	err := k.SetInActive(ctx, 1, 1)
	s.Require().NoError(err)

	member, err := k.GetMember(ctx, 1, 1)
	s.Require().NoError(err)
	s.Require().Equal(false, member.IsActive)

	// Failed case - no member
	err = k.SetInActive(ctx, 1, 300)
	s.Require().Error(err)
}

func (s *KeeperTestSuite) TestSetActive() {
	ctx, k := s.ctx, s.app.TSSKeeper
	s.SetupGroup(types.GROUP_STATUS_ACTIVE)

	// Success case
	err := k.SetInActive(ctx, 1, 1)
	s.Require().NoError(err)

	member, err := k.GetMember(ctx, 1, 1)
	s.Require().NoError(err)
	s.Require().Equal(false, member.IsActive)

	// Failed case - no member
	err = k.SetInActive(ctx, 1, 300)
	s.Require().Error(err)
}
