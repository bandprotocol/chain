package types

import "cosmossdk.io/errors"

// x/tss module sentinel errors
var (
	ErrInvalidAccAddressFormat = errors.Register(ModuleName, 2, "account address format is invalid")
	ErrGroupSizeTooLarge       = errors.Register(ModuleName, 3, "group size is too large")
	ErrGroupNotFound           = errors.Register(ModuleName, 4, "group not found")
	ErrMemberNotFound          = errors.Register(ModuleName, 5, "member not found")
	ErrNoActiveMember          = errors.Register(ModuleName, 6, "No active member in this group")
	ErrAlreadySubmit           = errors.Register(ModuleName, 7, "member is already submit message")
	ErrRound1InfoNotFound      = errors.Register(ModuleName, 8, "round 1 info not found")
	ErrDKGContextNotFound      = errors.Register(ModuleName, 9, "dkg context not found")
	ErrMemberNotAuthorized     = errors.Register(
		ModuleName,
		10,
		"member is not authorized for this group",
	)
	ErrInvalidStatus           = errors.Register(ModuleName, 11, "invalid status")
	ErrGroupExpired            = errors.Register(ModuleName, 12, "group expired")
	ErrVerifyOneTimeSigFailed  = errors.Register(ModuleName, 13, "failed to verify one time sign")
	ErrVerifyA0SigFailed       = errors.Register(ModuleName, 14, "failed to verify a0 sign")
	ErrAddCommit               = errors.Register(ModuleName, 15, "failed to add coefficient commit")
	ErrCommitsNotCorrectLength = errors.Register(
		ModuleName,
		16,
		"coefficients commit not correct length",
	)
	ErrRound2InfoNotFound                    = errors.Register(ModuleName, 17, "round 2 info not found")
	ErrEncryptedSecretSharesNotCorrectLength = errors.Register(
		ModuleName,
		18,
		"encrypted secret shares not correct length",
	)
	ErrComputeOwnPubKeyFailed           = errors.Register(ModuleName, 19, "failed to compute own public key")
	ErrMemberIsAlreadyComplainOrConfirm = errors.Register(ModuleName, 20, "member is already complain or confirm")
	ErrEncryptedSecretShareNotFound     = errors.Register(ModuleName, 21, "encrypted secret share not found")
	ErrComplainFailed                   = errors.Register(ModuleName, 22, "failed to complain")
	ErrConfirmFailed                    = errors.Register(ModuleName, 23, "failed to confirm")
	ErrConfirmNotFound                  = errors.Register(ModuleName, 24, "confirm not found")
	ErrComplainsWithStatusNotFound      = errors.Register(ModuleName, 25, "complaints with status not found")
	ErrDENotFound                       = errors.Register(ModuleName, 26, "de not found")
	ErrGroupIsNotActive                 = errors.Register(ModuleName, 27, "group is not active")
	ErrUnexpectedThreshold              = errors.Register(ModuleName, 28, "threshold value is unexpected")
	ErrBadDrbgInitialization            = errors.Register(ModuleName, 29, "bad drbg initialization")
	ErrPartialSigNotFound               = errors.Register(ModuleName, 30, "partial sig not found")
	ErrInvalidArgument                  = errors.Register(ModuleName, 31, "invalid argument")
	ErrSigningNotFound                  = errors.Register(ModuleName, 32, "signing not found")
	ErrAlreadySigned                    = errors.Register(ModuleName, 33, "already signed")
	ErrSigningAlreadySuccess            = errors.Register(ModuleName, 34, "signing already success")
	ErrPubNonceNotEqualToSigR           = errors.Register(ModuleName, 35, "public nonce not equal to signature r")
	ErrMemberNotAssigned                = errors.Register(ModuleName, 36, "member is not assigned participants")
	ErrVerifySigningSigFailed           = errors.Register(ModuleName, 37, "failed to verify signing signature")
	ErrCombineSigsFailed                = errors.Register(ModuleName, 38, "failed to combine signatures")
	ErrVerifyGroupSigningSigFailed      = errors.Register(
		ModuleName,
		39,
		"faileded to verify group signing signature",
	)
	ErrDEQueueFull                     = errors.Register(ModuleName, 40, "de queue is full")
	ErrSigningExpired                  = errors.Register(ModuleName, 41, "signing expired")
	ErrTooSoonToActivate               = errors.Register(ModuleName, 42, "too soon to activate")
	ErrStatusNotFound                  = errors.Register(ModuleName, 43, "status not found")
	ErrStatusIsNotActive               = errors.Register(ModuleName, 44, "status is not active")
	ErrNotEnoughFee                    = errors.Register(ModuleName, 45, "not enough fee")
	ErrInvalidRequestSignatureContent  = errors.Register(ModuleName, 46, "invalid request signature content")
	ErrNoRequestSignatureHandlerExists = errors.Register(
		ModuleName,
		47,
		"no handler exists for request signature type",
	)
	ErrReplacementNotFound      = errors.Register(ModuleName, 48, "replacement group not found")
	ErrRequestReplacementFailed = errors.Register(ModuleName, 49, "failed to request replacement")
)
