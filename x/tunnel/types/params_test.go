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
