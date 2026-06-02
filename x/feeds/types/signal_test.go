package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSumPower(t *testing.T) {
	sum, err := SumPower([]Signal{
		{
			ID:    "CS:BAND-USD",
			Power: 100000,
		},
		{
			ID:    "CS:ATOM-USD",
			Power: 150000,
		},
		{
			ID:    "CS:OSMO-USD",
			Power: 1000009,
		},
	})
	require.NoError(t, err)
	require.Equal(t, int64(1250009), sum)
}

// TestSumPowerOverflow verifies that SumPower rejects signal lists whose
// aggregate power would overflow int64.  The specific values below sum to
// exactly 100*2^64 which wraps to 0 in unchecked int64 arithmetic.
func TestSumPowerOverflow(t *testing.T) {
	signals := make([]Signal, 300)
	for i := 0; i < 200; i++ {
		signals[i] = Signal{ID: "ZZ:OVERFLOW", Power: 6148914691236517205}
	}
	for i := 0; i < 100; i++ {
		signals[200+i] = Signal{ID: "ZZ:OVERFLOW", Power: 6148914691236517206}
	}
	_, err := SumPower(signals)
	require.ErrorContains(t, err, "overflows int64")
}
