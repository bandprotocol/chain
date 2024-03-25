package types

import "cosmossdk.io/errors"

// x/bandtss module sentinel errors
var (
	ErrInvalidStatus         = errors.Register(ModuleName, 2, "invalid status")
	ErrStatusIsNotActive     = errors.Register(ModuleName, 3, "status is not active")
	ErrTooSoonToActivate     = errors.Register(ModuleName, 4, "too soon to activate")
	ErrNotEnoughFee          = errors.Register(ModuleName, 5, "not enough fee")
	ErrInvalidGroupID        = errors.Register(ModuleName, 6, "invalid groupID")
	ErrReplacementInProgress = errors.Register(ModuleName, 7, "group replacement is in progress")
	ErrSigningFeeNotFound    = errors.Register(ModuleName, 8, "signing fee not found")
)
