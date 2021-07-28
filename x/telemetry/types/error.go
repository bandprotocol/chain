package types

import sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

var (
	ErrInvalidDateInterval = sdkerrors.Register(ModuleName, 1, "Invalid Date interval")
)
