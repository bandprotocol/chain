package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewValidatorPrice creates new ValidatorPrice.
func NewValidatorPrice(
	val sdk.ValAddress,
	price SignalPrice,
	blockTime int64,
	blockHeight int64,
) ValidatorPrice {
	return ValidatorPrice{
		PriceStatus: price.PriceStatus,
		Validator:   val.String(),
		SignalID:    price.SignalID,
		Price:       price.Price,
		Timestamp:   blockTime,
		BlockHeight: blockHeight,
	}
}
