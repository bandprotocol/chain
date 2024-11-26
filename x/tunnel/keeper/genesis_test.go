package keeper_test

import (
	"go.uber.org/mock/gomock"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	feedstypes "github.com/bandprotocol/chain/v3/x/feeds/types"
	"github.com/bandprotocol/chain/v3/x/tunnel/keeper"
	"github.com/bandprotocol/chain/v3/x/tunnel/types"
)

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
			TotalBasePacketFee: sdk.NewCoins(sdk.NewCoin("uband", sdkmath.NewInt(100))),
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
