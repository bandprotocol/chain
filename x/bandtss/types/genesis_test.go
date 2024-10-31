package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bandprotocol/chain/v3/x/bandtss/types"
)

func TestGenesisStateValidate(t *testing.T) {
	validMembers := []types.Member{
		{
			Address: "cosmos1xxjxtce966clgkju06qp475j663tg8pmklxcy8",
			GroupID: 1,
		},
	}

	testCases := []struct {
		name         string
		genesisState types.GenesisState
		expErr       bool
	}{
		{
			"valid genesisState",
			types.GenesisState{
				Params:         types.DefaultParams(),
				Members:        validMembers,
				CurrentGroupID: 1,
			},
			false,
		},
		{
			"invalid genesisState - members not belongs to current group",
			types.GenesisState{
				Params:         types.DefaultParams(),
				Members:        validMembers,
				CurrentGroupID: 0,
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
