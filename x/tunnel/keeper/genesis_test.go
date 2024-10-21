package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/x/tunnel/types"
)

func TestValidateGenesis(t *testing.T) {
	cases := map[string]struct {
		genesis    *types.GenesisState
		requireErr bool
		errMsg     string
	}{
		"invalid port ID": {
			genesis: &types.GenesisState{
				PortID: "invalid/id",
				Params: types.DefaultParams(),
			},
			requireErr: true,
			errMsg:     "invalid port ID",
		},
		"length of tunnels does not match tunnel count": {
			genesis: &types.GenesisState{
				Tunnels: []types.Tunnel{
					{ID: 1},
				},
				TunnelCount: 2,
				PortID:      types.PortID,
			},
			requireErr: true,
			errMsg:     "length of tunnels does not match tunnel count",
		},
		"tunnel count mismatch in tunnels": {
			genesis: &types.GenesisState{
				Tunnels: []types.Tunnel{
					{ID: 1},
					{ID: 3},
				},
				TunnelCount: 2,
				PortID:      types.PortID,
			},
			requireErr: true,
			errMsg:     "tunnel count mismatch in tunnels",
		},
		"tunnel count mismatch in latest signal prices": {
			genesis: &types.GenesisState{
				Tunnels: []types.Tunnel{
					{ID: 1},
				},
				TunnelCount: 1,
				LatestSignalPricesList: []types.LatestSignalPrices{
					{TunnelID: 1},
					{TunnelID: 2},
				},
				PortID: types.PortID,
			},
			requireErr: true,
			errMsg:     "tunnel count mismatch in latest signal prices",
		},
		"invalid latest signal prices": {
			genesis: &types.GenesisState{
				Tunnels: []types.Tunnel{
					{ID: 1},
				},
				TunnelCount: 1,
				LatestSignalPricesList: []types.LatestSignalPrices{
					{
						TunnelID:     1,
						SignalPrices: []types.SignalPrice{},
					},
				},
				TotalFees: types.TotalFees{},
				PortID:    types.PortID,
			},
			requireErr: true,
			errMsg:     "invalid latest signal prices",
		},
		"invalid total fees": {
			genesis: &types.GenesisState{
				Tunnels: []types.Tunnel{
					{ID: 1},
				},

				TunnelCount: 1,
				LatestSignalPricesList: []types.LatestSignalPrices{
					{
						TunnelID: 1,
						SignalPrices: []types.SignalPrice{
							{SignalID: "ETH", Price: 5000},
						},
					},
				},
				TotalFees: types.TotalFees{
					TotalPacketFee: sdk.Coins{
						{Denom: "uband", Amount: sdkmath.NewInt(-100)},
					}, // Invalid coin
				},
				PortID: types.PortID,
			},
			requireErr: true,
			errMsg:     "invalid total fees",
		},
		"all good": {
			genesis: &types.GenesisState{
				Tunnels: []types.Tunnel{
					{ID: 1},
					{ID: 2},
				},
				TunnelCount: 2,
				LatestSignalPricesList: []types.LatestSignalPrices{
					{
						TunnelID: 1,
						SignalPrices: []types.SignalPrice{
							{SignalID: "ETH", Price: 5000},
						},
					},
					{
						TunnelID: 2,
						SignalPrices: []types.SignalPrice{
							{SignalID: "ETH", Price: 5000},
						},
					},
				},
				TotalFees: types.TotalFees{
					TotalPacketFee: sdk.NewCoins(sdk.NewCoin("uband", sdkmath.NewInt(100))),
				},
				Params: types.DefaultParams(),
				PortID: types.PortID,
			},
			requireErr: false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := types.ValidateGenesis(*tc.genesis)
			if tc.requireErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
