package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cmttime "github.com/cometbft/cometbft/types/time"

	capabilitykeeper "github.com/cosmos/ibc-go/modules/capability/keeper"

	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	owasm "github.com/bandprotocol/go-owasm/api"

	"github.com/bandprotocol/chain/v3/x/oracle/keeper"
	oracletestutil "github.com/bandprotocol/chain/v3/x/oracle/testutil"
	"github.com/bandprotocol/chain/v3/x/oracle/types"
)

type KeeperTestSuite struct {
	suite.Suite

	ctx           sdk.Context
	oracleKeeper  keeper.Keeper
	authKeeper    *oracletestutil.MockAccountKeeper
	bankKeeper    *oracletestutil.MockBankKeeper
	stakingKeeper *oracletestutil.MockStakingKeeper
	distrKeeper   *oracletestutil.MockDistrKeeper
	authzKeeper   *oracletestutil.MockAuthzKeeper

	queryClient types.QueryClient
	msgServer   types.MsgServer

	fileDir string

	encCfg moduletestutil.TestEncodingConfig
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) SetupTest() {
	key := storetypes.NewKVStoreKey(types.StoreKey)
	testCtx := testutil.DefaultContextWithDB(suite.T(), key, storetypes.NewTransientStoreKey("transient_test"))
	ctx := testCtx.Ctx.WithBlockHeader(cmtproto.Header{Time: cmttime.Now()})
	encCfg := moduletestutil.MakeTestEncodingConfig()

	// gomock initializations
	ctrl := gomock.NewController(suite.T())
	suite.authKeeper = oracletestutil.NewMockAccountKeeper(ctrl)
	suite.bankKeeper = oracletestutil.NewMockBankKeeper(ctrl)
	suite.stakingKeeper = oracletestutil.NewMockStakingKeeper(ctrl)
	suite.distrKeeper = oracletestutil.NewMockDistrKeeper(ctrl)
	suite.authzKeeper = oracletestutil.NewMockAuthzKeeper(ctrl)

	suite.fileDir = testutil.GetTempDir(suite.T())

	owasmVM, err := owasm.NewVm(100)
	suite.Require().NoError(err)

	suite.ctx = ctx
	suite.oracleKeeper = keeper.NewKeeper(
		encCfg.Codec,
		key,
		suite.fileDir,
		authtypes.FeeCollectorName,
		suite.authKeeper,
		suite.bankKeeper,
		suite.stakingKeeper,
		suite.distrKeeper,
		suite.authzKeeper,
		nil,
		nil,
		capabilitykeeper.ScopedKeeper{},
		owasmVM,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	suite.oracleKeeper.SetRequestCount(ctx, 0)
	suite.oracleKeeper.SetDataSourceCount(ctx, 0)
	suite.oracleKeeper.SetOracleScriptCount(ctx, 0)
	suite.oracleKeeper.SetRequestLastExpired(ctx, 0)

	err = suite.oracleKeeper.SetParams(ctx, types.DefaultParams())
	suite.Require().NoError(err)

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, encCfg.InterfaceRegistry)
	types.RegisterQueryServer(queryHelper, keeper.Querier{
		Keeper: suite.oracleKeeper,
	})
	suite.queryClient = types.NewQueryClient(queryHelper)

	suite.authzKeeper.EXPECT().
		GetAuthorization(gomock.Any(), reporterAddr, sdk.AccAddress(validators[0].Address), sdk.MsgTypeURL(&types.MsgReportData{})).
		Return(authz.NewGenericAuthorization(sdk.MsgTypeURL(&types.MsgReportData{})), nil).
		AnyTimes()
	suite.authzKeeper.EXPECT().
		GetAuthorization(gomock.Any(), reporterAddr, sdk.AccAddress(validators[1].Address), sdk.MsgTypeURL(&types.MsgReportData{})).
		Return(nil, nil).
		AnyTimes()

	authorization, err := codectypes.NewAnyWithValue(
		authz.NewGenericAuthorization(sdk.MsgTypeURL(&types.MsgReportData{})),
	)
	if err != nil {
		panic(err)
	}
	expiration := ctx.BlockTime().Add(time.Minute)
	suite.authzKeeper.EXPECT().
		GranterGrants(gomock.Any(), &authz.QueryGranterGrantsRequest{
			Granter: sdk.AccAddress(validators[0].Address).String(),
		}).
		Return(&authz.QueryGranterGrantsResponse{
			Grants: []*authz.GrantAuthorization{
				{
					Granter:       sdk.AccAddress(validators[0].Address).String(),
					Grantee:       reporterAddr.String(),
					Authorization: authorization,
					Expiration:    &expiration,
				},
			},
		}, nil).
		AnyTimes()
}

func (suite *KeeperTestSuite) activeAllValidators() {
	ctx := suite.ctx
	k := suite.oracleKeeper

	for _, v := range validators {
		err := k.Activate(ctx, v.Address)
		suite.Require().NoError(err)
	}
}

func (suite *KeeperTestSuite) TestGetSetRequestCount() {
	ctx := suite.ctx
	k := suite.oracleKeeper
	require := suite.Require()
	// Initially request count must be 0.
	require.Equal(uint64(0), k.GetRequestCount(ctx))
	// After we set the count manually, it should be reflected.
	k.SetRequestCount(suite.ctx, 42)
	require.Equal(uint64(42), k.GetRequestCount(ctx))
}

func (suite *KeeperTestSuite) TestGetDataSourceCount() {
	ctx := suite.ctx
	k := suite.oracleKeeper
	require := suite.Require()

	k.SetDataSourceCount(ctx, 42)
	require.Equal(uint64(42), k.GetDataSourceCount(ctx))
}

func (suite *KeeperTestSuite) TestGetSetOracleScriptCount() {
	ctx := suite.ctx
	k := suite.oracleKeeper
	require := suite.Require()

	k.SetOracleScriptCount(ctx, 42)
	require.Equal(uint64(42), k.GetOracleScriptCount(ctx))
}

func (suite *KeeperTestSuite) TestGetSetRollingSeed() {
	ctx := suite.ctx
	k := suite.oracleKeeper
	require := suite.Require()

	k.SetRollingSeed(ctx, []byte("HELLO_WORLD"))
	require.Equal([]byte("HELLO_WORLD"), k.GetRollingSeed(ctx))
}

func (suite *KeeperTestSuite) TestGetNextRequestID() {
	ctx := suite.ctx
	k := suite.oracleKeeper
	require := suite.Require()

	// First request id must be 1.
	require.Equal(types.RequestID(1), k.GetNextRequestID(ctx))
	// After we add new requests, the request count must increase accordingly.
	require.Equal(uint64(1), k.GetRequestCount(ctx))
	require.Equal(types.RequestID(2), k.GetNextRequestID(ctx))
	require.Equal(types.RequestID(3), k.GetNextRequestID(ctx))
	require.Equal(types.RequestID(4), k.GetNextRequestID(ctx))
	require.Equal(uint64(4), k.GetRequestCount(ctx))
}

func (suite *KeeperTestSuite) TestGetNextDataSourceID() {
	ctx := suite.ctx
	k := suite.oracleKeeper
	require := suite.Require()

	initialID := k.GetDataSourceCount(ctx)
	require.Equal(types.DataSourceID(initialID+1), k.GetNextDataSourceID(ctx))
	require.Equal(types.DataSourceID(initialID+2), k.GetNextDataSourceID(ctx))
	require.Equal(types.DataSourceID(initialID+3), k.GetNextDataSourceID(ctx))
}

func (suite *KeeperTestSuite) TestGetNextOracleScriptID() {
	ctx := suite.ctx
	k := suite.oracleKeeper
	require := suite.Require()

	initialID := k.GetOracleScriptCount(ctx)
	require.Equal(types.OracleScriptID(initialID+1), k.GetNextOracleScriptID(ctx))
	require.Equal(types.OracleScriptID(initialID+2), k.GetNextOracleScriptID(ctx))
	require.Equal(types.OracleScriptID(initialID+3), k.GetNextOracleScriptID(ctx))
}

func (suite *KeeperTestSuite) TestGetSetRequestLastExpiredID() {
	ctx := suite.ctx
	k := suite.oracleKeeper
	require := suite.Require()

	// Initially last expired request must be 0.
	require.Equal(types.RequestID(0), k.GetRequestLastExpired(ctx))
	k.SetRequestLastExpired(ctx, 20)
	require.Equal(types.RequestID(20), k.GetRequestLastExpired(ctx))
}
