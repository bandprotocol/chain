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
	ErrNotSupportedExecutor     = sdkerrors.Register(ModuleName, 156, "The executor is not supported")
	ErrWrongOutput              = sdkerrors.Register(ModuleName, 157, "Wrong output")
	ErrEventValueDoesNotExist   = sdkerrors.Register(ModuleName, 158, "Event value does not exist")
	ErrInvalidEventsCount       = sdkerrors.Register(ModuleName, 159, "Invalid events count")
	ErrUnknownEventType         = sdkerrors.Register(ModuleName, 160, "Unknown event type")
	ErrInconsistentCount        = sdkerrors.Register(ModuleName, 161, "Inconsistent count")
	ErrInvalidExecutionResult   = sdkerrors.Register(ModuleName, 162, "Invalid execution result")
	ErrInvalidCacheLoading      = sdkerrors.Register(ModuleName, 163, "Invalid cache loading")
)
