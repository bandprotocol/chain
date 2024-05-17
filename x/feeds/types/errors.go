package types

import (
	errorsmod "cosmossdk.io/errors"
)

// x/feeds module sentinel errors
var (
	ErrPriceNotFound             = errorsmod.Register(ModuleName, 2, "price not found")
	ErrValidatorPriceNotFound    = errorsmod.Register(ModuleName, 3, "validator price not found")
	ErrOracleStatusNotActive     = errorsmod.Register(ModuleName, 4, "oracle status not active")
	ErrPriceSubmitTooEarly       = errorsmod.Register(ModuleName, 5, "price is submitted too early")
	ErrInvalidTimestamp          = errorsmod.Register(ModuleName, 6, "invalid timestamp")
	ErrNotEnoughValidatorPrice   = errorsmod.Register(ModuleName, 7, "not enough validator price")
	ErrInvalidSigner             = errorsmod.Register(ModuleName, 8, "invalid signer")
	ErrNotBondedValidator        = errorsmod.Register(ModuleName, 9, "not bonded validator")
	ErrNotEnoughDelegation       = errorsmod.Register(ModuleName, 10, "not enough delegation")
	ErrUnableToUndelegate        = errorsmod.Register(ModuleName, 11, "unable to undelegate")
	ErrInvalidWeightedPriceArray = errorsmod.Register(ModuleName, 12, "invalid weighted price array")
	ErrPowerNegative             = errorsmod.Register(ModuleName, 13, "power is negative")
	ErrSignalIDNotSupported      = errorsmod.Register(ModuleName, 14, "signal id is not supported")
	ErrSubmitPricesTooLarge      = errorsmod.Register(ModuleName, 15, "submit prices list is too large")
	ErrSignalIDTooLarge          = errorsmod.Register(ModuleName, 16, "signal id is too large")
	ErrSignalTotalPowerNotFound  = errorsmod.Register(ModuleName, 17, "signal-total-power not found")
	ErrInvalidSignal             = errorsmod.Register(ModuleName, 18, "invalid signal")
)
