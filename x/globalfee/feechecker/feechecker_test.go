package feechecker_test

import (
	"math"
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
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/authz"

	bandtesting "github.com/bandprotocol/chain/v3/testing"
	feedstypes "github.com/bandprotocol/chain/v3/x/feeds/types"
	"github.com/bandprotocol/chain/v3/x/globalfee/feechecker"
	oracletypes "github.com/bandprotocol/chain/v3/x/oracle/types"
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
	err = app.AuthzKeeper.SaveGrant(
		ctx,
		bandtesting.Alice.Address,
		bandtesting.Validators[0].Address,
		authz.NewGenericAuthorization(
			sdk.MsgTypeURL(&feedstypes.MsgSubmitSignalPrices{}),
		),
		&expiration,
	)
	suite.Require().NoError(err)

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
	)
	suite.requestID = app.OracleKeeper.AddRequest(suite.ctx, req)

	suite.FeeChecker = feechecker.NewFeeChecker(
		app.AppCodec(),
		&app.AuthzKeeper,
		&app.OracleKeeper,
		&app.GlobalFeeKeeper,
		app.StakingKeeper,
		&app.FeedsKeeper,
	)
}

func (suite *FeeCheckerTestSuite) TestIsBypassMinFeeTxAndCheckTxFeeWithMinGasPrices() {
	testCases := []struct {
		name                string
		stubTx              func() *StubTx
		expIsBypassMinFeeTx bool
		expErr              error
		expFee              sdk.Coins
		expPriority         int64
	}{
		{
			name: "valid MsgReportData",
			stubTx: func() *StubTx {
				return &StubTx{
					Msgs: []sdk.Msg{
						oracletypes.NewMsgReportData(
							suite.requestID,
							[]oracletypes.RawReport{},
							bandtesting.Validators[0].ValAddress,
						),
					},
				}
			},
			expIsBypassMinFeeTx: true,
			expErr:              nil,
			expFee:              sdk.Coins{},
			expPriority:         math.MaxInt64,
		},
		{
			name: "valid MsgReportData in valid MsgExec",
			stubTx: func() *StubTx {
				msgExec := authz.NewMsgExec(bandtesting.Alice.Address, []sdk.Msg{
					oracletypes.NewMsgReportData(
						suite.requestID,
						[]oracletypes.RawReport{},
						bandtesting.Validators[0].ValAddress,
					),
				})

				return &StubTx{
					Msgs: []sdk.Msg{
						&msgExec,
					},
				}
			},
			expIsBypassMinFeeTx: true,
			expErr:              nil,
			expFee:              sdk.Coins{},
			expPriority:         math.MaxInt64,
		},
		{
			name: "invalid MsgReportData with not enough fee",
			stubTx: func() *StubTx {
				return &StubTx{
					Msgs: []sdk.Msg{
						oracletypes.NewMsgReportData(1, []oracletypes.RawReport{}, bandtesting.Alice.ValAddress),
					},
				}
			},
			expIsBypassMinFeeTx: false,
			expErr:              sdkerrors.ErrInsufficientFee,
			expFee:              nil,
			expPriority:         0,
		},
		{
			name: "invalid MsgReportData in valid MsgExec with not enough fee",
			stubTx: func() *StubTx {
				msgExec := authz.NewMsgExec(bandtesting.Alice.Address, []sdk.Msg{
					oracletypes.NewMsgReportData(
						suite.requestID+1,
						[]oracletypes.RawReport{},
						bandtesting.Validators[0].ValAddress,
					),
				})

				return &StubTx{
					Msgs: []sdk.Msg{
						&msgExec,
					},
				}
			},
			expIsBypassMinFeeTx: false,
			expErr:              sdkerrors.ErrInsufficientFee,
			expFee:              nil,
			expPriority:         0,
		},
		{
			name: "valid MsgReportData in invalid MsgExec with enough fee",
			stubTx: func() *StubTx {
				msgExec := authz.NewMsgExec(bandtesting.Bob.Address, []sdk.Msg{
					oracletypes.NewMsgReportData(
						suite.requestID,
						[]oracletypes.RawReport{},
						bandtesting.Validators[0].ValAddress,
					),
				})

				return &StubTx{
					Msgs: []sdk.Msg{
						&msgExec,
					},
					GasPrices: sdk.NewDecCoins(sdk.NewDecCoin("uband", sdkmath.NewInt(1))),
				}
			},
			expIsBypassMinFeeTx: false,
			expErr:              nil,
			expFee:              sdk.NewCoins(sdk.NewCoin("uband", sdkmath.NewInt(1000000))),
			expPriority:         10000,
		},
		{
			name: "valid MsgRequestData",
			stubTx: func() *StubTx {
				msgRequestData := oracletypes.NewMsgRequestData(
					1,
					BasicCalldata,
					1,
					1,
					BasicClientID,
					bandtesting.Coins100000000uband,
					bandtesting.TestDefaultPrepareGas,
					bandtesting.TestDefaultExecuteGas,
					bandtesting.FeePayer.Address,
				)

				return &StubTx{
					Msgs: []sdk.Msg{msgRequestData},
					GasPrices: sdk.NewDecCoins(
						sdk.NewDecCoinFromDec("uaaaa", sdkmath.LegacyNewDecWithPrec(100, 3)),
						sdk.NewDecCoinFromDec("uaaab", sdkmath.LegacyNewDecWithPrec(1, 3)),
						sdk.NewDecCoinFromDec("uaaac", sdkmath.LegacyNewDecWithPrec(0, 3)),
						sdk.NewDecCoinFromDec("uband", sdkmath.LegacyNewDecWithPrec(3, 3)),
						sdk.NewDecCoinFromDec("uccca", sdkmath.LegacyNewDecWithPrec(0, 3)),
						sdk.NewDecCoinFromDec("ucccb", sdkmath.LegacyNewDecWithPrec(1, 3)),
						sdk.NewDecCoinFromDec("ucccc", sdkmath.LegacyNewDecWithPrec(100, 3)),
					),
				}
			},
			expIsBypassMinFeeTx: false,
			expErr:              nil,
			expFee: sdk.NewCoins(
				sdk.NewCoin("uaaaa", sdkmath.NewInt(100000)),
				sdk.NewCoin("uaaab", sdkmath.NewInt(1000)),
				sdk.NewCoin("uband", sdkmath.NewInt(3000)),
				sdk.NewCoin("ucccb", sdkmath.NewInt(1000)),
				sdk.NewCoin("ucccc", sdkmath.NewInt(100000)),
			),
			expPriority: 30,
		},
		{
			name: "valid MsgRequestData and valid MsgReport in valid MsgExec with enough fee",
			stubTx: func() *StubTx {
				msgReportData := oracletypes.NewMsgReportData(
					suite.requestID,
					[]oracletypes.RawReport{},
					bandtesting.Validators[0].ValAddress,
				)
				msgRequestData := oracletypes.NewMsgRequestData(
					1,
					BasicCalldata,
					1,
					1,
					BasicClientID,
					bandtesting.Coins100000000uband,
					bandtesting.TestDefaultPrepareGas,
					bandtesting.TestDefaultExecuteGas,
					bandtesting.FeePayer.Address,
				)
				msgs := []sdk.Msg{msgReportData, msgRequestData}
				authzMsg := authz.NewMsgExec(bandtesting.Alice.Address, msgs)

				return &StubTx{
					Msgs:      []sdk.Msg{&authzMsg},
					GasPrices: sdk.NewDecCoins(sdk.NewDecCoin("uband", sdkmath.NewInt(1))),
				}
			},
			expIsBypassMinFeeTx: false,
			expErr:              nil,
			expFee: sdk.NewCoins(
				sdk.NewCoin("uband", sdkmath.NewInt(1000000)),
			),
			expPriority: 10000,
		},
		{
			name: "valid MsgRequestData and valid MsgReport with enough fee",
			stubTx: func() *StubTx {
				msgReportData := oracletypes.NewMsgReportData(
					suite.requestID,
					[]oracletypes.RawReport{},
					bandtesting.Validators[0].ValAddress,
				)
				msgRequestData := oracletypes.NewMsgRequestData(
					1,
					BasicCalldata,
					1,
					1,
					BasicClientID,
					bandtesting.Coins100000000uband,
					bandtesting.TestDefaultPrepareGas,
					bandtesting.TestDefaultExecuteGas,
					bandtesting.FeePayer.Address,
				)

				return &StubTx{
					Msgs:      []sdk.Msg{msgReportData, msgRequestData},
					GasPrices: sdk.NewDecCoins(sdk.NewDecCoin("uband", sdkmath.NewInt(1))),
				}
			},
			expIsBypassMinFeeTx: false,
			expErr:              nil,
			expFee: sdk.NewCoins(
				sdk.NewCoin("uband", sdkmath.NewInt(1000000)),
			),
			expPriority: 10000,
		},
		{
			name: "valid MsgSubmitSignalPrices",
			stubTx: func() *StubTx {
				return &StubTx{
					Msgs: []sdk.Msg{
						feedstypes.NewMsgSubmitSignalPrices(
							bandtesting.Validators[0].ValAddress.String(),
							suite.ctx.BlockTime().Unix(),
							[]feedstypes.SignalPrice{},
						),
					},
				}
			},
			expIsBypassMinFeeTx: true,
			expErr:              nil,
			expFee:              sdk.Coins{},
			expPriority:         math.MaxInt64,
		},
		{
			name: "valid MsgSubmitSignalPrices in valid MsgExec",
			stubTx: func() *StubTx {
				msgExec := authz.NewMsgExec(bandtesting.Alice.Address, []sdk.Msg{
					feedstypes.NewMsgSubmitSignalPrices(
						bandtesting.Validators[0].ValAddress.String(),
						suite.ctx.BlockTime().Unix(),
						[]feedstypes.SignalPrice{},
					),
				})

				return &StubTx{
					Msgs: []sdk.Msg{
						&msgExec,
					},
				}
			},
			expIsBypassMinFeeTx: true,
			expErr:              nil,
			expFee:              sdk.Coins{},
			expPriority:         math.MaxInt64,
		},
		{
			name: "invalid MsgSubmitSignalPrices with not enough fee",
			stubTx: func() *StubTx {
				return &StubTx{
					Msgs: []sdk.Msg{
						feedstypes.NewMsgSubmitSignalPrices(
							bandtesting.Alice.ValAddress.String(),
							suite.ctx.BlockTime().Unix(),
							[]feedstypes.SignalPrice{},
						),
					},
				}
			},
			expIsBypassMinFeeTx: false,
			expErr:              sdkerrors.ErrInsufficientFee,
			expFee:              nil,
			expPriority:         0,
		},
		{
			name: "invalid MsgSubmitSignalPrices in valid MsgExec with not enough fee",
			stubTx: func() *StubTx {
				msgExec := authz.NewMsgExec(bandtesting.Alice.Address, []sdk.Msg{
					feedstypes.NewMsgSubmitSignalPrices(
						bandtesting.Alice.ValAddress.String(),
						suite.ctx.BlockTime().Unix(),
						[]feedstypes.SignalPrice{},
					),
				})

				return &StubTx{
					Msgs: []sdk.Msg{
						&msgExec,
					},
				}
			},
			expIsBypassMinFeeTx: false,
			expErr:              sdkerrors.ErrInsufficientFee,
			expFee:              nil,
			expPriority:         0,
		},
		{
			name: "valid MsgSubmitSignalPrices in invalid MsgExec with enough fee",
			stubTx: func() *StubTx {
				msgExec := authz.NewMsgExec(bandtesting.Bob.Address, []sdk.Msg{
					feedstypes.NewMsgSubmitSignalPrices(
						bandtesting.Validators[0].ValAddress.String(),
						suite.ctx.BlockTime().Unix(),
						[]feedstypes.SignalPrice{},
					),
				})

				return &StubTx{
					Msgs: []sdk.Msg{
						&msgExec,
					},
					GasPrices: sdk.NewDecCoins(sdk.NewDecCoin("uband", sdkmath.NewInt(1))),
				}
			},
			expIsBypassMinFeeTx: false,
			expErr:              nil,
			expFee:              sdk.NewCoins(sdk.NewCoin("uband", sdkmath.NewInt(1000000))),
			expPriority:         10000,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			stubTx := tc.stubTx()

			// test - IsByPassMinFeeTx
			isByPassMinFeeTx := suite.FeeChecker.IsBypassMinFeeTx(suite.ctx, stubTx)
			suite.Require().Equal(tc.expIsBypassMinFeeTx, isByPassMinFeeTx)

			// test - CheckTxFee
			fee, priority, err := suite.FeeChecker.CheckTxFee(suite.ctx, stubTx)
			suite.Require().ErrorIs(err, tc.expErr)
			suite.Require().Equal(fee, tc.expFee)
			suite.Require().Equal(tc.expPriority, priority)
		})
	}
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
