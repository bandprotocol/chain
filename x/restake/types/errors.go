package types

import (
	errorsmod "cosmossdk.io/errors"
)

// x/restake module sentinel errors
var (
	ErrUnableToUndelegate  = errorsmod.Register(ModuleName, 2, "unable to undelegate")
	ErrKeyNotFound         = errorsmod.Register(ModuleName, 3, "key not found")
	ErrLockNotFound        = errorsmod.Register(ModuleName, 4, "lock not found")
	ErrDelegationNotEnough = errorsmod.Register(ModuleName, 5, "delegation not enough")
)
