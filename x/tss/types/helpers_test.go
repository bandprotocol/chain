package types_test

import (
	"testing"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

func TestVerifyMember(t *testing.T) {
	// Create a test member with a known address
	member := types.Member{
		Address:     "band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs",
		PubKey:      tss.PublicKey(nil),
		IsMalicious: false,
	}

	// Define test cases
	testCases := []struct {
		name        string
		member      types.Member
		address     string
		expectedRes bool
	}{
		{
			name:        "MatchingAddress",
			member:      member,
			address:     "band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs",
			expectedRes: true,
		},
		{
			name:        "NonMatchingAddress",
			member:      member,
			address:     "band1p40yh3zkmhcv0ecqp3mcazy83sa57rgjp07dun",
			expectedRes: false,
		},
	}

	// Run the test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := types.VerifyMember(tc.member, tc.address)
			if result != tc.expectedRes {
				t.Errorf("Expected %v for address %s, but got %v", tc.expectedRes, tc.address, result)
			}
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
			if result != tc.expectedSlot {
				t.Errorf("Expected slot %d for from %d and to %d, but got %d", tc.expectedSlot, tc.from, tc.to, result)
			}
		})
	}
}

func TestHaveMalicious(t *testing.T) {
	tests := []struct {
		name    string
		members []types.Member
		expect  bool
	}{
		{
			name: "No malicious members",
			members: []types.Member{
				{
					Address:     "band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs",
					PubKey:      tss.PublicKey(nil),
					IsMalicious: false,
				},
				{
					Address:     "band1p40yh3zkmhcv0ecqp3mcazy83sa57rgjp07dun",
					PubKey:      tss.PublicKey(nil),
					IsMalicious: false,
				},
			},
			expect: false,
		},
		{
			name: "Malicious members present",
			members: []types.Member{
				{
					Address:     "band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs",
					PubKey:      tss.PublicKey(nil),
					IsMalicious: false,
				},
				{
					Address:     "band1p40yh3zkmhcv0ecqp3mcazy83sa57rgjp07dun",
					PubKey:      tss.PublicKey(nil),
					IsMalicious: true,
				},
			},
			expect: true,
		},
		{
			name:    "Empty member list",
			members: []types.Member{},
			expect:  false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := types.HaveMalicious(test.members)
			if result != test.expect {
				t.Errorf("Expected %v, but got %v", test.expect, result)
			}
		})
	}
}

func TestDuplicateInArray(t *testing.T) {
	tests := []struct {
		name   string
		arr    []string
		expect bool
	}{
		{
			name:   "No duplicates",
			arr:    []string{"a", "b", "c", "d"},
			expect: false,
		},
		{
			name:   "Duplicates present",
			arr:    []string{"a", "b", "c", "a"},
			expect: true,
		},
		{
			name:   "Empty array",
			arr:    []string{},
			expect: false,
		},
		{
			name:   "Single element",
			arr:    []string{"a"},
			expect: false,
		},
		{
			name:   "Multiple duplicates",
			arr:    []string{"a", "b", "a", "b", "c"},
			expect: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := types.DuplicateInArray(test.arr)
			if result != test.expect {
				t.Errorf("Expected %v, but got %v", test.expect, result)
			}
		})
	}
}

func TestUint64ArrayContains(t *testing.T) {
	arr := []uint64{1, 2, 3, 4, 5}
	existing := uint64(3)
	nonExisting := uint64(6)

	if !types.Uint64ArrayContains(arr, existing) {
		t.Errorf("Expected arr to contain %d, but it did not.", existing)
	}

	if types.Uint64ArrayContains(arr, nonExisting) {
		t.Errorf("Expected arr to not contain %d, but it did.", nonExisting)
	}
}
