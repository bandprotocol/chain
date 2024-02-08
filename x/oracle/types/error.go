package types

import (
	errorsmod "cosmossdk.io/errors"
)

var (
	ErrOwasmCompilation         = errorsmod.Register(ModuleName, 1, "owasm compilation failed")
	ErrBadWasmExecution         = errorsmod.Register(ModuleName, 2, "bad wasm execution")
	ErrDataSourceNotFound       = errorsmod.Register(ModuleName, 3, "data source not found")
	ErrOracleScriptNotFound     = errorsmod.Register(ModuleName, 4, "oracle script not found")
	ErrRequestNotFound          = errorsmod.Register(ModuleName, 5, "request not found")
	ErrRawRequestNotFound       = errorsmod.Register(ModuleName, 6, "raw request not found")
	ErrReporterNotFound         = errorsmod.Register(ModuleName, 7, "reporter not found")
	ErrResultNotFound           = errorsmod.Register(ModuleName, 8, "result not found")
	ErrReporterAlreadyExists    = errorsmod.Register(ModuleName, 9, "reporter already exists")
	ErrValidatorNotRequested    = errorsmod.Register(ModuleName, 10, "validator not requested")
	ErrValidatorAlreadyReported = errorsmod.Register(ModuleName, 11, "validator already reported")
	ErrInvalidReportSize        = errorsmod.Register(ModuleName, 12, "invalid report size")
	ErrReporterNotAuthorized    = errorsmod.Register(ModuleName, 13, "reporter not authorized")
	ErrEditorNotAuthorized      = errorsmod.Register(ModuleName, 14, "editor not authorized")
	ErrValidatorAlreadyActive   = errorsmod.Register(ModuleName, 16, "validator already active")
	ErrTooSoonToActivate        = errorsmod.Register(ModuleName, 17, "too soon to activate")
	ErrTooLongName              = errorsmod.Register(ModuleName, 18, "too long name")
	ErrTooLongDescription       = errorsmod.Register(ModuleName, 19, "too long description")
	ErrEmptyExecutable          = errorsmod.Register(ModuleName, 20, "empty executable")
	ErrEmptyWasmCode            = errorsmod.Register(ModuleName, 21, "empty wasm code")
	ErrTooLargeExecutable       = errorsmod.Register(ModuleName, 22, "too large executable")
	ErrTooLargeWasmCode         = errorsmod.Register(ModuleName, 23, "too large wasm code")
	ErrInvalidMinCount          = errorsmod.Register(ModuleName, 24, "invalid min count")
	ErrInvalidAskCount          = errorsmod.Register(ModuleName, 25, "invalid ask count")
	ErrTooLargeCalldata         = errorsmod.Register(ModuleName, 26, "too large calldata")
	ErrTooLongClientID          = errorsmod.Register(ModuleName, 27, "too long client id")
	ErrEmptyRawRequests         = errorsmod.Register(ModuleName, 28, "empty raw requests")
	ErrEmptyReport              = errorsmod.Register(ModuleName, 29, "empty report")
	ErrDuplicateExternalID      = errorsmod.Register(ModuleName, 30, "duplicate external id")
	ErrTooLongSchema            = errorsmod.Register(ModuleName, 31, "too long schema")
	ErrTooLongURL               = errorsmod.Register(ModuleName, 32, "too long url")
	ErrTooLargeRawReportData    = errorsmod.Register(ModuleName, 33, "too large raw report data")
	ErrInsufficientValidators   = errorsmod.Register(ModuleName, 34, "insufficient available validators")
	ErrCreateWithDoNotModify    = errorsmod.Register(ModuleName, 35, "cannot create with [do-not-modify] content")
	ErrSelfReferenceAsReporter  = errorsmod.Register(ModuleName, 36, "cannot reference self as reporter")
	ErrOBIDecode                = errorsmod.Register(ModuleName, 37, "obi decode failed")
	ErrUncompressionFailed      = errorsmod.Register(ModuleName, 38, "uncompression failed")
	ErrRequestAlreadyExpired    = errorsmod.Register(ModuleName, 39, "request already expired")
	ErrBadDrbgInitialization    = errorsmod.Register(ModuleName, 40, "bad drbg initialization")
	ErrMaxOracleChannels        = errorsmod.Register(ModuleName, 41, "max oracle channels")
	ErrInvalidVersion           = errorsmod.Register(ModuleName, 42, "invalid ICS20 version")
	ErrNotEnoughFee             = errorsmod.Register(ModuleName, 43, "not enough fee")
	ErrInvalidOwasmGas          = errorsmod.Register(ModuleName, 44, "invalid owasm gas")
	ErrIBCRequestDisabled       = errorsmod.Register(ModuleName, 45, "sending oracle request via IBC is disabled")
	ErrSigningResultNotFound    = errorsmod.Register(ModuleName, 46, "signing result not found")
	ErrReportNotFound           = errorsmod.Register(ModuleName, 47, "report not found")
)

// WrapMaxError wraps an error message with additional info of the current and max values.
func WrapMaxError(err *errorsmod.Error, got int, max int) error {
	return err.Wrapf("got: %d, max: %d", got, max)
}
