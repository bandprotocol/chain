package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	bandtesting "github.com/bandprotocol/chain/v2/testing"
	"github.com/bandprotocol/chain/v2/x/bandtss/types"
)

func TestExportGenesis(t *testing.T) {
	s := NewKeeperTestSuite(t)
	ctx, k := s.Ctx, s.Keeper

	s.MockAccountKeeper.EXPECT().
		GetModuleAccount(ctx, gomock.Any()).
		Return(authtypes.AccountI(&authtypes.ModuleAccount{
			BaseAccount: &authtypes.BaseAccount{Address: "test"},
		})).
		AnyTimes()
	s.MockAccountKeeper.EXPECT().GetModuleAddress(types.ModuleName).Return(sdk.AccAddress{}).AnyTimes()
	s.MockAccountKeeper.EXPECT().SetModuleAccount(ctx, gomock.Any()).AnyTimes()
	s.MockBankKeeper.EXPECT().GetAllBalances(ctx, gomock.Any()).Return(sdk.Coins{}).AnyTimes()

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

	k.InitGenesis(ctx, &data)

	currentGroupID := k.GetCurrentGroupID(ctx)
	require.Equal(t, data.CurrentGroupID, currentGroupID)

	members := k.GetMembers(ctx)
	require.Equal(t, data.Members, members)

	exported := k.ExportGenesis(ctx)
	require.Equal(t, data.Params, exported.Params)
}

func TestExportGenesisGroupTransitionNil(t *testing.T) {
	s := NewKeeperTestSuite(t)
	ctx, k := s.Ctx, s.Keeper

	s.MockAccountKeeper.EXPECT().
		GetModuleAccount(ctx, gomock.Any()).
		Return(authtypes.AccountI(&authtypes.ModuleAccount{
			BaseAccount: &authtypes.BaseAccount{Address: "test"},
		})).
		AnyTimes()
	s.MockAccountKeeper.EXPECT().GetModuleAddress(types.ModuleName).Return(sdk.AccAddress{}).AnyTimes()
	s.MockAccountKeeper.EXPECT().SetModuleAccount(ctx, gomock.Any()).AnyTimes()
	s.MockBankKeeper.EXPECT().GetAllBalances(ctx, gomock.Any()).Return(sdk.Coins{}).AnyTimes()

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

	k.InitGenesis(ctx, &data)
	exported := k.ExportGenesis(ctx)
	require.Equal(t, data.Params, exported.Params)
}
