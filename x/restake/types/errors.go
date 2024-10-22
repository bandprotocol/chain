package types

import (
	errorsmod "cosmossdk.io/errors"
)

// x/restake module sentinel errors
var (
	ErrUnableToUndelegate     = errorsmod.Register(ModuleName, 2, "unable to undelegate")
	ErrVaultNotFound          = errorsmod.Register(ModuleName, 3, "vault not found")
	ErrVaultNotActive         = errorsmod.Register(ModuleName, 4, "vault not active")
	ErrLockNotFound           = errorsmod.Register(ModuleName, 5, "lock not found")
	ErrPowerNotEnough         = errorsmod.Register(ModuleName, 6, "power not enough")
	ErrInvalidPower           = errorsmod.Register(ModuleName, 7, "invalid power")
	ErrTotalPowerZero         = errorsmod.Register(ModuleName, 8, "total power is zero")
	ErrAccountAlreadyExist    = errorsmod.Register(ModuleName, 9, "account already exist")
	ErrInvalidLength          = errorsmod.Register(ModuleName, 10, "invalid length")
	ErrStakeNotEnough         = errorsmod.Register(ModuleName, 11, "stake not enough")
	ErrNotAllowedDenom        = errorsmod.Register(ModuleName, 12, "not allowed denom")
	ErrUnableToUnstake        = errorsmod.Register(ModuleName, 13, "unable to unstake")
	ErrLiquidStakerNotAllowed = errorsmod.Register(ModuleName, 14, "liquid staker not allowed")
)
