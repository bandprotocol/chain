package oracle_test

// TODO: Fix test
import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	abci "github.com/cometbft/cometbft/abci/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"

	bandtest "github.com/bandprotocol/chain/v3/app"
	bandtesting "github.com/bandprotocol/chain/v3/testing"
	"github.com/bandprotocol/chain/v3/x/oracle/types"
)

type AppTestSuite struct {
	suite.Suite

	app *bandtest.BandApp

	// For test teardown
	dir string
}

func TestAppTestSuite(t *testing.T) {
	suite.Run(t, new(AppTestSuite))
}

func (s *AppTestSuite) SetupTest() {
	dir := testutil.GetTempDir(s.T())
	s.app = bandtest.SetupWithCustomHome(false, dir)
	ctx := s.app.BaseApp.NewUncachedContext(false, tmproto.Header{})

	// Activate validators
	for _, v := range bandtest.Validators {
		err := s.app.OracleKeeper.Activate(ctx, v.ValAddress)
		s.Require().NoError(err)
	}

	_, err := s.app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: s.app.LastBlockHeight() + 1})
	s.Require().NoError(err)
	_, err = s.app.Commit()
	s.Require().NoError(err)
}

func (s *AppTestSuite) TestSuccessRequestOracleData() {
	require := s.Require()

	ctx := s.app.BaseApp.NewUncachedContext(false, tmproto.Header{})
	requestMsg := types.NewMsgRequestData(
		types.OracleScriptID(1),
		[]byte("calldata"),
		3,
		2,
		"app_test",
		sdk.NewCoins(sdk.NewInt64Coin("uband", 9000000)),
		bandtest.TestDefaultPrepareGas,
		bandtest.TestDefaultExecuteGas,
		bandtest.Validators[0].Address,
	)

	res1 := s.app.AccountKeeper.GetAccount(ctx, bandtest.Validators[0].Address)
	require.NotNil(res1)

	acc1Num := res1.GetAccountNumber()
	acc1Seq := res1.GetSequence()

	txConfig := moduletestutil.MakeTestTxConfig()
	_, res, _, err := bandtest.SignCheckDeliver(
		s.T(),
		txConfig,
		s.app.BaseApp,
		tmproto.Header{Height: s.app.LastBlockHeight() + 1, Time: time.Unix(1581589790, 0)},
		[]sdk.Msg{requestMsg},
		s.app.ChainID(),
		[]uint64{acc1Num},
		[]uint64{acc1Seq},
		true,
		true,
		bandtest.Validators[0].PrivKey,
	)
	require.NotNil(res)
	require.NoError(err)

	expectRequest := types.NewRequest(
		types.OracleScriptID(1),
		[]byte("calldata"),
		[]sdk.ValAddress{
			bandtest.Validators[2].ValAddress,
			bandtest.Validators[0].ValAddress,
			bandtest.Validators[1].ValAddress,
		},
		2,
		2,
		bandtesting.ParseTime(1581589790),
		"app_test",
		[]types.RawRequest{
			types.NewRawRequest(1, 1, []byte("test")),
			types.NewRawRequest(2, 2, []byte("test")),
			types.NewRawRequest(3, 3, []byte("test")),
		},
		nil,
		bandtest.TestDefaultExecuteGas,
	)

	request, err := s.app.OracleKeeper.GetRequest(ctx, types.RequestID(1))
	require.NoError(err)
	require.Equal(expectRequest, request)

	reportMsg1 := types.NewMsgReportData(
		types.RequestID(1), []types.RawReport{
			types.NewRawReport(1, 0, []byte("answer1")),
			types.NewRawReport(2, 0, []byte("answer2")),
			types.NewRawReport(3, 0, []byte("answer3")),
		},
		bandtest.Validators[0].ValAddress,
	)
	_, res, _, err = bandtest.SignCheckDeliver(
		s.T(),
		txConfig,
		s.app.BaseApp,
		tmproto.Header{Height: s.app.LastBlockHeight() + 1, Time: time.Unix(1581589791, 0)},
		[]sdk.Msg{reportMsg1},
		s.app.ChainID(),
		[]uint64{acc1Num},
		[]uint64{acc1Seq + 1},
		true,
		true,
		bandtest.Validators[0].PrivKey,
	)
	require.NotNil(res)
	require.NoError(err)

	ids := s.app.OracleKeeper.GetPendingResolveList(ctx)
	require.Equal([]types.RequestID{}, ids)
	_, err = s.app.OracleKeeper.GetResult(ctx, types.RequestID(1))
	require.Error(err)

	reportMsg2 := types.NewMsgReportData(
		types.RequestID(1), []types.RawReport{
			types.NewRawReport(1, 0, []byte("answer1")),
			types.NewRawReport(2, 0, []byte("answer2")),
			types.NewRawReport(3, 0, []byte("answer3")),
		},
		bandtest.Validators[1].ValAddress,
	)

	res2 := s.app.AccountKeeper.GetAccount(ctx, bandtest.Validators[1].Address)
	require.NotNil(res2)

	acc2Num := res2.GetAccountNumber()
	acc2Seq := res2.GetSequence()

	// res, err = handler(ctx, reportMsg2)
	_, res, endBlockEvent, err := bandtest.SignCheckDeliver(
		s.T(),
		txConfig,
		s.app.BaseApp,
		tmproto.Header{Height: s.app.LastBlockHeight() + 1, Time: time.Unix(1581589795, 0)},
		[]sdk.Msg{reportMsg2},
		s.app.ChainID(),
		[]uint64{acc2Num},
		[]uint64{acc2Seq},
		true,
		true,
		bandtest.Validators[1].PrivKey,
	)

	require.NotNil(res)
	require.NoError(err)

	resPacket := types.NewOracleResponsePacketData(
		expectRequest.ClientID, types.RequestID(1), 2, expectRequest.RequestTime, 1581589795,
		types.RESOLVE_STATUS_SUCCESS, []byte("test"),
	)
	expRes := types.NewResult(
		resPacket.ClientID,
		types.OracleScriptID(1),
		[]byte("calldata"),
		3,
		2,
		types.RequestID(1),
		2,
		time.Unix(1581589790, 0).Unix(),
		time.Unix(1581589795, 0).Unix(),
		resPacket.ResolveStatus,
		resPacket.Result,
	)

	// Resolve event must contain in block event
	expectEvent := abci.Event{Type: types.EventTypeResolve, Attributes: []abci.EventAttribute{
		{Key: types.AttributeKeyID, Value: fmt.Sprint(resPacket.RequestID), Index: true},
		{Key: types.AttributeKeyResolveStatus, Value: fmt.Sprint(uint32(resPacket.ResolveStatus)), Index: true},
		{Key: types.AttributeKeyResult, Value: "74657374", Index: true},
		{Key: types.AttributeKeyGasUsed, Value: "2485000000", Index: true},
		{Key: "mode", Value: "EndBlock", Index: true},
	}}

	require.Contains(endBlockEvent, expectEvent)

	ctx2 := s.app.BaseApp.NewContext(true)

	// Endblock should have been called and no pending request after endblock
	ids = s.app.OracleKeeper.GetPendingResolveList(ctx)
	require.Equal([]types.RequestID{}, ids)

	// Request 1 still remain until expired
	req, err := s.app.OracleKeeper.GetRequest(ctx, types.RequestID(1))
	require.NotEqual(types.Request{}, req)
	require.NoError(err)

	// Result 1 should be available
	result, err := s.app.OracleKeeper.GetResult(ctx2, types.RequestID(1))
	require.NoError(err)
	require.Equal(expRes, result)
}

// func TestExpiredRequestOracleData(t *testing.T) {
// 	app, ctx := bandtest.CreateTestApp(t, true)
// 	k := app.OracleKeeper

// 	ctx = ctx.WithBlockHeight(4).WithBlockTime(time.Unix(1581589790, 0))
// 	handler := oracle.NewHandler(k)
// 	requestMsg := types.NewMsgRequestData(
// 		types.OracleScriptID(1),
// 		[]byte("calldata"),
// 		3,
// 		2,
// 		"app_test",
// 		sdk.NewCoins(sdk.NewCoin("uband", math.NewInt(9000000))),
// 		bandtest.TestDefaultPrepareGas,
// 		bandtest.TestDefaultExecuteGas,
// 		bandtest.Validators[0].Address,
// 	)
// 	res, err := handler(ctx, requestMsg)
// 	require.NotNil(t, res)
// 	require.NoError(t, err)

// 	expectRequest := types.NewRequest(
// 		types.OracleScriptID(1),
// 		[]byte("calldata"),
// 		[]sdk.ValAddress{
// 			bandtest.Validators[2].ValAddress,
// 			bandtest.Validators[0].ValAddress,
// 			bandtest.Validators[1].ValAddress,
// 		},
// 		2,
// 		4,
// 		bandtest.ParseTime(1581589790),
// 		"app_test",
// 		[]types.RawRequest{
// 			types.NewRawRequest(1, 1, []byte("test")),
// 			types.NewRawRequest(2, 2, []byte("test")),
// 			types.NewRawRequest(3, 3, []byte("test")),
// 		},
// 		nil,
// 		bandtest.TestDefaultExecuteGas,
// 	)
// 	app.EndBlocker(ctx.WithBlockHeight(4))
// 	request, err := k.GetRequest(ctx, types.RequestID(1))
// 	require.NoError(t, err)
// 	require.Equal(t, expectRequest, request)

// 	result, err := app.EndBlocker(ctx.WithBlockHeight(132).WithBlockTime(ctx.BlockTime().Add(time.Minute)))
// 	require.NoError(t, err)
// 	resPacket := types.NewOracleResponsePacketData(
// 		expectRequest.ClientID, types.RequestID(1), 0, expectRequest.RequestTime, ctx.BlockTime().Unix(),
// 		types.RESOLVE_STATUS_EXPIRED, []byte{},
// 	)
// 	expectEvents := []abci.Event{{
// 		Type: types.EventTypeResolve,
// 		Attributes: []abci.EventAttribute{
// 			{Key: types.AttributeKeyID, Value: fmt.Sprint(resPacket.RequestID)},
// 			{
// 				Key:   types.AttributeKeyResolveStatus,
// 				Value: fmt.Sprint(uint32(resPacket.ResolveStatus)),
// 			},
// 		},
// 	}, {
// 		Type: types.EventTypeDeactivate,
// 		Attributes: []abci.EventAttribute{
// 			{
// 				Key:   types.AttributeKeyValidator,
// 				Value: fmt.Sprint(bandtest.Validators[2].ValAddress.String()),
// 			},
// 		},
// 	}, {
// 		Type: types.EventTypeDeactivate,
// 		Attributes: []abci.EventAttribute{
// 			{
// 				Key:   types.AttributeKeyValidator,
// 				Value: fmt.Sprint(bandtest.Validators[0].ValAddress.String()),
// 			},
// 		},
// 	}, {
// 		Type: types.EventTypeDeactivate,
// 		Attributes: []abci.EventAttribute{
// 			{
// 				Key:   types.AttributeKeyValidator,
// 				Value: fmt.Sprint(bandtest.Validators[1].ValAddress.String()),
// 			},
// 		},
// 	}}

// 	require.Equal(t, expectEvents, result.Events)
// }
