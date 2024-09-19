package types

import (
	errorsmod "cosmossdk.io/errors"
)

// x/tunnel module sentinel errors
var (
	ErrInvalidGenesis           = errorsmod.Register(ModuleName, 2, "invalid genesis")
	ErrMaxSignalsExceeded       = errorsmod.Register(ModuleName, 3, "max signals exceeded")
	ErrIntervalTooLow           = errorsmod.Register(ModuleName, 4, "interval too low")
	ErrTunnelNotFound           = errorsmod.Register(ModuleName, 5, "tunnel not found")
	ErrSignalPricesInfoNotFound = errorsmod.Register(ModuleName, 6, "signal prices info not found")
	ErrPacketNotFound           = errorsmod.Register(ModuleName, 7, "packet not found")
	ErrNoPacketContent          = errorsmod.Register(ModuleName, 8, "no packet content")
	ErrInvalidTunnelCreator     = errorsmod.Register(ModuleName, 9, "invalid creator of the tunnel")
	ErrAccountAlreadyExist      = errorsmod.Register(ModuleName, 10, "account already exist")
	ErrInvalidRoute             = errorsmod.Register(ModuleName, 11, "invalid tunnel route")
	ErrInactiveTunnel           = errorsmod.Register(ModuleName, 12, "inactive tunnel")
	ErrAlreadyActive            = errorsmod.Register(ModuleName, 13, "already active")
	ErrAlreadyInactive          = errorsmod.Register(ModuleName, 14, "already inactive")
)
