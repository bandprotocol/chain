package types_test

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bandprotocol/chain/v2/x/tss/types"
)

func TestEncodeTunnelOriginator(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name       string
		originator types.TunnelOriginator
		expected   string
		err        error
	}{
		{
			name: "EVM chain",
			originator: types.TunnelOriginator{
				TunnelID:        256,
				ContractAddress: "0x5662ac531A2737C3dB8901E982B43327a2fDe2ae",
				ChainID:         "BSC",
			},
			expected: "a466d3130000000000000100000000000000002a3078353636326163353331413237333743336442383930314539383242343333323761326644653261650000000000000003425343",
		},
	}

	// Run the test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := tc.originator.Encode()
			require.Equal(t, tc.err, err)
			require.Equal(t, tc.expected, hex.EncodeToString(result))
		})
	}
}

func TestEncodeDirectOriginator(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name       string
		originator types.DirectOriginator
		expected   string
		err        error
	}{
		{
			name: "Normal",
			originator: types.DirectOriginator{
				Requester: "band10d07y265gmmuvt4z0w9aw880jnsr700jrdn8wm",
				Memo:      "test",
			},
			expected: "3c839c26000000000000002b62616e64313064303779323635676d6d757674347a30773961773838306a6e73723730306a72646e38776d000000000000000474657374",
		},
	}

	// Run the test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := tc.originator.Encode()
			require.Equal(t, tc.err, err)
			require.Equal(t, tc.expected, hex.EncodeToString(result))
		})
	}
}
