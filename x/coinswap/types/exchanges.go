package types

import sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

func AddExchange(exchanges []Exchange, exchange Exchange) ([]Exchange, error) {
	index := IndexOfExchange(exchanges, exchange)
	if index != -1 {
		return nil, sdkerrors.Wrap(ErrExchangeAlreadyExist, "exchange already exists")
	}
	return append(exchanges, exchange), nil
}

func IndexOfExchange(exchanges []Exchange, exchange Exchange) int {
	for i, ex := range exchanges {
		if ex.To == exchange.To && ex.From == exchange.From {
			return i
		}
	}
	return -1
}

func RemoveExchange(exchanges []Exchange, exchange Exchange) ([]Exchange, error) {
	index := IndexOfExchange(exchanges, exchange)
	if index == -1 {
		return nil, sdkerrors.Wrap(ErrExchangeDoesNotExist, "failed to find exchange")
	}
	exchanges[index] = exchanges[len(exchanges)-1]
	return exchanges[:len(exchanges)-1], nil
}
