package types

import (
	errorsmod "cosmossdk.io/errors"
)

// x/restake module sentinel errors
var (
	ErrUnableToUndelegate    = errorsmod.Register(ModuleName, 2, "unable to undelegate")
	ErrKeyNotFound           = errorsmod.Register(ModuleName, 3, "key not found")
	ErrKeyNotActive          = errorsmod.Register(ModuleName, 4, "key not active")
	ErrKeyAlreadyDeactivated = errorsmod.Register(ModuleName, 5, "key already deactivated")
	ErrStakeNotFound         = errorsmod.Register(ModuleName, 6, "stake not found")
	ErrDelegationNotEnough   = errorsmod.Register(ModuleName, 7, "delegation not enough")
	ErrInvalidAmount         = errorsmod.Register(ModuleName, 8, "invalid amount")
	ErrRewardNotFound        = errorsmod.Register(ModuleName, 9, "reward not found")
	ErrTotalLockZero         = errorsmod.Register(ModuleName, 10, "total lock is zero")
	ErrAccountAlreadyExist   = errorsmod.Register(ModuleName, 11, "account already exist")
)
