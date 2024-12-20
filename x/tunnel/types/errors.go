package types

import (
	errorsmod "cosmossdk.io/errors"
)

// x/tunnel module sentinel errors
var (
	ErrInvalidGenesis            = errorsmod.Register(ModuleName, 2, "invalid genesis")
	ErrMaxSignalsExceeded        = errorsmod.Register(ModuleName, 3, "max signals exceeded")
	ErrIntervalOutOfRange        = errorsmod.Register(ModuleName, 4, "interval out of range")
	ErrDeviationOutOfRange       = errorsmod.Register(ModuleName, 5, "deviation out of range")
	ErrTunnelNotFound            = errorsmod.Register(ModuleName, 6, "tunnel not found")
	ErrLatestPricesNotFound      = errorsmod.Register(ModuleName, 7, "latest prices not found")
	ErrPacketNotFound            = errorsmod.Register(ModuleName, 8, "packet not found")
	ErrNoPacketReceipt           = errorsmod.Register(ModuleName, 9, "no packet receipt")
	ErrInvalidTunnelCreator      = errorsmod.Register(ModuleName, 10, "invalid creator of the tunnel")
	ErrAccountAlreadyExist       = errorsmod.Register(ModuleName, 11, "account already exist")
	ErrInvalidRoute              = errorsmod.Register(ModuleName, 12, "invalid tunnel route")
	ErrInactiveTunnel            = errorsmod.Register(ModuleName, 13, "inactive tunnel")
	ErrAlreadyActive             = errorsmod.Register(ModuleName, 14, "already active")
	ErrAlreadyInactive           = errorsmod.Register(ModuleName, 15, "already inactive")
	ErrInvalidDepositDenom       = errorsmod.Register(ModuleName, 16, "invalid deposit denom")
	ErrDepositNotFound           = errorsmod.Register(ModuleName, 17, "deposit not found")
	ErrInsufficientDeposit       = errorsmod.Register(ModuleName, 18, "insufficient deposit")
	ErrInsufficientFund          = errorsmod.Register(ModuleName, 19, "insufficient fund")
	ErrDeviationNotFound         = errorsmod.Register(ModuleName, 20, "deviation not found")
	ErrInvalidVersion            = errorsmod.Register(ModuleName, 21, "invalid version")
	ErrChannelCapabilityNotFound = errorsmod.Register(ModuleName, 22, "channel capability not found")
	ErrMaxTunnelChannels         = errorsmod.Register(ModuleName, 23, "max tunnel channels exceeded")
	ErrInvalidEncoder            = errorsmod.Register(ModuleName, 24, "invalid encoder")
	ErrSendPacketPanic           = errorsmod.Register(ModuleName, 25, "panic in sending packet")
	ErrInvalidChannelID          = errorsmod.Register(ModuleName, 26, "invalid channel id")
	ErrInvalidPortID             = errorsmod.Register(ModuleName, 27, "invalid port id")
	ErrRouteNotReady             = errorsmod.Register(ModuleName, 28, "route is not ready")
)
