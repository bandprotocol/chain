package types

import errorsmod "cosmossdk.io/errors"

// x/bandtss module sentinel errors
var (
	ErrInvalidStatus           = errorsmod.Register(ModuleName, 2, "invalid status")
	ErrTooSoonToActivate       = errorsmod.Register(ModuleName, 3, "too soon to activate")
	ErrFeeExceedsLimit         = errorsmod.Register(ModuleName, 4, "fee exceeds limit")
	ErrNoCurrentGroup          = errorsmod.Register(ModuleName, 5, "no current group")
	ErrTransitionInProgress    = errorsmod.Register(ModuleName, 6, "group transition is in progress")
	ErrInvalidExecTime         = errorsmod.Register(ModuleName, 7, "invalid exec time")
	ErrSigningNotFound         = errorsmod.Register(ModuleName, 8, "signing not found")
	ErrMemberNotFound          = errorsmod.Register(ModuleName, 9, "member not found")
	ErrMemberAlreadyExists     = errorsmod.Register(ModuleName, 10, "member already exists")
	ErrMemberAlreadyActive     = errorsmod.Register(ModuleName, 11, "member already active")
	ErrMemberDuplicate         = errorsmod.Register(ModuleName, 12, "duplicated member found within the list")
	ErrInvalidSigningThreshold = errorsmod.Register(ModuleName, 13, "invalid signing threshold number")
	ErrContentNotAllowed       = errorsmod.Register(ModuleName, 14, "content not allowed")
	ErrInvalidIncomingGroup    = errorsmod.Register(ModuleName, 15, "invalid incoming group")
	ErrNoActiveGroup           = errorsmod.Register(ModuleName, 16, "no active group supported")
	ErrNoIncomingGroup         = errorsmod.Register(ModuleName, 17, "no incoming group")
	ErrInvalidGroupID          = errorsmod.Register(ModuleName, 18, "invalid group ID")
)
