package types

import (
	errorsmod "cosmossdk.io/errors"
)

// x/feeds module sentinel errors
var (
	ErrPriceNotFound             = errorsmod.Register(ModuleName, 2, "price not found")
	ErrValidatorPriceNotFound    = errorsmod.Register(ModuleName, 3, "validator price not found")
	ErrFeedNotFound              = errorsmod.Register(ModuleName, 4, "feed not found")
	ErrOracleStatusNotActive     = errorsmod.Register(ModuleName, 5, "oracle status not active")
	ErrPriceTooFast              = errorsmod.Register(ModuleName, 6, "price is too fast")
	ErrInvalidTimestamp          = errorsmod.Register(ModuleName, 7, "invalid timestamp")
	ErrNotEnoughValidatorPrice   = errorsmod.Register(ModuleName, 8, "not enough validator price")
	ErrInvalidSigner             = errorsmod.Register(ModuleName, 9, "expected admin to be signer")
	ErrNotBondedValidator        = errorsmod.Register(ModuleName, 10, "not bonded validator")
	ErrNotEnoughDelegation       = errorsmod.Register(ModuleName, 11, "not enough delegation")
	ErrUnableToUndelegate        = errorsmod.Register(ModuleName, 12, "unable to undelegate")
	ErrInvalidWeightedPriceArray = errorsmod.Register(ModuleName, 13, "invalid weighted price array")
	ErrPowerNegative             = errorsmod.Register(ModuleName, 14, "power is negative")
)
