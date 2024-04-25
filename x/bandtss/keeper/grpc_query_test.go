package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	querytypes "github.com/cosmos/cosmos-sdk/types/query"

	"github.com/bandprotocol/chain/v2/x/bandtss/types"
	tsstypes "github.com/bandprotocol/chain/v2/x/tss/types"
)

func (s *KeeperTestSuite) TestGRPCQueryMembers() {
	ctx, q := s.ctx, s.queryClient

	s.SetupGroup(tsstypes.GROUP_STATUS_ACTIVE)

	result, err := s.queryClient.Members(ctx, &types.QueryMembersRequest{IsActive: true})
	s.Require().NoError(err)
	s.Require().Len(result.Members, 2)
	members := result.Members

	type expectOut struct {
		members []*types.Member
	}

	testCases := []struct {
		name        string
		preProcess  func()
		input       types.QueryMembersRequest
		expectOut   expectOut
		postProcess func()
	}{
		{
			name:       "get 2 active members",
			preProcess: func() {},
			input: types.QueryMembersRequest{
				IsActive: true,
			},
			expectOut:   expectOut{members: members},
			postProcess: func() {},
		},
		{
			name:       "get 1 active members; limit 1 offset 0",
			preProcess: func() {},
			input: types.QueryMembersRequest{
				IsActive:   true,
				Pagination: &querytypes.PageRequest{Limit: 1, Offset: 0},
			},
			expectOut:   expectOut{members: members[:1]},
			postProcess: func() {},
		},
		{
			name:       "get 1 active members limit 1 offset 1",
			preProcess: func() {},
			input: types.QueryMembersRequest{
				IsActive:   true,
				Pagination: &querytypes.PageRequest{Limit: 1, Offset: 1},
			},
			expectOut:   expectOut{members: members[1:]},
			postProcess: func() {},
		},
		{
			name:       "get 0 active members; out of pages limit 1 offset 5",
			preProcess: func() {},
			input: types.QueryMembersRequest{
				IsActive:   true,
				Pagination: &querytypes.PageRequest{Limit: 1, Offset: 5},
			},
			expectOut:   expectOut{members: nil},
			postProcess: func() {},
		},
		{
			name:       "get no active members",
			preProcess: func() {},
			input: types.QueryMembersRequest{
				IsActive: false,
			},
			expectOut:   expectOut{members: nil},
			postProcess: func() {},
		},
		{
			name: "get inactive members",
			preProcess: func() {
				err := s.app.BandtssKeeper.DeactivateMember(ctx, sdk.MustAccAddressFromBech32(members[0].Address))
				s.Require().NoError(err)
			},
			input: types.QueryMembersRequest{
				IsActive: false,
			},
			expectOut: expectOut{members: []*types.Member{
				{Address: members[0].Address, IsActive: false, Since: members[0].Since, LastActive: members[0].LastActive},
			}},
			postProcess: func() {
				ctx = ctx.WithBlockTime(ctx.BlockTime().Add(types.DefaultInactivePenaltyDuration))
				err := s.app.BandtssKeeper.ActivateMember(ctx, sdk.MustAccAddressFromBech32(members[0].Address))
				s.Require().NoError(err)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.name), func() {
			tc.preProcess()

			res, err := q.Members(ctx, &tc.input)
			s.Require().NoError(err)
			s.Require().Equal(tc.expectOut.members, res.Members)

			tc.postProcess()
		})
	}
}
