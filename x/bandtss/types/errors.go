package types

import "cosmossdk.io/errors"

// x/tss module sentinel errors
var (
	ErrInvalidAccAddressFormat = errors.Register(ModuleName, 1, "account address format is invalid")
	ErrGroupSizeTooLarge       = errors.Register(ModuleName, 2, "group size is too large")

	ErrCreateGroupTSSError   = errors.Register(ModuleName, 3, "failed to create")
	ErrUnexpectedThreshold   = errors.Register(ModuleName, 4, "threshold value is unexpected")
	ErrBadDrbgInitialization = errors.Register(ModuleName, 5, "bad drbg initialization")

	ErrInvalidStatus     = errors.Register(ModuleName, 6, "invalid status")
	ErrStatusIsNotActive = errors.Register(ModuleName, 7, "status is not active")
	ErrTooSoonToActivate = errors.Register(ModuleName, 8, "too soon to activate")

	ErrRequestReplacementFailed = errors.Register(ModuleName, 9, "failed to request replacement")

	ErrNotEnoughFee                  = errors.Register(ModuleName, 10, "not enough fee")
	ErrHandleSignatureOrderFailed    = errors.Register(ModuleName, 11, "failed to handle signature order")
	ErrNoSignatureOrderHandlerExists = errors.Register(ModuleName, 12, "no handler exists for signature order type")
)
