package executor

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	flagQueryTimeout   = "timeout"
	flagQueryMaxTry    = "maxTry"
	flagQueryPortRange = "portRange"
)

var (
	ErrExecutionimeout = errors.New("execution timeout")
	ErrRestNotOk       = errors.New("rest return non 2XX response")
	ErrReachMaxTry     = errors.New("execution reach max try")
)

type ExecResult struct {
	Output  []byte
	Code    uint32
	Version string
}

type Executor interface {
	Exec(exec []byte, arg string, env interface{}) (ExecResult, error)
}

var testProgram []byte = []byte(
	"#!/usr/bin/env python3\nimport os\nimport sys\nprint(sys.argv[1], os.getenv('BAND_CHAIN_ID'))",
)

// NewExecutor returns executor by name and executor URL
func NewExecutor(executor string) (exec Executor, err error) {
	name, base, timeout, maxTry, startPort, endPort, err := parseExecutor(executor)
	if err != nil {
		return nil, err
	}
	switch name {
	case "rest":
		exec = NewRestExec(base, timeout)
	case "docker":
		// Only use in testnet. No intensive testing, use at your own risk
		if endPort < startPort {
			return nil, fmt.Errorf("portRange invalid: startPort: %d, endPort: %d", startPort, endPort)
		}
		if maxTry < 1 {
			return nil, fmt.Errorf("maxTry invalid: %d", maxTry)
		}
		exec = NewDockerExec(base, timeout, maxTry, startPort, endPort)
	default:
		return nil, fmt.Errorf("Invalid executor name: %s, base: %s", name, base)
	}

	// TODO: Remove hardcode in test execution
	res, err := exec.Exec(testProgram, "TEST_ARG", map[string]interface{}{
		"BAND_CHAIN_ID":    "test-chain-id",
		"BAND_VALIDATOR":   "test-validator",
		"BAND_REQUEST_ID":  "test-request-id",
		"BAND_EXTERNAL_ID": "test-external-id",
		"BAND_REPORTER":    "test-reporter",
		"BAND_SIGNATURE":   "test-signature",
	})

	if err != nil {
		return nil, fmt.Errorf("failed to run test program: %s", err.Error())
	}
	if res.Code != 0 {
		return nil, fmt.Errorf("test program returned nonzero code: %d", res.Code)
	}
	if string(res.Output) != "TEST_ARG test-chain-id\n" {
		return nil, fmt.Errorf("test program returned wrong output: %s", res.Output)
	}
	return exec, nil
}

// parseExecutor splits the executor string in the form of "name:base?timeout=&maxTry=&portRange=" into parts.
func parseExecutor(
	executorStr string,
) (name string, base string, timeout time.Duration, maxTry int, startPort int, endPort int, err error) {
	executor := strings.SplitN(executorStr, ":", 2)
	if len(executor) != 2 {
		return "", "", 0, 0, 0, 0, fmt.Errorf("Invalid executor, cannot parse executor: %s", executorStr)
	}
	u, err := url.Parse(executor[1])
	if err != nil {
		return "", "", 0, 0, 0, 0, fmt.Errorf(
			"Invalid url, cannot parse %s to url with error: %s",
			executor[1],
			err.Error(),
		)
	}

	query := u.Query()
	timeoutStr := query.Get(flagQueryTimeout)
	if timeoutStr == "" {
		return "", "", 0, 0, 0, 0, fmt.Errorf("Invalid timeout, executor requires query timeout")
	}
	timeout, err = time.ParseDuration(timeoutStr)
	if err != nil {
		return "", "", 0, 0, 0, 0, fmt.Errorf("Invalid timeout, cannot parse duration with error: %s", err.Error())
	}

	maxTryStr := query.Get(flagQueryMaxTry)
	if maxTryStr == "" {
		maxTryStr = "1"
	}
	maxTry, err = strconv.Atoi(maxTryStr)
	if err != nil {
		return "", "", 0, 0, 0, 0, fmt.Errorf("Invalid maxTry, cannot parse integer with error: %s", err.Error())
	}

	portRangeStr := query.Get(flagQueryPortRange)
	ports := strings.SplitN(portRangeStr, "-", 2)
	if len(ports) != 2 {
		ports = []string{"0", "0"}
	}
	startPort, err = strconv.Atoi(ports[0])
	if err != nil {
		return "", "", 0, 0, 0, 0, fmt.Errorf("Invalid portRange, cannot parse integer with error: %s", err.Error())
	}
	endPort, err = strconv.Atoi(ports[1])
	if err != nil {
		return "", "", 0, 0, 0, 0, fmt.Errorf("Invalid portRange, cannot parse integer with error: %s", err.Error())
	}

	// Remove timeout from query because we need to return `base`
	query.Del(flagQueryTimeout)
	query.Del(flagQueryMaxTry)
	query.Del(flagQueryPortRange)

	u.RawQuery = query.Encode()
	return executor[0], u.String(), timeout, maxTry, startPort, endPort, nil
}
