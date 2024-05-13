package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

func (s *KeeperTestSuite) TestGetSetMember() {
	ctx, k := s.ctx, s.app.TSSKeeper
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

func (s *KeeperTestSuite) TestGetMembers() {
	ctx, k := s.ctx, s.app.TSSKeeper
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
	ctx, k := s.ctx, s.app.TSSKeeper

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
