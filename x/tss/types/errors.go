package types

import "cosmossdk.io/errors"

// x/tss module sentinel errors
var (
	ErrInvalidAccAddressFormat = errors.Register(ModuleName, 1, "account address format is invalid")
	ErrGroupNotFound           = errors.Register(ModuleName, 2, "group not found")
	ErrMemberNotFound          = errors.Register(ModuleName, 3, "member not found")
	ErrNoActiveMember          = errors.Register(ModuleName, 4, "no active member in this group")
	ErrMemberAlreadySubmit     = errors.Register(ModuleName, 5, "member is already submit message")
	ErrRound1InfoNotFound      = errors.Register(ModuleName, 6, "round 1 info not found")
	ErrDKGContextNotFound      = errors.Register(ModuleName, 7, "dkg context not found")
	ErrMemberNotAuthorized     = errors.Register(
		ModuleName,
		8,
		"member is not authorized for this group",
	)
	ErrInvalidStatus                = errors.Register(ModuleName, 9, "invalid status")
	ErrGroupExpired                 = errors.Register(ModuleName, 10, "group expired")
	ErrVerifyOneTimeSignatureFailed = errors.Register(ModuleName, 11, "failed to verify one time sign")
	ErrVerifyA0SignatureFailed      = errors.Register(ModuleName, 12, "failed to verify a0 sign")
	ErrAddCoefCommit                = errors.Register(ModuleName, 13, "failed to add coefficient commit")
	ErrInvalidLengthCoefCommits     = errors.Register(
		ModuleName,
		14,
		"coefficients commit length is invalid",
	)
	ErrRound2InfoNotFound                 = errors.Register(ModuleName, 15, "round 2 info not found")
	ErrInvalidLengthEncryptedSecretShares = errors.Register(
		ModuleName,
		16,
		"encrypted secret shares length is invalid ",
	)
	ErrComputeOwnPubKeyFailed           = errors.Register(ModuleName, 17, "failed to compute own public key")
	ErrMemberIsAlreadyComplainOrConfirm = errors.Register(ModuleName, 18, "member is already complain or confirm")
	ErrEncryptedSecretShareNotFound     = errors.Register(ModuleName, 19, "encrypted secret share not found")
	ErrComplainFailed                   = errors.Register(ModuleName, 20, "failed to complain")
	ErrConfirmFailed                    = errors.Register(ModuleName, 21, "failed to confirm")
	ErrConfirmNotFound                  = errors.Register(ModuleName, 22, "confirm not found")
	ErrComplaintsWithStatusNotFound     = errors.Register(ModuleName, 23, "complaints with status not found")
	ErrDENotFound                       = errors.Register(ModuleName, 24, "de not found")
	ErrGroupIsNotActive                 = errors.Register(ModuleName, 25, "group is not active")
	ErrUnexpectedThreshold              = errors.Register(ModuleName, 26, "threshold value is unexpected")
	ErrBadDrbgInitialization            = errors.Register(ModuleName, 27, "bad drbg initialization")
	ErrPartialSignatureNotFound         = errors.Register(ModuleName, 28, "partial signature not found")
	ErrInvalidArgument                  = errors.Register(ModuleName, 29, "invalid argument")
	ErrSigningNotFound                  = errors.Register(ModuleName, 30, "signing not found")
	ErrAlreadySigned                    = errors.Register(ModuleName, 31, "already signed")
	ErrSigningAlreadySuccess            = errors.Register(ModuleName, 32, "signing already success")
	ErrPubNonceNotEqualToSigR           = errors.Register(ModuleName, 33, "public nonce not equal to signature r")
	ErrMemberNotAssigned                = errors.Register(ModuleName, 34, "member is not assigned")
	ErrVerifySigningSigFailed           = errors.Register(ModuleName, 35, "failed to verify signing signature")
	ErrCombineSigsFailed                = errors.Register(ModuleName, 36, "failed to combine signatures")
	ErrVerifyGroupSigningSigFailed      = errors.Register(
		ModuleName,
		37,
		"failed to verify group signing signature",
	)
	ErrDEQueueFull         = errors.Register(ModuleName, 38, "de queue is full")
	ErrSigningExpired      = errors.Register(ModuleName, 39, "signing expired")
	ErrStatusNotFound      = errors.Register(ModuleName, 40, "status not found")
	ErrReplacementNotFound = errors.Register(ModuleName, 41, "replacement group not found")
)
