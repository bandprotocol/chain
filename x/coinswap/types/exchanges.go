package types

import sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

func AddExchange(exchanges []Exchange, exchange Exchange) ([]Exchange, error) {
	index := IndexOfExchange(exchanges, exchange)
	if index != -1 {
		return nil, sdkerrors.Wrap(ErrExchangeAlreadyExist, "failed to add new exchange")
	}
	return append(exchanges, exchange), nil
}

func RemoveExchange(exchanges []Exchange, exchange Exchange) ([]Exchange, error) {
	index := IndexOfExchange(exchanges, exchange)
	if index == -1 {
		return nil, sdkerrors.Wrap(ErrExchangeDoesNotExist, "failed to find exchange")
	}
	newExchanges := make([]Exchange, 0, len(exchanges)-1)
	for _, ex := range exchanges {
		if ex != exchange {
			newExchanges = append(newExchanges, ex)
		}
	}
	return newExchanges, nil
}

func IndexOfExchange(exchanges []Exchange, exchange Exchange) int {
	for i, ex := range exchanges {
		if ex == exchange {
			return i
		}
	}
	return -1
}
