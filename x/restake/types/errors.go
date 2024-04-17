package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/restake module sentinel errors
var (
	ErrEmptyValidatorAddr = sdkerrors.Register(ModuleName, 2, "empty validator address")
)
