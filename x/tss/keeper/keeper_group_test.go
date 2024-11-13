package keeper_test

import (
	"github.com/bandprotocol/chain/v3/x/tss/types"
)

func (s *KeeperTestSuite) TestGetSetGroupCount() {
	ctx, k := s.ctx, s.keeper
	k.AddGroup(ctx, 3, 2, "test")
	k.AddGroup(ctx, 4, 3, "test")

	groupCount := k.GetGroupCount(ctx)
	s.Require().Equal(uint64(2), groupCount)
}

func (s *KeeperTestSuite) TestGetGroups() {
	ctx, k := s.ctx, s.keeper
	group := types.Group{
		ID:        1,
		Size_:     5,
		Threshold: 3,
		PubKey:    nil,
		Status:    types.GROUP_STATUS_ROUND_1,
	}

	// Set new group
	k.SetGroup(ctx, group)

	// Get group from chain state
	got := k.GetGroups(ctx)
	s.Require().Equal([]types.Group{group}, got)
}

func (s *KeeperTestSuite) TestGetSetDKGContext() {
	ctx, k := s.ctx, s.keeper

	dkgContext := []byte("dkg-context sample")
	k.SetDKGContext(ctx, 1, dkgContext)

	got, err := k.GetDKGContext(ctx, 1)
	s.Require().NoError(err)
	s.Require().Equal(dkgContext, got)
}

func (s *KeeperTestSuite) TestCreateNewGroup() {
	ctx, k := s.ctx, s.keeper

	group := types.Group{
		Size_:       5,
		Threshold:   3,
		PubKey:      nil,
		Status:      types.GROUP_STATUS_ROUND_1,
		ModuleOwner: "test",
	}

	// Create new group
	groupID := k.AddGroup(ctx, group.Size_, group.Threshold, group.ModuleOwner)

	// init group ID
	group.ID = groupID

	// Get group by id
	got, err := k.GetGroup(ctx, groupID)
	s.Require().NoError(err)
	s.Require().Equal(group, got)
}

func (s *KeeperTestSuite) TestSetGroup() {
	ctx, k := s.ctx, s.keeper
	group := types.Group{
		Size_:       5,
		Threshold:   3,
		PubKey:      nil,
		Status:      types.GROUP_STATUS_ROUND_1,
		ModuleOwner: "test",
	}

	// Set new group
	groupID := k.AddGroup(ctx, group.Size_, group.Threshold, group.ModuleOwner)

	// Update group size value
	group.Size_ = 6

	// Add group ID
	group.ID = groupID

	k.SetGroup(ctx, group)

	// Get group from chain state
	got, err := k.GetGroup(ctx, groupID)

	// Validate group size value
	s.Require().NoError(err)
	s.Require().Equal(group.Size_, got.Size_)
}
