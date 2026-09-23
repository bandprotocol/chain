package simulation_test

import (
	"encoding/hex"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	abci "github.com/cometbft/cometbft/abci/types"

	"github.com/cosmos/gogoproto/proto"

	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	band "github.com/bandprotocol/chain/v3/app"
	bandtesting "github.com/bandprotocol/chain/v3/testing"
	"github.com/bandprotocol/chain/v3/testing/testdata"
	"github.com/bandprotocol/chain/v3/x/oracle/simulation"
	"github.com/bandprotocol/chain/v3/x/oracle/types"
)

type SimTestSuite struct {
	suite.Suite

	ctx  sdk.Context
	app  *band.BandApp
	r    *rand.Rand
	accs []simtypes.Account
}

func init() {
	band.SetBech32AddressPrefixesAndBip44CoinTypeAndSeal(sdk.GetConfig())
}

func (suite *SimTestSuite) SetupTest() {
	dir := testutil.GetTempDir(suite.T())
	suite.app = bandtesting.SetupWithCustomHome(false, dir)
	suite.ctx = suite.app.BaseApp.NewContext(false).WithChainID(bandtesting.ChainID)

	s := rand.NewSource(1)
	suite.r = rand.New(s)
	suite.accs = suite.getTestingAccounts(suite.r, 10)

	_, err := suite.app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: suite.app.LastBlockHeight() + 1,
		Hash:   suite.app.LastCommitID().Hash,
	})
	suite.NoError(err)
}

// TestWeightedOperations tests the weights of the operations.
func (suite *SimTestSuite) TestWeightedOperations() {
	cdc := suite.app.AppCodec()
	appParams := make(simtypes.AppParams)

	weightesOps := simulation.WeightedOperations(
		appParams,
		cdc,
		suite.app.AccountKeeper,
		suite.app.BankKeeper,
		suite.app.StakingKeeper,
		suite.app.OracleKeeper,
	)

	expected := []struct {
		weight    int
		opMsgName string
	}{
		{simulation.DefaultWeightMsgRequestData, sdk.MsgTypeURL(&types.MsgRequestData{})},
		{simulation.DefaultWeightMsgReportData, sdk.MsgTypeURL(&types.MsgReportData{})},
		{simulation.DefaultWeightMsgActivate, sdk.MsgTypeURL(&types.MsgActivate{})},
	}

	for i, w := range weightesOps {
		operationMsg, _, _ := w.Op()(suite.r, suite.app.BaseApp, suite.ctx, suite.accs, "")
		// the following checks are very much dependent from the ordering of the output given
		// by WeightedOperations. if the ordering in WeightedOperations changes some tests
		// will fail
		suite.Require().Equal(expected[i].weight, w.Weight(), "weight should be the same")
		suite.Require().Equal(expected[i].opMsgName, operationMsg.Name, "operation Msg name should be the same")
	}
}

// TestSimulateMsgRequestData tests the normal scenario of a valid message of type TypeMsgRequestData
func (suite *SimTestSuite) TestSimulateMsgRequestData() {
	// Prepare oracle script for request, owned by one of the simulation accounts.
	oracleScriptFilename, err := suite.app.OracleKeeper.AddOracleScriptFile(testdata.Wasm1)
	suite.Require().NoError(err)
	suite.app.OracleKeeper.AddOracleScript(
		suite.ctx,
		types.NewOracleScript(
			suite.accs[0].Address,
			"name",
			"description",
			oracleScriptFilename,
			"schema",
			"sourceCodeURL",
		),
	)

	// Prepare data sources 1-3 for request. The test oracle script asks for data from
	// these data sources, and the generated request carries no fee, so their fee is zero.
	executableFilename := suite.app.OracleKeeper.AddExecutableFile([]byte("executable"))
	for i := 1; i <= 3; i++ {
		suite.app.OracleKeeper.SetDataSource(
			suite.ctx,
			types.DataSourceID(i),
			types.NewDataSource(
				suite.accs[0].Address,
				"name",
				"description",
				executableFilename,
				sdk.NewCoins(),
				suite.accs[0].Address,
			),
		)
	}

	// Prepare active validators
	err = suite.app.StakingKeeper.IterateBondedValidatorsByPower(suite.ctx,
		func(idx int64, val stakingtypes.ValidatorI) (stop bool) {
			operator, err := sdk.ValAddressFromBech32(val.GetOperator())
			if err != nil {
				return false
			}

			_ = suite.app.OracleKeeper.Activate(suite.ctx, operator)

			return false
		},
	)
	suite.Require().NoError(err)

	// Simulate MsgRequestData
	op := simulation.SimulateMsgRequestData(
		suite.app.AccountKeeper,
		suite.app.BankKeeper,
		suite.app.StakingKeeper,
		suite.app.OracleKeeper,
	)
	operationMsg, futureOperations, err := op(suite.r, suite.app.BaseApp, suite.ctx, suite.accs, "")
	suite.Require().NoError(err)

	// Verify the fields of the message
	var msg types.MsgRequestData
	err = proto.Unmarshal(operationMsg.Msg, &msg)
	suite.Require().NoError(err)

	suite.Require().True(operationMsg.OK)
	suite.Require().Equal(types.OracleScriptID(10), msg.OracleScriptID)
	suite.Require().
		Equal("424f5249755a716f586c5a755476416a45646c4557444f444652726567445471474e6f464249487876696d6d495a774c6646794b55664557416e4e426474647a446d545058747048524764496275756366546a4f79675a735478506a667765586853556b", hex.EncodeToString(msg.Calldata))
	suite.Require().Equal(uint64(2), msg.AskCount)
	suite.Require().Equal(uint64(2), msg.MinCount)
	suite.Require().
		Equal("UScOXvYthRXpPfKwMhptXaxIxgqBoUqzrWbaoLTVpQoottZyPFfNOoMioXHRuFwMRYUiKvcWPkrayyTLOCFJlAyslDameIuqVAux", msg.ClientID)
	suite.Require().Equal(sdk.Coins(nil), msg.FeeLimit)
	suite.Require().Equal(uint64(127189), msg.PrepareGas)
	suite.Require().Equal(uint64(182199), msg.ExecuteGas)
	suite.Require().Equal("band1n5sqxutsmk6eews5z9z673wv7n9wah8hjlxyuf", msg.Sender)
	suite.Require().Equal(sdk.MsgTypeURL(&types.MsgRequestData{}), sdk.MsgTypeURL(&msg))
	suite.Require().Len(futureOperations, 0)
}

// TestSimulateMsgReportData tests the normal scenario of a valid message of type TypeMsgReportData
func (suite *SimTestSuite) TestSimulateMsgReportData() {
	// Prepare request that we will simulate to send report to
	suite.app.OracleKeeper.AddRequest(
		suite.ctx,
		types.NewRequest(types.OracleScriptID(1),
			[]byte("calldata"),
			[]sdk.ValAddress{sdk.ValAddress(suite.accs[0].Address)},
			1,
			1,
			time.Now().UTC(),
			"clientID",
			[]types.RawRequest{
				types.NewRawRequest(types.ExternalID(1), types.DataSourceID(1), []byte("data")),
				types.NewRawRequest(types.ExternalID(2), types.DataSourceID(2), []byte("data")),
				types.NewRawRequest(types.ExternalID(3), types.DataSourceID(3), []byte("data")),
			},
			nil,
			300000,
			0,
			suite.accs[0].PubKey.String(),
			sdk.NewCoins(sdk.NewInt64Coin("band", 1000)),
		),
	)

	// Simulate MsgReportData
	op := simulation.SimulateMsgReportData(
		suite.app.AccountKeeper,
		suite.app.BankKeeper,
		suite.app.StakingKeeper,
		suite.app.OracleKeeper,
	)
	operationMsg, futureOperations, err := op(suite.r, suite.app.BaseApp, suite.ctx, suite.accs, "")
	suite.Require().NoError(err)

	// Verify the fields of the message
	var msg types.MsgReportData
	err = proto.Unmarshal(operationMsg.Msg, &msg)
	suite.Require().NoError(err)

	suite.Require().True(operationMsg.OK)
	suite.Require().Equal(types.RequestID(1), msg.RequestID)
	suite.Require().Equal(3, len(msg.RawReports))
	suite.Require().Equal("bandvaloper1tnh2q55v8wyygtt9srz5safamzdengsn4qqe0j", msg.Validator)
	suite.Require().Equal(sdk.MsgTypeURL(&types.MsgReportData{}), sdk.MsgTypeURL(&msg))
	suite.Require().Len(futureOperations, 0)
}

// TestSimulateMsgActivate tests the normal scenario of a valid message of type TypeMsgActivate
func (suite *SimTestSuite) TestSimulateMsgActivate() {
	// Simulate MsgActivate
	op := simulation.SimulateMsgActivate(
		suite.app.AccountKeeper,
		suite.app.BankKeeper,
		suite.app.StakingKeeper,
		suite.app.OracleKeeper,
	)
	operationMsg, futureOperations, err := op(suite.r, suite.app.BaseApp, suite.ctx, suite.accs, "")
	suite.Require().NoError(err)

	// Verify the fields of the message
	var msg types.MsgActivate
	err = proto.Unmarshal(operationMsg.Msg, &msg)
	suite.Require().NoError(err)

	suite.Require().True(operationMsg.OK)
	suite.Require().Equal("bandvaloper1n5sqxutsmk6eews5z9z673wv7n9wah8h7fz8ez", msg.Validator)
	suite.Require().Equal(sdk.MsgTypeURL(&types.MsgActivate{}), sdk.MsgTypeURL(&msg))
	suite.Require().Len(futureOperations, 0)
}

func (suite *SimTestSuite) getTestingAccounts(r *rand.Rand, n int) []simtypes.Account {
	accounts := simtypes.RandomAccounts(r, n)

	initAmt := sdk.TokensFromConsensusPower(200, sdk.DefaultPowerReduction)
	initCoins := sdk.NewCoins(sdk.NewCoin("uband", initAmt))

	// add coins to the accounts
	for _, account := range accounts {
		acc := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, account.Address)
		suite.app.AccountKeeper.SetAccount(suite.ctx, acc)
		suite.Require().
			NoError(banktestutil.FundAccount(suite.ctx, suite.app.BankKeeper, account.Address, initCoins))
	}

	return accounts
}

func TestSimTestSuite(t *testing.T) {
	suite.Run(t, new(SimTestSuite))
}
