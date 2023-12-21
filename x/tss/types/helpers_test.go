package types_test

import (
	"testing"

	"github.com/bandprotocol/chain/v2/x/tss/types"
	"github.com/stretchr/testify/require"
)

func TestDuplicateInArray(t *testing.T) {
	testCases := []struct {
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

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := types.DuplicateInArray(tc.arr)
			require.Equal(t, tc.expect, result)
		})
	}
}
