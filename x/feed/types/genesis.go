package types

import fmt "fmt"

func NewGenesisState(params Params, symbols []Symbol) *GenesisState {
	return &GenesisState{
		Params:  params,
		Symbols: symbols,
	}
}

// DefaultGenesis returns the default genesis state
// TODO: what should be default one?
func DefaultGenesis() *GenesisState {
	return NewGenesisState(DefaultParams(), []Symbol{{
		Symbol:   "BAND",
		Interval: 1,
	}})
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
