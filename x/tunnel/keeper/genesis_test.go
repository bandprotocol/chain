package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/bandprotocol/chain/v3/x/tunnel/keeper"
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

func (s *KeeperTestSuite) TestInitExportGenesis() {
	ctx, k := s.ctx, s.keeper

	s.accountKeeper.EXPECT().
		GetModuleAccount(ctx, gomock.Any()).
		Return(sdk.AccountI(&authtypes.ModuleAccount{
			BaseAccount: &authtypes.BaseAccount{Address: "test"},
		})).
		AnyTimes()
	s.accountKeeper.EXPECT().GetModuleAddress(types.ModuleName).Return(sdk.AccAddress{}).AnyTimes()
	s.accountKeeper.EXPECT().SetModuleAccount(ctx, gomock.Any()).AnyTimes()
	s.bankKeeper.EXPECT().GetAllBalances(ctx, gomock.Any()).Return(sdk.Coins{}).AnyTimes()
	// s.scopedKeeper.EXPECT().GetCapability(ctx, gomock.Any()).Return(nil, true).AnyTimes()

	// Create a valid genesis state
	genesisState := &types.GenesisState{
		PortID:      types.PortID,
		Params:      types.DefaultParams(),
		TunnelCount: 1,
		Tunnels: []types.Tunnel{
			{ID: 1},
		},
		LatestSignalPricesList: []types.LatestSignalPrices{
			{
				TunnelID: 1,
				SignalPrices: []types.SignalPrice{
					{SignalID: "ETH", Price: 5000},
				},
				Timestamp: 0,
			},
		},
		TotalFees: types.TotalFees{
			TotalPacketFee: sdk.NewCoins(sdk.NewCoin("uband", sdkmath.NewInt(100))),
		},
	}

	// Initialize the genesis state
	keeper.InitGenesis(ctx, k, genesisState)

	// Export the genesis state
	exportedGenesisState := keeper.ExportGenesis(ctx, k)

	// Verify the exported state matches the initialized state
	s.Require().Equal(genesisState, exportedGenesisState)
}
