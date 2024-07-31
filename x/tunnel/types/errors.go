package types

import (
	errorsmod "cosmossdk.io/errors"
)

// x/tunnel module sentinel errors
var (
	ErrMaxTunnelChannels   = errorsmod.Register(ModuleName, 2, "max tunnel channels")
	ErrTunnelNotFound      = errorsmod.Register(ModuleName, 3, "tunnel not found")
	ErrPacketNotFound      = errorsmod.Register(ModuleName, 4, "packet not found")
	ErrAccountAlreadyExist = errorsmod.Register(ModuleName, 6, "account already exist")
)
