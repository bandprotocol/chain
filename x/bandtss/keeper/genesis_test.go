package keeper_test

import (
	"time"

	"go.uber.org/mock/gomock"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	bandtesting "github.com/bandprotocol/chain/v3/testing"
	"github.com/bandprotocol/chain/v3/x/bandtss/types"
)

func (s *KeeperTestSuite) TestExportGenesis() {
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

	data := types.GenesisState{
		Params: types.DefaultParams(),
		Members: []types.Member{
			{
				Address:    bandtesting.Alice.Address.String(),
				Since:      ctx.BlockTime(),
				IsActive:   true,
				LastActive: ctx.BlockTime(),
			},
		},
		CurrentGroup: types.NewCurrentGroup(1, ctx.BlockTime()),
	}

	k.InitGenesis(ctx, data)

	currentGroup := k.GetCurrentGroup(ctx)
	s.Require().Equal(data.CurrentGroup, currentGroup)

	members := k.GetMembers(ctx)
	s.Require().Equal(data.Members, members)

	exported := k.ExportGenesis(ctx)
	s.Require().Equal(data.Params, exported.Params)
}

func (s *KeeperTestSuite) TestImportGenesisInvalidActiveTime() {
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

	data := types.GenesisState{
		Params: types.DefaultParams(),
		Members: []types.Member{
			{
				Address:    bandtesting.Alice.Address.String(),
				Since:      ctx.BlockTime(),
				IsActive:   true,
				LastActive: ctx.BlockTime(),
			},
		},
		CurrentGroup: types.NewCurrentGroup(1, ctx.BlockTime().Add(time.Hour)),
	}

	s.Require().Panics(func() {
		k.InitGenesis(ctx, data)
	})
}

func (s *KeeperTestSuite) TestImportGenesisInvalidCurrentGroupInfo() {
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

	data := types.GenesisState{
		Params: types.DefaultParams(),
		Members: []types.Member{
			{
				Address:    bandtesting.Alice.Address.String(),
				Since:      ctx.BlockTime(),
				IsActive:   true,
				LastActive: ctx.BlockTime(),
			},
		},
		CurrentGroup: types.NewCurrentGroup(0, ctx.BlockTime().Add(-1*time.Hour)),
	}

	s.Require().Panics(func() {
		k.InitGenesis(ctx, data)
	})
}
