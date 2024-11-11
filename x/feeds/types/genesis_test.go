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
				Params:                DefaultParams(),
				Votes:                 []Vote{},
				ReferenceSourceConfig: DefaultReferenceSourceConfig(),
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
				Params:                Params{},
				Votes:                 []Vote{},
				ReferenceSourceConfig: DefaultReferenceSourceConfig(),
			},
			true,
		},
		{
			"invalid reference source config",
			GenesisState{
				Params:                DefaultParams(),
				Votes:                 []Vote{},
				ReferenceSourceConfig: ReferenceSourceConfig{},
			},
			true,
		},
	}

	for _, tc := range testCases {
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
