package types

// NewGenesisState creates new GenesisState
func NewGenesisState(
	params Params,
	votes []Vote,
	rs ReferenceSourceConfig,
) *GenesisState {
	return &GenesisState{
		Params:                params,
		Votes:                 votes,
		ReferenceSourceConfig: rs,
	}
}

// DefaultGenesisState returns the default genesis state
func DefaultGenesisState() *GenesisState {
	return NewGenesisState(DefaultParams(), []Vote{}, DefaultReferenceSourceConfig())
}

// Validate performs basic genesis state validation
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}

	for _, v := range gs.Votes {
		if err := v.Validate(); err != nil {
			return err
		}
	}

	if err := gs.ReferenceSourceConfig.Validate(); err != nil {
		return err
	}

	return nil
}
