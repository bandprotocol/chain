package keeper_test

import (
	"time"

	"github.com/bandprotocol/chain/v3/x/bandtss/types"
	tsstypes "github.com/bandprotocol/chain/v3/x/tss/types"
)

func (s *AppTestSuite) TestSetInactive() {
	ctx, k, tssKeeper := s.ctx, s.app.BandtssKeeper, s.app.TSSKeeper
	groupCtx := s.SetupNewGroup(5, 3)
	address := groupCtx.Accounts[0].Address

	err := k.DeactivateMember(ctx, address, groupCtx.GroupID)
	s.Require().NoError(err)

	member, err := k.GetMember(ctx, address, groupCtx.GroupID)
	s.Require().NoError(err)
	s.Require().False(member.IsActive)

	tssMember, err := tssKeeper.GetMemberByAddress(ctx, groupCtx.GroupID, address.String())
	s.Require().NoError(err)
	s.Require().False(tssMember.IsActive)
}

func (s *AppTestSuite) TestActivateMember() {
	ctx, k, tssKeeper := s.ctx, s.app.BandtssKeeper, s.app.TSSKeeper
	groupCtx := s.SetupNewGroup(5, 3)
	address := groupCtx.Accounts[0].Address

	// Success case
	err := k.DeactivateMember(ctx, address, groupCtx.GroupID)
	s.Require().NoError(err)
	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(k.GetParams(ctx).ActiveDuration))

	err = k.ActivateMember(ctx, address, groupCtx.GroupID)
	s.Require().NoError(err)

	member, err := k.GetMember(ctx, address, groupCtx.GroupID)
	s.Require().NoError(err)
	s.Require().True(member.IsActive)

	tssMember, err := tssKeeper.GetMemberByAddress(ctx, groupCtx.GroupID, address.String())
	s.Require().NoError(err)
	s.Require().True(tssMember.IsActive)

	// Failed case - penalty
	err = k.DeactivateMember(ctx, address, groupCtx.GroupID)
	s.Require().NoError(err)

	err = k.ActivateMember(ctx, address, groupCtx.GroupID)
	s.Require().ErrorIs(err, types.ErrTooSoonToActivate)

	// Failed case - no member
	err = k.ActivateMember(ctx, address, groupCtx.GroupID)
	s.Require().Error(err)
}

func (s *AppTestSuite) TestSetLastActive() {
	ctx, k := s.ctx, s.app.BandtssKeeper
	groupCtx := s.SetupNewGroup(5, 3)
	address := groupCtx.Accounts[0].Address

	// Success case
	err := k.SetLastActive(ctx, address, groupCtx.GroupID)
	s.Require().NoError(err)

	member, err := k.GetMember(ctx, address, groupCtx.GroupID)
	s.Require().NoError(err)
	s.Require().Equal(ctx.BlockTime(), member.LastActive)

	// Failed case
	err = k.DeactivateMember(ctx, address, groupCtx.GroupID)
	s.Require().NoError(err)

	err = k.SetLastActive(ctx, address, groupCtx.GroupID)
	s.Require().Error(err)
}

func (s *AppTestSuite) TestHandleInactiveMembers() {
	ctx, k := s.ctx, s.app.BandtssKeeper
	groupCtx := s.SetupNewGroup(3, 2)
	address := groupCtx.Accounts[0].Address

	m, err := k.GetMember(ctx, address, groupCtx.GroupID)
	s.Require().NoError(err)
	s.Require().True(m.IsActive)

	m.LastActive = time.Time{}
	k.SetMember(ctx, m)
	s.app.TSSKeeper.SetMember(ctx, tsstypes.Member{
		ID:       1,
		GroupID:  1,
		Address:  address.String(),
		IsActive: true,
	})
	ctx = ctx.WithBlockTime(time.Now())

	k.HandleInactiveMembers(ctx)

	member, err := k.GetMember(ctx, address, groupCtx.GroupID)
	s.Require().NoError(err)
	s.Require().False(member.IsActive)
}
