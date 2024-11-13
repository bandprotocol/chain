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
		{simulation.DefaultWeightMsgCreateDataSource, sdk.MsgTypeURL(&types.MsgCreateDataSource{})},
		{simulation.DefaultWeightMsgEditDataSource, sdk.MsgTypeURL(&types.MsgEditDataSource{})},
		{
			simulation.DefaultWeightMsgCreateOracleScript,
			sdk.MsgTypeURL(&types.MsgCreateOracleScript{}),
		},
		{simulation.DefaultWeightMsgEditOracleScript, sdk.MsgTypeURL(&types.MsgEditOracleScript{})},
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
	// Prepare oracle script for request
	suite.TestSimulateMsgCreateOracleScript()
	// Prepare data sources for request
	for i := 1; i <= 3; i++ {
		ds, _ := suite.app.OracleKeeper.GetDataSource(suite.ctx, types.DataSourceID(i))
		ds.Fee = sdk.NewCoins()
		suite.app.OracleKeeper.SetDataSource(suite.ctx, types.DataSourceID(i), ds)
	}
	// Prepare active validators
	err := suite.app.StakingKeeper.IterateBondedValidatorsByPower(suite.ctx,
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
		Equal("6f7857727a526e54566a5374506164687345536c45526e4b68704550736644784e767871634f7949756c61436b6d5064616d624c48764768545a7a7973767146617545676b4652497450667669736568466d6f426851716d6b6662485673676648584450", hex.EncodeToString(msg.Calldata))
	suite.Require().Equal(uint64(3), msg.AskCount)
	suite.Require().Equal(uint64(3), msg.MinCount)
	suite.Require().
		Equal("RTRnuwdBeuOGgFbJLbDksHVapaRayWzwoYBEpmrlAxrUxYMUekKbpjPNfjUCjhbdMAnJmYQVZBQZkFVweHDAlaqJjRqoQPoOMLhy", msg.ClientID)
	suite.Require().Equal(sdk.Coins(nil), msg.FeeLimit)
	suite.Require().Equal(uint64(169271), msg.PrepareGas)
	suite.Require().Equal(uint64(115894), msg.ExecuteGas)
	suite.Require().Equal("band1ghekyjucln7y67ntx7cf27m9dpuxxemnvh82dt", msg.Sender)
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
	err = proto.Unmarshal(operationMsg.Msg, &msg)
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
	suite.Require().Equal(sdk.MsgTypeURL(&types.MsgCreateDataSource{}), sdk.MsgTypeURL(&msg))
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
	err = proto.Unmarshal(operationMsg.Msg, &msg)
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
	suite.Require().Equal(sdk.MsgTypeURL(&types.MsgEditDataSource{}), sdk.MsgTypeURL(&msg))
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
	err = proto.Unmarshal(operationMsg.Msg, &msg)
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
		Equal("0061736d0100000001100360000060047e7e7e7e0060027e7e00022f0203656e761161736b5f65787465726e616c5f64617461000103656e760f7365745f72657475726e5f6461746100020303020000040501700101010503010011071e030770726570617265000207657865637574650003066d656d6f727902000a4e022601017e42014201418008ad22004204100042024202200042041000420342032000420410000b2501017f4100210002400340200041016a2100200041e400490d000b0b418008ad420410010b0b0b01004180080b0474657374006f046e616d65013704001161736b5f65787465726e616c5f64617461010f7365745f72657475726e5f64617461020770726570617265030765786563757465020e02020100026c3003010003696478040d030002743001027431020274320505010002543006090100066d656d6f7279", hex.EncodeToString(msg.Code))
	suite.Require().Equal("band1n5sqxutsmk6eews5z9z673wv7n9wah8hjlxyuf", msg.Owner)
	suite.Require().Equal(sdk.MsgTypeURL(&types.MsgCreateOracleScript{}), sdk.MsgTypeURL(&msg))
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
	err = proto.Unmarshal(operationMsg.Msg, &msg)
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
		Equal("0061736d0100000001100360000060047e7e7e7e0060027e7e00022f0203656e761161736b5f65787465726e616c5f64617461000103656e760f7365745f72657475726e5f6461746100020303020000040501700101010503010011071e030770726570617265000207657865637574650003066d656d6f727902000a4e022601017e42014201418008ad22004204100042024202200042041000420342032000420410000b2501017f4100210002400340200041016a2100200041e400490d000b0b418008ad420410010b0b0b01004180080b0474657374006f046e616d65013704001161736b5f65787465726e616c5f64617461010f7365745f72657475726e5f64617461020770726570617265030765786563757465020e02020100026c3003010003696478040d030002743001027431020274320505010002543006090100066d656d6f7279", hex.EncodeToString(msg.Code))
	suite.Require().Equal("band1tnh2q55v8wyygtt9srz5safamzdengsneky62e", msg.Owner)
	suite.Require().Equal(sdk.MsgTypeURL(&types.MsgEditOracleScript{}), sdk.MsgTypeURL(&msg))
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
