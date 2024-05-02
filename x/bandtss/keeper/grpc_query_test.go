package keeper_test

import (
	"fmt"
	"testing"
	"time"

	querytypes "github.com/cosmos/cosmos-sdk/types/query"
	"github.com/stretchr/testify/require"

	"github.com/bandprotocol/chain/v2/x/bandtss/testutil"
	"github.com/bandprotocol/chain/v2/x/bandtss/types"
)

func TestGRPCQueryMembers(t *testing.T) {
	type expectOut struct {
		members []*types.Member
	}

	since := time.Now().UTC()
	lastActive := time.Now().UTC()

	members := []*types.Member{
		{
			Address:    "band1t5x8hrmht463eq4m0xhfgz95h62dyvkq049eek",
			IsActive:   true,
			Since:      since,
			LastActive: lastActive,
		},
		{
			Address:    "band1a22hgwm4tz8gj82y6zad3de2dcg5dpymtj20m5",
			IsActive:   true,
			Since:      since,
			LastActive: lastActive,
		},
	}

	testCases := []struct {
		name       string
		preProcess func(s *testutil.TestSuite)
		input      *types.QueryMembersRequest
		expectOut  expectOut
	}{
		{
			name: "get 2 active members",
			input: &types.QueryMembersRequest{
				IsActive: true,
			},
			expectOut: expectOut{members: members},
		},
		{
			name: "get 1 active members; limit 1 offset 0",
			input: &types.QueryMembersRequest{
				IsActive:   true,
				Pagination: &querytypes.PageRequest{Limit: 1, Offset: 0},
			},
			expectOut: expectOut{members: members[:1]},
		},
		{
			name: "get 1 active members limit 1 offset 1",
			input: &types.QueryMembersRequest{
				IsActive:   true,
				Pagination: &querytypes.PageRequest{Limit: 1, Offset: 1},
			},
			expectOut: expectOut{members: members[1:]},
		},
		{
			name: "get 0 active members; out of pages limit 1 offset 5",
			input: &types.QueryMembersRequest{
				IsActive:   true,
				Pagination: &querytypes.PageRequest{Limit: 1, Offset: 5},
			},
			expectOut: expectOut{members: []*types.Member{}},
		},
		{
			name: "get no inactive members",
			input: &types.QueryMembersRequest{
				IsActive: false,
			},
			expectOut: expectOut{members: []*types.Member{}},
		},
		{
			name: "get inactive members",
			preProcess: func(s *testutil.TestSuite) {
				s.Keeper.SetMember(s.Ctx, types.Member{
					Address:    members[0].Address,
					IsActive:   false,
					Since:      members[0].Since,
					LastActive: members[0].LastActive,
				})
			},
			input: &types.QueryMembersRequest{
				IsActive: false,
			},
			expectOut: expectOut{members: []*types.Member{
				{Address: members[0].Address, IsActive: false, Since: members[0].Since, LastActive: members[0].LastActive},
			}},
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Case %s", tc.name), func(t *testing.T) {
			s := testutil.NewTestSuite(t)
			q := s.QueryServer

			for _, member := range members {
				s.Keeper.SetMember(s.Ctx, *member)
			}

			if tc.preProcess != nil {
				tc.preProcess(&s)
			}

			res, err := q.Members(s.Ctx, tc.input)
			require.NoError(t, err)
			require.Equal(t, tc.expectOut.members, res.Members)
		})
	}
}
