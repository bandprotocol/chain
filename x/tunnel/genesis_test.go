package tunnel_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/bandprotocol/chain/v2/x/tunnel"
	"github.com/bandprotocol/chain/v2/x/tunnel/testutil"
	"github.com/bandprotocol/chain/v2/x/tunnel/types"
)

func TestValidateGenesis(t *testing.T) {
	tests := []struct {
		name       string
		genesis    *types.GenesisState
		requireErr bool
	}{
		{
			name: "valid genesis state",
			genesis: &types.GenesisState{
				Tunnels: []types.Tunnel{
					{ID: 1},
					{ID: 2},
				},
				TunnelCount: 2,
				SignalPricesInfos: []types.SignalPricesInfo{
					{TunnelID: 1},
					{TunnelID: 2},
				},
				TotalFees: types.TotalFees{
					TotalPacketFee: sdk.NewCoins(sdk.NewCoin("uband", sdk.NewInt(100))),
				},
				Params: types.DefaultParams(),
			},
			requireErr: false,
		},
		{
			name: "invalid tunnel count",
			genesis: &types.GenesisState{
				Tunnels: []types.Tunnel{
					{ID: 1},
				},
				TunnelCount: 2,
			},
			requireErr: true,
		},
		{
			name: "invalid tunnel ID",
			genesis: &types.GenesisState{
				Tunnels: []types.Tunnel{
					{ID: 3},
				},
				TunnelCount: 2,
			},
			requireErr: true,
		},
		{
			name: "invalid signal prices info",
			genesis: &types.GenesisState{
				Tunnels: []types.Tunnel{
					{ID: 1},
				},
				TunnelCount: 1,
				SignalPricesInfos: []types.SignalPricesInfo{
					{TunnelID: 0},
				},
			},
			requireErr: true,
		},
		{
			name: "invalid total fee",
			genesis: &types.GenesisState{
				Tunnels: []types.Tunnel{
					{ID: 1},
				},
				TunnelCount: 1,
				TotalFees: types.TotalFees{
					TotalPacketFee: sdk.Coins{}, // Invalid coin
				},
			},
			requireErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tunnel.ValidateGenesis(tt.genesis)
			if tt.requireErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestInitExportGenesis(t *testing.T) {
	s := testutil.NewTestSuite(t)
	ctx, k := s.Ctx, s.Keeper

	// Mock the account keeper
	s.MockAccountKeeper.EXPECT().
		GetModuleAccount(ctx, gomock.Any()).
		Return(authtypes.AccountI(&authtypes.ModuleAccount{
			BaseAccount: &authtypes.BaseAccount{Address: "test"},
		})).
		AnyTimes()
	s.MockAccountKeeper.EXPECT().GetModuleAddress(types.ModuleName).Return(sdk.AccAddress{}).AnyTimes()
	s.MockAccountKeeper.EXPECT().SetModuleAccount(ctx, gomock.Any()).AnyTimes()
	s.MockBankKeeper.EXPECT().GetAllBalances(ctx, gomock.Any()).Return(sdk.Coins{}).AnyTimes()

	// Create a valid genesis state
	genesisState := &types.GenesisState{
		Params:      types.DefaultParams(),
		TunnelCount: 1,
		Tunnels: []types.Tunnel{
			{ID: 1},
		},
		SignalPricesInfos: []types.SignalPricesInfo{
			{
				TunnelID: 1,
				SignalPrices: []types.SignalPrice{
					{SignalID: "ETH", Price: 5000},
				},
				LastIntervalTimestamp: 0,
			},
		},
		TotalFees: types.TotalFees{
			TotalPacketFee: sdk.NewCoins(sdk.NewCoin("uband", sdk.NewInt(100))),
		},
	}

	// Initialize the genesis state
	tunnel.InitGenesis(ctx, k, genesisState)

	// Export the genesis state
	exportedGenesisState := tunnel.ExportGenesis(ctx, k)

	fmt.Printf("genesisState: %v\n", genesisState)
	fmt.Printf("exportedGenesisState: %v\n", exportedGenesisState)

	// Verify the exported state matches the initialized state
	require.Equal(t, genesisState, exportedGenesisState)
}