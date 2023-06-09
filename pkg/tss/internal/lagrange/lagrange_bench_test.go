package lagrange_test

import (
	"fmt"
	"testing"

	"github.com/bandprotocol/chain/v2/pkg/tss/internal/lagrange"
)

func BenchmarkComputeCoefficient(b *testing.B) {
	tests := []struct {
		name string
		i    int64   // test description
		s    []int64 // hex encoded signature to parse
		exp  string
	}{{
		name: "1 from 1,2",
		i:    1,
		s:    []int64{1, 2},
	}, {
		name: "2 from 1,2",
		i:    2,
		s:    []int64{1, 2},
	}, {
		name: "1 from 1,2,3",
		i:    1,
		s:    []int64{1, 2, 3},
	}, {
		name: "2 from 1,2,4",
		i:    2,
		s:    []int64{1, 2, 4},
	}, {
		name: "3 from 1,3,4",
		i:    3,
		s:    []int64{1, 3, 4},
	}, {
		name: "4 from 2,3,4",
		i:    4,
		s:    []int64{2, 3, 4},
	}, {
		name: "18 in |S| = 10",
		i:    18,
		s:    []int64{3, 5, 8, 10, 13, 14, 16, 18, 19, 20},
	}, {
		name: "7 in |S| = 15",
		i:    7,
		s:    []int64{1, 4, 5, 7, 8, 9, 10, 11, 12, 13, 15, 16, 17, 18, 19},
	}, {
		name: "9 in |S| = 15",
		i:    9,
		s:    []int64{1, 4, 5, 7, 8, 9, 10, 11, 12, 13, 15, 16, 17, 18, 19},
	}, {
		name: "11 in |S| = 15",
		i:    11,
		s:    []int64{1, 4, 5, 7, 8, 9, 10, 11, 12, 13, 15, 16, 17, 18, 19},
	}, {
		name: "12 in |S| = 15",
		i:    12,
		s:    []int64{1, 4, 5, 7, 8, 9, 10, 11, 12, 13, 15, 16, 17, 18, 19},
	}, {
		name: "6 in |S| = 16",
		i:    6,
		s:    []int64{2, 4, 5, 6, 7, 8, 9, 10, 12, 13, 14, 15, 17, 18, 19, 20},
	}, {
		name: "13 in |S| = 20",
		i:    13,
		s:    []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
	}}

	for _, test := range tests {
		b.Run(fmt.Sprintf("Direct - %s", test.name), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				lagrange.ComputeCoefficient(test.i, test.s)
			}
		})

		b.Run(fmt.Sprintf("PreCompute - %s", test.name), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				lagrange.ComputeCoefficientPreCompute(test.i, test.s)
			}
		})
	}
}
