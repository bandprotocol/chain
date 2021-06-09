package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	ErrInvalidExchangeDenom   = sdkerrors.New(ModuleName, 111, "unsupported exchange denomination")
	ErrExchangeDenomMissmatch = sdkerrors.New(ModuleName, 112, "exchange denomination does not match the amount provided")
)
