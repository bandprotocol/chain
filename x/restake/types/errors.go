package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/restake module sentinel errors
var (
	ErrUnableToUndelegate = sdkerrors.Register(ModuleName, 2, "unable to undelegate")
)
