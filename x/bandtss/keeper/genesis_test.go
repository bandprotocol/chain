package keeper_test

import (
	"go.uber.org/mock/gomock"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/bandprotocol/chain/v3/pkg/tss"
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
		CurrentGroupID: tss.GroupID(1),
	}

	k.InitGenesis(ctx, data)

	currentGroupID := k.GetCurrentGroupID(ctx)
	s.Require().Equal(data.CurrentGroupID, currentGroupID)

	members := k.GetMembers(ctx)
	s.Require().Equal(data.Members, members)

	exported := k.ExportGenesis(ctx)
	s.Require().Equal(data.Params, exported.Params)
}

func (s *KeeperTestSuite) TestExportGenesisGroupTransitionNil() {
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
		CurrentGroupID: tss.GroupID(1),
	}

	k.InitGenesis(ctx, data)
	exported := k.ExportGenesis(ctx)
	s.Require().Equal(data.Params, exported.Params)
}
