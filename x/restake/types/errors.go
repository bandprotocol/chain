package types

import (
	errorsmod "cosmossdk.io/errors"
)

// x/restake module sentinel errors
var (
	ErrUnableToUndelegate      = errorsmod.Register(ModuleName, 2, "unable to undelegate")
	ErrVaultNotFound           = errorsmod.Register(ModuleName, 3, "vault not found")
	ErrVaultNotActive          = errorsmod.Register(ModuleName, 4, "vault not active")
	ErrVaultAlreadyDeactivated = errorsmod.Register(ModuleName, 5, "vault already deactivated")
	ErrLockNotFound            = errorsmod.Register(ModuleName, 6, "lock not found")
	ErrDelegationNotEnough     = errorsmod.Register(ModuleName, 7, "delegation not enough")
	ErrInvalidPower            = errorsmod.Register(ModuleName, 8, "invalid power")
	ErrTotalPowerZero          = errorsmod.Register(ModuleName, 9, "total power is zero")
	ErrAccountAlreadyExist     = errorsmod.Register(ModuleName, 10, "account already exist")
	ErrInvalidLength           = errorsmod.Register(ModuleName, 11, "invalid length")
)
