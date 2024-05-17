package keeper_test

import (
	"testing"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	tmtime "github.com/cometbft/cometbft/types/time"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/bandprotocol/chain/v2/x/feeds/keeper"
	feedstestutil "github.com/bandprotocol/chain/v2/x/feeds/testutil"
	"github.com/bandprotocol/chain/v2/x/feeds/types"
	oracletypes "github.com/bandprotocol/chain/v2/x/oracle/types"
)

var (
	ValidValidator   = sdk.ValAddress("1234567890")
	ValidValidator2  = sdk.ValAddress("2345678901")
	ValidDelegator   = sdk.AccAddress("3456789012")
	ValidDelegator2  = sdk.AccAddress("4567890123")
	InvalidValidator = sdk.ValAddress("9876543210")
	InvalidDelegator = sdk.AccAddress("8765432109")
)

type KeeperTestSuite struct {
	suite.Suite

	ctx           sdk.Context
	feedsKeeper   keeper.Keeper
	oracleKeeper  *feedstestutil.MockOracleKeeper
	stakingKeeper *feedstestutil.MockStakingKeeper

	queryClient types.QueryClient
	msgServer   types.MsgServer
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) SetupTest() {
	key := sdk.NewKVStoreKey(types.StoreKey)
	testCtx := testutil.DefaultContextWithDB(suite.T(), key, sdk.NewTransientStoreKey("transient_test"))
	suite.ctx = testCtx.Ctx.WithBlockHeader(tmproto.Header{Time: tmtime.Now()})
	encCfg := moduletestutil.MakeTestEncodingConfig()

	// gomock initializations
	ctrl := gomock.NewController(suite.T())
	oracleKeeper := feedstestutil.NewMockOracleKeeper(ctrl)
	oracleKeeper.EXPECT().
		GetValidatorStatus(gomock.Any(), gomock.Eq(ValidValidator)).
		Return(oracletypes.NewValidatorStatus(true, suite.ctx.BlockHeader().Time)).
		AnyTimes()
	oracleKeeper.EXPECT().
		GetValidatorStatus(gomock.Any(), gomock.Eq(ValidValidator2)).
		Return(oracletypes.NewValidatorStatus(true, suite.ctx.BlockHeader().Time)).
		AnyTimes()
	oracleKeeper.EXPECT().
		GetValidatorStatus(gomock.Any(), gomock.Eq(InvalidValidator)).
		Return(oracletypes.NewValidatorStatus(false, suite.ctx.BlockHeader().Time)).
		AnyTimes()
	suite.oracleKeeper = oracleKeeper

	stakingKeeper := feedstestutil.NewMockStakingKeeper(ctrl)
	stakingKeeper.EXPECT().
		GetValidator(gomock.Any(), gomock.Eq(ValidValidator)).
		Return(stakingtypes.Validator{Status: stakingtypes.Bonded}, true).
		AnyTimes()
	stakingKeeper.EXPECT().
		GetValidator(gomock.Any(), gomock.Eq(ValidValidator2)).
		Return(stakingtypes.Validator{Status: stakingtypes.Bonded}, true).
		AnyTimes()
	stakingKeeper.EXPECT().
		GetValidator(gomock.Any(), gomock.Eq(InvalidValidator)).
		Return(stakingtypes.Validator{Status: stakingtypes.Unbonded}, true).
		AnyTimes()
	stakingKeeper.EXPECT().
		IterateBondedValidatorsByPower(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx sdk.Context, fn func(index int64, validator stakingtypes.ValidatorI) bool) {
			vals := []stakingtypes.Validator{
				{
					OperatorAddress: ValidValidator.String(),
					Tokens:          sdk.NewInt(5000),
				},
				{
					OperatorAddress: ValidValidator2.String(),
					Tokens:          sdk.NewInt(3000),
				},
			}

			for i, val := range vals {
				stop := fn(int64(i), val)
				if stop {
					break
				}
			}
		}).
		AnyTimes()
	stakingKeeper.EXPECT().
		GetDelegatorBonded(gomock.Any(), ValidDelegator).
		Return(sdk.NewInt(1e10)).
		AnyTimes()
	stakingKeeper.EXPECT().
		GetDelegatorBonded(gomock.Any(), InvalidDelegator).
		Return(sdk.NewInt(0)).
		AnyTimes()
	suite.stakingKeeper = stakingKeeper

	suite.feedsKeeper = keeper.NewKeeper(
		encCfg.Codec,
		key,
		oracleKeeper,
		stakingKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	suite.feedsKeeper.InitGenesis(suite.ctx, *types.DefaultGenesisState())

	queryHelper := baseapp.NewQueryServerTestHelper(suite.ctx, encCfg.InterfaceRegistry)
	queryServer := keeper.NewQueryServer(suite.feedsKeeper)

	types.RegisterInterfaces(encCfg.InterfaceRegistry)
	types.RegisterQueryServer(queryHelper, queryServer)
	queryClient := types.NewQueryClient(queryHelper)
	suite.queryClient = queryClient
	suite.msgServer = keeper.NewMsgServerImpl(suite.feedsKeeper)
}
