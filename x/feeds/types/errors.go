package types

import (
	"cosmossdk.io/errors"
)

// x/feeds module sentinel errors
var (
	ErrPriceNotFound           = errors.Register(ModuleName, 2, "price not found")
	ErrPriceServiceNotFound    = errors.Register(ModuleName, 3, "price-service not found")
	ErrOracleStatusNotActive   = errors.Register(ModuleName, 4, "oracle status not active")
	ErrPriceTooFast            = errors.Register(ModuleName, 5, "price is too fast")
	ErrInvalidTimestamp        = errors.Register(ModuleName, 6, "invalid timestamp")
	ErrNotEnoughPriceValidator = errors.Register(ModuleName, 7, "not enough price validator")
	ErrInvalidSigner           = errors.Register(ModuleName, 8, "expected admin to be signer")
	ErrNotTopValidator         = errors.Register(ModuleName, 9, "not top validator")
)
