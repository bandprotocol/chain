package executor

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestParseExecutor(t *testing.T) {
	name, url, timeout, maxTry, startPort, endPort, err := parseExecutor("beeb:www.beebprotocol.com?timeout=3s")
	require.Equal(t, name, "beeb")
	require.Equal(t, timeout, 3*time.Second)
	require.Equal(t, url, "www.beebprotocol.com")
	require.Equal(t, maxTry, 1)
	require.Equal(t, startPort, 0)
	require.Equal(t, endPort, 0)
	require.NoError(t, err)

	name, url, timeout, _, _, _, err = parseExecutor(
		"beeb2:www.beeb.com/anna/kondanna?timeout=300ms",
	)
	require.Equal(t, name, "beeb2")
	require.Equal(t, timeout, 300*time.Millisecond)
	require.Equal(t, url, "www.beeb.com/anna/kondanna")
	require.NoError(t, err)

	name, url, timeout, _, _, _, err = parseExecutor(
		"beeb3:https://bandprotocol.com/gg/gg2/bandchain?timeout=1s300ms",
	)
	require.Equal(t, name, "beeb3")
	require.Equal(t, timeout, 1*time.Second+300*time.Millisecond)
	require.Equal(t, url, "https://bandprotocol.com/gg/gg2/bandchain")
	require.NoError(t, err)
}

func TestParseExecutorWithoutRawQuery(t *testing.T) {
	_, _, _, _, _, _, err := parseExecutor("beeb:www.beebprotocol.com")
	require.EqualError(t, err, "Invalid timeout, executor requires query timeout")
}

func TestParseExecutorInvalidExecutorError(t *testing.T) {
	_, _, _, _, _, _, err := parseExecutor("beeb")
	require.EqualError(t, err, "Invalid executor, cannot parse executor: beeb")
}

func TestParseExecutorInvalidTimeoutError(t *testing.T) {
	_, _, _, _, _, _, err := parseExecutor("beeb:www.beebprotocol.com?timeout=beeb")
	require.EqualError(t, err, "Invalid timeout, cannot parse duration with error: time: invalid duration \"beeb\"")
}

func TestExecuteDockerExecutorSuccess(t *testing.T) {
	e, err := NewExecutor(
		"docker:ongartbandprotocol/band-testing:python-runtime?timeout=120s&maxTry=10&portRange=5000-5009",
	)
	require.NoError(t, err)
	for i := 0; i < 20; i++ {
		res, err := e.Exec([]byte(
			"#!/usr/bin/env python3\nimport os\nimport sys\nprint(sys.argv[1], os.getenv('BAND_CHAIN_ID'))",
		), "TEST_ARG", map[string]interface{}{
			"BAND_CHAIN_ID":    "test-chain-id",
			"BAND_VALIDATOR":   "test-validator",
			"BAND_REQUEST_ID":  "test-request-id",
			"BAND_EXTERNAL_ID": "test-external-id",
			"BAND_REPORTER":    "test-reporter",
			"BAND_SIGNATURE":   "test-signature",
		})
		require.Equal(t, []byte("TEST_ARG test-chain-id\n"), res.Output)
		require.Equal(t, uint32(0), res.Code)
		require.NoError(t, err)
	}
}
