package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	tmtime "github.com/cometbft/cometbft/types/time"

	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/bandprotocol/chain/v3/x/feeds/keeper"
	feedstestutil "github.com/bandprotocol/chain/v3/x/feeds/testutil"
	"github.com/bandprotocol/chain/v3/x/feeds/types"
	oracletypes "github.com/bandprotocol/chain/v3/x/oracle/types"
	restaketypes "github.com/bandprotocol/chain/v3/x/restake/types"
)

var (
	ValidValidator   = sdk.ValAddress("1000000001")
	ValidValidator2  = sdk.ValAddress("1000000002")
	ValidValidator3  = sdk.ValAddress("1000000003")
	ValidVoter       = sdk.AccAddress("2000000001")
	ValidVoter2      = sdk.AccAddress("2000000002")
	ValidFeeder      = sdk.AccAddress("3000000001")
	InvalidValidator = sdk.ValAddress("9000000001")
	InvalidVoter     = sdk.AccAddress("9000000002")
)

type KeeperTestSuite struct {
	suite.Suite

	ctx           sdk.Context
	feedsKeeper   keeper.Keeper
	oracleKeeper  *feedstestutil.MockOracleKeeper
	stakingKeeper *feedstestutil.MockStakingKeeper
	restakeKeeper *feedstestutil.MockRestakeKeeper

	queryClient types.QueryClient
	msgServer   types.MsgServer
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) SetupTest() {
	key := storetypes.NewKVStoreKey(types.StoreKey)
	testCtx := testutil.DefaultContextWithDB(suite.T(), key, storetypes.NewTransientStoreKey("transient_test"))
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
		GetValidatorStatus(gomock.Any(), gomock.Eq(ValidValidator3)).
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
		Return(stakingtypes.Validator{Status: stakingtypes.Bonded}, nil).
		AnyTimes()
	stakingKeeper.EXPECT().
		GetValidator(gomock.Any(), gomock.Eq(ValidValidator2)).
		Return(stakingtypes.Validator{Status: stakingtypes.Bonded}, nil).
		AnyTimes()
	stakingKeeper.EXPECT().
		GetValidator(gomock.Any(), gomock.Eq(ValidValidator3)).
		Return(stakingtypes.Validator{Status: stakingtypes.Bonded}, nil).
		AnyTimes()
	stakingKeeper.EXPECT().
		GetValidator(gomock.Any(), gomock.Eq(InvalidValidator)).
		Return(stakingtypes.Validator{Status: stakingtypes.Unbonded}, nil).
		AnyTimes()
	stakingKeeper.EXPECT().
		IterateBondedValidatorsByPower(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx sdk.Context, fn func(index int64, validator stakingtypes.ValidatorI) bool) error {
			vals := []stakingtypes.Validator{
				{
					OperatorAddress: ValidValidator.String(),
					Tokens:          math.NewInt(5000),
				},
				{
					OperatorAddress: ValidValidator2.String(),
					Tokens:          math.NewInt(3000),
				},
				{
					OperatorAddress: ValidValidator3.String(),
					Tokens:          math.NewInt(3000),
				},
			}

			for i, val := range vals {
				stop := fn(int64(i), val)
				if stop {
					break
				}
			}

			return nil
		}).
		AnyTimes()
	stakingKeeper.EXPECT().
		TotalBondedTokens(gomock.Any()).
		Return(math.NewInt(11000), nil).
		AnyTimes()

	suite.stakingKeeper = stakingKeeper

	restakeKeeper := feedstestutil.NewMockRestakeKeeper(ctrl)
	restakeKeeper.EXPECT().
		SetLockedPower(gomock.Any(), ValidVoter, types.ModuleName, gomock.Any()).
		DoAndReturn(func(_ sdk.Context, _ sdk.AccAddress, _ string, amount math.Int) error {
			if amount.GT(math.NewInt(1e10)) {
				return restaketypes.ErrPowerNotEnough
			}
			return nil
		}).
		AnyTimes()
	restakeKeeper.EXPECT().
		SetLockedPower(gomock.Any(), InvalidVoter, types.ModuleName, gomock.Any()).
		Return(restaketypes.ErrPowerNotEnough).
		AnyTimes()
	suite.restakeKeeper = restakeKeeper

	authzKeeper := feedstestutil.NewMockAuthzKeeper(ctrl)

	suite.feedsKeeper = keeper.NewKeeper(
		encCfg.Codec,
		key,
		oracleKeeper,
		stakingKeeper,
		restakeKeeper,
		authzKeeper,
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
