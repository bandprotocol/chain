package executor

import (
	"github.com/GeoDB-Limited/odin-core/yoda/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"net/url"
	"strings"
	"time"
)

const (
	flagQueryTimeout = "timeout"

	RestExecutorType   = "rest"
	DockerExecutorType = "docker"
)

const (
	EnvVarReporter    = "ODIN_REPORTER"
	EnvVarChainID     = "ODIN_CHAIN_ID"
	EnvVarValidator   = "ODIN_VALIDATOR"
	EnvVarRequestID   = "ODIN_REQUEST_ID"
	EnvVarExternalID  = "ODIN_EXTERNAL_ID"
	EnvVarSignature   = "ODIN_SIGNATURE"
)

type ExecResult struct {
	Output  []byte
	Code    uint32
	Version string
}

type Executor interface {
	Exec(exec []byte, arg string, env interface{}) (ExecResult, error)
}

// NewExecutor returns executor by name and executor URL
func NewExecutor(executor string) (exec Executor, err error) {
	name, base, timeout, err := parseExecutor(executor)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "failed to parse executor")
	}
	switch name {
	case RestExecutorType:
		exec = NewRestExec(base, timeout)
	case DockerExecutorType:
		return nil, sdkerrors.Wrap(errors.ErrNotSupportedExecutor, "docker executor is currently not supported")
	default:
		return nil, sdkerrors.Wrapf(errors.ErrNotSupportedExecutor, "executor name: %s, base: %s", name, base)
	}
	return exec, nil
}

// parseExecutor splits the executor string in the form of "name:base?timeout=" into parts.
func parseExecutor(executorStr string) (name string, base string, timeout time.Duration, err error) {
	executor := strings.SplitN(executorStr, ":", 2)
	if len(executor) != 2 {
		return "", "", 0,
			sdkerrors.Wrapf(errors.ErrNotSupportedExecutor, "cannot parse executor: %s", executorStr)
	}
	u, err := url.Parse(executor[1])
	if err != nil {
		return "", "", 0,
			sdkerrors.Wrapf(err, "invalid url, cannot parse %s to url", executor[1])
	}

	query := u.Query()
	timeoutStr := query.Get(flagQueryTimeout)
	if timeoutStr == "" {
		return "", "", 0, sdkerrors.Wrap(errors.ErrExecutionTimeout, "executor requires query timeout")
	}
	// Remove timeout from query because we need to return `base`
	query.Del(flagQueryTimeout)
	u.RawQuery = query.Encode()

	timeout, err = time.ParseDuration(timeoutStr)
	if err != nil {
		return "", "", 0, sdkerrors.Wrap(err, "invalid timeout, cannot parse duration")
	}
	return executor[0], u.String(), timeout, nil
}
