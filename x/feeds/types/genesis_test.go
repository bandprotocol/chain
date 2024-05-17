package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenesisStateValidate(t *testing.T) {
	testCases := []struct {
		name         string
		genesisState GenesisState
		expErr       bool
	}{
		{
			"valid genesisState",
			GenesisState{
				Params:           DefaultParams(),
				DelegatorSignals: []DelegatorSignals{},
				PriceService:     DefaultPriceService(),
			},
			false,
		},
		{
			"empty genesisState",
			GenesisState{},
			true,
		},
		{
			"invalid params",
			GenesisState{
				Params:           Params{},
				DelegatorSignals: []DelegatorSignals{},
				PriceService:     DefaultPriceService(),
			},
			true,
		},
		{
			"invalid price service",
			GenesisState{
				Params:           DefaultParams(),
				DelegatorSignals: []DelegatorSignals{},
				PriceService:     PriceService{},
			},
			true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(tt *testing.T) {
			err := tc.genesisState.Validate()

			if tc.expErr {
				require.Error(tt, err)
			} else {
				require.NoError(tt, err)
			}
		})
	}
}
