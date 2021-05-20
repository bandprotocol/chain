package types

import (
	"fmt"
	coinswaptypes "github.com/GeoDB-Limited/odin-core/x/coinswap/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

const (
	DefaultThresholdDenom = "loki"

	DefaultFromExchange = DefaultThresholdDenom
	DefaultToExchange   = "minigeo"
)

var (
	KeyThreshold    = []byte("Threshold")
	KeyExchangeRate = []byte("ExchangeRate")
)

var (
	DefaultThreshold = sdk.NewCoins(sdk.NewInt64Coin(DefaultThresholdDenom, 100000000000000))
)

// ParamKeyTable param table for auction module.
func ParamKeyTable() paramstypes.KeyTable {
	return paramstypes.NewKeyTable().RegisterParamSet(&Params{})
}

func (p *Params) ParamSetPairs() paramstypes.ParamSetPairs {
	return paramstypes.ParamSetPairs{
		paramstypes.NewParamSetPair(KeyExchangeRate, &p.ExchangeRate, validateExchangeRate),
		paramstypes.NewParamSetPair(KeyThreshold, &p.Threshold, validateThreshold),
	}
}

func DefaultParams() Params {
	return Params{
		ExchangeRate: coinswaptypes.Exchange{
			From:           DefaultFromExchange,
			To:             DefaultToExchange,
			RateMultiplier: sdk.NewDec(1),
		},
		Threshold: DefaultThreshold,
	}
}

func (p Params) Validate() error {
	if err := validateExchangeRate(p.ExchangeRate); err != nil {
		return err
	}
	return validateThreshold(p.Threshold)
}

func validateExchangeRate(i interface{}) error {
	ex, ok := i.(coinswaptypes.Exchange)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if !ex.RateMultiplier.IsPositive() && !ex.RateMultiplier.IsZero() {
		return fmt.Errorf("rate multiplier %s must be positive or zero", ex)
	}
	if ex.From == "" || ex.To == "" {
		return fmt.Errorf("one or both denoms are empty. From: %s, To: %s", ex.From, ex.To)
	}

	return nil
}

func validateThreshold(i interface{}) error {
	v, ok := i.(sdk.Coins)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v.IsAnyNegative() {
		return fmt.Errorf("threshold amount must be positive: %v", v)
	}
	if v.IsZero() {
		return fmt.Errorf("threshold amount must be greater than zero: %v", v)
	}

	return nil
}
