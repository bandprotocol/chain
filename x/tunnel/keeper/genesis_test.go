package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/bandprotocol/chain/v2/x/tunnel/keeper"
	"github.com/bandprotocol/chain/v2/x/tunnel/types"
)

func TestValidateGenesis(t *testing.T) {
	tests := []struct {
		name       string
		genesis    *types.GenesisState
		requireErr bool
		errMsg     string
	}{
		{
			name: "invalid port ID",
			genesis: &types.GenesisState{
				PortID: "invalid/id",
			},
			requireErr: true,
			errMsg:     "invalid identifier",
		},
		{
			name: "length of tunnels does not match tunnel count",
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
		{
			name: "tunnel ID greater than tunnel count",
			genesis: &types.GenesisState{
				Tunnels: []types.Tunnel{
					{ID: 1},
					{ID: 2},
				},
				TunnelCount: 1,
				PortID:      types.PortID,
			},
			requireErr: true,
			errMsg:     "length of tunnels does not match tunnel count:",
		},
		{
			name: "latest signal prices does not match tunnel count",
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
			errMsg:     "latest signal prices does not match tunnel count",
		},
		{
			name: "tunnel count mismatch in latest signal prices",
			genesis: &types.GenesisState{
				Tunnels: []types.Tunnel{
					{ID: 1},
				},
				TunnelCount: 1,
				LatestSignalPricesList: []types.LatestSignalPrices{
					{TunnelID: 0},
				},
				PortID: types.PortID,
			},
			requireErr: true,
			errMsg:     "tunnel count mismatch",
		},
		{
			name: "tunnel id is zero in latest signal prices",
			genesis: &types.GenesisState{
				Tunnels: []types.Tunnel{
					{ID: 1},
				},
				TunnelCount: 1,
				LatestSignalPricesList: []types.LatestSignalPrices{
					{TunnelID: 0},
				},
				PortID: types.PortID,
			},
			requireErr: true,
			errMsg:     "tunnel count mismatch",
		},
		{
			name: "invalid total fee",
			genesis: &types.GenesisState{
				Tunnels: []types.Tunnel{
					{ID: 1},
				},

				TunnelCount: 1,
				LatestSignalPricesList: []types.LatestSignalPrices{
					{TunnelID: 1},
				},
				TotalFees: types.TotalFees{
					TotalPacketFee: sdk.Coins{
						{Denom: "uband", Amount: sdk.NewInt(-100)},
					}, // Invalid coin
				},
				PortID: types.PortID,
			},
			requireErr: true,
			errMsg:     "invalid total fees",
		},

		{
			name: "all good",
			genesis: &types.GenesisState{
				Tunnels: []types.Tunnel{
					{ID: 1},
					{ID: 2},
				},
				TunnelCount: 2,
				LatestSignalPricesList: []types.LatestSignalPrices{
					{TunnelID: 1},
					{TunnelID: 2},
				},
				TotalFees: types.TotalFees{
					TotalPacketFee: sdk.NewCoins(sdk.NewCoin("uband", sdk.NewInt(100))),
				},
				PortID: types.PortID,
				Params: types.DefaultParams(),
			},
			requireErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := keeper.ValidateGenesis(tt.genesis)
			if tt.requireErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
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
		Return(authtypes.AccountI(&authtypes.ModuleAccount{
			BaseAccount: &authtypes.BaseAccount{Address: "test"},
		})).
		AnyTimes()
	s.accountKeeper.EXPECT().GetModuleAddress(types.ModuleName).Return(sdk.AccAddress{}).AnyTimes()
	s.accountKeeper.EXPECT().SetModuleAccount(ctx, gomock.Any()).AnyTimes()
	s.bankKeeper.EXPECT().GetAllBalances(ctx, gomock.Any()).Return(sdk.Coins{}).AnyTimes()
	s.scopedKeeper.EXPECT().GetCapability(ctx, gomock.Any()).Return(nil, true).AnyTimes()

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
			TotalPacketFee: sdk.NewCoins(sdk.NewCoin("uband", sdk.NewInt(100))),
		},
	}

	// Initialize the genesis state
	keeper.InitGenesis(ctx, k, genesisState)

	// Export the genesis state
	exportedGenesisState := keeper.ExportGenesis(ctx, k)

	// Verify the exported state matches the initialized state
	s.Require().Equal(genesisState, exportedGenesisState)
}