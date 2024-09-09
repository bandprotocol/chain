package oracle_test

// TODO: Fix test
// import (
// 	"fmt"
// 	"testing"
// 	"time"

// 	"cosmossdk.io/math"
// 	abci "github.com/cometbft/cometbft/abci/types"
// 	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
// 	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
// 	sdk "github.com/cosmos/cosmos-sdk/types"
// 	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
// 	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
// 	"github.com/stretchr/testify/require"

// 	bandtesting "github.com/bandprotocol/chain/v3/testing"
// 	"github.com/bandprotocol/chain/v3/x/oracle"
// 	"github.com/bandprotocol/chain/v3/x/oracle/types"
// )

// func TestSuccessRequestOracleData(t *testing.T) {
// 	app, ctx := bandtesting.CreateTestApp(t, true)
// 	k := app.OracleKeeper

// 	ctx = ctx.WithBlockHeight(4).WithBlockTime(time.Unix(1581589790, 0))
// 	requestMsg := types.NewMsgRequestData(
// 		types.OracleScriptID(1),
// 		[]byte("calldata"),
// 		3,
// 		2,
// 		"app_test",
// 		sdk.NewCoins(sdk.NewCoin("uband", math.NewInt(9000000))),
// 		bandtesting.TestDefaultPrepareGas,
// 		bandtesting.TestDefaultExecuteGas,
// 		bandtesting.Validators[0].Address,
// 	)

// 	acc := &authtypes.BaseAccount{
// 		Address: bandtesting.Validators[0].Address.String(),
// 	}

// 	res1 := app.AccountKeeper.GetAccount(ctx, bandtesting.Validators[0].Address)
// 	require.NotNil(t, res1)
// 	require.Equal(t, acc, res1.(*authtypes.BaseAccount))

// 	origAccNum := res1.GetAccountNumber()
// 	origSeq := res1.GetSequence()

// 	header := cmtproto.Header{Height: app.BaseApp.LastBlockHeight() + 1}
// 	txConfig := moduletestutil.MakeTestTxConfig()
// 	// res, err := handler(ctx, requestMsg)
// 	_, res, err := simtestutil.SignCheckDeliver(
// 		t,
// 		txConfig,
// 		app.BaseApp,
// 		header,
// 		[]sdk.Msg{requestMsg},
// 		"",
// 		[]uint64{origAccNum},
// 		[]uint64{origSeq},
// 		false,
// 		false,
// 		bandtesting.Validators[0].PrivKey,
// 	)
// 	require.NotNil(t, res)
// 	require.NoError(t, err)

// 	expectRequest := types.NewRequest(
// 		types.OracleScriptID(1),
// 		[]byte("calldata"),
// 		[]sdk.ValAddress{
// 			bandtesting.Validators[2].ValAddress,
// 			bandtesting.Validators[0].ValAddress,
// 			bandtesting.Validators[1].ValAddress,
// 		},
// 		2,
// 		4,
// 		bandtesting.ParseTime(1581589790),
// 		"app_test",
// 		[]types.RawRequest{
// 			types.NewRawRequest(1, 1, []byte("beeb")),
// 			types.NewRawRequest(2, 2, []byte("beeb")),
// 			types.NewRawRequest(3, 3, []byte("beeb")),
// 		},
// 		nil,
// 		bandtesting.TestDefaultExecuteGas,
// 	)
// 	_, err = app.EndBlocker(ctx.WithBlockHeight(4))
// 	require.NoError(t, err)
// 	request, err := k.GetRequest(ctx, types.RequestID(1))
// 	require.NoError(t, err)
// 	require.Equal(t, expectRequest, request)

// 	reportMsg1 := types.NewMsgReportData(
// 		types.RequestID(1), []types.RawReport{
// 			types.NewRawReport(1, 0, []byte("answer1")),
// 			types.NewRawReport(2, 0, []byte("answer2")),
// 			types.NewRawReport(3, 0, []byte("answer3")),
// 		},
// 		bandtesting.Validators[0].ValAddress,
// 	)
// 	res, err = handler(ctx, reportMsg1)
// 	require.NotNil(t, res)
// 	require.NoError(t, err)

// 	ids := k.GetPendingResolveList(ctx)
// 	require.Equal(t, []types.RequestID{}, ids)
// 	_, err = k.GetResult(ctx, types.RequestID(1))
// 	require.Error(t, err)

// 	result, err := app.EndBlocker(ctx.WithBlockHeight(6))
// 	require.NoError(t, err)
// 	expectEvents := []abci.Event{}

// 	require.Equal(t, expectEvents, result.Events)

// 	ctx = ctx.WithBlockTime(time.Unix(1581589795, 0))
// 	reportMsg2 := types.NewMsgReportData(
// 		types.RequestID(1), []types.RawReport{
// 			types.NewRawReport(1, 0, []byte("answer1")),
// 			types.NewRawReport(2, 0, []byte("answer2")),
// 			types.NewRawReport(3, 0, []byte("answer3")),
// 		},
// 		bandtesting.Validators[1].ValAddress,
// 	)
// 	res, err = handler(ctx, reportMsg2)
// 	require.NotNil(t, res)
// 	require.NoError(t, err)

// 	ids = k.GetPendingResolveList(ctx)
// 	require.Equal(t, []types.RequestID{1}, ids)
// 	_, err = k.GetResult(ctx, types.RequestID(1))
// 	require.Error(t, err)

// 	result, err = app.EndBlocker(ctx.WithBlockHeight(8))
// 	require.NoError(t, err)
// 	resPacket := types.NewOracleResponsePacketData(
// 		expectRequest.ClientID, types.RequestID(1), 2, expectRequest.RequestTime, 1581589795,
// 		types.RESOLVE_STATUS_SUCCESS, []byte("beeb"),
// 	)
// 	expectEvents = []abci.Event{{Type: types.EventTypeResolve, Attributes: []abci.EventAttribute{
// 		{Key: types.AttributeKeyID, Value: fmt.Sprint(resPacket.RequestID)},
// 		{Key: types.AttributeKeyResolveStatus, Value: fmt.Sprint(uint32(resPacket.ResolveStatus))},
// 		{Key: types.AttributeKeyResult, Value: "62656562"},
// 		{Key: types.AttributeKeyGasUsed, Value: "2485000000"},
// 	}}}

// 	require.Equal(t, expectEvents, result.Events)

// 	ids = k.GetPendingResolveList(ctx)
// 	require.Equal(t, []types.RequestID{}, ids)

// 	req, err := k.GetRequest(ctx, types.RequestID(1))
// 	require.NotEqual(t, types.Request{}, req)
// 	require.NoError(t, err)

// 	app.EndBlocker(ctx.WithBlockHeight(32).WithBlockTime(ctx.BlockTime().Add(time.Minute)))
// }

// func TestExpiredRequestOracleData(t *testing.T) {
// 	app, ctx := bandtesting.CreateTestApp(t, true)
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
// 		bandtesting.TestDefaultPrepareGas,
// 		bandtesting.TestDefaultExecuteGas,
// 		bandtesting.Validators[0].Address,
// 	)
// 	res, err := handler(ctx, requestMsg)
// 	require.NotNil(t, res)
// 	require.NoError(t, err)

// 	expectRequest := types.NewRequest(
// 		types.OracleScriptID(1),
// 		[]byte("calldata"),
// 		[]sdk.ValAddress{
// 			bandtesting.Validators[2].ValAddress,
// 			bandtesting.Validators[0].ValAddress,
// 			bandtesting.Validators[1].ValAddress,
// 		},
// 		2,
// 		4,
// 		bandtesting.ParseTime(1581589790),
// 		"app_test",
// 		[]types.RawRequest{
// 			types.NewRawRequest(1, 1, []byte("beeb")),
// 			types.NewRawRequest(2, 2, []byte("beeb")),
// 			types.NewRawRequest(3, 3, []byte("beeb")),
// 		},
// 		nil,
// 		bandtesting.TestDefaultExecuteGas,
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
// 				Value: fmt.Sprint(bandtesting.Validators[2].ValAddress.String()),
// 			},
// 		},
// 	}, {
// 		Type: types.EventTypeDeactivate,
// 		Attributes: []abci.EventAttribute{
// 			{
// 				Key:   types.AttributeKeyValidator,
// 				Value: fmt.Sprint(bandtesting.Validators[0].ValAddress.String()),
// 			},
// 		},
// 	}, {
// 		Type: types.EventTypeDeactivate,
// 		Attributes: []abci.EventAttribute{
// 			{
// 				Key:   types.AttributeKeyValidator,
// 				Value: fmt.Sprint(bandtesting.Validators[1].ValAddress.String()),
// 			},
// 		},
// 	}}

// 	require.Equal(t, expectEvents, result.Events)
// }
