package types

import (
	errorsmod "cosmossdk.io/errors"
)

// x/tunnel module sentinel errors
var (
	ErrMaxTunnelChannels = errorsmod.Register(ModuleName, 2, "max tunnel channels")
)
