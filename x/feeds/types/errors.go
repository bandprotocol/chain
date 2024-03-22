package types

import (
	errorsmod "cosmossdk.io/errors"
)

// x/feeds module sentinel errors
var (
	ErrPriceNotFound             = errorsmod.Register(ModuleName, 1, "price not found")
	ErrPriceValidatorNotFound    = errorsmod.Register(ModuleName, 2, "price validator not found")
	ErrFeedNotFound              = errorsmod.Register(ModuleName, 3, "feed not found")
	ErrOracleStatusNotActive     = errorsmod.Register(ModuleName, 4, "oracle status not active")
	ErrPriceTooFast              = errorsmod.Register(ModuleName, 5, "price is too fast")
	ErrInvalidTimestamp          = errorsmod.Register(ModuleName, 6, "invalid timestamp")
	ErrNotEnoughPriceValidator   = errorsmod.Register(ModuleName, 7, "not enough price validator")
	ErrInvalidSigner             = errorsmod.Register(ModuleName, 8, "expected admin to be signer")
	ErrNotTopValidator           = errorsmod.Register(ModuleName, 9, "not top validator")
	ErrNotEnoughDelegation       = errorsmod.Register(ModuleName, 10, "not enough delegation")
	ErrUnableToUndelegate        = errorsmod.Register(ModuleName, 11, "unable to undelegate")
	ErrInvalidWeightedPriceArray = errorsmod.Register(ModuleName, 12, "invalid weighted price array")
)
