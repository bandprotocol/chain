package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gopkg.in/yaml.v2"
)

var (
	DefaultMinDeposit            = sdk.NewCoins(sdk.NewInt64Coin("uband", 10))
	DefaultMinDeviationBPS       = uint64(100)
	DefaultTSSRouteFee           = sdk.NewCoins(sdk.NewInt64Coin("uband", 1000000))
	DefaultAxelarRouteFee        = sdk.NewCoins(sdk.NewInt64Coin("uband", 1000000))
	DefaultTSSSupportedChains    = []string{"ethereum", "cosmos", "terra", "band"}
	DefaultAxelarSupportedChains = []string{"ethereum", "cosmos", "terra", "band"}
)

// NewParams creates a new Params instance
func NewParams(
	minDeposit sdk.Coins,
	minDeviationBPS uint64,
	tssRouteFee sdk.Coins,
	axelarRouteFee sdk.Coins,
	tssSupportedChains []string,
	axelarSupportedChains []string,
) Params {
	return Params{
		MinDeposit:            minDeposit,
		MinDeviationBPS:       minDeviationBPS,
		TSSRouteFee:           tssRouteFee,
		AxelarRouteFee:        axelarRouteFee,
		TssSupportedChains:    tssSupportedChains,
		AxelarSupportedChains: axelarSupportedChains,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(
		DefaultMinDeposit,
		DefaultMinDeviationBPS,
		DefaultTSSRouteFee,
		DefaultAxelarRouteFee,
		DefaultTSSSupportedChains,
		DefaultAxelarSupportedChains,
	)
}

// Validate validates the set of params
func (p Params) Validate() error {
	return nil
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}
