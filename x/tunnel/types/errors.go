package types

import (
	errorsmod "cosmossdk.io/errors"
)

// x/tunnel module sentinel errors
var (
	ErrInvalidGenesis           = errorsmod.Register(ModuleName, 2, "invalid genesis")
	ErrMaxTunnelChannels        = errorsmod.Register(ModuleName, 3, "max tunnel channels")
	ErrTunnelNotFound           = errorsmod.Register(ModuleName, 4, "tunnel not found")
	ErrSignalPricesInfoNotFound = errorsmod.Register(ModuleName, 5, "signal prices info not found")
	ErrPacketNotFound           = errorsmod.Register(ModuleName, 6, "packet not found")
	ErrNoPacketContent          = errorsmod.Register(ModuleName, 7, "no packet content")
	ErrInvalidTunnelCreator     = errorsmod.Register(ModuleName, 8, "invalid creator of tunnel")
	ErrAccountAlreadyExist      = errorsmod.Register(ModuleName, 9, "account already exist")
)
