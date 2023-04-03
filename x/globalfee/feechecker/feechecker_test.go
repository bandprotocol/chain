package feechecker_test

import (
	"math"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/suite"

	"github.com/bandprotocol/chain/v2/testing/testapp"
	"github.com/bandprotocol/chain/v2/x/globalfee"
	"github.com/bandprotocol/chain/v2/x/globalfee/feechecker"
	"github.com/bandprotocol/chain/v2/x/oracle/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

var (
	BasicCalldata = []byte("BASIC_CALLDATA")
	BasicClientID = "BASIC_CLIENT_ID"
)

type StubTx struct {
	sdk.Tx
	sdk.FeeTx
	Msgs      []sdk.Msg
	GasPrices sdk.DecCoins
}

func (st *StubTx) GetMsgs() []sdk.Msg {
	return st.Msgs
}

func (st *StubTx) ValidateBasic() error {
	return nil
}

func (st *StubTx) GetGas() uint64 {
	return 1000000
}
func (st *StubTx) GetFee() sdk.Coins {
	fees := make(sdk.Coins, len(st.GasPrices))

	// Determine the fees by multiplying each gas prices
	glDec := sdk.NewDec(int64(st.GetGas()))
	for i, gp := range st.GasPrices {
		fee := gp.Amount.Mul(glDec)
		fees[i] = sdk.NewCoin(gp.Denom, fee.Ceil().RoundInt())
	}

	return fees
}

type FeeCheckerTestSuite struct {
	suite.Suite
	FeeChecker feechecker.FeeChecker
	ctx        sdk.Context
	requestId  types.RequestID
}

func (suite *FeeCheckerTestSuite) SetupTest() {
	app, ctx, oracleKeeper := testapp.CreateTestInput(true)
	suite.ctx = ctx.WithBlockHeight(999).
		WithIsCheckTx(true).
		WithMinGasPrices(sdk.DecCoins{{Denom: "uband", Amount: sdk.NewDecWithPrec(1, 4)}})

	oracleKeeper.GrantReporter(suite.ctx, testapp.Validators[0].ValAddress, testapp.Alice.Address)

	req := types.NewRequest(
		1,
		BasicCalldata,
		[]sdk.ValAddress{testapp.Validators[0].ValAddress},
		1,
		1,
		testapp.ParseTime(0),
		"",
		nil,
		nil,
		0,
	)
	suite.requestId = oracleKeeper.AddRequest(suite.ctx, req)

	suite.FeeChecker = feechecker.NewFeeChecker(
		&oracleKeeper,
		app.GetSubspace(globalfee.ModuleName),
		app.GetSubspace(stakingtypes.ModuleName),
	)
}

func (suite *FeeCheckerTestSuite) TestValidRawReport() {
	msgs := []sdk.Msg{types.NewMsgReportData(suite.requestId, []types.RawReport{}, testapp.Validators[0].ValAddress)}
	stubTx := &StubTx{Msgs: msgs}

	// test - check report tx
	isReportTx, err := suite.FeeChecker.CheckReportTx(suite.ctx, stubTx)
	suite.Require().NoError(err)
	suite.Require().True(isReportTx)

	// test - check tx fee with min gas prices
	fee, priority, err := suite.FeeChecker.CheckTxFeeWithMinGasPrices(suite.ctx, stubTx)
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.Coins{}, fee)
	suite.Require().Equal(int64(math.MaxInt64), priority)
}

func (suite *FeeCheckerTestSuite) TestNotValidRawReport() {
	msgs := []sdk.Msg{types.NewMsgReportData(1, []types.RawReport{}, testapp.Alice.ValAddress)}
	stubTx := &StubTx{Msgs: msgs}

	// test - check report tx
	isReportTx, err := suite.FeeChecker.CheckReportTx(suite.ctx, stubTx)
	suite.Require().Error(err)
	suite.Require().False(isReportTx)

	// test - check tx fee with min gas prices
	_, _, err = suite.FeeChecker.CheckTxFeeWithMinGasPrices(suite.ctx, stubTx)
	suite.Require().Error(err)
}

func (suite *FeeCheckerTestSuite) TestValidReport() {
	reportMsgs := []sdk.Msg{
		types.NewMsgReportData(suite.requestId, []types.RawReport{}, testapp.Validators[0].ValAddress),
	}
	authzMsg := authz.NewMsgExec(testapp.Alice.Address, reportMsgs)
	stubTx := &StubTx{Msgs: []sdk.Msg{&authzMsg}}

	// test - check report tx
	isReportTx, err := suite.FeeChecker.CheckReportTx(suite.ctx, stubTx)
	suite.Require().NoError(err)
	suite.Require().True(isReportTx)

	// test - check tx fee with min gas prices
	fee, priority, err := suite.FeeChecker.CheckTxFeeWithMinGasPrices(suite.ctx, stubTx)
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.Coins{}, fee)
	suite.Require().Equal(int64(math.MaxInt64), priority)
}

func (suite *FeeCheckerTestSuite) TestNoAuthzReport() {
	reportMsgs := []sdk.Msg{
		types.NewMsgReportData(suite.requestId, []types.RawReport{}, testapp.Validators[0].ValAddress),
	}
	authzMsg := authz.NewMsgExec(testapp.Bob.Address, reportMsgs)
	stubTx := &StubTx{Msgs: []sdk.Msg{&authzMsg}}

	// test - check report tx
	isReportTx, err := suite.FeeChecker.CheckReportTx(suite.ctx, stubTx)
	suite.Require().EqualError(err, sdkerrors.ErrUnauthorized.Wrap("authorization not found").Error())
	suite.Require().False(isReportTx)

	// test - check tx fee with min gas prices
	_, _, err = suite.FeeChecker.CheckTxFeeWithMinGasPrices(suite.ctx, stubTx)
	suite.Require().EqualError(err, sdkerrors.ErrUnauthorized.Wrap("authorization not found").Error())
}

func (suite *FeeCheckerTestSuite) TestNotValidReport() {
	reportMsgs := []sdk.Msg{
		types.NewMsgReportData(suite.requestId+1, []types.RawReport{}, testapp.Validators[0].ValAddress),
	}
	authzMsg := authz.NewMsgExec(testapp.Alice.Address, reportMsgs)
	stubTx := &StubTx{Msgs: []sdk.Msg{&authzMsg}}

	// test - check report tx
	isReportTx, err := suite.FeeChecker.CheckReportTx(suite.ctx, stubTx)
	suite.Require().Error(err)
	suite.Require().False(isReportTx)

	// test - check tx fee with min gas prices
	_, _, err = suite.FeeChecker.CheckTxFeeWithMinGasPrices(suite.ctx, stubTx)
	suite.Require().Error(err)
}

func (suite *FeeCheckerTestSuite) TestNotReportMsg() {
	requestMsg := types.NewMsgRequestData(
		1,
		BasicCalldata,
		1,
		1,
		BasicClientID,
		testapp.Coins100000000uband,
		testapp.TestDefaultPrepareGas,
		testapp.TestDefaultExecuteGas,
		testapp.FeePayer.Address,
	)
	stubTx := &StubTx{Msgs: []sdk.Msg{requestMsg}, GasPrices: sdk.NewDecCoins(sdk.NewDecCoin("uband", sdk.NewInt(1)))}

	// test - check report tx
	isReportTx, err := suite.FeeChecker.CheckReportTx(suite.ctx, stubTx)
	suite.Require().NoError(err)
	suite.Require().False(isReportTx)

	// test - check tx fee with min gas prices
	fee, priority, err := suite.FeeChecker.CheckTxFeeWithMinGasPrices(suite.ctx, stubTx)
	suite.Require().NoError(err)
	suite.Require().Equal(stubTx.GetFee(), fee)
	suite.Require().Equal(int64(1000), priority)
}

func (suite *FeeCheckerTestSuite) TestReportMsgAndOthersTypeMsgInTheSameAuthzMsgs() {
	reportMsg := types.NewMsgReportData(suite.requestId, []types.RawReport{}, testapp.Validators[0].ValAddress)
	requestMsg := types.NewMsgRequestData(
		1,
		BasicCalldata,
		1,
		1,
		BasicClientID,
		testapp.Coins100000000uband,
		testapp.TestDefaultPrepareGas,
		testapp.TestDefaultExecuteGas,
		testapp.FeePayer.Address,
	)
	msgs := []sdk.Msg{reportMsg, requestMsg}
	authzMsg := authz.NewMsgExec(testapp.Alice.Address, msgs)
	stubTx := &StubTx{Msgs: []sdk.Msg{&authzMsg}, GasPrices: sdk.NewDecCoins(sdk.NewDecCoin("uband", sdk.NewInt(1)))}

	// test - check report tx
	isReportTx, err := suite.FeeChecker.CheckReportTx(suite.ctx, stubTx)
	suite.Require().NoError(err)
	suite.Require().False(isReportTx)

	// test - check tx fee with min gas prices
	fee, priority, err := suite.FeeChecker.CheckTxFeeWithMinGasPrices(suite.ctx, stubTx)
	suite.Require().NoError(err)
	suite.Require().Equal(stubTx.GetFee(), fee)
	suite.Require().Equal(int64(1000), priority)
}

func (suite *FeeCheckerTestSuite) TestReportMsgAndOthersTypeMsgInTheSameTx() {
	reportMsg := types.NewMsgReportData(suite.requestId, []types.RawReport{}, testapp.Validators[0].ValAddress)
	requestMsg := types.NewMsgRequestData(
		1,
		BasicCalldata,
		1,
		1,
		BasicClientID,
		testapp.Coins100000000uband,
		testapp.TestDefaultPrepareGas,
		testapp.TestDefaultExecuteGas,
		testapp.FeePayer.Address,
	)
	stubTx := &StubTx{
		Msgs:      []sdk.Msg{reportMsg, requestMsg},
		GasPrices: sdk.NewDecCoins(sdk.NewDecCoin("uband", sdk.NewInt(1))),
	}

	// test - check report tx
	isReportTx, err := suite.FeeChecker.CheckReportTx(suite.ctx, stubTx)
	suite.Require().NoError(err)
	suite.Require().False(isReportTx)

	// test - check tx fee with min gas prices
	fee, priority, err := suite.FeeChecker.CheckTxFeeWithMinGasPrices(suite.ctx, stubTx)
	suite.Require().NoError(err)
	suite.Require().Equal(stubTx.GetFee(), fee)
	suite.Require().Equal(int64(1000), priority)
}

func (suite *FeeCheckerTestSuite) TestGetBondDenom() {
	denom := suite.FeeChecker.GetBondDenom(suite.ctx)
	suite.Require().Equal("uband", denom)
}

func (suite *FeeCheckerTestSuite) TestDefaultZeroGlobalFee() {
	coins, err := suite.FeeChecker.DefaultZeroGlobalFee(suite.ctx)

	suite.Require().Equal(1, len(coins))
	suite.Require().Equal("uband", coins[0].Denom)
	suite.Require().Equal(sdk.NewDec(0), coins[0].Amount)
	suite.Require().NoError(err)
}

func TestFeeCheckerTestSuite(t *testing.T) {
	suite.Run(t, new(FeeCheckerTestSuite))
}
