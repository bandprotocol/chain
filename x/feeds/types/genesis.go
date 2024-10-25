package types

// NewGenesisState creates new GenesisState
func NewGenesisState(
	params Params,
	ds []DelegatorSignals,
	rs ReferenceSourceConfig,
) *GenesisState {
	return &GenesisState{
		Params:                params,
		DelegatorSignals:      ds,
		ReferenceSourceConfig: rs,
	}
}

// DefaultGenesisState returns the default genesis state
func DefaultGenesisState() *GenesisState {
	return NewGenesisState(DefaultParams(), []DelegatorSignals{}, DefaultReferenceSourceConfig())
}

// Validate performs basic genesis state validation
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}

	for _, ds := range gs.DelegatorSignals {
		if err := ds.Validate(); err != nil {
			return err
		}
	}

	if err := gs.ReferenceSourceConfig.Validate(); err != nil {
		return err
	}

	return nil
}
