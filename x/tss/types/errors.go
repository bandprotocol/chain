package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/tss module sentinel errors
var (
	ErrRound1NoteNotFound        = sdkerrors.Register(ModuleName, 2, "round 1 note not found")
	ErrRound1CommitmentsNotFound = sdkerrors.Register(ModuleName, 3, "round 1 commitments not found")
	ErrMemberNotAuthorized       = sdkerrors.Register(ModuleName, 4, "member is not authorized for this group")
	ErrRound1AlreadyExpired      = sdkerrors.Register(ModuleName, 5, "round 1 already expired")
	ErrAlreadyCommitRound1       = sdkerrors.Register(ModuleName, 6, "this sender already commit round 1 message")
	ErrVerifyOneTimeSigFailed    = sdkerrors.Register(ModuleName, 7, "fail to verify one time sign")
	ErrVerifyA0SigFailed         = sdkerrors.Register(ModuleName, 8, "fail to verify a0 sign")
)
