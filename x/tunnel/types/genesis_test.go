package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/x/tunnel/types"
)

func TestValidateGenesis(t *testing.T) {
	cases := map[string]struct {
		genesisState types.GenesisState
		expErr       bool
		expErrMsg    string
	}{
		"length of tunnels does not match tunnel count": {
			genesisState: types.GenesisState{
				Params:      types.DefaultParams(),
				TunnelCount: 2,
				Tunnels: []types.Tunnel{
					{ID: 1},
				},
			},
			expErr:    true,
			expErrMsg: "length of tunnels does not match tunnel count",
		},
		"invalid tunnel count": {
			genesisState: types.GenesisState{
				Params:      types.DefaultParams(),
				TunnelCount: 2,
				Tunnels: []types.Tunnel{
					{ID: 1},
				},
			},
			expErr:    true,
			expErrMsg: "length of tunnels does not match tunnel count",
		},
		"duplicate tunnel ID": {
			genesisState: types.GenesisState{
				Params:      types.DefaultParams(),
				TunnelCount: 2,
				Tunnels: []types.Tunnel{
					{ID: 1},
					{ID: 1}, // Duplicate ID
				},
			},
			expErr:    true,
			expErrMsg: "duplicate tunnel ID found",
		},
		"deposit has non-existent": {
			genesisState: types.GenesisState{
				Params:      types.DefaultParams(),
				TunnelCount: 2,
				Tunnels: []types.Tunnel{
					{ID: 1},
					{ID: 2},
				},
				Deposits: []types.Deposit{
					{
						TunnelID:  3,
						Depositor: "addr1",
						Amount:    sdk.NewCoins(sdk.NewInt64Coin("uband", 100)),
					}, // Non-existent tunnel ID
				},
			},
			expErr:    true,
			expErrMsg: "deposit has non-existent",
		},
		"duplicate deposit": {
			genesisState: types.GenesisState{
				Params:      types.DefaultParams(),
				TunnelCount: 2,
				Tunnels: []types.Tunnel{
					{ID: 1},
					{ID: 2},
				},
				Deposits: []types.Deposit{
					{
						TunnelID:  1,
						Depositor: "addr1",
						Amount:    sdk.NewCoins(sdk.NewInt64Coin("uband", 100)),
					},
					{
						TunnelID:  1,
						Depositor: "addr1",
						Amount:    sdk.NewCoins(sdk.NewInt64Coin("uband", 100)),
					}, // Duplicate deposit
				},
			},
			expErr:    true,
			expErrMsg: "duplicate deposit",
		},
		"deposits mismatch total deposit for tunnel": {
			genesisState: types.GenesisState{
				Params:      types.DefaultParams(),
				TunnelCount: 2,
				Tunnels: []types.Tunnel{
					{ID: 1, TotalDeposit: sdk.NewCoins(sdk.NewInt64Coin("uband", 100))},
					{ID: 2, TotalDeposit: sdk.NewCoins(sdk.NewInt64Coin("uband", 200))},
				},
				Deposits: []types.Deposit{
					{
						TunnelID:  1,
						Depositor: "addr1",
						Amount:    sdk.NewCoins(sdk.NewInt64Coin("uband", 50)),
					},
				},
			},
			expErr:    true,
			expErrMsg: "deposits mismatch total deposit for tunnel",
		},
		"all good": {
			genesisState: types.GenesisState{
				Params:      types.DefaultParams(),
				TunnelCount: 2,
				Tunnels: []types.Tunnel{
					{ID: 1, TotalDeposit: sdk.NewCoins(sdk.NewInt64Coin("uband", 100))},
					{ID: 2, TotalDeposit: sdk.NewCoins(sdk.NewInt64Coin("uband", 200))},
				},
				Deposits: []types.Deposit{
					{
						TunnelID:  1,
						Depositor: "addr1",
						Amount:    sdk.NewCoins(sdk.NewInt64Coin("uband", 100)),
					},
					{
						TunnelID:  2,
						Depositor: "addr2",
						Amount:    sdk.NewCoins(sdk.NewInt64Coin("uband", 200)),
					},
				},
			},
			expErr: false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := types.ValidateGenesis(tc.genesisState)
			if tc.expErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expErrMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
