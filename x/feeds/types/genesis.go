package types

import (
	"fmt"
)

func NewGenesisState(params Params, symbols []Symbol, ps PriceService) *GenesisState {
	return &GenesisState{
		Params:       params,
		Symbols:      symbols,
		PriceService: ps,
	}
}

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return NewGenesisState(DefaultParams(), []Symbol{}, DefaultPriceService())
}

// Validate performs basic genesis state validation
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}

	for _, symbol := range gs.Symbols {
		if err := validateInt64("minInterval", true)(symbol.MinInterval); err != nil {
			return err
		}
		if err := validateInt64("maxInterval", true)(symbol.MaxInterval); err != nil {
			return err
		}
		if err := validateInt64("timestamp", true)(symbol.Timestamp); err != nil {
			return err
		}
	}

	return nil
}

func validateInt64(name string, positiveOnly bool) func(interface{}) error {
	return func(i interface{}) error {
		v, ok := i.(int64)
		if !ok {
			return fmt.Errorf("invalid parameter type: %T", i)
		}
		if v <= 0 && positiveOnly {
			return fmt.Errorf("%s must be positive: %d", name, v)
		}
		return nil
	}
}
