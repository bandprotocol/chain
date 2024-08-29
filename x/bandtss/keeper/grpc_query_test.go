package keeper_test

import (
	"fmt"
	"testing"
	"time"

	querytypes "github.com/cosmos/cosmos-sdk/types/query"
	"github.com/stretchr/testify/require"

	"github.com/bandprotocol/chain/v2/pkg/tss"
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
			Address:    "band1p40yh3zkmhcv0ecqp3mcazy83sa57rgjp07dun",
			GroupID:    tss.GroupID(1),
			IsActive:   false,
			Since:      since,
			LastActive: lastActive,
		},
		{
			Address:    "band1t5x8hrmht463eq4m0xhfgz95h62dyvkq049eek",
			GroupID:    tss.GroupID(1),
			IsActive:   true,
			Since:      since,
			LastActive: lastActive,
		},
		{
			Address:    "band1a22hgwm4tz8gj82y6zad3de2dcg5dpymtj20m5",
			GroupID:    tss.GroupID(1),
			IsActive:   true,
			Since:      since,
			LastActive: lastActive,
		},
	}

	testCases := []struct {
		name       string
		preProcess func(s *KeeperTestSuite)
		input      *types.QueryMembersRequest
		expectOut  expectOut
	}{
		{
			name: "get all members",
			input: &types.QueryMembersRequest{
				Status: types.MEMBER_STATUS_FILTER_UNSPECIFIED,
			},
			expectOut: expectOut{members: members},
		},
		{
			name: "get 1 active members; limit 1 offset 0",
			input: &types.QueryMembersRequest{
				Status:     types.MEMBER_STATUS_FILTER_ACTIVE,
				Pagination: &querytypes.PageRequest{Limit: 1, Offset: 0},
			},
			expectOut: expectOut{members: members[1:2]},
		},
		{
			name: "get 1 active members limit 1 offset 1",
			input: &types.QueryMembersRequest{
				Status:     types.MEMBER_STATUS_FILTER_ACTIVE,
				Pagination: &querytypes.PageRequest{Limit: 1, Offset: 1},
			},
			expectOut: expectOut{members: members[2:]},
		},
		{
			name: "get 0 active members; out of pages limit 1 offset 5",
			input: &types.QueryMembersRequest{
				Status:     types.MEMBER_STATUS_FILTER_ACTIVE,
				Pagination: &querytypes.PageRequest{Limit: 1, Offset: 5},
			},
			expectOut: expectOut{members: []*types.Member{}},
		},
		{
			name: "get inactive members",
			input: &types.QueryMembersRequest{
				Status: types.MEMBER_STATUS_FILTER_INACTIVE,
			},
			expectOut: expectOut{members: members[0:1]},
		},
		{
			name: "get incoming members; error",
			input: &types.QueryMembersRequest{
				Status:          types.MEMBER_STATUS_FILTER_INACTIVE,
				IsIncomingGroup: true,
			},
			expectOut: expectOut{members: []*types.Member{}},
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Case %s", tc.name), func(t *testing.T) {
			s := NewKeeperTestSuite(t)
			q := s.QueryServer
			s.Keeper.SetCurrentGroupID(s.Ctx, 1)

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
