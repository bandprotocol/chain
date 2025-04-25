package client_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/bandprotocol/chain/v3/cylinder/client"
	bandtsstypes "github.com/bandprotocol/chain/v3/x/bandtss/types"
)

func TestNewMemberResponse(t *testing.T) {
	// Use a fixed time for all test cases
	fixedTime := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name          string
		queryResponse *bandtsstypes.QueryMemberResponse
		expectedAddrs []string
	}{
		{
			name: "both members present",
			queryResponse: &bandtsstypes.QueryMemberResponse{
				CurrentGroupMember: bandtsstypes.Member{
					Address:  "band address 1",
					GroupID:  1,
					IsActive: true,
					Since:    fixedTime,
				},
				IncomingGroupMember: bandtsstypes.Member{
					Address:  "band address 2",
					GroupID:  2,
					IsActive: true,
					Since:    fixedTime,
				},
			},
			expectedAddrs: []string{"band address 1", "band address 2"},
		},
		{
			name: "only current member present",
			queryResponse: &bandtsstypes.QueryMemberResponse{
				CurrentGroupMember: bandtsstypes.Member{
					Address:  "band address 1",
					GroupID:  1,
					IsActive: true,
					Since:    fixedTime,
				},
				IncomingGroupMember: bandtsstypes.Member{
					Address:  "",
					GroupID:  0,
					IsActive: false,
					Since:    time.Time{},
				},
			},
			expectedAddrs: []string{"band address 1"},
		},
		{
			name: "only incoming member present",
			queryResponse: &bandtsstypes.QueryMemberResponse{
				CurrentGroupMember: bandtsstypes.Member{
					Address:  "",
					GroupID:  0,
					IsActive: false,
					Since:    time.Time{},
				},
				IncomingGroupMember: bandtsstypes.Member{
					Address:  "band address 2",
					GroupID:  2,
					IsActive: true,
					Since:    fixedTime,
				},
			},
			expectedAddrs: []string{"band address 2"},
		},
		{
			name: "no members present",
			queryResponse: &bandtsstypes.QueryMemberResponse{
				CurrentGroupMember: bandtsstypes.Member{
					Address:  "",
					GroupID:  0,
					IsActive: false,
					Since:    time.Time{},
				},
				IncomingGroupMember: bandtsstypes.Member{
					Address:  "",
					GroupID:  0,
					IsActive: false,
					Since:    time.Time{},
				},
			},
			expectedAddrs: []string{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := client.NewMemberResponse(test.queryResponse)

			// Check the number of members
			expectedCount := len(test.expectedAddrs)
			resultCount := len(result.Members)
			assert.Equal(
				t,
				expectedCount,
				resultCount,
				"expected %d members, got %d",
				expectedCount,
				resultCount,
			)

			// Check member addresses if any are expected
			if len(test.expectedAddrs) > 0 {
				for i, expectedAddr := range test.expectedAddrs {
					assert.Equal(t, expectedAddr, result.Members[i].Address, "member address mismatch at index %d", i)
				}
			}
		})
	}
}
