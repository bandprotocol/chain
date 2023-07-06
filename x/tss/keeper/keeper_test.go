package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	"github.com/stretchr/testify/suite"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/pkg/tss/testutil"
	"github.com/bandprotocol/chain/v2/testing/testapp"
	"github.com/bandprotocol/chain/v2/x/tss/keeper"
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

type KeeperTestSuite struct {
	suite.Suite

	app     *testapp.TestingApp
	ctx     sdk.Context
	querier keeper.Querier
	msgSrvr types.MsgServer
}

func (s *KeeperTestSuite) SetupTest() {
	app, ctx, _ := testapp.CreateTestInput(false)
	s.app = app
	s.ctx = ctx
	s.querier = keeper.Querier{
		app.TSSKeeper,
	}
	s.msgSrvr = app.TSSKeeper
}

func (s *KeeperTestSuite) SetupGroup(groupStatus types.GroupStatus) {
	ctx, msgSrvr, k := s.ctx, s.msgSrvr, s.app.TSSKeeper

	// Create group from testutil
	for _, tc := range testutil.TestCases {
		// Init members
		var members []string
		for _, m := range tc.Group.Members {
			members = append(members, sdk.AccAddress(m.PubKey()).String())
		}

		// Create group
		_, err := msgSrvr.CreateGroup(ctx, &types.MsgCreateGroup{
			Members:   members,
			Threshold: tc.Group.Threshold,
			Sender:    sdk.AccAddress(tc.Group.Members[0].PubKey()).String(),
		})
		s.Require().NoError(err)

		// Set dkg context
		s.app.TSSKeeper.SetDKGContext(ctx, tc.Group.ID, tc.Group.DKGContext)

		// Get group
		group, err := k.GetGroup(ctx, tc.Group.ID)
		s.Require().NoError(err)

		switch groupStatus {
		case types.GROUP_STATUS_ROUND_2:
			group.Status = types.GROUP_STATUS_ROUND_2
			k.SetGroup(ctx, group)
		case types.GROUP_STATUS_ROUND_3:
			group.Status = types.GROUP_STATUS_ROUND_3
			k.SetGroup(ctx, group)
			// Update member public key
			for i, m := range tc.Group.Members {
				member := types.Member{
					MemberID:    tss.MemberID(i + 1),
					Address:     sdk.AccAddress(m.PubKey()).String(),
					PubKey:      m.PubKey(),
					IsMalicious: false,
				}
				k.SetMember(ctx, tc.Group.ID, member)
			}
		case types.GROUP_STATUS_ACTIVE:
			group.Status = types.GROUP_STATUS_ACTIVE
			group.PubKey = tc.Group.PubKey
			k.SetGroup(ctx, group)

			// Update member public key
			for i, m := range tc.Group.Members {
				member := types.Member{
					MemberID:    tss.MemberID(i + 1),
					Address:     sdk.AccAddress(m.PubKey()).String(),
					PubKey:      m.PubKey(),
					IsMalicious: false,
				}
				k.SetMember(ctx, tc.Group.ID, member)
			}

			for _, signing := range tc.Signings {
				for _, am := range signing.AssignedMembers {
					pubD := am.PrivD.Point()
					pubE := am.PrivE.Point()

					member, err := k.GetMember(ctx, tc.Group.ID, am.ID)
					s.Require().NoError(err)
					address, err := sdk.AccAddressFromBech32(member.Address)
					s.Require().NoError(err)

					k.HandleSetDEs(ctx, address, []types.DE{
						{
							PubD: pubD,
							PubE: pubE,
						},
					})
				}
			}
		}
	}
}

func (s *KeeperTestSuite) TestGetSetGroupCount() {
	ctx, k := s.ctx, s.app.TSSKeeper
	k.SetGroupCount(ctx, 1)

	groupCount := k.GetGroupCount(ctx)
	s.Require().Equal(uint64(1), groupCount)
}

func (s *KeeperTestSuite) TestGetNextGroupID() {
	ctx, k := s.ctx, s.app.TSSKeeper

	// Initial group count
	k.SetGroupCount(ctx, 0)

	groupID1 := k.GetNextGroupID(ctx)
	s.Require().Equal(tss.GroupID(1), groupID1)
	groupID2 := k.GetNextGroupID(ctx)
	s.Require().Equal(tss.GroupID(2), groupID2)
}

func (s *KeeperTestSuite) TestIsGrantee() {
	ctx, k := s.ctx, s.app.TSSKeeper
	expTime := time.Unix(0, 0)

	// Init grantee address
	grantee, _ := sdk.AccAddressFromBech32("band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs")

	// Init granter address
	granter, _ := sdk.AccAddressFromBech32("band1p40yh3zkmhcv0ecqp3mcazy83sa57rgjp07dun")

	// Save grant msgs to grantee
	for _, m := range types.GetMsgGrants() {
		s.app.AuthzKeeper.SaveGrant(ctx, grantee, granter, authz.NewGenericAuthorization(m), &expTime)
	}

	isGrantee := k.IsGrantee(ctx, granter, grantee)
	s.Require().True(isGrantee)
}

func (s *KeeperTestSuite) TestCreateNewGroup() {
	ctx, k := s.ctx, s.app.TSSKeeper
	expiration := ctx.BlockHeader().Time.Add(k.RoundPeriod(ctx))

	group := types.Group{
		Size_:      5,
		Threshold:  3,
		PubKey:     nil,
		Status:     types.GROUP_STATUS_ROUND_1,
		Expiration: &expiration,
	}

	// Create new group
	groupID := k.CreateNewGroup(ctx, group)

	// init group ID
	group.GroupID = tss.GroupID(1)

	// Get group by id
	got, err := k.GetGroup(ctx, groupID)
	s.Require().NoError(err)
	s.Require().Equal(group, got)
}

func (s *KeeperTestSuite) TestSetGroup() {
	ctx, k := s.ctx, s.app.TSSKeeper
	group := types.Group{
		Size_:     5,
		Threshold: 3,
		PubKey:    nil,
		Status:    types.GROUP_STATUS_ROUND_1,
	}

	// Set new group
	groupID := k.CreateNewGroup(ctx, group)

	// Update group size value
	group.Size_ = 6

	// Add group ID
	group.GroupID = 1

	k.SetGroup(ctx, group)

	// Get group from chain state
	got, err := k.GetGroup(ctx, groupID)

	// Validate group size value
	s.Require().NoError(err)
	s.Require().Equal(group.Size_, got.Size_)
}

func (s *KeeperTestSuite) TestGetGroups() {
	ctx, k := s.ctx, s.app.TSSKeeper
	group := types.Group{
		GroupID:   1,
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
	ctx, k := s.ctx, s.app.TSSKeeper

	dkgContext := []byte("dkg-context sample")
	k.SetDKGContext(ctx, 1, dkgContext)

	got, err := k.GetDKGContext(ctx, 1)
	s.Require().NoError(err)
	s.Require().Equal(dkgContext, got)
}

func (s *KeeperTestSuite) TestGetSetMember() {
	ctx, k := s.ctx, s.app.TSSKeeper
	groupID, memberID := tss.GroupID(1), tss.MemberID(1)
	member := types.Member{
		MemberID:    1,
		Address:     "band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs",
		PubKey:      nil,
		IsMalicious: false,
	}
	k.SetMember(ctx, groupID, member)

	got, err := k.GetMember(ctx, groupID, memberID)
	s.Require().NoError(err)
	s.Require().Equal(member, got)
}

func (s *KeeperTestSuite) TestGetMembers() {
	ctx, k := s.ctx, s.app.TSSKeeper
	groupID := tss.GroupID(1)
	members := []types.Member{
		{
			MemberID:    1,
			Address:     "band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs",
			PubKey:      nil,
			IsMalicious: false,
		},
		{
			MemberID:    2,
			Address:     "band1p40yh3zkmhcv0ecqp3mcazy83sa57rgjp07dun",
			PubKey:      nil,
			IsMalicious: false,
		},
	}

	// Set members
	for _, m := range members {
		k.SetMember(ctx, groupID, m)
	}

	got, err := k.GetMembers(ctx, groupID)
	s.Require().NoError(err)
	s.Require().Equal(members, got)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
