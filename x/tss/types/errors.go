package types

import "cosmossdk.io/errors"

// x/tss module sentinel errors
var (
	ErrInvalidAccAddressFormat = errors.Register(ModuleName, 1, "account address format is invalid")
	ErrGroupSizeTooLarge       = errors.Register(ModuleName, 2, "group size is too large")
	ErrGroupNotFound           = errors.Register(ModuleName, 3, "group not found")
	ErrMemberNotFound          = errors.Register(ModuleName, 4, "member not found")
	ErrNoActiveMember          = errors.Register(ModuleName, 5, "no active member in this group")
	ErrMemberAlreadySubmit     = errors.Register(ModuleName, 6, "member is already submit message")
	ErrRound1InfoNotFound      = errors.Register(ModuleName, 7, "round 1 info not found")
	ErrDKGContextNotFound      = errors.Register(ModuleName, 8, "dkg context not found")
	ErrMemberNotAuthorized     = errors.Register(
		ModuleName,
		9,
		"member is not authorized for this group",
	)
	ErrInvalidStatus                = errors.Register(ModuleName, 10, "invalid status")
	ErrGroupExpired                 = errors.Register(ModuleName, 11, "group expired")
	ErrVerifyOneTimeSignatureFailed = errors.Register(ModuleName, 12, "failed to verify one time sign")
	ErrVerifyA0SignatureFailed      = errors.Register(ModuleName, 13, "failed to verify a0 sign")
	ErrAddCoefCommit                = errors.Register(ModuleName, 14, "failed to add coefficient commit")
	ErrInvalidLengthCoefCommits     = errors.Register(
		ModuleName,
		15,
		"coefficients commit length is invalid",
	)
	ErrRound2InfoNotFound                 = errors.Register(ModuleName, 16, "round 2 info not found")
	ErrInvalidLengthEncryptedSecretShares = errors.Register(
		ModuleName,
		17,
		"encrypted secret shares length is invalid ",
	)
	ErrComputeOwnPubKeyFailed           = errors.Register(ModuleName, 18, "failed to compute own public key")
	ErrMemberIsAlreadyComplainOrConfirm = errors.Register(ModuleName, 19, "member is already complain or confirm")
	ErrEncryptedSecretShareNotFound     = errors.Register(ModuleName, 20, "encrypted secret share not found")
	ErrComplainFailed                   = errors.Register(ModuleName, 21, "failed to complain")
	ErrConfirmFailed                    = errors.Register(ModuleName, 22, "failed to confirm")
	ErrConfirmNotFound                  = errors.Register(ModuleName, 23, "confirm not found")
	ErrComplaintsWithStatusNotFound     = errors.Register(ModuleName, 24, "complaints with status not found")
	ErrDENotFound                       = errors.Register(ModuleName, 25, "de not found")
	ErrGroupIsNotActive                 = errors.Register(ModuleName, 26, "group is not active")
	ErrUnexpectedThreshold              = errors.Register(ModuleName, 27, "threshold value is unexpected")
	ErrBadDrbgInitialization            = errors.Register(ModuleName, 28, "bad drbg initialization")
	ErrPartialSignatureNotFound         = errors.Register(ModuleName, 29, "partial signature not found")
	ErrInvalidArgument                  = errors.Register(ModuleName, 30, "invalid argument")
	ErrSigningNotFound                  = errors.Register(ModuleName, 31, "signing not found")
	ErrAlreadySigned                    = errors.Register(ModuleName, 32, "already signed")
	ErrSigningAlreadySuccess            = errors.Register(ModuleName, 33, "signing already success")
	ErrPubNonceNotEqualToSigR           = errors.Register(ModuleName, 34, "public nonce not equal to signature r")
	ErrMemberNotAssigned                = errors.Register(ModuleName, 35, "member is not assigned")
	ErrVerifySigningSigFailed           = errors.Register(ModuleName, 36, "failed to verify signing signature")
	ErrCombineSigsFailed                = errors.Register(ModuleName, 37, "failed to combine signatures")
	ErrVerifyGroupSigningSigFailed      = errors.Register(
		ModuleName,
		38,
		"failed to verify group signing signature",
	)
	ErrDEQueueFull                   = errors.Register(ModuleName, 39, "de queue is full")
	ErrSigningExpired                = errors.Register(ModuleName, 40, "signing expired")
	ErrStatusNotFound                = errors.Register(ModuleName, 41, "status not found")
	ErrHandleSignatureOrderFailed    = errors.Register(ModuleName, 42, "failed to handle signature order")
	ErrNoSignatureOrderHandlerExists = errors.Register(ModuleName, 43, "no handler exists for signature order type")
	ErrReplacementNotFound           = errors.Register(ModuleName, 44, "replacement group not found")
	ErrRequestReplacementFailed      = errors.Register(ModuleName, 45, "failed to request replacement")
)
