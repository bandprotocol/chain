package keeper_test

import (
	"github.com/bandprotocol/chain/v2/pkg/tss/testutil"
	"github.com/bandprotocol/chain/v2/x/tss/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (s *KeeperTestSuite) TestSetInActive() {
	ctx, k := s.ctx, s.app.TSSKeeper
	s.SetupGroup(types.GROUP_STATUS_ACTIVE)
	address := sdk.AccAddress(testutil.TestCases[0].Group.Members[0].PubKey())

	// Success case
	err := k.SetInActive(ctx, address, 1)
	s.Require().NoError(err)

	status, err := k.GetStatus(ctx, address, 1)
	s.Require().NoError(err)
	s.Require().Equal(false, status.IsActive)

	// Failed case - no member
	err = k.SetInActive(ctx, address, 300)
	s.Require().Error(err)
}

func (s *KeeperTestSuite) TestSetActive() {
	ctx, k := s.ctx, s.app.TSSKeeper
	s.SetupGroup(types.GROUP_STATUS_ACTIVE)
	address := sdk.AccAddress(testutil.TestCases[0].Group.Members[0].PubKey())

	// Success case
	err := k.SetActive(ctx, address, 1)
	s.Require().NoError(err)

	status, err := k.GetStatus(ctx, address, 1)
	s.Require().NoError(err)
	s.Require().Equal(true, status.IsActive)

	// Failed case - penalty
	err = k.SetInActive(ctx, address, 1)
	s.Require().NoError(err)

	err = k.SetActive(ctx, address, 1)
	s.Require().ErrorIs(err, types.ErrTooSoonToActivate)

	// Failed case - no member
	err = k.SetActive(ctx, address, 300)
	s.Require().Error(err)
}
