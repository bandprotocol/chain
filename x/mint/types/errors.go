package types

import sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

var (
	ErrInvalidMintDenom                     = sdkerrors.Register(ModuleName, 121, "The given mint denom is invalid")
	ErrAccountIsNotEligible                 = sdkerrors.Register(ModuleName, 122, "The given account is not eligible to mint")
	ErrInvalidWithdrawalAmount              = sdkerrors.Register(ModuleName, 123, "The given withdrawal amount is invalid")
	ErrExceedsWithdrawalLimitPerTime        = sdkerrors.Register(ModuleName, 124, "The given amount exceeds the withdrawal limit per time")
	ErrWithdrawalAmountExceedsModuleBalance = sdkerrors.Register(ModuleName, 125, "The given amount to withdraw exceeds module balance")
)
