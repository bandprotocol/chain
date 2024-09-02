package types

import (
	errorsmod "cosmossdk.io/errors"
)

// x/tunnel module sentinel errors
var (
	ErrInvalidGenesis           = errorsmod.Register(ModuleName, 2, "invalid genesis")
	ErrMaxSignalsExceeded       = errorsmod.Register(ModuleName, 3, "max signals exceeded")
	ErrMinIntervalExceeded      = errorsmod.Register(ModuleName, 4, "min interval exceeded")
	ErrMaxTunnelChannels        = errorsmod.Register(ModuleName, 5, "max tunnel channels")
	ErrTunnelNotFound           = errorsmod.Register(ModuleName, 6, "tunnel not found")
	ErrActiveTunnelIDsNotFound  = errorsmod.Register(ModuleName, 7, "active tunnel IDs not found")
	ErrSignalPricesInfoNotFound = errorsmod.Register(ModuleName, 8, "signal prices info not found")
	ErrPacketNotFound           = errorsmod.Register(ModuleName, 9, "packet not found")
	ErrNoPacketContent          = errorsmod.Register(ModuleName, 10, "no packet content")
	ErrInvalidTunnelCreator     = errorsmod.Register(ModuleName, 11, "invalid creator of tunnel")
	ErrAccountAlreadyExist      = errorsmod.Register(ModuleName, 12, "account already exist")
	ErrTunnelAlreadyActive      = errorsmod.Register(ModuleName, 13, "tunnel already active")
	ErrTunnelNotActive          = errorsmod.Register(ModuleName, 14, "tunnel not active")
)
