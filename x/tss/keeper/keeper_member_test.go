package keeper_test

import (
	"go.uber.org/mock/gomock"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/pkg/tss"
	tsstestutil "github.com/bandprotocol/chain/v3/x/tss/testutil"
	"github.com/bandprotocol/chain/v3/x/tss/types"
)

func (s *KeeperTestSuite) TestGetSetMember() {
	ctx, k := s.ctx, s.keeper
	groupID, memberID := tss.GroupID(1), tss.MemberID(1)
	member := types.Member{
		ID:          1,
		GroupID:     groupID,
		Address:     "band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs",
		PubKey:      nil,
		IsMalicious: false,
	}
	k.SetMember(ctx, member)

	got, err := k.GetMember(ctx, groupID, memberID)
	s.Require().NoError(err)
	s.Require().Equal(member, got)
}

func (s *KeeperTestSuite) TestGetGroupMembers() {
	ctx, k := s.ctx, s.keeper
	groupID := tss.GroupID(1)
	members := []types.Member{
		{
			ID:          1,
			GroupID:     groupID,
			Address:     "band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs",
			PubKey:      nil,
			IsMalicious: false,
		},
		{
			ID:          2,
			GroupID:     groupID,
			Address:     "band1p40yh3zkmhcv0ecqp3mcazy83sa57rgjp07dun",
			PubKey:      nil,
			IsMalicious: false,
		},
	}

	k.SetMembers(ctx, members)

	got, err := k.GetGroupMembers(ctx, groupID)
	s.Require().NoError(err)
	s.Require().Equal(members, got)
}

func (s *KeeperTestSuite) TestGetSetMemberIsActive() {
	ctx, k := s.ctx, s.keeper

	groupID := tss.GroupID(10)
	address := sdk.MustAccAddressFromBech32("band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs")
	k.SetMember(ctx, types.Member{
		ID:       tss.MemberID(1),
		GroupID:  groupID,
		Address:  address.String(),
		PubKey:   nil,
		IsActive: true,
	})

	// check when being set to active
	members, err := k.GetGroupMembers(ctx, groupID)
	s.Require().NoError(err)
	s.Require().Len(members, 1)

	for _, member := range members {
		s.Require().True(member.IsActive)
	}

	err = k.SetMemberIsActive(ctx, groupID, address, false)
	s.Require().NoError(err)

	members, err = k.GetGroupMembers(ctx, groupID)
	s.Require().NoError(err)
	s.Require().Len(members, 1)

	for _, member := range members {
		s.Require().False(member.IsActive)
	}
}

func (s *KeeperTestSuite) TestMarkMalicious() {
	ctx, k := s.ctx, s.keeper
	groupID := tss.GroupID(1)
	memberID := tss.MemberID(1)

	// Set member
	k.SetMember(ctx, types.Member{
		ID:          memberID,
		GroupID:     groupID,
		Address:     "member_address",
		PubKey:      []byte("pub_key"),
		IsMalicious: false,
	})

	// Mark malicious
	err := k.MarkMemberMalicious(ctx, groupID, memberID)
	s.Require().NoError(err)

	// Get members
	members, err := k.GetGroupMembers(ctx, groupID)
	s.Require().NoError(err)

	got := types.Members(members).HaveMalicious()
	s.Require().Equal(got, true)
}

func (s *KeeperTestSuite) TestGetRandomMembersSuccess() {
	ctx, k := s.ctx, s.keeper

	// create a group context and mock the result.
	groupCtx, err := tsstestutil.CompleteGroupCreation(ctx, k, 4, 2)
	s.Require().NoError(err)
	s.rollingseedKeeper.EXPECT().GetRollingSeed(gomock.Any()).
		Return([]byte("RandomStringThatShouldBeLongEnough"))

	// call GetRandomMembers and check results
	ams, err := k.GetRandomMembers(ctx, groupCtx.GroupID, []byte("test_nonce"))
	s.Require().NoError(err)

	memberIDs := []tss.MemberID{}
	for _, am := range ams {
		memberIDs = append(memberIDs, am.ID)
	}
	s.Require().Equal(memberIDs, []tss.MemberID{3, 4})
}

func (s *KeeperTestSuite) TestGetRandomMembersInsufficientActiveMembers() {
	ctx, k := s.ctx, s.keeper

	// create a group context.
	groupCtx, err := tsstestutil.CompleteGroupCreation(ctx, k, 4, 2)
	s.Require().NoError(err)

	// set every member to inactive.
	members, err := k.GetGroupMembers(ctx, groupCtx.GroupID)
	s.Require().NoError(err)
	for _, member := range members[:len(members)-1] {
		member.IsActive = false
		k.SetMember(ctx, member)
	}

	_, err = k.GetRandomMembers(ctx, tss.GroupID(1), []byte("test_nonce"))
	s.Require().ErrorIs(err, types.ErrInsufficientActiveMembers)
}

func (s *KeeperTestSuite) TestGetRandomMembersNoActiveMember() {
	ctx, k := s.ctx, s.keeper

	// create a group context.
	groupCtx, err := tsstestutil.CompleteGroupCreation(ctx, k, 4, 2)
	s.Require().NoError(err)

	// set every member to inactive.
	members, err := k.GetGroupMembers(ctx, groupCtx.GroupID)
	s.Require().NoError(err)
	for _, member := range members {
		member.IsActive = false
		k.SetMember(ctx, member)
	}

	_, err = k.GetRandomMembers(ctx, tss.GroupID(1), []byte("test_nonce"))
	s.Require().ErrorIs(err, types.ErrNoActiveMember)
}
