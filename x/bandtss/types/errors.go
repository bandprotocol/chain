package types

import errorsmod "cosmossdk.io/errors"

// x/bandtss module sentinel errors
var (
	ErrPenaltyDurationNotElapsed = errorsmod.Register(ModuleName, 2, "not allowed to activate due to penalty duration")
	ErrFeeExceedsLimit           = errorsmod.Register(ModuleName, 3, "fee exceeds limit")
	ErrNoCurrentGroup            = errorsmod.Register(ModuleName, 4, "no current group")
	ErrTransitionInProgress      = errorsmod.Register(ModuleName, 5, "group transition is in progress")
	ErrInvalidExecTime           = errorsmod.Register(ModuleName, 6, "invalid exec time")
	ErrSigningNotFound           = errorsmod.Register(ModuleName, 7, "signing not found")
	ErrMemberNotFound            = errorsmod.Register(ModuleName, 8, "member not found")
	ErrMemberAlreadyExists       = errorsmod.Register(ModuleName, 9, "member already exists")
	ErrMemberAlreadyActive       = errorsmod.Register(ModuleName, 10, "member already active")
	ErrMemberDuplicated          = errorsmod.Register(ModuleName, 11, "duplicated member found within the list")
	ErrInvalidThreshold          = errorsmod.Register(ModuleName, 12, "invalid threshold number")
	ErrContentNotAllowed         = errorsmod.Register(ModuleName, 13, "content not allowed")
	ErrInvalidIncomingGroup      = errorsmod.Register(ModuleName, 14, "invalid incoming group")
	ErrNoActiveGroup             = errorsmod.Register(ModuleName, 15, "no active group supported")
	ErrNoIncomingGroup           = errorsmod.Register(ModuleName, 16, "no incoming group")
	ErrInvalidGroupID            = errorsmod.Register(ModuleName, 17, "invalid group ID")
	ErrInvalidMember             = errorsmod.Register(ModuleName, 18, "invalid member")
)
