package executor

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestParseExecutor(t *testing.T) {
	name, url, timeout, err := parseExecutor("test:www.bandprotocol.com?timeout=3s")
	require.Equal(t, name, "test")
	require.Equal(t, timeout, 3*time.Second)
	require.Equal(t, url, "www.bandprotocol.com")
	require.NoError(t, err)

	name, url, timeout, err = parseExecutor("test2:www.test.com/anna/kondanna?timeout=300ms")
	require.Equal(t, name, "test2")
	require.Equal(t, timeout, 300*time.Millisecond)
	require.Equal(t, url, "www.test.com/anna/kondanna")
	require.NoError(t, err)

	name, url, timeout, err = parseExecutor("test3:https://bandprotocol.com/gg/gg2/bandchain?timeout=1s300ms")
	require.Equal(t, name, "test3")
	require.Equal(t, timeout, 1*time.Second+300*time.Millisecond)
	require.Equal(t, url, "https://bandprotocol.com/gg/gg2/bandchain")
	require.NoError(t, err)
}

func TestParseExecutorWithoutRawQuery(t *testing.T) {
	_, _, _, err := parseExecutor("test:www.bandprotocol.com")
	require.EqualError(t, err, "invalid timeout, executor requires query timeout")
}

func TestParseExecutorInvalidExecutorError(t *testing.T) {
	_, _, _, err := parseExecutor("test")
	require.EqualError(t, err, "invalid executor, cannot parse executor: test")
}

func TestParseExecutorInvalidTimeoutError(t *testing.T) {
	_, _, _, err := parseExecutor("test:www.bandprotocol.com?timeout=test")
	require.EqualError(t, err, "invalid timeout, cannot parse duration with error: time: invalid duration \"test\"")
}
