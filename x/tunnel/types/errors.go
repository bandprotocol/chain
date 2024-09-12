package types

import (
	errorsmod "cosmossdk.io/errors"
)

// x/tunnel module sentinel errors
var (
	ErrInvalidGenesis           = errorsmod.Register(ModuleName, 2, "invalid genesis")
	ErrMaxSignalsExceeded       = errorsmod.Register(ModuleName, 3, "max signals exceeded")
	ErrMinIntervalExceeded      = errorsmod.Register(ModuleName, 4, "min interval exceeded")
	ErrTunnelNotFound           = errorsmod.Register(ModuleName, 5, "tunnel not found")
	ErrActiveTunnelIDsNotFound  = errorsmod.Register(ModuleName, 6, "active tunnel IDs not found")
	ErrSignalPricesInfoNotFound = errorsmod.Register(ModuleName, 7, "signal prices info not found")
	ErrPacketNotFound           = errorsmod.Register(ModuleName, 8, "packet not found")
	ErrNoPacketContent          = errorsmod.Register(ModuleName, 9, "no packet content")
	ErrInvalidTunnelCreator     = errorsmod.Register(ModuleName, 10, "invalid creator of tunnel")
	ErrAccountAlreadyExist      = errorsmod.Register(ModuleName, 11, "account already exist")
	ErrInvalidRoute             = errorsmod.Register(ModuleName, 12, "invalid tunnel route")
	ErrInactiveTunnel           = errorsmod.Register(ModuleName, 13, "inactive tunnel")
	ErrMinDepositTooSmall       = errorsmod.Register(ModuleName, 14, "min deposit too small")
	ErrInvalidDepositDenom      = errorsmod.Register(ModuleName, 15, "invalid deposit denom")
	ErrDepositNotFound          = errorsmod.Register(ModuleName, 16, "deposit not found")
	ErrInsufficientFunds        = errorsmod.Register(ModuleName, 17, "insufficient funds")
)
