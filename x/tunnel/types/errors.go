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
	ErrNoRoute                   = errorsmod.Register(ModuleName, 7, "no route")
	ErrNoReceipt                 = errorsmod.Register(ModuleName, 8, "no receipt")
	ErrLatestPricesNotFound      = errorsmod.Register(ModuleName, 9, "latest prices not found")
	ErrPacketNotFound            = errorsmod.Register(ModuleName, 10, "packet not found")
	ErrNoPacketReceipt           = errorsmod.Register(ModuleName, 11, "no packet receipt")
	ErrInvalidTunnelCreator      = errorsmod.Register(ModuleName, 12, "invalid creator of the tunnel")
	ErrAccountAlreadyExist       = errorsmod.Register(ModuleName, 13, "account already exist")
	ErrInvalidRoute              = errorsmod.Register(ModuleName, 14, "invalid tunnel route")
	ErrInactiveTunnel            = errorsmod.Register(ModuleName, 15, "inactive tunnel")
	ErrAlreadyActive             = errorsmod.Register(ModuleName, 16, "already active")
	ErrAlreadyInactive           = errorsmod.Register(ModuleName, 17, "already inactive")
	ErrInvalidDepositDenom       = errorsmod.Register(ModuleName, 18, "invalid deposit denom")
	ErrDepositNotFound           = errorsmod.Register(ModuleName, 19, "deposit not found")
	ErrInsufficientDeposit       = errorsmod.Register(ModuleName, 20, "insufficient deposit")
	ErrInsufficientFund          = errorsmod.Register(ModuleName, 21, "insufficient fund")
	ErrDeviationNotFound         = errorsmod.Register(ModuleName, 22, "deviation not found")
	ErrInvalidVersion            = errorsmod.Register(ModuleName, 23, "invalid ICS20 version")
	ErrChannelCapabilityNotFound = errorsmod.Register(ModuleName, 24, "channel capability not found")
)
