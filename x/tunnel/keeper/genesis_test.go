package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	feedstypes "github.com/bandprotocol/chain/v3/x/feeds/types"
	"github.com/bandprotocol/chain/v3/x/tunnel/keeper"
	"github.com/bandprotocol/chain/v3/x/tunnel/types"
)

func TestValidateGenesis(t *testing.T) {
	cases := map[string]struct {
		genesis    *types.GenesisState
		requireErr bool
		errMsg     string
	}{
		"length of tunnels does not match tunnel count": {
			genesis: &types.GenesisState{
				TunnelCount: 2,
				Tunnels: []types.Tunnel{
					{ID: 1},
				},
			},
			requireErr: true,
			errMsg:     "length of tunnels does not match tunnel count",
		},
		"tunnel count mismatch in tunnels": {
			genesis: &types.GenesisState{
				TunnelCount: 1,
				Tunnels: []types.Tunnel{
					{ID: 3},
				},
			},
			requireErr: true,
			errMsg:     "tunnel count mismatch in tunnels",
		},
		"invalid total fees": {
			genesis: &types.GenesisState{
				TunnelCount: 1,
				Tunnels: []types.Tunnel{
					{ID: 1},
				},
				TotalFees: types.TotalFees{
					TotalPacketFee: sdk.Coins{
						{Denom: "uband", Amount: sdkmath.NewInt(-100)},
					}, // Invalid coin
				},
			},
			requireErr: true,
			errMsg:     "invalid total fees",
		},
		"deposits mismatch total deposit for tunnel": {
			genesis: &types.GenesisState{
				TunnelCount: 1,
				Tunnels: []types.Tunnel{
					{ID: 1, TotalDeposit: sdk.NewCoins()},
				},
				TotalFees: types.TotalFees{},
				Deposits: []types.Deposit{
					{
						TunnelID:  1,
						Depositor: sdk.AccAddress("account1").String(),
						Amount:    sdk.NewCoins(sdk.NewCoin("uband", sdkmath.NewInt(100))),
					},
				},
			},
			requireErr: true,
			errMsg:     "deposits mismatch total deposit for tunnel",
		},
		"all good": {
			genesis: &types.GenesisState{
				TunnelCount: 2,
				Tunnels: []types.Tunnel{
					{ID: 1, TotalDeposit: sdk.NewCoins(sdk.NewCoin("uband", sdkmath.NewInt(100)))},
					{ID: 2, TotalDeposit: sdk.NewCoins(sdk.NewCoin("uband", sdkmath.NewInt(100)))},
				},
				TotalFees: types.TotalFees{
					TotalPacketFee: sdk.NewCoins(sdk.NewCoin("uband", sdkmath.NewInt(100))),
				},
				Deposits: []types.Deposit{
					{
						TunnelID:  1,
						Depositor: sdk.AccAddress("account1").String(),
						Amount:    sdk.NewCoins(sdk.NewCoin("uband", sdkmath.NewInt(100))),
					},
					{
						TunnelID:  2,
						Depositor: sdk.AccAddress("account2").String(),
						Amount:    sdk.NewCoins(sdk.NewCoin("uband", sdkmath.NewInt(100))),
					},
				},
				Params: types.DefaultParams(),
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
	s.bankKeeper.EXPECT().
		GetAllBalances(ctx, gomock.Any()).
		Return(sdk.NewCoins(sdk.NewCoin("uband", sdkmath.NewInt(100)))).
		AnyTimes()

	// Create a valid genesis state
	genesisState := &types.GenesisState{
		Params:      types.DefaultParams(),
		TunnelCount: 1,
		Tunnels: []types.Tunnel{
			{ID: 1},
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

	// check latest price on chain.
	for _, t := range genesisState.Tunnels {
		latestPrices, err := k.GetLatestPrices(ctx, t.ID)
		s.Require().NoError(err)
		s.Require().Equal(types.LatestPrices{
			TunnelID:     t.ID,
			Prices:       []feedstypes.Price(nil),
			LastInterval: 0,
		}, latestPrices)
	}
}
