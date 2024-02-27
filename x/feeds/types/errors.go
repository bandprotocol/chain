package types

import (
	"cosmossdk.io/errors"
)

// x/feeds module sentinel errors
var (
	ErrPriceNotFound           = errors.Register(ModuleName, 2, "price not found")
	ErrPriceValidatorNotFound  = errors.Register(ModuleName, 3, "price validator not found")
	ErrSymbolNotFound          = errors.Register(ModuleName, 4, "symbol not found")
	ErrPriceServiceNotFound    = errors.Register(ModuleName, 5, "price-service not found")
	ErrOracleStatusNotActive   = errors.Register(ModuleName, 6, "oracle status not active")
	ErrPriceTooFast            = errors.Register(ModuleName, 7, "price is too fast")
	ErrInvalidTimestamp        = errors.Register(ModuleName, 8, "invalid timestamp")
	ErrNotEnoughPriceValidator = errors.Register(ModuleName, 9, "not enough price validator")
	ErrInvalidSigner           = errors.Register(ModuleName, 10, "expected admin to be signer")
	ErrNotTopValidator         = errors.Register(ModuleName, 11, "not top validator")
	ErrNotEnoughDelegation     = errors.Register(ModuleName, 12, "not enough delegation")
	ErrSymbolPowerNotFound     = errors.Register(ModuleName, 13, "symbol power not found")
	ErrUnableToUndelegate      = errors.Register(ModuleName, 14, "unable to undelegate")
)
