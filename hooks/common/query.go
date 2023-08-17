package common

import (
	abci "github.com/cometbft/cometbft/abci/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func QueryResultError(err error) abci.ResponseQuery {
	space, code, log := sdkerrors.ABCIInfo(err, true)
	return abci.ResponseQuery{
		Code:      code,
		Codespace: space,
		Log:       log,
	}
}

func QueryResultSuccess(value []byte, height int64) abci.ResponseQuery {
	space, code, log := sdkerrors.ABCIInfo(nil, true)
	return abci.ResponseQuery{
		Code:      code,
		Codespace: space,
		Log:       log,
		Height:    height,
		Value:     value,
	}
}
