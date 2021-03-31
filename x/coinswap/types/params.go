package types

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

const (
	DefaultFromExchange = "geo"
	DefaultToExchange   = "odin"
)

// nolint
var (
	KeyRateMultiplier = []byte("RateMultiplier")
	KeyValidExchanges = []byte("ValidExchanges")
)

// ParamTable for coinswap module.
func ParamKeyTable() paramstypes.KeyTable {
	return paramstypes.NewKeyTable().RegisterParamSet(&Params{})
}

func (p *Params) ParamSetPairs() paramstypes.ParamSetPairs {
	return paramstypes.ParamSetPairs{
		paramstypes.NewParamSetPair(KeyRateMultiplier, &p.RateMultiplier, validateRateMultiplier),
		paramstypes.NewParamSetPair(KeyValidExchanges, &p.ValidExchanges, validatePossibleExchanges),
	}
}

func DefaultParams() Params {
	return Params{
		RateMultiplier: sdk.NewDec(1),
		ValidExchanges: ValidExchanges{
			Exchanges: map[string]*Exchange{
				DefaultFromExchange: {Value: []string{DefaultToExchange}},
			},
		},
	}
}

func validateRateMultiplier(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if !v.IsPositive() && !v.IsZero() {
		return fmt.Errorf("rate multiplier %s must be positive or zero", v)
	}
	return nil
}

func validatePossibleExchanges(i interface{}) error {
	exchanges, ok := i.(ValidExchanges)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	for k, valid := range exchanges.Exchanges {
		for _, v := range valid.Value {
			if k == "" || v == "" {
				return fmt.Errorf("one or both denoms are empty. From: %s, To: %s", k, v)
			}
		}
	}
	return nil
}
