package types

import errorsmod "cosmossdk.io/errors"

// x/bandtss module sentinel errors
var (
	ErrInvalidStatus           = errorsmod.Register(ModuleName, 2, "invalid status")
	ErrStatusIsNotActive       = errorsmod.Register(ModuleName, 3, "status is not active")
	ErrTooSoonToActivate       = errorsmod.Register(ModuleName, 4, "too soon to activate")
	ErrFeeExceedsLimit         = errorsmod.Register(ModuleName, 5, "fee exceeds limit")
	ErrInvalidGroupID          = errorsmod.Register(ModuleName, 6, "invalid groupID")
	ErrNoActiveGroup           = errorsmod.Register(ModuleName, 7, "no active group")
	ErrReplacementInProgress   = errorsmod.Register(ModuleName, 8, "group replacement is in progress")
	ErrInvalidExecTime         = errorsmod.Register(ModuleName, 9, "invalid exec time")
	ErrSigningNotFound         = errorsmod.Register(ModuleName, 10, "signing not found")
	ErrMemberNotFound          = errorsmod.Register(ModuleName, 11, "member not found")
	ErrMemberAlreadyExists     = errorsmod.Register(ModuleName, 12, "member already exists")
	ErrMemberAlreadyActive     = errorsmod.Register(ModuleName, 13, "member already active")
	ErrMemberDuplicate         = errorsmod.Register(ModuleName, 14, "duplicated member found within the list")
	ErrInvalidSigningThreshold = errorsmod.Register(ModuleName, 15, "invalid signing threshold number")
)
