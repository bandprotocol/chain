package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	band "github.com/bandprotocol/chain/v3/app"
	"github.com/bandprotocol/chain/v3/x/bandtss/types"
)

func init() {
	band.SetBech32AddressPrefixesAndBip44CoinTypeAndSeal(sdk.GetConfig())
	sdk.DefaultBondDenom = "uband"
}

func TestGenesisStateValidate(t *testing.T) {
	validMembers := []types.Member{
		{
			Address: "band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs",
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
