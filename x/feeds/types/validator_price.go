package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewValidatorPrice creates new ValidatorPrice.
func NewValidatorPrice(
	val sdk.ValAddress,
	signalPrice SignalPrice,
	blockTime int64,
	blockHeight int64,
) ValidatorPrice {
	return ValidatorPrice{
		SignalPriceStatus: signalPrice.Status,
		Validator:         val.String(),
		SignalID:          signalPrice.SignalID,
		Price:             signalPrice.Price,
		Timestamp:         blockTime,
		BlockHeight:       blockHeight,
	}
}

// NewValidatorPriceList creates new ValidatorPriceList.
func NewValidatorPriceList(
	val sdk.ValAddress,
	prices []ValidatorPrice,
) ValidatorPriceList {
	return ValidatorPriceList{
		Validator:       val.String(),
		ValidatorPrices: prices,
	}
}
