package simulation_test

import (
	"encoding/hex"
	"math/rand"
	"testing"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"
	"github.com/stretchr/testify/suite"

	"github.com/bandprotocol/chain/v2/testing/testapp"
	"github.com/bandprotocol/chain/v2/x/oracle/simulation"
	"github.com/bandprotocol/chain/v2/x/oracle/types"
)

type SimTestSuite struct {
	suite.Suite

	ctx  sdk.Context
	app  *testapp.TestingApp
	r    *rand.Rand
	accs []simtypes.Account
}

func (suite *SimTestSuite) SetupTest() {
	app, _, _ := testapp.CreateTestInput(true)
	suite.app = app
	suite.ctx = app.BaseApp.NewContext(false, tmproto.Header{ChainID: testapp.ChainID})
	s := rand.NewSource(1)
	suite.r = rand.New(s)
	suite.accs = suite.getTestingAccounts(suite.r, 10)

	// begin a new block
	suite.app.BeginBlock(
		abci.RequestBeginBlock{
			Header: tmproto.Header{
				ChainID: testapp.ChainID,
				Height:  suite.app.LastBlockHeight() + 1,
				AppHash: suite.app.LastCommitID().Hash,
			},
		},
	)
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
		weight     int
		opMsgRoute string
		opMsgName  string
	}{
		{simulation.DefaultWeightMsgRequestData, types.ModuleName, types.TypeMsgRequestData},
		{simulation.DefaultWeightMsgReportData, types.ModuleName, types.TypeMsgReportData},
		{simulation.DefaultWeightMsgCreateDataSource, types.ModuleName, types.TypeMsgCreateDataSource},
		{simulation.DefaultWeightMsgEditDataSource, types.ModuleName, types.TypeMsgEditDataSource},
		{simulation.DefaultWeightMsgCreateOracleScript, types.ModuleName, types.TypeMsgCreateOracleScript},
		{simulation.DefaultWeightMsgEditOracleScript, types.ModuleName, types.TypeMsgEditOracleScript},
		{simulation.DefaultWeightMsgActivate, types.ModuleName, types.TypeMsgActivate},
	}

	for i, w := range weightesOps {
		operationMsg, _, _ := w.Op()(suite.r, suite.app.BaseApp, suite.ctx, suite.accs, "")
		// the following checks are very much dependent from the ordering of the output given
		// by WeightedOperations. if the ordering in WeightedOperations changes some tests
		// will fail
		suite.Require().Equal(expected[i].weight, w.Weight(), "weight should be the same")
		suite.Require().Equal(expected[i].opMsgRoute, operationMsg.Route, "route should be the same")
		suite.Require().Equal(expected[i].opMsgName, operationMsg.Name, "operation Msg name should be the same")
	}
}

// TestSimulateMsgRequestData tests the normal scenario of a valid message of type TypeMsgRequestData
func (suite *SimTestSuite) TestSimulateMsgRequestData() {
	// Prepare oracle script for request
	suite.TestSimulateMsgCreateOracleScript()
	// Prepare data sources for request
	for i := 1; i <= 3; i++ {
		ds, _ := suite.app.OracleKeeper.GetDataSource(suite.ctx, types.DataSourceID(i))
		ds.Fee = sdk.NewCoins()
		suite.app.OracleKeeper.SetDataSource(suite.ctx, types.DataSourceID(i), ds)
	}

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
	err = types.AminoCdc.UnmarshalJSON(operationMsg.Msg, &msg)
	suite.Require().NoError(err)

	suite.Require().True(operationMsg.OK)
	suite.Require().Equal(types.OracleScriptID(10), msg.OracleScriptID)
	suite.Require().
		Equal("6f7857727a526e54566a5374506164687345536c45526e4b68704550736644784e767871634f7949756c61436b6d5064616d624c48764768545a7a7973767146617545676b4652497450667669736568466d6f426851716d6b6662485673676648584450", hex.EncodeToString(msg.Calldata))
	suite.Require().Equal(uint64(3), msg.AskCount)
	suite.Require().Equal(uint64(3), msg.MinCount)
	suite.Require().
		Equal("RTRnuwdBeuOGgFbJLbDksHVapaRayWzwoYBEpmrlAxrUxYMUekKbpjPNfjUCjhbdMAnJmYQVZBQZkFVweHDAlaqJjRqoQPoOMLhy", msg.ClientID)
	suite.Require().Equal(sdk.Coins(nil), msg.FeeLimit)
	suite.Require().Equal(uint64(169271), msg.PrepareGas)
	suite.Require().Equal(uint64(115894), msg.ExecuteGas)
	suite.Require().Equal("band1ghekyjucln7y67ntx7cf27m9dpuxxemnvh82dt", msg.Sender)
	suite.Require().Equal(types.TypeMsgRequestData, msg.Type())
	suite.Require().Equal(types.ModuleName, msg.Route())
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
	err = types.AminoCdc.UnmarshalJSON(operationMsg.Msg, &msg)
	suite.Require().NoError(err)

	suite.Require().True(operationMsg.OK)
	suite.Require().Equal(types.RequestID(1), msg.RequestID)
	suite.Require().Equal(3, len(msg.RawReports))
	suite.Require().Equal("bandvaloper1tnh2q55v8wyygtt9srz5safamzdengsn4qqe0j", msg.Validator)
	suite.Require().Equal(types.TypeMsgReportData, msg.Type())
	suite.Require().Equal(types.ModuleName, msg.Route())
	suite.Require().Len(futureOperations, 0)
}

// TestSimulateMsgCreateDataSource tests the normal scenario of a valid message of type TypeMsgCreateDataSource
func (suite *SimTestSuite) TestSimulateMsgCreateDataSource() {
	// Simulate MsgCreateDataSource
	op := simulation.SimulateMsgCreateDataSource(
		suite.app.AccountKeeper,
		suite.app.BankKeeper,
		suite.app.StakingKeeper,
		suite.app.OracleKeeper,
	)
	operationMsg, futureOperations, err := op(suite.r, suite.app.BaseApp, suite.ctx, suite.accs, "")
	suite.Require().NoError(err)

	// Verify the fields of the message
	var msg types.MsgCreateDataSource
	err = types.AminoCdc.UnmarshalJSON(operationMsg.Msg, &msg)
	suite.Require().NoError(err)

	suite.Require().True(operationMsg.OK)
	suite.Require().Equal("band1n5sqxutsmk6eews5z9z673wv7n9wah8hjlxyuf", msg.Sender)
	suite.Require().Equal("OygZsTxPjf", msg.Name)
	suite.Require().
		Equal("lDameIuqVAuxErqFPEWIScKpBORIuZqoXlZuTvAjEdlEWDODFRregDTqGNoFBIHxvimmIZwLfFyKUfEWAnNBdtdzDmTPXtpHRGdI", msg.Description)
	suite.Require().
		Equal("4f6a4346754976547968584b4c79685553634f587659746852587050664b774d68707458617849786771426f55717a725762616f4c545670516f6f74745a795046664e4f6f4d696f5848527546774d525955694b766357506b72617979544c4f43464a6c", hex.EncodeToString(msg.Executable))
	suite.Require().Equal(sdk.Coins(nil), msg.Fee)
	suite.Require().Equal("band13rmqzzysyz4qh3yg6rvknd6u9rvrd98qvy9azu", msg.Treasury)
	suite.Require().Equal("band1n5sqxutsmk6eews5z9z673wv7n9wah8hjlxyuf", msg.Owner)
	suite.Require().Equal(types.TypeMsgCreateDataSource, msg.Type())
	suite.Require().Equal(types.ModuleName, msg.Route())
	suite.Require().Len(futureOperations, 0)
}

// TestSimulateMsgEditDataSource tests the normal scenario of a valid message of type TypeMsgEditDataSource
func (suite *SimTestSuite) TestSimulateMsgEditDataSource() {
	// Prepare data source for us to edit by message
	suite.app.OracleKeeper.SetDataSource(
		suite.ctx,
		1,
		types.NewDataSource(
			suite.accs[0].Address,
			"name",
			"description",
			"filename",
			sdk.NewCoins(),
			suite.accs[0].Address,
		),
	)

	// Simulate MsgEditDataSource
	op := simulation.SimulateMsgEditDataSource(
		suite.app.AccountKeeper,
		suite.app.BankKeeper,
		suite.app.StakingKeeper,
		suite.app.OracleKeeper,
	)
	operationMsg, futureOperations, err := op(suite.r, suite.app.BaseApp, suite.ctx, suite.accs, "")
	suite.Require().NoError(err)

	// Verify the fields of the message
	var msg types.MsgEditDataSource
	err = types.AminoCdc.UnmarshalJSON(operationMsg.Msg, &msg)
	suite.Require().NoError(err)

	suite.Require().True(operationMsg.OK)
	suite.Require().Equal(types.DataSourceID(1), msg.DataSourceID)
	suite.Require().Equal("band1tnh2q55v8wyygtt9srz5safamzdengsneky62e", msg.Sender)
	suite.Require().Equal("PjfweXhSUk", msg.Name)
	suite.Require().
		Equal("VAuxErqFPEWIScKpBORIuZqoXlZuTvAjEdlEWDODFRregDTqGNoFBIHxvimmIZwLfFyKUfEWAnNBdtdzDmTPXtpHRGdIbuucfTjO", msg.Description)
	suite.Require().
		Equal("7968584b4c79685553634f587659746852587050664b774d68707458617849786771426f55717a725762616f4c545670516f6f74745a795046664e4f6f4d696f5848527546774d525955694b766357506b72617979544c4f43464a6c4179736c44616d65", hex.EncodeToString(msg.Executable))
	suite.Require().Equal(sdk.Coins(nil), msg.Fee)
	suite.Require().Equal("band1n5sqxutsmk6eews5z9z673wv7n9wah8hjlxyuf", msg.Treasury)
	suite.Require().Equal("band1n5sqxutsmk6eews5z9z673wv7n9wah8hjlxyuf", msg.Owner)
	suite.Require().Equal(types.TypeMsgEditDataSource, msg.Type())
	suite.Require().Equal(types.ModuleName, msg.Route())
	suite.Require().Len(futureOperations, 0)
}

// TestSimulateMsgCreateOracleScript tests the normal scenario of a valid message of type TypeMsgCreateOracleScript
func (suite *SimTestSuite) TestSimulateMsgCreateOracleScript() {
	// Simulate MsgCreateOracleScript
	op := simulation.SimulateMsgCreateOracleScript(
		suite.app.AccountKeeper,
		suite.app.BankKeeper,
		suite.app.StakingKeeper,
		suite.app.OracleKeeper,
	)
	operationMsg, futureOperations, err := op(suite.r, suite.app.BaseApp, suite.ctx, suite.accs, "")
	suite.Require().NoError(err)

	// Verify the fields of the message
	var msg types.MsgCreateOracleScript
	err = types.AminoCdc.UnmarshalJSON(operationMsg.Msg, &msg)
	suite.Require().NoError(err)

	suite.Require().True(operationMsg.OK)
	suite.Require().Equal("band1n5sqxutsmk6eews5z9z673wv7n9wah8hjlxyuf", msg.Sender)
	suite.Require().Equal("PjfweXhSUk", msg.Name)
	suite.Require().
		Equal("VAuxErqFPEWIScKpBORIuZqoXlZuTvAjEdlEWDODFRregDTqGNoFBIHxvimmIZwLfFyKUfEWAnNBdtdzDmTPXtpHRGdIbuucfTjO", msg.Description)
	suite.Require().
		Equal("yhXKLyhUScOXvYthRXpPfKwMhptXaxIxgqBoUqzrWbaoLTVpQoottZyPFfNOoMioXHRuFwMRYUiKvcWPkrayyTLOCFJlAyslDame", msg.Schema)
	suite.Require().
		Equal("nDQfwRLGIWozYaOAilMBcObErwgTDNGWnwQMUgFFSKtPDMEoEQCTKVREqrXZSGLqwTMcxHfWotDllNkIJPMbXzjDVjPOOjCFuIvT", msg.SourceCodeURL)
	suite.Require().
		Equal("0061736d0100000001100360000060047e7e7e7e0060027e7e00022f0203656e761161736b5f65787465726e616c5f64617461000103656e760f7365745f72657475726e5f6461746100020303020000040501700101010503010011071e030770726570617265000207657865637574650003066d656d6f727902000a4e022601017e42014201418008ad22004204100042024202200042041000420342032000420410000b2501017f4100210002400340200041016a2100200041e400490d000b0b418008ad420410010b0b0b01004180080b0462656562", hex.EncodeToString(msg.Code))
	suite.Require().Equal("band1n5sqxutsmk6eews5z9z673wv7n9wah8hjlxyuf", msg.Owner)
	suite.Require().Equal(types.TypeMsgCreateOracleScript, msg.Type())
	suite.Require().Equal(types.ModuleName, msg.Route())
	suite.Require().Len(futureOperations, 0)
}

// TestSimulateMsgEditOracleScript tests the normal scenario of a valid message of type TypeMsgEditOracleScript
func (suite *SimTestSuite) TestSimulateMsgEditOracleScript() {
	// Prepare oracle script for us to edit by message
	suite.app.OracleKeeper.SetOracleScript(
		suite.ctx,
		1,
		types.NewOracleScript(
			suite.accs[0].Address,
			"name",
			"description",
			"filename",
			"schema",
			"sourceCodeURL",
		),
	)

	// Simulate MSgEditOracleScript
	op := simulation.SimulateMsgEditOracleScript(
		suite.app.AccountKeeper,
		suite.app.BankKeeper,
		suite.app.StakingKeeper,
		suite.app.OracleKeeper,
	)
	operationMsg, futureOperations, err := op(suite.r, suite.app.BaseApp, suite.ctx, suite.accs, "")
	suite.Require().NoError(err)

	// Verify the fields of the message
	var msg types.MsgEditOracleScript
	err = types.AminoCdc.UnmarshalJSON(operationMsg.Msg, &msg)
	suite.Require().NoError(err)

	suite.Require().True(operationMsg.OK)
	suite.Require().Equal(types.OracleScriptID(1), msg.OracleScriptID)
	suite.Require().Equal("band1tnh2q55v8wyygtt9srz5safamzdengsneky62e", msg.Sender)
	suite.Require().Equal("MaxKlMIJMO", msg.Name)
	suite.Require().
		Equal("BORIuZqoXlZuTvAjEdlEWDODFRregDTqGNoFBIHxvimmIZwLfFyKUfEWAnNBdtdzDmTPXtpHRGdIbuucfTjOygZsTxPjfweXhSUk", msg.Description)
	suite.Require().
		Equal("YthRXpPfKwMhptXaxIxgqBoUqzrWbaoLTVpQoottZyPFfNOoMioXHRuFwMRYUiKvcWPkrayyTLOCFJlAyslDameIuqVAuxErqFPE", msg.Schema)
	suite.Require().
		Equal("WozYaOAilMBcObErwgTDNGWnwQMUgFFSKtPDMEoEQCTKVREqrXZSGLqwTMcxHfWotDllNkIJPMbXzjDVjPOOjCFuIvTyhXKLyhUS", msg.SourceCodeURL)
	suite.Require().
		Equal("0061736d0100000001100360000060047e7e7e7e0060027e7e00022f0203656e761161736b5f65787465726e616c5f64617461000103656e760f7365745f72657475726e5f6461746100020303020000040501700101010503010011071e030770726570617265000207657865637574650003066d656d6f727902000a4e022601017e42014201418008ad22004204100042024202200042041000420342032000420410000b2501017f4100210002400340200041016a2100200041e400490d000b0b418008ad420410010b0b0b01004180080b0462656562", hex.EncodeToString(msg.Code))
	suite.Require().Equal("band1tnh2q55v8wyygtt9srz5safamzdengsneky62e", msg.Owner)
	suite.Require().Equal(types.TypeMsgEditOracleScript, msg.Type())
	suite.Require().Equal(types.ModuleName, msg.Route())
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
	err = types.AminoCdc.UnmarshalJSON(operationMsg.Msg, &msg)
	suite.Require().NoError(err)

	suite.Require().True(operationMsg.OK)
	suite.Require().Equal("bandvaloper1n5sqxutsmk6eews5z9z673wv7n9wah8h7fz8ez", msg.Validator)
	suite.Require().Equal(types.TypeMsgActivate, msg.Type())
	suite.Require().Equal(types.ModuleName, msg.Route())
	suite.Require().Len(futureOperations, 0)
}

func (suite *SimTestSuite) getTestingAccounts(r *rand.Rand, n int) []simtypes.Account {
	accounts := simtypes.RandomAccounts(r, n)

	initAmt := suite.app.StakingKeeper.TokensFromConsensusPower(suite.ctx, 200)
	initCoins := sdk.NewCoins(sdk.NewCoin("uband", initAmt))

	// add coins to the accounts
	for _, account := range accounts {
		acc := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, account.Address)
		suite.app.AccountKeeper.SetAccount(suite.ctx, acc)
		suite.Require().NoError(testutil.FundAccount(suite.app.BankKeeper, suite.ctx, account.Address, initCoins))
	}

	return accounts
}

func TestSimTestSuite(t *testing.T) {
	suite.Run(t, new(SimTestSuite))
}
