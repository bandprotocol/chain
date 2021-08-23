package ante_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/authz"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/bandprotocol/chain/v2/testing/testapp"
	bandante "github.com/bandprotocol/chain/v2/x/oracle/ante"
	"github.com/bandprotocol/chain/v2/x/oracle/keeper"
	"github.com/bandprotocol/chain/v2/x/oracle/types"
)

var (
	BasicCalldata = []byte("BASIC_CALLDATA")
	BasicClientID = "BASIC_CLIENT_ID"
)

type MyStubTx struct {
	sdk.Tx
	Msgs []sdk.Msg
}

func (mst *MyStubTx) GetMsgs() []sdk.Msg {
	return mst.Msgs
}

//mock object for tracking behavior of ante function
type MyMockAnte struct {
	mock.Mock
}

func (m *MyMockAnte) Ante(ctx sdk.Context, tx sdk.Tx, simulate bool) (newCtx sdk.Context, err error) {
	m.Called(ctx, tx, simulate)
	//make the function return the same contex as in argument list and no error for the ease of testing
	return ctx, nil
}

type AnteTestSuit struct {
	suite.Suite
	ctx          sdk.Context
	oracleKeeper keeper.Keeper
	mockAnte     *MyMockAnte
	feelessAnte  sdk.AnteHandler
	requestId    types.RequestID
}

func (suite *AnteTestSuit) SetupTest() {
	_, suite.ctx, suite.oracleKeeper = testapp.CreateTestInput(true)
	suite.ctx = suite.ctx.WithBlockHeight(999).WithIsCheckTx(true).WithMinGasPrices(sdk.DecCoins{{Denom: "uband", Amount: sdk.NewDec(14000)}})

	suite.oracleKeeper.GrantReporter(suite.ctx, testapp.Validators[0].ValAddress, testapp.Alice.Address)

	req := types.NewRequest(1, BasicCalldata, []sdk.ValAddress{testapp.Validators[0].ValAddress}, 1, 1, testapp.ParseTime(0), "", nil, nil, 0)
	suite.requestId = suite.oracleKeeper.AddRequest(suite.ctx, req)

	suite.mockAnte = new(MyMockAnte)
	suite.feelessAnte = bandante.NewFeelessReportsAnteHandler(suite.mockAnte.Ante, suite.oracleKeeper)
}

func (suite *AnteTestSuit) TestValidRawReport() {
	msgs := []sdk.Msg{types.NewMsgReportData(suite.requestId, []types.RawReport{}, testapp.Validators[0].ValAddress)}
	stubTx := &MyStubTx{Msgs: msgs}

	//makes an expectation when call function 'Ante' of 'mockAnte' object
	//valid report Msg was called by ante function without minGasPrice
	suite.mockAnte.On("Ante", suite.ctx.WithMinGasPrices(sdk.DecCoins{}), stubTx, false)
	ctx, err := suite.feelessAnte(suite.ctx, stubTx, false)

	//asserts all everything specificed with 'On' was in fact called as expected of the 'mockAnte' object
	suite.mockAnte.AssertExpectations(suite.T())
	//the contex's minGasPrice should be the same as before had been validated by ante function
	suite.Require().Equal(ctx.MinGasPrices(), suite.ctx.MinGasPrices())
	suite.Require().NoError(err)
}

func (suite *AnteTestSuit) TestNotValidRawReport() {
	msgs := []sdk.Msg{types.NewMsgReportData(1, []types.RawReport{}, testapp.Alice.ValAddress)}
	stubTx := &MyStubTx{Msgs: msgs}

	//no need to make an expectaion because ante function will not be called by this condition
	ctx, err := suite.feelessAnte(suite.ctx, stubTx, false)

	//make sure that ante function was not called
	suite.mockAnte.AssertNumberOfCalls(suite.T(), "Ante", 0)
	suite.Require().Equal(ctx, suite.ctx)
	suite.Require().Error(err)
}

func (suite *AnteTestSuit) TestValidReport() {
	reportMsgs := []sdk.Msg{types.NewMsgReportData(suite.requestId, []types.RawReport{}, testapp.Validators[0].ValAddress)}
	authzMsg := authz.NewMsgExec(testapp.Alice.Address, reportMsgs)
	stubTx := &MyStubTx{Msgs: []sdk.Msg{&authzMsg}}

	//makes an expectation when call function 'Ante' of 'mockAnte' object
	//valid report Msg was called by ante function without minGasPrice
	suite.mockAnte.On("Ante", suite.ctx.WithMinGasPrices(sdk.DecCoins{}), stubTx, false)
	ctx, err := suite.feelessAnte(suite.ctx, stubTx, false)

	//asserts all everything specificed with 'On' was in fact called as expected of the 'mockAnte' object
	suite.mockAnte.AssertExpectations(suite.T())
	//the contex's minGasPrice should be the same as before had been validated by ante function
	suite.Require().Equal(ctx.MinGasPrices(), suite.ctx.MinGasPrices())
	suite.Require().NoError(err)
}

func (suite *AnteTestSuit) TestNoAuthzReport() {
	reportMsgs := []sdk.Msg{types.NewMsgReportData(suite.requestId, []types.RawReport{}, testapp.Validators[0].ValAddress)}
	authzMsg := authz.NewMsgExec(testapp.Bob.Address, reportMsgs)
	stubTx := &MyStubTx{Msgs: []sdk.Msg{&authzMsg}}

	//no need to make an expectaion because ante function will not be called by this condition
	_, err := suite.feelessAnte(suite.ctx, stubTx, false)

	//make sure that ante function was not called
	suite.mockAnte.AssertNumberOfCalls(suite.T(), "Ante", 0)
	suite.Require().EqualError(err, sdkerrors.ErrUnauthorized.Wrap("authorization not found").Error())
}

func (suite *AnteTestSuit) TestNotValidReport() {
	reportMsgs := []sdk.Msg{types.NewMsgReportData(suite.requestId+1, []types.RawReport{}, testapp.Validators[0].ValAddress)}
	authzMsg := authz.NewMsgExec(testapp.Alice.Address, reportMsgs)
	stubTx := &MyStubTx{Msgs: []sdk.Msg{&authzMsg}}

	//no need to make an expectaion because ante function will not be called by this condition
	_, err := suite.feelessAnte(suite.ctx, stubTx, false)

	//make sure that ante function was not called
	suite.mockAnte.AssertNumberOfCalls(suite.T(), "Ante", 0)
	suite.Require().Error(err)
}

func (suite *AnteTestSuit) TestNotReportMsg() {
	requestMsg := types.NewMsgRequestData(1, BasicCalldata, 1, 1, BasicClientID, testapp.Coins100000000uband, testapp.TestDefaultPrepareGas, testapp.TestDefaultExecuteGas, testapp.FeePayer.Address)
	stubTx := &MyStubTx{Msgs: []sdk.Msg{requestMsg}}

	//makes an expectation when call function 'Ante' of 'mockAnte' object
	//others type Msg was normally called by ante function
	suite.mockAnte.On("Ante", suite.ctx, stubTx, false)
	ctx, err := suite.feelessAnte(suite.ctx, stubTx, false)

	//asserts all everything specificed with 'On' was in fact called as expected of the 'mockAnte' object
	suite.mockAnte.AssertExpectations(suite.T())
	suite.Require().Equal(ctx, suite.ctx)
	suite.Require().NoError(err)
}

func (suite *AnteTestSuit) TestNotReportMsgOnReportOnlyBlockByCash() {
	reportMsgs := []sdk.Msg{types.NewMsgReportData(suite.requestId, []types.RawReport{}, testapp.Validators[0].ValAddress)}
	authzMsg := authz.NewMsgExec(testapp.Alice.Address, reportMsgs)
	stubTxReport := &MyStubTx{Msgs: []sdk.Msg{&authzMsg}}
	requestMsg := types.NewMsgRequestData(1, BasicCalldata, 1, 1, BasicClientID, testapp.Coins100000000uband, testapp.TestDefaultPrepareGas, testapp.TestDefaultExecuteGas, testapp.FeePayer.Address)
	stubTxNotReport := &MyStubTx{Msgs: []sdk.Msg{requestMsg}}

	//makes an expectation when call function 'Ante' of 'mockAnte' object
	//valid report Msg was called by ante function without minGasPrice
	suite.mockAnte.On("Ante", suite.ctx.WithMinGasPrices(sdk.DecCoins{}), stubTxReport, false)
	suite.feelessAnte(suite.ctx, stubTxReport, false)

	//do the simulating as the proposal block had been passed for 21 blocks
	suite.ctx = suite.ctx.WithBlockHeight(suite.ctx.BlockHeight() + 21)
	//makes an expectation when call function 'Ante' of 'mockAnte' object
	//valid report Msg was called by ante function without minGasPrice
	//need to make another expectation because blockHeight of 'suite.ctx' have been changed
	suite.mockAnte.On("Ante", suite.ctx.WithMinGasPrices(sdk.DecCoins{}), stubTxReport, false)
	suite.feelessAnte(suite.ctx, stubTxReport, false)

	//do the simulating as the proposal block had passed for a block
	suite.ctx = suite.ctx.WithBlockHeight(suite.ctx.BlockHeight() + 1)
	//no need to make an expectaion because ante function will not be called by this condition
	_, err := suite.feelessAnte(suite.ctx, stubTxNotReport, false)

	//asserts all everything specificed with 'On' was in fact called as expected of the 'mockAnte' object
	suite.mockAnte.AssertExpectations(suite.T())
	//the method 'Ante' was called only 2 times because the last 'feelessAnte' execution had been rejected by the block reserved for report txs only reason
	suite.mockAnte.AssertNumberOfCalls(suite.T(), "Ante", 2)
	suite.Require().EqualError(err, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "Block reserved for report txs").Error())
}

func (suite *AnteTestSuit) TestReportMsgAndOthersTypeMsgInTheSameAuthzMsgs() {
	reportMsg := types.NewMsgReportData(suite.requestId, []types.RawReport{}, testapp.Validators[0].ValAddress)
	requestMsg := types.NewMsgRequestData(1, BasicCalldata, 1, 1, BasicClientID, testapp.Coins100000000uband, testapp.TestDefaultPrepareGas, testapp.TestDefaultExecuteGas, testapp.FeePayer.Address)
	msgs := []sdk.Msg{reportMsg, requestMsg}
	authzMsg := authz.NewMsgExec(testapp.Alice.Address, msgs)
	stubTx := &MyStubTx{Msgs: []sdk.Msg{&authzMsg}}

	//makes an expectation when call function 'Ante' of 'mockAnte' object
	//the authzMsgs have others type Msg was normally called by ante function
	suite.mockAnte.On("Ante", suite.ctx, stubTx, false)
	ctx, err := suite.feelessAnte(suite.ctx, stubTx, false)

	//asserts all everything specificed with 'On' was in fact called as expected of the 'mockAnte' object
	suite.mockAnte.AssertExpectations(suite.T())
	suite.Require().Equal(ctx, suite.ctx)
	suite.Require().NoError(err)
}

func (suite *AnteTestSuit) TestReportMsgAndOthersTypeMsgInTheSameTx() {
	reportMsg := types.NewMsgReportData(suite.requestId, []types.RawReport{}, testapp.Validators[0].ValAddress)
	requestMsg := types.NewMsgRequestData(1, BasicCalldata, 1, 1, BasicClientID, testapp.Coins100000000uband, testapp.TestDefaultPrepareGas, testapp.TestDefaultExecuteGas, testapp.FeePayer.Address)
	stubTx := &MyStubTx{Msgs: []sdk.Msg{reportMsg, requestMsg}}

	//makes an expectation when call function 'Ante' of 'mockAnte' object
	//the tx has others type Msg was normally called by ante function
	suite.mockAnte.On("Ante", suite.ctx, stubTx, false)
	ctx, err := suite.feelessAnte(suite.ctx, stubTx, false)

	//asserts all everything specificed with 'On' was in fact called as expected of the 'mockAnte' object
	suite.mockAnte.AssertExpectations(suite.T())
	suite.Require().Equal(ctx, suite.ctx)
	suite.Require().NoError(err)
}

func TestAnteTestSuite(t *testing.T) {
	suite.Run(t, new(AnteTestSuit))
}
