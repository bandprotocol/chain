package types

import errorsmod "cosmossdk.io/errors"

// x/tss module sentinel errors
var (
	ErrGroupSizeTooLarge   = errorsmod.Register(ModuleName, 2, "group size is too large")
	ErrGroupNotFound       = errorsmod.Register(ModuleName, 3, "group not found")
	ErrMemberNotFound      = errorsmod.Register(ModuleName, 4, "member not found")
	ErrNoActiveMember      = errorsmod.Register(ModuleName, 5, "no active member in this group")
	ErrMemberAlreadySubmit = errorsmod.Register(ModuleName, 6, "member is already submit message")
	ErrRound1InfoNotFound  = errorsmod.Register(ModuleName, 7, "round 1 info not found")
	ErrDKGContextNotFound  = errorsmod.Register(ModuleName, 8, "dkg context not found")
	ErrMemberNotAuthorized = errorsmod.Register(
		ModuleName,
		9,
		"member is not authorized for this group",
	)
	ErrInvalidStatus                = errorsmod.Register(ModuleName, 10, "invalid status")
	ErrVerifyOneTimeSignatureFailed = errorsmod.Register(ModuleName, 11, "failed to verify one time sign")
	ErrVerifyA0SignatureFailed      = errorsmod.Register(ModuleName, 12, "failed to verify a0 sign")
	ErrAddCoeffCommit               = errorsmod.Register(ModuleName, 13, "failed to add coefficient commit")
	ErrInvalidLengthCoeffCommits    = errorsmod.Register(
		ModuleName,
		14,
		"coefficients commit length is invalid",
	)
	ErrRound2InfoNotFound                 = errorsmod.Register(ModuleName, 15, "round 2 info not found")
	ErrInvalidLengthEncryptedSecretShares = errorsmod.Register(
		ModuleName,
		16,
		"encrypted secret shares length is invalid",
	)
	ErrComputeOwnPubKeyFailed        = errorsmod.Register(ModuleName, 17, "failed to compute own public key")
	ErrComplainFailed                = errorsmod.Register(ModuleName, 18, "failed to complain")
	ErrConfirmFailed                 = errorsmod.Register(ModuleName, 19, "failed to confirm")
	ErrConfirmNotFound               = errorsmod.Register(ModuleName, 20, "confirm not found")
	ErrComplaintsWithStatusNotFound  = errorsmod.Register(ModuleName, 21, "complaints with status not found")
	ErrDENotFound                    = errorsmod.Register(ModuleName, 22, "DE not found")
	ErrGroupIsNotActive              = errorsmod.Register(ModuleName, 23, "group is not active")
	ErrInsufficientActiveMembers     = errorsmod.Register(ModuleName, 24, "insufficient active members for signing")
	ErrBadDrbgInitialization         = errorsmod.Register(ModuleName, 25, "bad drbg initialization")
	ErrPartialSignatureNotFound      = errorsmod.Register(ModuleName, 26, "partial signature not found")
	ErrInvalidArgument               = errorsmod.Register(ModuleName, 27, "invalid argument")
	ErrSigningNotFound               = errorsmod.Register(ModuleName, 28, "signing not found")
	ErrAlreadySigned                 = errorsmod.Register(ModuleName, 29, "already signed")
	ErrSigningAlreadySuccess         = errorsmod.Register(ModuleName, 30, "signing already success")
	ErrPubNonceNotEqualToSigR        = errorsmod.Register(ModuleName, 31, "public nonce not equal to signature r")
	ErrMemberNotAssigned             = errorsmod.Register(ModuleName, 32, "member is not assigned")
	ErrVerifySigningSigFailed        = errorsmod.Register(ModuleName, 33, "failed to verify signing signature")
	ErrDEReachMaxLimit               = errorsmod.Register(ModuleName, 34, "the number of existing DEs reach the limit")
	ErrHandleSignatureOrderFailed    = errorsmod.Register(ModuleName, 35, "failed to handle signature order")
	ErrNoSignatureOrderHandlerExists = errorsmod.Register(ModuleName, 36, "no handler exists for signature order type")
	ErrInvalidCoefficientCommit      = errorsmod.Register(ModuleName, 37, "invalid coefficient commit")
	ErrInvalidPublicKey              = errorsmod.Register(ModuleName, 38, "invalid public key")
	ErrInvalidSignature              = errorsmod.Register(ModuleName, 39, "invalid signature")
	ErrInvalidSecretShare            = errorsmod.Register(ModuleName, 40, "invalid secret share")
	ErrInvalidComplaint              = errorsmod.Register(ModuleName, 41, "invalid complaint")
	ErrInvalidSymmetricKey           = errorsmod.Register(ModuleName, 42, "invalid symmetric key")
	ErrInvalidDE                     = errorsmod.Register(ModuleName, 43, "invalid DE")
	ErrMaxSigningAttemptReached      = errorsmod.Register(ModuleName, 44, "max signing attempt reached")
	ErrEncodeOriginatorFailed        = errorsmod.Register(ModuleName, 45, "failed to encode originator")
	ErrInvalidMemo                   = errorsmod.Register(ModuleName, 46, "invalid memo")
	ErrInvalidMessage                = errorsmod.Register(ModuleName, 47, "invalid message")
	ErrSigningAttemptNotFound        = errorsmod.Register(ModuleName, 48, "signing attempt not found")
	ErrInvalidMember                 = errorsmod.Register(ModuleName, 49, "invalid member")
)
