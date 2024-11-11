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
				LatestSignalPricesList: []types.LatestSignalPrices{
					{TunnelID: 1},
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
				LatestSignalPricesList: []types.LatestSignalPrices{
					{TunnelID: 1},
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
				LatestSignalPricesList: []types.LatestSignalPrices{
					{TunnelID: 1},
					{TunnelID: 1},
				},
			},
			expErr:    true,
			expErrMsg: "duplicate tunnel ID found",
		},
		"tunnel count mismatch in latest signal prices": {
			genesisState: types.GenesisState{
				Params:      types.DefaultParams(),
				TunnelCount: 2,
				Tunnels: []types.Tunnel{
					{ID: 1},
					{ID: 2},
				},
				LatestSignalPricesList: []types.LatestSignalPrices{
					{TunnelID: 1},
				},
			},
			expErr:    true,
			expErrMsg: "tunnel count mismatch in latest signal prices",
		},
		"invalid latest signal prices": {
			genesisState: types.GenesisState{
				Params:      types.DefaultParams(),
				TunnelCount: 2,
				Tunnels: []types.Tunnel{
					{ID: 1},
					{ID: 2},
				},
				LatestSignalPricesList: []types.LatestSignalPrices{
					{TunnelID: 1},
					{TunnelID: 2},
				},
			},
			expErr:    true,
			expErrMsg: "invalid latest signal prices",
		},
		"deposit has non-existent": {
			genesisState: types.GenesisState{
				Params:      types.DefaultParams(),
				TunnelCount: 2,
				Tunnels: []types.Tunnel{
					{ID: 1},
					{ID: 2},
				},
				LatestSignalPricesList: []types.LatestSignalPrices{
					{TunnelID: 1, SignalPrices: []types.SignalPrice{
						{SignalID: "signal1", Price: 100},
					}, LastInterval: 0},
					{TunnelID: 2, SignalPrices: []types.SignalPrice{
						{SignalID: "signal1", Price: 100},
					}, LastInterval: 0},
				},
				Deposits: []types.Deposit{
					{
						TunnelID:  3,
						Depositor: "addr1",
						Amount:    sdk.NewCoins(sdk.NewInt64Coin("stake", 100)),
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
				LatestSignalPricesList: []types.LatestSignalPrices{
					{TunnelID: 1, SignalPrices: []types.SignalPrice{
						{SignalID: "signal1", Price: 100},
					}, LastInterval: 0},
					{TunnelID: 2, SignalPrices: []types.SignalPrice{
						{SignalID: "signal1", Price: 100},
					}, LastInterval: 0},
				},
				Deposits: []types.Deposit{
					{
						TunnelID:  1,
						Depositor: "addr1",
						Amount:    sdk.NewCoins(sdk.NewInt64Coin("stake", 100)),
					},
					{
						TunnelID:  1,
						Depositor: "addr1",
						Amount:    sdk.NewCoins(sdk.NewInt64Coin("stake", 100)),
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
					{ID: 1, TotalDeposit: sdk.NewCoins(sdk.NewInt64Coin("stake", 100))},
					{ID: 2, TotalDeposit: sdk.NewCoins(sdk.NewInt64Coin("stake", 200))},
				},
				LatestSignalPricesList: []types.LatestSignalPrices{
					{TunnelID: 1, SignalPrices: []types.SignalPrice{
						{SignalID: "signal1", Price: 100},
					}, LastInterval: 0},
					{TunnelID: 2, SignalPrices: []types.SignalPrice{
						{SignalID: "signal1", Price: 100},
					}, LastInterval: 0},
				},
				Deposits: []types.Deposit{
					{
						TunnelID:  1,
						Depositor: "addr1",
						Amount:    sdk.NewCoins(sdk.NewInt64Coin("stake", 50)),
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
					{ID: 1, TotalDeposit: sdk.NewCoins(sdk.NewInt64Coin("stake", 100))},
					{ID: 2, TotalDeposit: sdk.NewCoins(sdk.NewInt64Coin("stake", 200))},
				},
				LatestSignalPricesList: []types.LatestSignalPrices{
					{TunnelID: 1, SignalPrices: []types.SignalPrice{
						{SignalID: "signal1", Price: 100},
					}, LastInterval: 0},
					{TunnelID: 2, SignalPrices: []types.SignalPrice{
						{SignalID: "signal1", Price: 100},
					}, LastInterval: 0},
				},
				Deposits: []types.Deposit{
					{
						TunnelID:  1,
						Depositor: "addr1",
						Amount:    sdk.NewCoins(sdk.NewInt64Coin("stake", 100)),
					},
					{
						TunnelID:  2,
						Depositor: "addr2",
						Amount:    sdk.NewCoins(sdk.NewInt64Coin("stake", 200)),
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
