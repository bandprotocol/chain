package oracle_test

import (
	"fmt"
	"testing"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/bandprotocol/chain/v2/testing/testapp"
	"github.com/bandprotocol/chain/v2/x/oracle"
	"github.com/bandprotocol/chain/v2/x/oracle/types"
)

func TestSuccessRequestOracleData(t *testing.T) {
	app, ctx, k := testapp.CreateTestInput(true)

	ctx = ctx.WithBlockHeight(4).WithBlockTime(time.Unix(1581589790, 0))
	handler := oracle.NewHandler(k)
	requestMsg := types.NewMsgRequestData(
		types.OracleScriptID(1),
		[]byte("calldata"),
		3,
		2,
		"app_test",
		sdk.NewCoins(sdk.NewCoin("uband", sdk.NewInt(9000000))),
		testapp.TestDefaultPrepareGas,
		testapp.TestDefaultExecuteGas,
		testapp.Validators[0].Address,
	)
	res, err := handler(ctx, requestMsg)
	require.NotNil(t, res)
	require.NoError(t, err)

	expectRequest := types.NewRequest(
		types.OracleScriptID(1),
		[]byte("calldata"),
		[]sdk.ValAddress{
			testapp.Validators[2].ValAddress,
			testapp.Validators[0].ValAddress,
			testapp.Validators[1].ValAddress,
		},
		2,
		4,
		testapp.ParseTime(1581589790),
		"app_test",
		[]types.RawRequest{
			types.NewRawRequest(1, 1, []byte("beeb")),
			types.NewRawRequest(2, 2, []byte("beeb")),
			types.NewRawRequest(3, 3, []byte("beeb")),
		},
		nil,
		testapp.TestDefaultExecuteGas,
	)
	app.EndBlocker(ctx, abci.RequestEndBlock{Height: 4})
	request, err := k.GetRequest(ctx, types.RequestID(1))
	require.NoError(t, err)
	require.Equal(t, expectRequest, request)

	reportMsg1 := types.NewMsgReportData(
		types.RequestID(1), []types.RawReport{
			types.NewRawReport(1, 0, []byte("answer1")),
			types.NewRawReport(2, 0, []byte("answer2")),
			types.NewRawReport(3, 0, []byte("answer3")),
		},
		testapp.Validators[0].ValAddress,
	)
	res, err = handler(ctx, reportMsg1)
	require.NotNil(t, res)
	require.NoError(t, err)

	ids := k.GetPendingResolveList(ctx)
	require.Equal(t, []types.RequestID{}, ids)
	_, err = k.GetResult(ctx, types.RequestID(1))
	require.Error(t, err)

	result := app.EndBlocker(ctx, abci.RequestEndBlock{Height: 6})
	expectEvents := []abci.Event{}

	require.Equal(t, expectEvents, result.GetEvents())

	ctx = ctx.WithBlockTime(time.Unix(1581589795, 0))
	reportMsg2 := types.NewMsgReportData(
		types.RequestID(1), []types.RawReport{
			types.NewRawReport(1, 0, []byte("answer1")),
			types.NewRawReport(2, 0, []byte("answer2")),
			types.NewRawReport(3, 0, []byte("answer3")),
		},
		testapp.Validators[1].ValAddress,
	)
	res, err = handler(ctx, reportMsg2)
	require.NotNil(t, res)
	require.NoError(t, err)

	ids = k.GetPendingResolveList(ctx)
	require.Equal(t, []types.RequestID{1}, ids)
	_, err = k.GetResult(ctx, types.RequestID(1))
	require.Error(t, err)

	result = app.EndBlocker(ctx, abci.RequestEndBlock{Height: 8})
	resPacket := types.NewOracleResponsePacketData(
		expectRequest.ClientID, types.RequestID(1), 2, int64(expectRequest.RequestTime), 1581589795,
		types.RESOLVE_STATUS_SUCCESS, []byte("beeb"),
	)
	expectEvents = []abci.Event{{Type: types.EventTypeResolve, Attributes: []abci.EventAttribute{
		{Key: types.AttributeKeyID, Value: fmt.Sprint(resPacket.RequestID)},
		{Key: types.AttributeKeyResolveStatus, Value: fmt.Sprint(uint32(resPacket.ResolveStatus))},
		{Key: types.AttributeKeyResult, Value: "62656562"},
		{Key: types.AttributeKeyGasUsed, Value: "2485000000"},
	}}}

	require.Equal(t, expectEvents, result.GetEvents())

	ids = k.GetPendingResolveList(ctx)
	require.Equal(t, []types.RequestID{}, ids)

	req, err := k.GetRequest(ctx, types.RequestID(1))
	require.NotEqual(t, types.Request{}, req)
	require.NoError(t, err)

	ctx = ctx.WithBlockHeight(32).WithBlockTime(ctx.BlockTime().Add(time.Minute))
	app.EndBlocker(ctx, abci.RequestEndBlock{Height: 32})
}

func TestExpiredRequestOracleData(t *testing.T) {
	app, ctx, k := testapp.CreateTestInput(true)

	ctx = ctx.WithBlockHeight(4).WithBlockTime(time.Unix(1581589790, 0))
	handler := oracle.NewHandler(k)
	requestMsg := types.NewMsgRequestData(
		types.OracleScriptID(1),
		[]byte("calldata"),
		3,
		2,
		"app_test",
		sdk.NewCoins(sdk.NewCoin("uband", sdk.NewInt(9000000))),
		testapp.TestDefaultPrepareGas,
		testapp.TestDefaultExecuteGas,
		testapp.Validators[0].Address,
	)
	res, err := handler(ctx, requestMsg)
	require.NotNil(t, res)
	require.NoError(t, err)

	expectRequest := types.NewRequest(
		types.OracleScriptID(1),
		[]byte("calldata"),
		[]sdk.ValAddress{
			testapp.Validators[2].ValAddress,
			testapp.Validators[0].ValAddress,
			testapp.Validators[1].ValAddress,
		},
		2,
		4,
		testapp.ParseTime(1581589790),
		"app_test",
		[]types.RawRequest{
			types.NewRawRequest(1, 1, []byte("beeb")),
			types.NewRawRequest(2, 2, []byte("beeb")),
			types.NewRawRequest(3, 3, []byte("beeb")),
		},
		nil,
		testapp.TestDefaultExecuteGas,
	)
	app.EndBlocker(ctx, abci.RequestEndBlock{Height: 4})
	request, err := k.GetRequest(ctx, types.RequestID(1))
	require.NoError(t, err)
	require.Equal(t, expectRequest, request)

	ctx = ctx.WithBlockHeight(132).WithBlockTime(ctx.BlockTime().Add(time.Minute))
	result := app.EndBlocker(ctx, abci.RequestEndBlock{Height: 132})
	resPacket := types.NewOracleResponsePacketData(
		expectRequest.ClientID, types.RequestID(1), 0, int64(expectRequest.RequestTime), ctx.BlockTime().Unix(),
		types.RESOLVE_STATUS_EXPIRED, []byte{},
	)
	expectEvents := []abci.Event{{
		Type: types.EventTypeResolve,
		Attributes: []abci.EventAttribute{
			{Key: types.AttributeKeyID, Value: fmt.Sprint(resPacket.RequestID)},
			{
				Key:   types.AttributeKeyResolveStatus,
				Value: fmt.Sprint(uint32(resPacket.ResolveStatus)),
			},
		},
	}, {
		Type: types.EventTypeDeactivate,
		Attributes: []abci.EventAttribute{
			{
				Key:   types.AttributeKeyValidator,
				Value: fmt.Sprint(testapp.Validators[2].ValAddress.String()),
			},
		},
	}, {
		Type: types.EventTypeDeactivate,
		Attributes: []abci.EventAttribute{
			{
				Key:   types.AttributeKeyValidator,
				Value: fmt.Sprint(testapp.Validators[0].ValAddress.String()),
			},
		},
	}, {
		Type: types.EventTypeDeactivate,
		Attributes: []abci.EventAttribute{
			{
				Key:   types.AttributeKeyValidator,
				Value: fmt.Sprint(testapp.Validators[1].ValAddress.String()),
			},
		},
	}}

	require.Equal(t, expectEvents, result.GetEvents())
}
