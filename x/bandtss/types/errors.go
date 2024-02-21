package types

import "cosmossdk.io/errors"

// x/tss module sentinel errors
var (
	ErrInvalidAccAddressFormat = errors.Register(ModuleName, 1, "account address format is invalid")
	ErrGroupSizeTooLarge       = errors.Register(ModuleName, 2, "group size is too large")

	ErrCreateGroupTSSError   = errors.Register(ModuleName, 3, "failed to create")
	ErrUnexpectedThreshold   = errors.Register(ModuleName, 4, "threshold value is unexpected")
	ErrBadDrbgInitialization = errors.Register(ModuleName, 5, "bad drbg initialization")

	ErrStatusIsNotActive = errors.Register(ModuleName, 6, "status is not active")

	ErrRequestReplacementFailed = errors.Register(ModuleName, 7, "failed to request replacement")

	ErrNotEnoughFee                  = errors.Register(ModuleName, 8, "not enough fee")
	ErrHandleSignatureOrderFailed    = errors.Register(ModuleName, 9, "failed to handle signature order")
	ErrNoSignatureOrderHandlerExists = errors.Register(ModuleName, 10, "no handler exists for signature order type")
)
