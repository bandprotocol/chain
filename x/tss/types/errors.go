package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/tss module sentinel errors
var (
	ErrInvalidAccAddressFormat   = sdkerrors.Register(ModuleName, 2, "account address format is invalid")
	ErrGroupNotFound             = sdkerrors.Register(ModuleName, 3, "group not found")
	ErrMemberNotFound            = sdkerrors.Register(ModuleName, 4, "member not found")
	ErrRound1CommitmentsNotFound = sdkerrors.Register(ModuleName, 5, "round 1 commitments not found")
	ErrDKGContextNotFound        = sdkerrors.Register(ModuleName, 6, "dkg context not found")
	ErrMemberNotAuthorized       = sdkerrors.Register(
		ModuleName,
		7,
		"member is not authorized for this group",
	)
	ErrRoundExpired                          = sdkerrors.Register(ModuleName, 7, "round expired")
	ErrAlreadySubmit                         = sdkerrors.Register(ModuleName, 8, "member is already submit message")
	ErrVerifyOneTimeSigFailed                = sdkerrors.Register(ModuleName, 9, "fail to verify one time sign")
	ErrVerifyA0SigFailed                     = sdkerrors.Register(ModuleName, 10, "fail to verify a0 sign")
	ErrRound2ShareNotFound                   = sdkerrors.Register(ModuleName, 11, "round 2 share not found")
	ErrRound2ShareNotCorrectLength           = sdkerrors.Register(ModuleName, 12, "round 2 share not correct length")
	ErrDKGMaliciousIndexesNotFound           = sdkerrors.Register(ModuleName, 13, "dkg malicious indexes not found")
	ErrMemberIsAlreadyMalicious              = sdkerrors.Register(ModuleName, 14, "member is already malicious")
	ErrComplainsFailed                       = sdkerrors.Register(ModuleName, 15, "complains failed")
	ErrConfirmationsNotFound                 = sdkerrors.Register(ModuleName, 16, "confirmations not found")
	ErrPendingRoundNoteNotFound              = sdkerrors.Register(ModuleName, 17, "pending round note not found")
	ErrEncryptedSecretSharesNotCorrectLength = sdkerrors.Register(
		ModuleName,
		18,
		"encrypted secret shares not correct length",
	)
	ErrRound1DataNotFound = sdkerrors.Register(ModuleName, 19, "round1 data not found")
	ErrRound2DataNotFound = sdkerrors.Register(ModuleName, 20, "round2 data not found")
)
