package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

func TestVerify(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name          string
		member        types.Member
		address       string
		expectedMatch bool
	}{
		{
			name: "MatchingAddress",
			member: types.Member{
				Address: "band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs",
			},
			address:       "band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs",
			expectedMatch: true,
		},
		{
			name: "NonMatchingAddress",
			member: types.Member{
				Address: "band1p40yh3zkmhcv0ecqp3mcazy83sa57rgjp07dun",
			},
			address:       "band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs",
			expectedMatch: false,
		},
	}

	// Run the test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.member.Verify(tc.address)
			require.Equal(t, tc.expectedMatch, result)
		})
	}
}

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
				{ID: 1},
				{ID: 2},
				{ID: 3},
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
				{IsMalicious: true},
				{IsMalicious: false},
			},
			expectedValue: true,
		},
		{
			name: "NoMalicious",
			members: types.Members{
				{IsMalicious: false},
				{IsMalicious: false},
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
