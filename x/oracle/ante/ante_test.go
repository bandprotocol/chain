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

type MyStubTx struct {
	sdk.Tx
	Msgs []sdk.Msg
}

func (mst *MyStubTx) GetMsgs() []sdk.Msg {
	return mst.Msgs
}

type MyMockAnte struct {
	mock.Mock
}

func (m *MyMockAnte) Ante(ctx sdk.Context, tx sdk.Tx, simulate bool) (newCtx sdk.Context, err error) {

	m.Called(ctx, tx, simulate)
	return ctx, nil
}

type AnteTestSuit struct {
	suite.Suite
	ctx          sdk.Context
	oracleKeeper keeper.Keeper
	mockAnte     *MyMockAnte
	gaslessAnte  sdk.AnteHandler
	requestId    types.RequestID
}

func (suite *AnteTestSuit) SetupTest() {
	_, suite.ctx, suite.oracleKeeper = testapp.CreateTestInput(true)
	suite.ctx = suite.ctx.WithBlockHeight(999).WithIsCheckTx(true).WithMinGasPrices(sdk.DecCoins{{Denom: "uband", Amount: sdk.NewDec(14000)}})

	suite.oracleKeeper.GrantReporter(suite.ctx, testapp.Validators[0].ValAddress, testapp.Alice.Address)

	req := types.NewRequest(1, []byte("BASIC_CALLDATA"), []sdk.ValAddress{testapp.Validators[0].ValAddress}, 1, 1, testapp.ParseTime(0), "", nil, nil, 0)
	suite.requestId = suite.oracleKeeper.AddRequest(suite.ctx, req)

	suite.mockAnte = new(MyMockAnte)
	suite.gaslessAnte = bandante.NewFeelessReportsAnteHandler(suite.mockAnte.Ante, suite.oracleKeeper)
}

func (suite *AnteTestSuit) TestValidRawReport() {

	msgs := []sdk.Msg{types.NewMsgReportData(suite.requestId, []types.RawReport{}, testapp.Validators[0].ValAddress)}
	stubTx := &MyStubTx{Msgs: msgs}

	suite.mockAnte.On("Ante", suite.ctx.WithMinGasPrices(sdk.DecCoins{}), stubTx, false)
	ctx, err := suite.gaslessAnte(suite.ctx, stubTx, false)

	suite.mockAnte.AssertExpectations(suite.T())
	suite.Require().Equal(ctx.MinGasPrices(), suite.ctx.MinGasPrices())
	suite.Require().Equal(err, nil)
}

func (suite *AnteTestSuit) TestNotValidRawReport() {
	msgs := []sdk.Msg{types.NewMsgReportData(1, []types.RawReport{}, testapp.Alice.ValAddress)}
	stubTx := &MyStubTx{Msgs: msgs}

	ctx, err := suite.gaslessAnte(suite.ctx, stubTx, false)

	suite.Require().Equal(ctx, suite.ctx)
	suite.Require().Error(err)
}

func (suite *AnteTestSuit) TestValidReport() {
	reportMsgs := []sdk.Msg{types.NewMsgReportData(suite.requestId, []types.RawReport{}, testapp.Validators[0].ValAddress)}
	autzMsg := authz.NewMsgExec(testapp.Alice.Address, reportMsgs)
	stubTx := &MyStubTx{Msgs: []sdk.Msg{&autzMsg}}

	suite.mockAnte.On("Ante", suite.ctx.WithMinGasPrices(sdk.DecCoins{}), stubTx, false)
	ctx, err := suite.gaslessAnte(suite.ctx, stubTx, false)

	suite.mockAnte.AssertExpectations(suite.T())
	suite.Require().Equal(ctx.MinGasPrices(), suite.ctx.MinGasPrices())
	suite.Require().Equal(err, nil)
}

func (suite *AnteTestSuit) TestNoAuthzReport() {
	reportMsgs := []sdk.Msg{types.NewMsgReportData(suite.requestId, []types.RawReport{}, testapp.Validators[0].ValAddress)}
	autzMsg := authz.NewMsgExec(testapp.Bob.Address, reportMsgs)
	stubTx := &MyStubTx{Msgs: []sdk.Msg{&autzMsg}}

	_, err := suite.gaslessAnte(suite.ctx, stubTx, false)

	suite.mockAnte.AssertNumberOfCalls(suite.T(), "Ante", 0)
	suite.Require().EqualError(err, sdkerrors.ErrUnauthorized.Wrap("authorization not found").Error())
}

func (suite *AnteTestSuit) TestNotValidReport() {
	reportMsgs := []sdk.Msg{types.NewMsgReportData(suite.requestId+1, []types.RawReport{}, testapp.Validators[0].ValAddress)}
	autzMsg := authz.NewMsgExec(testapp.Alice.Address, reportMsgs)
	stubTx := &MyStubTx{Msgs: []sdk.Msg{&autzMsg}}

	_, err := suite.gaslessAnte(suite.ctx, stubTx, false)

	suite.mockAnte.AssertNumberOfCalls(suite.T(), "Ante", 0)
	suite.Require().Error(err)
}

func (suite *AnteTestSuit) TestNotReportMsg() {
	requetMsg := types.NewMsgRequestData(1, []byte("BASIC_CALLDATA"), 1, 1, "BASIC_CLIENT_ID", testapp.Coins100000000uband, testapp.TestDefaultPrepareGas, testapp.TestDefaultExecuteGas, testapp.FeePayer.Address)
	stubTx := &MyStubTx{Msgs: []sdk.Msg{requetMsg}}

	suite.mockAnte.On("Ante", suite.ctx, stubTx, false)
	suite.gaslessAnte(suite.ctx, stubTx, false)

	suite.mockAnte.AssertExpectations(suite.T())
}

func (suite *AnteTestSuit) TestNotReportMsgButReportOnlyBlock() {
	suite.ctx = suite.ctx.WithBlockHeight(0)
	requetMsg := types.NewMsgRequestData(1, []byte("BASIC_CALLDATA"), 1, 1, "BASIC_CLIENT_ID", testapp.Coins100000000uband, testapp.TestDefaultPrepareGas, testapp.TestDefaultExecuteGas, testapp.FeePayer.Address)
	stubTx := &MyStubTx{Msgs: []sdk.Msg{requetMsg}}

	_, err := suite.gaslessAnte(suite.ctx, stubTx, false)

	suite.mockAnte.AssertNumberOfCalls(suite.T(), "Ante", 0)
	suite.Require().EqualError(err, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "Block reserved for report txs").Error())
}

func TestAnteTestSuite(t *testing.T) {
	suite.Run(t, new(AnteTestSuit))
}
