package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/tss module sentinel errors
var (
	ErrInvalidAccAddressFormat = sdkerrors.Register(ModuleName, 2, "account address format is invalid")
	ErrGroupSizeTooLarge       = sdkerrors.Register(ModuleName, 3, "group size is too large")
	ErrGroupNotFound           = sdkerrors.Register(ModuleName, 4, "group not found")
	ErrMemberNotFound          = sdkerrors.Register(ModuleName, 5, "member not found")
	ErrAlreadySubmit           = sdkerrors.Register(ModuleName, 6, "member is already submit message")
	ErrRound1DataNotFound      = sdkerrors.Register(ModuleName, 7, "round 1 data not found")
	ErrDKGContextNotFound      = sdkerrors.Register(ModuleName, 8, "dkg context not found")
	ErrMemberNotAuthorized     = sdkerrors.Register(
		ModuleName,
		9,
		"member is not authorized for this group",
	)
	ErrRoundExpired            = sdkerrors.Register(ModuleName, 10, "round expired")
	ErrVerifyOneTimeSigFailed  = sdkerrors.Register(ModuleName, 11, "fail to verify one time sign")
	ErrVerifyA0SigFailed       = sdkerrors.Register(ModuleName, 12, "fail to verify a0 sign")
	ErrAddCommit               = sdkerrors.Register(ModuleName, 13, "fail to add coefficient commit")
	ErrCommitsNotCorrectLength = sdkerrors.Register(
		ModuleName,
		14,
		"coefficients commit not correct length",
	)
	ErrRound2DataNotFound                    = sdkerrors.Register(ModuleName, 15, "round 2 data not found")
	ErrEncryptedSecretSharesNotCorrectLength = sdkerrors.Register(
		ModuleName,
		16,
		"encrypted secret shares not correct length",
	)
	ErrComputeOwnPubKeyFailed           = sdkerrors.Register(ModuleName, 17, "fail to compute own public key")
	ErrMemberIsAlreadyComplainOrConfirm = sdkerrors.Register(ModuleName, 18, "member is already complain or confirm")
	ErrComplainFailed                   = sdkerrors.Register(ModuleName, 19, "complain failed")
	ErrConfirmFailed                    = sdkerrors.Register(ModuleName, 20, "confirm failed")
	ErrConfirmNotFound                  = sdkerrors.Register(ModuleName, 21, "confirm not found")
	ErrComplainsWithStatusNotFound      = sdkerrors.Register(ModuleName, 22, "complains with status not found")
	ErrDENotFound                       = sdkerrors.Register(ModuleName, 23, "DE not found")
	ErrGroupIsNotActive                 = sdkerrors.Register(ModuleName, 24, "group is not active")
	ErrBadDrbgInitialization            = sdkerrors.Register(ModuleName, 25, "bad drbg initialization")
	ErrPartialSigNotFound               = sdkerrors.Register(ModuleName, 26, "partial sig not found")
	ErrInvalidArgument                  = sdkerrors.Register(ModuleName, 27, "invalid argument")
	ErrSigningNotFound                  = sdkerrors.Register(ModuleName, 28, "signing not found")
	ErrAlreadySigned                    = sdkerrors.Register(ModuleName, 29, "already signed")
	ErrSigningAlreadySuccess            = sdkerrors.Register(ModuleName, 30, "signing already success")
	ErrPubNonceNotEqualToSigR           = sdkerrors.Register(ModuleName, 31, "public nonce not equal to signature r")
	ErrMemberNotAssigned                = sdkerrors.Register(ModuleName, 32, "member is not assigned participants")
	ErrVerifySigningSigFailed           = sdkerrors.Register(ModuleName, 33, "failed to verify signing signature")
	ErrCombineSigsFailed                = sdkerrors.Register(ModuleName, 34, "failed to combine signatures")
	ErrVerifyGroupSigningSigFailed      = sdkerrors.Register(
		ModuleName,
		35,
		"failed to verify group signing signature",
	)
)
