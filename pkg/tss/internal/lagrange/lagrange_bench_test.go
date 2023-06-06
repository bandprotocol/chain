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
	}}

	for _, test := range tests {
		b.Run(fmt.Sprintf("Compute coefficient - %s", test.name), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				lagrange.ComputeCoefficient(test.i, test.s)
			}
		})

		b.Run(fmt.Sprintf("Compute coefficient optimized - %s", test.name), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				lagrange.ComputeCoefficientOptimize(test.i, test.s)
			}
		})
	}
}
