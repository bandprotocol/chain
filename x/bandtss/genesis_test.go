package bandtss_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	bandtesting "github.com/bandprotocol/chain/v2/testing"
	"github.com/bandprotocol/chain/v2/x/bandtss"
	"github.com/bandprotocol/chain/v2/x/bandtss/testutil"
	"github.com/bandprotocol/chain/v2/x/bandtss/types"
)

func TestExportGenesis(t *testing.T) {
	s := testutil.NewTestSuite(t)
	ctx, k := s.Ctx, s.Keeper

	s.MockAccountKeeper.EXPECT().GetModuleAccount(ctx, gomock.Any()).Return(authtypes.AccountI(&authtypes.ModuleAccount{
		BaseAccount: &authtypes.BaseAccount{Address: "test"},
	})).AnyTimes()
	s.MockAccountKeeper.EXPECT().GetModuleAddress(types.ModuleName).Return(sdk.AccAddress{}).AnyTimes()
	s.MockAccountKeeper.EXPECT().SetModuleAccount(ctx, gomock.Any()).AnyTimes()
	s.MockBankKeeper.EXPECT().GetAllBalances(ctx, gomock.Any()).Return(sdk.Coins{}).AnyTimes()

	data := types.GenesisState{
		Params: types.DefaultParams(),
		Members: []types.Member{
			{Address: bandtesting.Alice.Address.String(), Since: ctx.BlockTime(), IsActive: true, LastActive: ctx.BlockTime()},
		},
		CurrentGroupID: tss.GroupID(1),
		SigningCount:   1,
		Signings: []types.Signing{
			{ID: types.SigningID(1), Requester: bandtesting.Alice.Address.String(), CurrentGroupSigningID: tss.SigningID(3)},
		},
		SigningIDMappings: []types.SigningIDMappingGenesis{
			{SigningID: tss.SigningID(1), BandtssSigningID: types.SigningID(3)},
		},
		Replacement: types.Replacement{
			SigningID:      tss.SigningID(1),
			CurrentGroupID: tss.GroupID(1),
			NewGroupID:     tss.GroupID(2),
			Status:         types.REPLACEMENT_STATUS_WAITING_SIGN,
		},
	}

	bandtss.InitGenesis(ctx, k, &data)

	exported := bandtss.ExportGenesis(ctx, k)
	require.Equal(t, data.Params, exported.Params)
}
