package keeper_test

import (
	"github.com/bandprotocol/chain/v3/x/bandtss/types"
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
	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(k.GetParams(ctx).InactivePenaltyDuration))

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
	s.Require().ErrorIs(err, types.ErrPenaltyDurationNotElapsed)

	// Failed case - no member
	err = k.ActivateMember(ctx, address, groupCtx.GroupID)
	s.Require().Error(err)
}
