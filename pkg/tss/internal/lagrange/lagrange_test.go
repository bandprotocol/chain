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
		i    int64   // test description
		s    []int64 // hex encoded signature to parse
		exp  string
	}{{
		name: "1 from 1,2",
		i:    1,
		s:    []int64{1, 2},
		exp:  "2",
	}, {
		name: "2 from 1,2",
		i:    2,
		s:    []int64{1, 2},
		exp:  "115792089237316195423570985008687907852837564279074904382605163141518161494336",
	}, {
		name: "1 from 1,2,3",
		i:    1,
		s:    []int64{1, 2, 3},
		exp:  "3",
	}, {
		name: "2 from 1,2,4",
		i:    2,
		s:    []int64{1, 2, 4},
		exp:  "115792089237316195423570985008687907852837564279074904382605163141518161494335",
	}, {
		name: "3 from 1,3,4",
		i:    3,
		s:    []int64{1, 3, 4},
		exp:  "115792089237316195423570985008687907852837564279074904382605163141518161494335",
	}, {
		name: "4 from 2,3,4",
		i:    4,
		s:    []int64{2, 3, 4},
		exp:  "3",
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res := lagrange.ComputeCoefficient(test.i, test.s)
			exp, _ := new(big.Int).SetString(test.exp, 10)

			assert.Equal(t, exp, res)

			res = lagrange.ComputeCoefficientOptimize(test.i, test.s)
			assert.Equal(t, exp, res)
		})
	}
}
