package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bandprotocol/chain/v3/x/tunnel/types"
)

func TestParams_Validate(t *testing.T) {
	cases := map[string]struct {
		genesisState types.Params
		expErr       bool
		expErrMsg    string
	}{
		"invalid MinInterval": {
			genesisState: func() types.Params {
				p := types.DefaultParams()
				p.MinInterval = 0
				return p
			}(),
			expErr:    true,
			expErrMsg: "min interval must be positive",
		},
		"invalid MaxInterval": {
			genesisState: func() types.Params {
				p := types.DefaultParams()
				p.MaxInterval = 0
				return p
			}(),
			expErr:    true,
			expErrMsg: "max interval must be positive",
		},
		"invalid interval range": {
			genesisState: func() types.Params {
				p := types.DefaultParams()
				p.MinInterval = 10
				p.MaxInterval = 5
				return p
			}(),
			expErr:    true,
			expErrMsg: "max interval must be greater than min interval: 5 <= 10",
		},
		"invalid MinDeviationBPS": {
			genesisState: func() types.Params {
				p := types.DefaultParams()
				p.MinDeviationBPS = 0
				return p
			}(),
			expErr:    true,
			expErrMsg: "min deviation bps must be positive",
		},
		"invalid MaxDeviationBPS": {
			genesisState: func() types.Params {
				p := types.DefaultParams()
				p.MaxDeviationBPS = 0
				return p
			}(),
			expErr:    true,
			expErrMsg: "max deviation bps must be positive",
		},
		"invalid deviation range": {
			genesisState: func() types.Params {
				p := types.DefaultParams()
				p.MinDeviationBPS = 10
				p.MaxDeviationBPS = 5
				return p
			}(),
			expErr:    true,
			expErrMsg: "max deviation bps must be greater than min deviation bps: 5 <= 10",
		},
		"invalid MaxSignals": {
			genesisState: func() types.Params {
				p := types.DefaultParams()
				p.MaxSignals = 0
				return p
			}(),
			expErr:    true,
			expErrMsg: "max signals must be positive",
		},
		"valid params": {
			genesisState: types.DefaultParams(),
			expErr:       false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := tc.genesisState.Validate()
			if tc.expErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expErrMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
