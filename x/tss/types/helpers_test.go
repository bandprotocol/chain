package types_test

import (
	"testing"

	"github.com/bandprotocol/chain/v2/x/tss/types"
)

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
