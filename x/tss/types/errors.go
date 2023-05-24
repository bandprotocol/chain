package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/tss module sentinel errors
var (
	ErrInvalidAccAddressFormat = sdkerrors.Register(ModuleName, 2, "account address format is invalid")
	ErrGroupNotFound           = sdkerrors.Register(ModuleName, 3, "group not found")
	ErrMemberNotFound          = sdkerrors.Register(ModuleName, 4, "member not found")
	ErrRound1DataNotFound      = sdkerrors.Register(ModuleName, 5, "round 1 data not found")
	ErrAlreadySubmitRound1     = sdkerrors.Register(ModuleName, 6, "already submit round1 message")
	ErrDKGContextNotFound      = sdkerrors.Register(ModuleName, 7, "dkg context not found")
	ErrMemberNotAuthorized     = sdkerrors.Register(
		ModuleName,
		8,
		"member is not authorized for this group",
	)
	ErrRoundExpired                          = sdkerrors.Register(ModuleName, 9, "round expired")
	ErrAlreadySubmit                         = sdkerrors.Register(ModuleName, 10, "member is already submit message")
	ErrVerifyOneTimeSigFailed                = sdkerrors.Register(ModuleName, 11, "fail to verify one time sign")
	ErrVerifyA0SigFailed                     = sdkerrors.Register(ModuleName, 12, "fail to verify a0 sign")
	ErrRound2DataNotFound                    = sdkerrors.Register(ModuleName, 13, "round 2 data not found")
	ErrEncryptedSecretSharesNotCorrectLength = sdkerrors.Register(
		ModuleName,
		14,
		"encrypted secret shares not correct length",
	)
)
