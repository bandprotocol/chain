package executor

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func SetupDockerTest(t *testing.T) {
}

func TestDockerSuccess(t *testing.T) {
	// TODO: Enable test when CI has docker installed.
	// Prerequisite: please build docker image before running test
	e := NewDockerExec("ongartbandprotocol/band-testing", 120*time.Second, 10, 5000, 5009)
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
