package types_test

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bandprotocol/chain/v3/pkg/tss"
	"github.com/bandprotocol/chain/v3/x/tss/types"
)

func TestOriginatorPrefix(t *testing.T) {
	require.Equal(t, []byte(types.DirectOriginatorPrefix), tss.Hash([]byte("DirectOriginator"))[:4])
	require.Equal(t, []byte(types.TunnelOriginatorPrefix), tss.Hash([]byte("TunnelOriginator"))[:4])
}

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
			originator: types.NewTunnelOriginator(
				"bandchain",
				256,
				"0x5662ac531A2737C3dB8901E982B43327a2fDe2ae",
				"BSC",
			),
			expected: "72ebe83d" +
				"0e1ac2c4a50a82aa49717691fc1ae2e5fa68eff45bd8576b0f2be7a0850fa7c6" +
				"0000000000000100" +
				"1b791f9b381ec74bd523be18b5d02eacfc1811c3817f87e7981664ccabe31e00" +
				"4602a37e2aeaf2820d53eaeb5ab645d0d45172d006889d176509ed9e7cfa6144",
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
			originator: types.NewDirectOriginator(
				"bandchain",
				"band10d07y265gmmuvt4z0w9aw880jnsr700jrdn8wm",
				"test",
			),
			expected: "b39fa5d2" +
				"0e1ac2c4a50a82aa49717691fc1ae2e5fa68eff45bd8576b0f2be7a0850fa7c6" +
				"ae646b8a6bc479924298c56a0f6beed904198e60195eb75460970c0bf879010e" +
				"9c22ff5f21f0b81b113e63f7db6da94fedef11b2119b4088b89664fb9a3cb658",
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
