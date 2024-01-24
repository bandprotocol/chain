package executor

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/bandprotocol/chain/v2/x/feeds/types"
)

const (
	flagQueryTimeout = "timeout"
)

var (
	ErrExecutionimeout = errors.New("execution timeout")
	ErrRestNotOk       = errors.New("rest return non 2XX response")
)

type ExecResult struct {
	Output  []byte
	Code    uint32
	Version string
}

type Executor interface {
	Exec(params map[string]string) ([]types.SubmitPrice, error)
}

// NewExecutor returns executor by name and executor URL
func NewExecutor(executor string) (exec Executor, err error) {
	name, base, timeout, err := parseExecutor(executor)
	if err != nil {
		return nil, err
	}
	switch name {
	case "rest":
		exec = NewRestExec(base, timeout)
	default:
		return nil, fmt.Errorf("invalid executor name: %s, base: %s", name, base)
	}

	// TODO: Remove hardcode in test execution
	_, err = exec.Exec(map[string]string{
		"symbols": "BTC",
	})

	if err != nil {
		return nil, fmt.Errorf("failed to run test program: %s", err.Error())
	}
	return exec, nil
}

// parseExecutor splits the executor string in the form of "name:base?timeout=" into parts.
func parseExecutor(executorStr string) (name string, base string, timeout time.Duration, err error) {
	executor := strings.SplitN(executorStr, ":", 2)
	if len(executor) != 2 {
		return "", "", 0, fmt.Errorf("invalid executor, cannot parse executor: %s", executorStr)
	}
	u, err := url.Parse(executor[1])
	if err != nil {
		return "", "", 0, fmt.Errorf("invalid url, cannot parse %s to url with error: %s", executor[1], err.Error())
	}

	query := u.Query()
	timeoutStr := query.Get(flagQueryTimeout)
	if timeoutStr == "" {
		return "", "", 0, fmt.Errorf("invalid timeout, executor requires query timeout")
	}
	// Remove timeout from query because we need to return `base`
	query.Del(flagQueryTimeout)
	u.RawQuery = query.Encode()

	timeout, err = time.ParseDuration(timeoutStr)
	if err != nil {
		return "", "", 0, fmt.Errorf("invalid timeout, cannot parse duration with error: %s", err.Error())
	}
	return executor[0], u.String(), timeout, nil
}
