package v039

import (
	"github.com/bandprotocol/chain/v2/x/oracle/types"
)

// DataSource represents data source information
type DataSource struct {
	Owner       string `json:"owner,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Filename    string `json:"filename,omitempty"`
}

// Params represents parameters for oracle module
type Params struct {
	MaxRawRequestCount      uint64 `json:"max_raw_request_count,omitempty"`
	MaxAskCount             uint64 `son:"max_ask_count,omitempty"`
	ExpirationBlockCount    uint64 `json:"expiration_block_count,omitempty"`
	BaseRequestGas          uint64 `json:"base_request_gas,omitempty"`
	PerValidatorRequestGas  uint64 `json:"per_validator_request_gas,omitempty"`
	SamplingTryCount        uint64 `json:"sampling_try_count,omitempty"`
	OracleRewardPercentage  uint64 `json:"oracle_reward_percentage,omitempty"`
	InactivePenaltyDuration uint64 `json:"inactive_penalty_duration,omitempty"`
}

type GenesisState struct {
	Params        Params                        `json:"params" yaml:"params"`
	DataSources   []DataSource                  `json:"data_sources"  yaml:"data_sources"`
	OracleScripts []types.OracleScript          `json:"oracle_scripts"  yaml:"oracle_scripts"`
	Reporters     []types.ReportersPerValidator `json:"reporters" yaml:"reporters"`
}
