package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/pkg/tss/testutil"
	"github.com/bandprotocol/chain/v2/x/bandtss/types"
	tsstypes "github.com/bandprotocol/chain/v2/x/tss/types"
)

func (s *KeeperTestSuite) TestSetInActive() {
	ctx, k := s.ctx, s.app.BandTSSKeeper
	s.SetupGroup(tsstypes.GROUP_STATUS_ACTIVE)
	address := sdk.AccAddress(testutil.TestCases[0].Group.Members[0].PubKey())

	k.SetInactiveStatus(ctx, address)

	status := k.GetStatus(ctx, address)
	s.Require().Equal(types.MEMBER_STATUS_INACTIVE, status.Status)
}

func (s *KeeperTestSuite) TestSetActive() {
	ctx, k := s.ctx, s.app.BandTSSKeeper
	s.SetupGroup(tsstypes.GROUP_STATUS_ACTIVE)
	address := sdk.AccAddress(testutil.TestCases[0].Group.Members[0].PubKey())

	// Success case
	err := k.SetActiveStatus(ctx, address)
	s.Require().NoError(err)

	status := k.GetStatus(ctx, address)
	s.Require().Equal(types.MEMBER_STATUS_ACTIVE, status.Status)

	// Failed case - penalty
	k.SetInactiveStatus(ctx, address)

	err = k.SetActiveStatus(ctx, address)
	s.Require().ErrorIs(err, tsstypes.ErrTooSoonToActivate)

	// Failed case - no member
	err = k.SetActiveStatus(ctx, address)
	s.Require().Error(err)
}

func (s *KeeperTestSuite) TestSetLastActive() {
	ctx, k := s.ctx, s.app.BandTSSKeeper
	s.SetupGroup(tsstypes.GROUP_STATUS_ACTIVE)
	address := sdk.AccAddress(testutil.TestCases[0].Group.Members[0].PubKey())

	// Success case
	err := k.SetLastActive(ctx, address)
	s.Require().NoError(err)

	status := k.GetStatus(ctx, address)
	s.Require().Equal(ctx.BlockTime(), status.LastActive)

	// Failed case
	k.SetInactiveStatus(ctx, address)

	err = k.SetLastActive(ctx, address)
	s.Require().Error(err)
}
