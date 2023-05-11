package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/tss module sentinel errors
var (
	ErrGroupNotFound             = sdkerrors.Register(ModuleName, 2, "fail to verify a0 sign")
	ErrMemberNotFound            = sdkerrors.Register(ModuleName, 3, "fail to verify a0 sign")
	ErrRound1CommitmentsNotFound = sdkerrors.Register(ModuleName, 4, "round 1 commitments not found")
	ErrDKGContextNotFound        = sdkerrors.Register(ModuleName, 5, "dkg context not found")
	ErrMemberNotAuthorized       = sdkerrors.Register(ModuleName, 6, "member is not authorized for this group")
	ErrRound1AlreadyExpired      = sdkerrors.Register(ModuleName, 7, "round 1 already expired")
	ErrAlreadyCommitRound1       = sdkerrors.Register(ModuleName, 8, "already commit round 1 message")
	ErrVerifyOneTimeSigFailed    = sdkerrors.Register(ModuleName, 9, "fail to verify one time sign")
	ErrVerifyA0SigFailed         = sdkerrors.Register(ModuleName, 10, "fail to verify a0 sign")
)
