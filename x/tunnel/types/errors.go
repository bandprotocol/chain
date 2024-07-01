package types

import (
	errorsmod "cosmossdk.io/errors"
)

// x/tunnel module sentinel errors
var (
	ErrMaxTunnelChannels    = errorsmod.Register(ModuleName, 2, "max tunnel channels")
	ErrTunnelNotFound       = errorsmod.Register(ModuleName, 3, "tunnel not found")
	ErrTSSPacketNotFound    = errorsmod.Register(ModuleName, 4, "tss packet not found")
	ErrAxelarPacketNotFound = errorsmod.Register(ModuleName, 5, "axelar packet not found")
)
