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
				Vaults: []Vault{},
				Locks:  []Lock{},
			},
			false,
		},
		{
			"valid genesisState - normal",
			GenesisState{
				Vaults: []Vault{
					{
						Key:             "key",
						VaultAddress:    "vault_address",
						IsActive:        true,
						RewardsPerPower: sdk.NewDecCoins(),
						TotalPower:      sdkmath.NewInt(10),
						Remainders:      sdk.NewDecCoins(),
					},
				},
				Locks: []Lock{
					{
						StakerAddress:  "address1",
						Key:            "key",
						Power:          sdkmath.NewInt(4),
						PosRewardDebts: sdk.NewDecCoins(),
						NegRewardDebts: sdk.NewDecCoins(),
					},
					{
						StakerAddress:  "address2",
						Key:            "key",
						Power:          sdkmath.NewInt(6),
						PosRewardDebts: sdk.NewDecCoins(),
						NegRewardDebts: sdk.NewDecCoins(),
					},
				},
			},
			false,
		},
		{
			"valid genesisState - diff total power on inactive vault",
			GenesisState{
				Vaults: []Vault{
					{
						Key:             "key",
						VaultAddress:    "vault_address",
						IsActive:        false,
						RewardsPerPower: sdk.NewDecCoins(),
						TotalPower:      sdkmath.NewInt(20),
						Remainders:      sdk.NewDecCoins(),
					},
				},
				Locks: []Lock{
					{
						StakerAddress:  "address1",
						Key:            "key",
						Power:          sdkmath.NewInt(4),
						PosRewardDebts: sdk.NewDecCoins(),
						NegRewardDebts: sdk.NewDecCoins(),
					},
					{
						StakerAddress:  "address2",
						Key:            "key",
						Power:          sdkmath.NewInt(6),
						PosRewardDebts: sdk.NewDecCoins(),
						NegRewardDebts: sdk.NewDecCoins(),
					},
				},
			},
			false,
		},
		{
			"invalid genesisState - duplicate vault name",
			GenesisState{
				Vaults: []Vault{
					{
						Key: "test",
					},
					{
						Key: "test",
					},
				},
				Locks: []Lock{},
			},
			true,
		},
		{
			"invalid genesisState - diff total power on active vault",
			GenesisState{
				Vaults: []Vault{
					{
						Key:        "test",
						TotalPower: sdkmath.NewInt(5),
						IsActive:   true,
					},
				},
				Locks: []Lock{
					{
						Key:   "test",
						Power: sdkmath.NewInt(4),
					},
					{
						Key:   "test",
						Power: sdkmath.NewInt(6),
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
