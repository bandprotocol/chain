package keeper_test

import (
	"github.com/bandprotocol/chain/v3/pkg/tss"
	"github.com/bandprotocol/chain/v3/x/tss/types"
)

func (s *KeeperTestSuite) TestGetSetGroupCount() {
	ctx, k := s.ctx, s.keeper
	k.CreateNewGroup(ctx, 3, 2, "test")
	k.CreateNewGroup(ctx, 4, 3, "test")

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
	groupID := k.CreateNewGroup(ctx, group.Size_, group.Threshold, group.ModuleOwner)

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
	groupID := k.CreateNewGroup(ctx, group.Size_, group.Threshold, group.ModuleOwner)

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

func (s *KeeperTestSuite) TestSetLastExpiredGroupID() {
	ctx, k := s.ctx, s.keeper
	groupID := tss.GroupID(1)
	k.SetLastExpiredGroupID(ctx, groupID)

	got := k.GetLastExpiredGroupID(ctx)
	s.Require().Equal(groupID, got)
}

func (s *KeeperTestSuite) TestGetSetLastExpiredGroupID() {
	ctx, k := s.ctx, s.keeper

	// Set the last expired group ID
	groupID := tss.GroupID(98765)
	k.SetLastExpiredGroupID(ctx, groupID)

	// Get the last expired group ID
	got := k.GetLastExpiredGroupID(ctx)

	// Assert equality
	s.Require().Equal(groupID, got)
}

func (s *KeeperTestSuite) TestProcessExpiredGroups() {
	ctx, k := s.ctx, s.keeper

	// Create group
	groupID := k.CreateNewGroup(ctx, 3, 2, "test")
	k.SetMember(ctx, types.Member{
		ID:          1,
		GroupID:     groupID,
		Address:     "band1p40yh3zkmhcv0ecqp3mcazy83sa57rgjp07dun",
		PubKey:      nil,
		IsMalicious: false,
	})

	// Set the current block height
	blockHeight := int64(30001)
	ctx = ctx.WithBlockHeight(blockHeight)

	// Handle expired groups
	k.HandleExpiredGroups(ctx)

	// Assert that the last expired group ID is updated correctly
	lastExpiredGroupID := k.GetLastExpiredGroupID(ctx)
	s.Require().Equal(groupID, lastExpiredGroupID)
}

func (s *KeeperTestSuite) TestGetSetPendingProcessGroups() {
	ctx, k := s.ctx, s.keeper
	groupID := tss.GroupID(1)

	// Set the pending process group in the store
	k.SetPendingProcessGroups(ctx, types.PendingProcessGroups{
		GroupIDs: []tss.GroupID{groupID},
	})

	got := k.GetPendingProcessGroups(ctx)

	// Check if the retrieved pending process groups match the original sample
	s.Require().Len(got, 1)
	s.Require().Equal(groupID, got[0])
}

func (s *KeeperTestSuite) TestHandleProcessGroup() {
	ctx, k := s.ctx, s.keeper
	groupID, memberID := tss.GroupID(1), tss.MemberID(1)
	member := types.Member{
		ID:          memberID,
		GroupID:     groupID,
		IsMalicious: false,
	}

	k.SetMember(ctx, member)

	k.SetGroup(ctx, types.Group{
		ID:     groupID,
		Status: types.GROUP_STATUS_ROUND_1,
	})
	k.HandleProcessGroup(ctx, groupID)
	group := k.MustGetGroup(ctx, groupID)
	s.Require().Equal(types.GROUP_STATUS_ROUND_2, group.Status)

	k.SetGroup(ctx, types.Group{
		ID:     groupID,
		Status: types.GROUP_STATUS_ROUND_2,
	})
	k.HandleProcessGroup(ctx, groupID)
	group = k.MustGetGroup(ctx, groupID)
	s.Require().Equal(types.GROUP_STATUS_ROUND_3, group.Status)

	k.SetGroup(ctx, types.Group{
		ID:     groupID,
		Status: types.GROUP_STATUS_FALLEN,
	})
	k.HandleProcessGroup(ctx, groupID)
	group = k.MustGetGroup(ctx, groupID)
	s.Require().Equal(types.GROUP_STATUS_FALLEN, group.Status)

	k.SetGroup(ctx, types.Group{
		ID:     groupID,
		Status: types.GROUP_STATUS_ROUND_3,
	})
	k.HandleProcessGroup(ctx, groupID)
	group = k.MustGetGroup(ctx, groupID)
	s.Require().Equal(types.GROUP_STATUS_ACTIVE, group.Status)

	// if member is malicious
	k.SetGroup(ctx, types.Group{
		ID:     groupID,
		Status: types.GROUP_STATUS_ROUND_3,
	})
	member.IsMalicious = true
	k.SetMember(ctx, member)
	k.HandleProcessGroup(ctx, groupID)
	group = k.MustGetGroup(ctx, groupID)
	s.Require().Equal(types.GROUP_STATUS_FALLEN, group.Status)
}
