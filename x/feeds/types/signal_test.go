package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSumPower(t *testing.T) {
	require.Equal(t, int64(1250009), SumPower([]Signal{
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
	}))
}
