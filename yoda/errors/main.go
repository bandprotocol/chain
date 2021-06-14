package errors

import sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

const (
	ModuleName = "yoda"
)

var (
	ErrEmptyChainIDParam        = sdkerrors.Register(ModuleName, 150, "Empty chain ID parameter")
	ErrNoKeyAvailable           = sdkerrors.Register(ModuleName, 151, "There is no key available")
	ErrNotReportedDataMsg       = sdkerrors.Register(ModuleName, 152, "Unsupported non-report data message")
	ErrExecutionTimeout         = sdkerrors.Register(ModuleName, 153, "Execution timeout")
	ErrNotOkResponse            = sdkerrors.Register(ModuleName, 154, "The response status is not OK")
	ErrUnknownMultiExecStrategy = sdkerrors.Register(ModuleName, 155, "Unknown multi execution strategy")
	ErrUnknownExecutor          = sdkerrors.Register(ModuleName, 156, "Unknown executor")
	ErrWrongOutput              = sdkerrors.Register(ModuleName, 157, "Wrong output")
)
