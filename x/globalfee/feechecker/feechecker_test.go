package feechecker_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	protov2 "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/protoadapt"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"

	bandtesting "github.com/bandprotocol/chain/v3/testing"
	bandtsstypes "github.com/bandprotocol/chain/v3/x/bandtss/types"
	feedstypes "github.com/bandprotocol/chain/v3/x/feeds/types"
	"github.com/bandprotocol/chain/v3/x/globalfee/feechecker"
	oracletypes "github.com/bandprotocol/chain/v3/x/oracle/types"
	tsstypes "github.com/bandprotocol/chain/v3/x/tss/types"
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

func (st *StubTx) GetMsgsV2() (ms []protov2.Message, err error) {
	for _, msg := range st.Msgs {
		ms = append(ms, protoadapt.MessageV2Of(msg))
	}

	return
}

func (st *StubTx) GetGas() uint64 {
	return 1000000
}

func (st *StubTx) GetFee() sdk.Coins {
	fees := make(sdk.Coins, len(st.GasPrices))

	// Determine the fees by multiplying each gas prices
	glDec := sdkmath.LegacyNewDec(int64(st.GetGas()))
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
	requestID  oracletypes.RequestID
}

func (suite *FeeCheckerTestSuite) SetupTest() {
	dir := testutil.GetTempDir(suite.T())
	app := bandtesting.SetupWithCustomHome(false, dir)
	ctx := app.BaseApp.NewUncachedContext(false, cmtproto.Header{})

	// Activate validators
	for _, v := range bandtesting.Validators {
		err := app.OracleKeeper.Activate(ctx, v.ValAddress)
		suite.Require().NoError(err)
	}

	_, err := app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: app.LastBlockHeight() + 1})
	suite.Require().NoError(err)
	_, err = app.Commit()
	suite.Require().NoError(err)

	suite.ctx = ctx.WithBlockHeight(999).
		WithIsCheckTx(true).
		WithMinGasPrices(sdk.DecCoins{{Denom: "uband", Amount: sdkmath.LegacyNewDecWithPrec(1, 4)}})

	err = app.OracleKeeper.GrantReporter(suite.ctx, bandtesting.Validators[0].ValAddress, bandtesting.Alice.Address)
	suite.Require().NoError(err)

	expiration := ctx.BlockTime().Add(1000 * time.Hour)

	msgTypeURLs := []sdk.Msg{&tsstypes.MsgSubmitDEs{}, &feedstypes.MsgSubmitSignalPrices{}}
	for _, msg := range msgTypeURLs {
		err = app.AuthzKeeper.SaveGrant(
			ctx,
			bandtesting.Alice.Address,
			bandtesting.Validators[0].Address,
			authz.NewGenericAuthorization(sdk.MsgTypeURL(msg)),
			&expiration,
		)
		suite.Require().NoError(err)
	}

	// mock setup bandtss module
	app.BandtssKeeper.SetCurrentGroup(ctx, bandtsstypes.NewCurrentGroup(1, ctx.BlockTime()))
	app.BandtssKeeper.SetMember(ctx, bandtsstypes.Member{
		Address:  bandtesting.Validators[0].Address.String(),
		IsActive: true,
		GroupID:  1,
	})

	req := oracletypes.NewRequest(
		1,
		BasicCalldata,
		[]sdk.ValAddress{bandtesting.Validators[0].ValAddress},
		1,
		1,
		bandtesting.ParseTime(0),
		"",
		nil,
		nil,
		0,
		0,
		bandtesting.FeePayer.Address.String(),
		bandtesting.Coins100band,
	)
	suite.requestID = app.OracleKeeper.AddRequest(suite.ctx, req)

	suite.FeeChecker = feechecker.NewFeeChecker(
		app.AppCodec(),
		&app.AuthzKeeper,
		&app.OracleKeeper,
		&app.GlobalFeeKeeper,
		app.StakingKeeper,
		app.TSSKeeper,
		&app.BandtssKeeper,
		&app.FeedsKeeper,
	)
}

func (suite *FeeCheckerTestSuite) TestDefaultZeroGlobalFee() {
	coins, err := suite.FeeChecker.DefaultZeroGlobalFee(suite.ctx)

	suite.Require().Equal(1, len(coins))
	suite.Require().Equal("uband", coins[0].Denom)
	suite.Require().Equal(sdkmath.LegacyNewDec(0), coins[0].Amount)
	suite.Require().NoError(err)
}

func TestFeeCheckerTestSuite(t *testing.T) {
	suite.Run(t, new(FeeCheckerTestSuite))
}
