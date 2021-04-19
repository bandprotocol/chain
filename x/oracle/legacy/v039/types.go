package v039

import (
	"github.com/bandprotocol/chain/x/oracle/types"
)

type DataSource struct {
	Owner       string `json:"owner,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Filename    string `json:"filename,omitempty"`
}

type GenesisState struct {
	Params        types.Params                  `json:"params" yaml:"params"`
	DataSources   []DataSource                  `json:"data_sources"  yaml:"data_sources"`
	OracleScripts []types.OracleScript          `json:"oracle_scripts"  yaml:"oracle_scripts"`
	Reporters     []types.ReportersPerValidator `json:"reporters" yaml:"reporters"`
}
