package common

import (
	errorsmod "cosmossdk.io/errors"
	abci "github.com/cometbft/cometbft/abci/types"
)

func QueryResultError(err error) abci.ResponseQuery {
	space, code, log := errorsmod.ABCIInfo(err, true)
	return abci.ResponseQuery{
		Code:      code,
		Codespace: space,
		Log:       log,
	}
}

func QueryResultSuccess(value []byte, height int64) abci.ResponseQuery {
	space, code, log := errorsmod.ABCIInfo(nil, true)
	return abci.ResponseQuery{
		Code:      code,
		Codespace: space,
		Log:       log,
		Height:    height,
		Value:     value,
	}
}
