package lagrange_test

import (
	"math/big"
	"testing"

	"github.com/bandprotocol/chain/v2/pkg/tss/internal/lagrange"
	"github.com/stretchr/testify/assert"
)

func TestComputeCoefficient(t *testing.T) {
	tests := []struct {
		name string
		i    int64    // test description
		s    []int64  // hex encoded signature to parse
		exp  *big.Int // expected error
	}{{
		name: "1 from 1,2",
		i:    1,
		s:    []int64{1, 2},
		exp:  big.NewInt(2),
	}, {
		name: "2 from 1,2",
		i:    2,
		s:    []int64{1, 2},
		exp: big.NewInt(0).
			SetBits([]big.Word{13822214165235122496, 13451932020343611451, 18446744073709551614, 18446744073709551615}),
	}, {
		name: "1 from 1,2,3",
		i:    1,
		s:    []int64{1, 2, 3},
		exp:  big.NewInt(3),
	}, {
		name: "2 from 1,2,4",
		i:    2,
		s:    []int64{1, 2, 4},
		exp: big.NewInt(0).
			SetBits([]big.Word{13822214165235122495, 13451932020343611451, 18446744073709551614, 18446744073709551615}),
	}, {
		name: "3 from 1,3,4",
		i:    3,
		s:    []int64{1, 3, 4},
		exp: big.NewInt(0).
			SetBits([]big.Word{13822214165235122495, 13451932020343611451, 18446744073709551614, 18446744073709551615}),
	}, {
		name: "4 from 2,3,4",
		i:    4,
		s:    []int64{2, 3, 4},
		exp:  big.NewInt(3),
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res := lagrange.ComputeCoefficient(test.i, test.s)
			assert.Equal(t, test.exp, res)

			res = lagrange.ComputeCoefficientOptimize(test.i, test.s)
			assert.Equal(t, test.exp, res)
		})
	}
}
