package types

import (
	errorsmod "cosmossdk.io/errors"
)

// x/feeds module sentinel errors
var (
	ErrPriceNotFound           = errorsmod.Register(ModuleName, 2, "price not found")
	ErrPriceValidatorNotFound  = errorsmod.Register(ModuleName, 3, "price validator not found")
	ErrSymbolNotFound          = errorsmod.Register(ModuleName, 4, "symbol not found")
	ErrPriceServiceNotFound    = errorsmod.Register(ModuleName, 5, "price-service not found")
	ErrOracleStatusNotActive   = errorsmod.Register(ModuleName, 6, "oracle status not active")
	ErrPriceTooFast            = errorsmod.Register(ModuleName, 7, "price is too fast")
	ErrInvalidTimestamp        = errorsmod.Register(ModuleName, 8, "invalid timestamp")
	ErrNotEnoughPriceValidator = errorsmod.Register(ModuleName, 9, "not enough price validator")
	ErrInvalidSigner           = errorsmod.Register(ModuleName, 10, "expected admin to be signer")
	ErrNotTopValidator         = errorsmod.Register(ModuleName, 11, "not top validator")
	ErrNotEnoughDelegation     = errorsmod.Register(ModuleName, 12, "not enough delegation")
	ErrSymbolPowerNotFound     = errorsmod.Register(ModuleName, 13, "symbol power not found")
	ErrUnableToUndelegate      = errorsmod.Register(ModuleName, 14, "unable to undelegate")
)
