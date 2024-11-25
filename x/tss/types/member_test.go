package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bandprotocol/chain/v3/pkg/tss"
	"github.com/bandprotocol/chain/v3/x/tss/types"
)

func TestGetIDs(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name        string
		members     types.Members
		expectedIDs []tss.MemberID
	}{
		{
			name: "MultipleMembers",
			members: types.Members{
				types.Member{ID: 1},
				types.Member{ID: 2},
				types.Member{ID: 3},
			},
			expectedIDs: []tss.MemberID{1, 2, 3},
		},
		{
			name:        "EmptyMembers",
			members:     types.Members{},
			expectedIDs: []tss.MemberID(nil),
		},
	}

	// Run the test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.members.GetIDs()
			require.Equal(t, tc.expectedIDs, result)
		})
	}
}

func TestHaveMalicious(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name          string
		members       types.Members
		expectedValue bool
	}{
		{
			name: "ContainsMalicious",
			members: types.Members{
				types.Member{ID: 1, IsMalicious: true},
				types.Member{ID: 2, IsMalicious: false},
			},
			expectedValue: true,
		},
		{
			name: "NoMalicious",
			members: types.Members{
				types.Member{ID: 1, IsMalicious: false},
				types.Member{ID: 2, IsMalicious: false},
			},
			expectedValue: false,
		},
	}

	// Run the test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.members.HaveMalicious()
			require.Equal(t, tc.expectedValue, result)
		})
	}
}

func TestFindMemberSlot(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name         string
		from         tss.MemberID
		to           tss.MemberID
		expectedSlot tss.MemberID
	}{
		{
			name:         "FromLessThanTo",
			from:         2,
			to:           5,
			expectedSlot: 3,
		},
		{
			name:         "FromGreaterThanTo",
			from:         7,
			to:           4,
			expectedSlot: 3,
		},
		{
			name:         "FromEqualsTo",
			from:         5,
			to:           5,
			expectedSlot: 4,
		},
	}

	// Run the test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := types.FindMemberSlot(tc.from, tc.to)
			require.Equal(t, tc.expectedSlot, result)
		})
	}
}
