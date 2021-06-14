package types

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/codec"
)

// NewGenesisState creates a new GenesisState instanc e
func NewGenesisState(params Params, dataSources []DataSource, oracleScripts []OracleScript, reporters []ReportersPerValidator) *GenesisState {
	return &GenesisState{
		Params:        params,
		DataSources:   dataSources,
		OracleScripts: oracleScripts,
		Reporters:     reporters,
	}
}

// DefaultGenesisState returns the default oracle genesis state.
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params:        DefaultParams(),
		DataSources:   []DataSource{},
		OracleScripts: []OracleScript{},
		Reporters:     []ReportersPerValidator{},
	}
}

// GetGenesisStateFromAppState returns oracle GenesisState given raw application genesis state.
func GetGenesisStateFromAppState(cdc codec.JSONMarshaler, appState map[string]json.RawMessage) *GenesisState {
	var genesisState GenesisState

	if appState[ModuleName] != nil {
		cdc.MustUnmarshalJSON(appState[ModuleName], &genesisState)
	}

	return &genesisState
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (g GenesisState) Validate() error { return nil }
