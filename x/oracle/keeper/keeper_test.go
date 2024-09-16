package keeper_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cmttime "github.com/cometbft/cometbft/types/time"

	capabilitykeeper "github.com/cosmos/ibc-go/modules/capability/keeper"

	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
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

	encCfg  moduletestutil.TestEncodingConfig
	fileDir string
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

	var err error
	suite.fileDir, err = os.MkdirTemp(".", "files-*")
	suite.Require().NoError(err)

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
}

func (suite *KeeperTestSuite) TearDownTest() {
	os.RemoveAll(suite.fileDir)
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
