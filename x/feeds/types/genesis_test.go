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
				Params:       DefaultParams(),
				Feeds:        []Feed{},
				PriceService: DefaultPriceService(),
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
				Params:       Params{},
				Feeds:        []Feed{},
				PriceService: DefaultPriceService(),
			},
			true,
		},
		{
			"invalid symbol",
			GenesisState{
				Params: DefaultParams(),
				Feeds: []Feed{
					{
						SignalID:                    "crypto_price.bandusd",
						Power:                       10,
						Interval:                    -5,
						LastIntervalUpdateTimestamp: 1234567890,
					},
				},
				PriceService: DefaultPriceService(),
			},
			true,
		},
		{
			"invalid price service",
			GenesisState{
				Params:       DefaultParams(),
				Feeds:        []Feed{},
				PriceService: PriceService{},
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
