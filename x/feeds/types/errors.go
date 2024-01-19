package types

import (
	"cosmossdk.io/errors"
)

// x/feed module sentinel errors
var (
	ErrPriceNotFound           = errors.Register(ModuleName, 2, "price not found")
	ErrOracleStatusNotActive   = errors.Register(ModuleName, 3, "oracle status not active")
	ErrTimestampOlder          = errors.Register(ModuleName, 4, "timestamp order")
	ErrInvalidTimestamp        = errors.Register(ModuleName, 5, "invalid timestamp")
	ErrNotEnoughPriceValidator = errors.Register(ModuleName, 6, "not enough price validator")
	ErrInvalidSigner           = errors.Register(ModuleName, 7, "expected admin to be signer")
)
