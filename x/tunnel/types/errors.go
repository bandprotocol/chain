package types

import (
	errorsmod "cosmossdk.io/errors"
)

// x/tunnel module sentinel errors
var (
	ErrInvalidGenesis      = errorsmod.Register(ModuleName, 2, "invalid genesis")
	ErrMaxTunnelChannels   = errorsmod.Register(ModuleName, 3, "max tunnel channels")
	ErrTunnelNotFound      = errorsmod.Register(ModuleName, 4, "tunnel not found")
	ErrPacketNotFound      = errorsmod.Register(ModuleName, 5, "packet not found")
	ErrAccountAlreadyExist = errorsmod.Register(ModuleName, 6, "account already exist")
)
