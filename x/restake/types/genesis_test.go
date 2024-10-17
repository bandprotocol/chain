package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestGenesisStateValidate(t *testing.T) {
	testCases := []struct {
		name         string
		genesisState GenesisState
		expErr       bool
	}{
		{
			"valid genesisState - empty",
			GenesisState{
				Params: Params{
					AllowedDenoms: []string{},
				},
				Vaults: []Vault{},
				Locks:  []Lock{},
				Stakes: []Stake{},
			},
			false,
		},
		{
			"valid genesisState - default",
			*DefaultGenesisState(),
			false,
		},
		{
			"valid genesisState - normal",
			GenesisState{
				Params: DefaultParams(),
				Vaults: []Vault{
					{
						Key:      "key",
						IsActive: true,
					},
				},
				Locks: []Lock{
					{
						StakerAddress: "address1",
						Key:           "key",
						Power:         sdkmath.NewInt(4),
					},
					{
						StakerAddress: "address2",
						Key:           "key",
						Power:         sdkmath.NewInt(6),
					},
				},
				Stakes: []Stake{},
			},
			false,
		},
		{
			"invalid genesisState - duplicate vault name",
			GenesisState{
				Params: DefaultParams(),
				Vaults: []Vault{
					{
						Key: "test",
					},
					{
						Key: "test",
					},
				},
				Locks:  []Lock{},
				Stakes: []Stake{},
			},
			true,
		},
		{
			"invalid genesisState - lock without vault",
			GenesisState{
				Params: DefaultParams(),
				Vaults: []Vault{},
				Locks: []Lock{
					{
						Key:   "test",
						Power: sdkmath.NewInt(4),
					},
				},
				Stakes: []Stake{},
			},
			true,
		},
		{
			"invalid genesisState - wrong params",
			GenesisState{
				Params: NewParams([]string{""}),
				Vaults: []Vault{},
				Locks:  []Lock{},
				Stakes: []Stake{},
			},
			true,
		},
		{
			"invalid genesisState - invalid staker address",
			GenesisState{
				Params: DefaultParams(),
				Vaults: []Vault{},
				Locks:  []Lock{},
				Stakes: []Stake{
					{
						StakerAddress: "invalidAddress",
						Coins:         sdk.NewCoins(sdk.NewCoin("uband", sdkmath.NewInt(1))),
					},
				},
			},
			true,
		},
		{
			"invalid genesisState - invalid staked coins",
			GenesisState{
				Params: DefaultParams(),
				Vaults: []Vault{},
				Locks:  []Lock{},
				Stakes: []Stake{
					{
						StakerAddress: ValidAddress,
						Coins: []sdk.Coin{
							{Denom: "", Amount: sdkmath.NewInt(1)},
						},
					},
				},
			},
			true,
		},
		{
			"invalid genesisState - wrong params",
			GenesisState{
				Params: NewParams([]string{""}),
				Vaults: []Vault{},
				Locks:  []Lock{},
				Stakes: []Stake{},
			},
			true,
		},
		{
			"invalid genesisState - invalid staker address",
			GenesisState{
				Params: DefaultParams(),
				Vaults: []Vault{},
				Locks:  []Lock{},
				Stakes: []Stake{
					{
						StakerAddress: "invalidAddress",
						Coins:         sdk.NewCoins(sdk.NewCoin("uband", sdkmath.NewInt(1))),
					},
				},
			},
			true,
		},
		{
			"invalid genesisState - invalid staked coins",
			GenesisState{
				Params: DefaultParams(),
				Vaults: []Vault{},
				Locks:  []Lock{},
				Stakes: []Stake{
					{
						StakerAddress: ValidAddress,
						Coins: []sdk.Coin{
							{Denom: "", Amount: sdkmath.NewInt(1)},
						},
					},
				},
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
