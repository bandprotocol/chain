package types

import errorsmod "cosmossdk.io/errors"

// x/tss module sentinel errors
var (
	ErrInvalidAccAddressFormat = errorsmod.Register(ModuleName, 2, "account address format is invalid")
	ErrGroupSizeTooLarge       = errorsmod.Register(ModuleName, 3, "group size is too large")
	ErrGroupNotFound           = errorsmod.Register(ModuleName, 4, "group not found")
	ErrMemberNotFound          = errorsmod.Register(ModuleName, 5, "member not found")
	ErrNoActiveMember          = errorsmod.Register(ModuleName, 6, "no active member in this group")
	ErrMemberAlreadySubmit     = errorsmod.Register(ModuleName, 7, "member is already submit message")
	ErrRound1InfoNotFound      = errorsmod.Register(ModuleName, 8, "round 1 info not found")
	ErrDKGContextNotFound      = errorsmod.Register(ModuleName, 9, "dkg context not found")
	ErrMemberNotAuthorized     = errorsmod.Register(
		ModuleName,
		10,
		"member is not authorized for this group",
	)
	ErrInvalidStatus                = errorsmod.Register(ModuleName, 11, "invalid status")
	ErrVerifyOneTimeSignatureFailed = errorsmod.Register(ModuleName, 12, "failed to verify one time sign")
	ErrVerifyA0SignatureFailed      = errorsmod.Register(ModuleName, 13, "failed to verify a0 sign")
	ErrAddCoeffCommit               = errorsmod.Register(ModuleName, 14, "failed to add coefficient commit")
	ErrInvalidLengthCoefCommits     = errorsmod.Register(
		ModuleName,
		15,
		"coefficients commit length is invalid",
	)
	ErrRound2InfoNotFound                 = errorsmod.Register(ModuleName, 16, "round 2 info not found")
	ErrInvalidLengthEncryptedSecretShares = errorsmod.Register(
		ModuleName,
		17,
		"encrypted secret shares length is invalid ",
	)
	ErrComputeOwnPubKeyFailed           = errorsmod.Register(ModuleName, 18, "failed to compute own public key")
	ErrMemberIsAlreadyComplainOrConfirm = errorsmod.Register(ModuleName, 19, "member is already complain or confirm")
	ErrComplainFailed                   = errorsmod.Register(ModuleName, 20, "failed to complain")
	ErrConfirmFailed                    = errorsmod.Register(ModuleName, 21, "failed to confirm")
	ErrConfirmNotFound                  = errorsmod.Register(ModuleName, 22, "confirm not found")
	ErrComplaintsWithStatusNotFound     = errorsmod.Register(ModuleName, 23, "complaints with status not found")
	ErrDENotFound                       = errorsmod.Register(ModuleName, 24, "de not found")
	ErrGroupIsNotActive                 = errorsmod.Register(ModuleName, 25, "group is not active")
	ErrUnexpectedThreshold              = errorsmod.Register(ModuleName, 26, "threshold value is unexpected")
	ErrBadDrbgInitialization            = errorsmod.Register(ModuleName, 27, "bad drbg initialization")
	ErrPartialSignatureNotFound         = errorsmod.Register(ModuleName, 28, "partial signature not found")
	ErrInvalidArgument                  = errorsmod.Register(ModuleName, 29, "invalid argument")
	ErrSigningNotFound                  = errorsmod.Register(ModuleName, 30, "signing not found")
	ErrAlreadySigned                    = errorsmod.Register(ModuleName, 31, "already signed")
	ErrSigningAlreadySuccess            = errorsmod.Register(ModuleName, 32, "signing already success")
	ErrPubNonceNotEqualToSigR           = errorsmod.Register(ModuleName, 33, "public nonce not equal to signature r")
	ErrMemberNotAssigned                = errorsmod.Register(ModuleName, 34, "member is not assigned")
	ErrVerifySigningSigFailed           = errorsmod.Register(ModuleName, 35, "failed to verify signing signature")
	ErrDEQueueFull                      = errorsmod.Register(ModuleName, 36, "de queue is full")
	ErrHandleSignatureOrderFailed       = errorsmod.Register(ModuleName, 37, "failed to handle signature order")
	ErrNoSignatureOrderHandlerExists    = errorsmod.Register(ModuleName, 38, "no handler exists for signature order type")
	ErrInvalidCoefficientCommit         = errorsmod.Register(ModuleName, 39, "invalid coefficient commit")
	ErrInvalidPublicKey                 = errorsmod.Register(ModuleName, 40, "invalid public key")
	ErrInvalidSignature                 = errorsmod.Register(ModuleName, 41, "invalid signature")
	ErrInvalidSecretShare               = errorsmod.Register(ModuleName, 42, "invalid secret share")
	ErrInvalidComplaint                 = errorsmod.Register(ModuleName, 43, "invalid complaint")
	ErrInvalidSymmetricKey              = errorsmod.Register(ModuleName, 44, "invalid symmetric key")
	ErrInvalidDE                        = errorsmod.Register(ModuleName, 45, "invalid de")
)
