package types

// NewGenesisState creates new GenesisState
func NewGenesisState(
	params Params,
	ds []DelegatorSignals,
	ps PriceService,
) *GenesisState {
	return &GenesisState{
		Params:           params,
		DelegatorSignals: ds,
		PriceService:     ps,
	}
}

// DefaultGenesisState returns the default genesis state
func DefaultGenesisState() *GenesisState {
	return NewGenesisState(DefaultParams(), []DelegatorSignals{}, DefaultPriceService())
}

// Validate performs basic genesis state validation
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}

	maxSignalIDCharacters := gs.Params.MaxSignalIDCharacters
	for _, ds := range gs.DelegatorSignals {
		if err := ds.Validate(maxSignalIDCharacters); err != nil {
			return err
		}
	}

	if err := gs.PriceService.Validate(); err != nil {
		return err
	}

	return nil
}
