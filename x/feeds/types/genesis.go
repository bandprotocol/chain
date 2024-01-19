package types

import (
	"fmt"
)

func NewGenesisState(params Params, symbols []Symbol, offChain OffChain) *GenesisState {
	return &GenesisState{
		Params:   params,
		Symbols:  symbols,
		OffChain: offChain,
	}
}

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return NewGenesisState(DefaultParams(), []Symbol{}, DefaultOffChain())
}

// Validate performs basic genesis state validation
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}

	for _, symbol := range gs.Symbols {
		if err := validateUint64("interval", true)(symbol.Interval); err != nil {
			return err
		}
	}

	return nil
}

func validateUint64(name string, positiveOnly bool) func(interface{}) error {
	return func(i interface{}) error {
		v, ok := i.(uint64)
		if !ok {
			return fmt.Errorf("invalid parameter type: %T", i)
		}
		if v <= 0 && positiveOnly {
			return fmt.Errorf("%s must be positive: %d", name, v)
		}
		return nil
	}
}
