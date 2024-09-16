package types

import (
	errorsmod "cosmossdk.io/errors"
)

// x/tunnel module sentinel errors
var (
	ErrInvalidGenesis           = errorsmod.Register(ModuleName, 2, "invalid genesis")
	ErrMaxTunnelChannels        = errorsmod.Register(ModuleName, 3, "max tunnel channels")
	ErrMaxSignalsExceeded       = errorsmod.Register(ModuleName, 4, "max signals exceeded")
	ErrMinIntervalExceeded      = errorsmod.Register(ModuleName, 5, "min interval exceeded")
	ErrTunnelNotFound           = errorsmod.Register(ModuleName, 6, "tunnel not found")
	ErrActiveTunnelIDsNotFound  = errorsmod.Register(ModuleName, 7, "active tunnel IDs not found")
	ErrSignalPricesInfoNotFound = errorsmod.Register(ModuleName, 8, "signal prices info not found")
	ErrPacketNotFound           = errorsmod.Register(ModuleName, 9, "packet not found")
	ErrNoPacketContent          = errorsmod.Register(ModuleName, 10, "no packet content")
	ErrInvalidTunnelCreator     = errorsmod.Register(ModuleName, 11, "invalid creator of tunnel")
	ErrAccountAlreadyExist      = errorsmod.Register(ModuleName, 12, "account already exist")
	ErrInvalidRoute             = errorsmod.Register(ModuleName, 13, "invalid tunnel route")
	ErrInactiveTunnel           = errorsmod.Register(ModuleName, 14, "inactive tunnel")
)
