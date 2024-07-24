package types

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
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
				Keys:  []Key{},
				Locks: []Lock{},
			},
			false,
		},
		{
			"valid genesisState - normal",
			GenesisState{
				Keys: []Key{
					{
						Name:            "key",
						PoolAddress:     "pool_address",
						IsActive:        true,
						RewardPerPowers: sdk.NewDecCoins(),
						TotalPower:      sdkmath.NewInt(10),
						Remainders:      sdk.NewDecCoins(),
					},
				},
				Locks: []Lock{
					{
						LockerAddress:  "address1",
						Key:            "key",
						Amount:         sdkmath.NewInt(4),
						PosRewardDebts: sdk.NewDecCoins(),
						NegRewardDebts: sdk.NewDecCoins(),
					},
					{
						LockerAddress:  "address2",
						Key:            "key",
						Amount:         sdkmath.NewInt(6),
						PosRewardDebts: sdk.NewDecCoins(),
						NegRewardDebts: sdk.NewDecCoins(),
					},
				},
			},
			false,
		},
		{
			"valid genesisState - diff total power on inactive key",
			GenesisState{
				Keys: []Key{
					{
						Name:            "key",
						PoolAddress:     "pool_address",
						IsActive:        false,
						RewardPerPowers: sdk.NewDecCoins(),
						TotalPower:      sdkmath.NewInt(20),
						Remainders:      sdk.NewDecCoins(),
					},
				},
				Locks: []Lock{
					{
						LockerAddress:  "address1",
						Key:            "key",
						Amount:         sdkmath.NewInt(4),
						PosRewardDebts: sdk.NewDecCoins(),
						NegRewardDebts: sdk.NewDecCoins(),
					},
					{
						LockerAddress:  "address2",
						Key:            "key",
						Amount:         sdkmath.NewInt(6),
						PosRewardDebts: sdk.NewDecCoins(),
						NegRewardDebts: sdk.NewDecCoins(),
					},
				},
			},
			false,
		},
		{
			"invalid genesisState - duplicate key name",
			GenesisState{
				Keys: []Key{
					{
						Name: "test",
					},
					{
						Name: "test",
					},
				},
				Locks: []Lock{},
			},
			true,
		},
		{
			"invalid genesisState - diff total power on active key",
			GenesisState{
				Keys: []Key{
					{
						Name:       "test",
						TotalPower: sdkmath.NewInt(5),
						IsActive:   true,
					},
				},
				Locks: []Lock{
					{
						Key:    "test",
						Amount: sdkmath.NewInt(4),
					},
					{
						Key:    "test",
						Amount: sdkmath.NewInt(6),
					},
				},
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
