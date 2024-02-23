package types

import "cosmossdk.io/errors"

// x/bandtss module sentinel errors
var (
	ErrInvalidAccAddressFormat = errors.Register(ModuleName, 1, "account address format is invalid")
	ErrGroupSizeTooLarge       = errors.Register(ModuleName, 2, "group size is too large")

	ErrCreateGroupTSSError = errors.Register(ModuleName, 3, "failed to create")

	ErrInvalidStatus     = errors.Register(ModuleName, 6, "invalid status")
	ErrStatusIsNotActive = errors.Register(ModuleName, 7, "status is not active")
	ErrTooSoonToActivate = errors.Register(ModuleName, 8, "too soon to activate")

	ErrRequestReplacementFailed = errors.Register(ModuleName, 9, "failed to request replacement")
)
