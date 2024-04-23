package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/pkg/tss/testutil"
	"github.com/bandprotocol/chain/v2/x/bandtss/types"
	tsstypes "github.com/bandprotocol/chain/v2/x/tss/types"
)

func (s *KeeperTestSuite) TestSetInActive() {
	ctx, k, tssKeeper := s.ctx, s.app.BandtssKeeper, s.app.TSSKeeper
	s.SetupGroup(tsstypes.GROUP_STATUS_ACTIVE)
	address := sdk.AccAddress(testutil.TestCases[0].Group.Members[0].PubKey())

	err := k.DeactivateMember(ctx, address)
	s.Require().NoError(err)

	member, err := k.GetMember(ctx, address)
	s.Require().NoError(err)
	s.Require().False(member.IsActive)

	tssMember, err := tssKeeper.GetMemberByAddress(ctx, testutil.TestCases[0].Group.ID, address.String())
	s.Require().NoError(err)
	s.Require().False(tssMember.IsActive)
}

func (s *KeeperTestSuite) TestActivateMember() {
	ctx, k, tssKeeper := s.ctx, s.app.BandtssKeeper, s.app.TSSKeeper
	s.SetupGroup(tsstypes.GROUP_STATUS_ACTIVE)
	address := sdk.AccAddress(testutil.TestCases[0].Group.Members[0].PubKey())

	// Success case
	err := k.DeactivateMember(ctx, address)
	s.Require().NoError(err)
	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(k.GetParams(ctx).ActiveDuration))

	err = k.ActivateMember(ctx, address)
	s.Require().NoError(err)

	member, err := k.GetMember(ctx, address)
	s.Require().NoError(err)
	s.Require().True(member.IsActive)

	tssMember, err := tssKeeper.GetMemberByAddress(ctx, testutil.TestCases[0].Group.ID, address.String())
	s.Require().NoError(err)
	s.Require().True(tssMember.IsActive)

	// Failed case - penalty
	err = k.DeactivateMember(ctx, address)
	s.Require().NoError(err)

	err = k.ActivateMember(ctx, address)
	s.Require().ErrorIs(err, types.ErrTooSoonToActivate)

	// Failed case - no member
	err = k.ActivateMember(ctx, address)
	s.Require().Error(err)
}

func (s *KeeperTestSuite) TestSetLastActive() {
	ctx, k := s.ctx, s.app.BandtssKeeper
	s.SetupGroup(tsstypes.GROUP_STATUS_ACTIVE)
	address := sdk.AccAddress(testutil.TestCases[0].Group.Members[0].PubKey())

	// Success case
	err := k.SetLastActive(ctx, address)
	s.Require().NoError(err)

	member, err := k.GetMember(ctx, address)
	s.Require().NoError(err)
	s.Require().Equal(ctx.BlockTime(), member.LastActive)

	// Failed case
	err = k.DeactivateMember(ctx, address)
	s.Require().NoError(err)

	err = k.SetLastActive(ctx, address)
	s.Require().Error(err)
}
