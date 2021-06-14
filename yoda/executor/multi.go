package executor

import (
	"github.com/GeoDB-Limited/odin-core/yoda/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"strings"
	"sync/atomic"
)

const (
	OrderStrategyType      = "order"
	RoundRobinStrategyType = "round-robin"
)

// MultiExec is a higher-order executor that utlizes the underlying executors to perform Exec.
type MultiExec struct {
	execs    []Executor // The underlying executors (duplicated if strategy is round-robin).
	strategy string     // Execution strategy. Can be "order" or "round-robin".

	// Round-robin specific state variables.
	rIndex  int64 // Current round-robin starting index (need to mod rLength).
	rLength int64 // Total number of available executors.
}

// MultiError encapsulates error messages from the underlying executors into one error.
type MultiError struct {
	errs []error
}

// NewMultiExec creates a new MultiExec instance.
func NewMultiExec(execs []Executor, strategy string) (*MultiExec, error) {
	switch strategy {
	case OrderStrategyType:
		return &MultiExec{execs: execs, strategy: strategy, rIndex: 0, rLength: 0}, nil
	case RoundRobinStrategyType:
		rExecs := append(append([]Executor{}, execs...), execs...)
		return &MultiExec{
			execs:    rExecs,
			strategy: strategy,
			rIndex:   -1,
			rLength:  int64(len(execs)),
		}, nil
	default:
		return &MultiExec{}, sdkerrors.Wrapf(errors.ErrUnknownMultiExecStrategy, "strategy: %s", strategy)
	}
}

// Error implements error interface for MultiError by returning all error messages concatenated.
func (e *MultiError) Error() string {
	var s strings.Builder
	s.WriteString("MultiError: ")
	for idx, each := range e.errs {
		if idx != 0 {
			s.WriteString(", ")
		}
		s.WriteString(each.Error())
	}
	return s.String()
}

// nextExecOrder returns the next order of executors to be used by MultiExec.
func (e *MultiExec) nextExecOrder() []Executor {
	switch e.strategy {
	case OrderStrategyType:
		return e.execs
	case RoundRobinStrategyType:
		rIndex := atomic.AddInt64(&e.rIndex, 1) % e.rLength
		return e.execs[rIndex : rIndex+e.rLength]
	default:
		panic(sdkerrors.Wrapf(errors.ErrUnknownMultiExecStrategy, "strategy: %s", e.strategy)) // We should never reach here.
	}
}

// Exec implements Executor interface for MultiExec.
func (e *MultiExec) Exec(code []byte, arg string, env interface{}) (ExecResult, error) {
	var errs []error
	for _, each := range e.nextExecOrder() {
		res, err := each.Exec(code, arg, env)
		if err == nil || err == errors.ErrExecutionTimeout {
			return res, sdkerrors.Wrap(err, "execution failed")
		} else {
			errs = append(errs, err)
		}
	}
	return ExecResult{}, &MultiError{errs: errs}
}
