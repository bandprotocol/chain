package oracle_test

import (
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/pkg/obi"
	bandtesting "github.com/bandprotocol/chain/v3/testing"
	"github.com/bandprotocol/chain/v3/testing/testdata"
	"github.com/bandprotocol/chain/v3/x/oracle/types"
)

const (
	chainID            = "BANDCHAIN"
	basicName          = "BASIC_NAME"
	basicDesc          = "BASIC_DESCRIPTION"
	basicSchema        = "BASIC_SCHEMA"
	basicSourceCodeURL = "BASIC_SOURCE_CODE_URL"
	basicFilename      = "BASIC_FILENAME"
	basicClientID      = "BASIC_CLIENT_ID"
)

var basicCalldata = []byte("BASIC_CALLDATA")

func (s *AppTestSuite) TestPrepareRequestNotEnoughMaxFee() {
	require := s.Require()
	app := s.app
	ctx := s.app.BaseApp.NewUncachedContext(false, cmtproto.Header{})
	k := app.OracleKeeper

	ctx = ctx.WithBlockTime(bandtesting.ParseTime(1581589790)).WithBlockHeight(42)
	// OracleScript#1: Prepare asks for DS#1,2,3 with ExtID#1,2,3 and calldata "test"
	m := types.NewMsgRequestData(
		1,
		basicCalldata,
		1,
		1,
		basicClientID,
		bandtesting.EmptyCoins,
		bandtesting.TestDefaultPrepareGas,
		bandtesting.TestDefaultExecuteGas,
		bandtesting.FeePayer.Address,
	)
	_, err := k.PrepareRequest(ctx, m, bandtesting.FeePayer.Address, nil)
	require.EqualError(err, "require: 1000000uband, max: 0uband: not enough fee")
	m = types.NewMsgRequestData(
		1,
		basicCalldata,
		1,
		1,
		basicClientID,
		sdk.NewCoins(sdk.NewInt64Coin("uband", 1000000)),
		bandtesting.TestDefaultPrepareGas,
		bandtesting.TestDefaultExecuteGas,
		bandtesting.FeePayer.Address,
	)
	_, err = k.PrepareRequest(ctx, m, bandtesting.FeePayer.Address, nil)
	require.EqualError(err, "require: 2000000uband, max: 1000000uband: not enough fee")
	m = types.NewMsgRequestData(
		1,
		basicCalldata,
		1,
		1,
		basicClientID,
		sdk.NewCoins(sdk.NewInt64Coin("uband", 2000000)),
		bandtesting.TestDefaultPrepareGas,
		bandtesting.TestDefaultExecuteGas,
		bandtesting.FeePayer.Address,
	)
	_, err = k.PrepareRequest(ctx, m, bandtesting.FeePayer.Address, nil)
	require.EqualError(err, "require: 3000000uband, max: 2000000uband: not enough fee")
	m = types.NewMsgRequestData(
		1,
		basicCalldata,
		1,
		1,
		basicClientID,
		sdk.NewCoins(sdk.NewInt64Coin("uband", 2999999)),
		bandtesting.TestDefaultPrepareGas,
		bandtesting.TestDefaultExecuteGas,
		bandtesting.FeePayer.Address,
	)
	_, err = k.PrepareRequest(ctx, m, bandtesting.FeePayer.Address, nil)
	require.EqualError(err, "require: 3000000uband, max: 2999999uband: not enough fee")
	m = types.NewMsgRequestData(
		1,
		basicCalldata,
		1,
		1,
		basicClientID,
		sdk.NewCoins(sdk.NewInt64Coin("uband", 3000000)),
		bandtesting.TestDefaultPrepareGas,
		bandtesting.TestDefaultExecuteGas,
		bandtesting.FeePayer.Address,
	)
	id, err := k.PrepareRequest(ctx, m, bandtesting.FeePayer.Address, nil)
	require.NoError(err)
	require.Equal(types.RequestID(1), id)
}

func (s *AppTestSuite) TestPrepareRequestNotEnoughFund() {
	require := s.Require()
	app := s.app
	ctx := s.app.BaseApp.NewUncachedContext(false, cmtproto.Header{})
	k := app.OracleKeeper

	ctx = ctx.WithBlockTime(bandtesting.ParseTime(1581589790)).WithBlockHeight(42)
	// OracleScript#1: Prepare asks for DS#1,2,3 with ExtID#1,2,3 and calldata "test"
	m := types.NewMsgRequestData(
		1,
		basicCalldata,
		1,
		1,
		basicClientID,
		bandtesting.Coins100000000uband,
		bandtesting.TestDefaultPrepareGas,
		bandtesting.TestDefaultExecuteGas,
		bandtesting.Alice.Address,
	)
	_, err := k.PrepareRequest(ctx, m, bandtesting.Alice.Address, nil)
	require.EqualError(err, "spendable balance 0uband is smaller than 1000000uband: insufficient funds")
}

func (s *AppTestSuite) TestPrepareRequestNotEnoughPrepareGas() {
	require := s.Require()
	app := s.app
	ctx := s.app.BaseApp.NewUncachedContext(false, cmtproto.Header{})
	k := app.OracleKeeper

	ctx = ctx.WithBlockTime(bandtesting.ParseTime(1581589790)).WithBlockHeight(42)

	wrappedGasMeter := bandtesting.NewGasMeterWrapper(ctx.GasMeter())
	ctx = ctx.WithGasMeter(wrappedGasMeter)

	m := types.NewMsgRequestData(
		1,
		basicCalldata,
		1,
		1,
		basicClientID,
		bandtesting.EmptyCoins,
		1,
		bandtesting.TestDefaultExecuteGas,
		bandtesting.Alice.Address,
	)
	_, err := k.PrepareRequest(ctx, m, bandtesting.FeePayer.Address, nil)
	require.ErrorIs(err, types.ErrBadWasmExecution)
	require.Contains(err.Error(), "out-of-gas")

	params := k.GetParams(ctx)
	require.Equal(1, wrappedGasMeter.CountRecord(params.BaseOwasmGas, "BASE_OWASM_FEE"))
	require.Equal(0, wrappedGasMeter.CountRecord(100, "OWASM_PREPARE_FEE"))
	require.Equal(0, wrappedGasMeter.CountDescriptor("OWASM_EXECUTE_FEE"))
}

func (s *AppTestSuite) TestPrepareRequestInvalidAskCountFail() {
	require := s.Require()
	app := s.app
	ctx := s.app.BaseApp.NewUncachedContext(false, cmtproto.Header{})
	k := app.OracleKeeper

	params := k.GetParams(ctx)
	params.MaxAskCount = 5
	err := k.SetParams(ctx, params)
	require.NoError(err)

	wrappedGasMeter := bandtesting.NewGasMeterWrapper(ctx.GasMeter())
	ctx = ctx.WithGasMeter(wrappedGasMeter)

	m := types.NewMsgRequestData(
		1,
		basicCalldata,
		10,
		1,
		basicClientID,
		bandtesting.Coins100000000uband,
		bandtesting.TestDefaultPrepareGas,
		bandtesting.TestDefaultExecuteGas,
		bandtesting.Alice.Address,
	)
	_, err = k.PrepareRequest(ctx, m, bandtesting.FeePayer.Address, nil)
	require.ErrorIs(err, types.ErrInvalidAskCount)

	require.Equal(0, wrappedGasMeter.CountDescriptor("BASE_OWASM_FEE"))
	require.Equal(0, wrappedGasMeter.CountDescriptor("OWASM_PREPARE_FEE"))
	require.Equal(0, wrappedGasMeter.CountDescriptor("OWASM_EXECUTE_FEE"))

	m = types.NewMsgRequestData(
		1,
		basicCalldata,
		4,
		1,
		basicClientID,
		bandtesting.Coins100000000uband,
		bandtesting.TestDefaultPrepareGas,
		bandtesting.TestDefaultExecuteGas,
		bandtesting.Alice.Address,
	)
	_, err = k.PrepareRequest(ctx, m, bandtesting.FeePayer.Address, nil)
	require.ErrorIs(err, types.ErrInsufficientValidators)

	require.Equal(0, wrappedGasMeter.CountDescriptor("BASE_OWASM_FEE"))
	require.Equal(0, wrappedGasMeter.CountDescriptor("OWASM_PREPARE_FEE"))
	require.Equal(0, wrappedGasMeter.CountDescriptor("OWASM_EXECUTE_FEE"))

	m = types.NewMsgRequestData(
		1,
		basicCalldata,
		1,
		1,
		basicClientID,
		bandtesting.Coins100000000uband,
		bandtesting.TestDefaultPrepareGas,
		bandtesting.TestDefaultExecuteGas,
		bandtesting.Alice.Address,
	)
	id, err := k.PrepareRequest(ctx, m, bandtesting.FeePayer.Address, nil)
	require.Equal(types.RequestID(1), id)
	require.NoError(err)
	require.Equal(2, wrappedGasMeter.CountDescriptor("BASE_OWASM_FEE"))
	require.Equal(1, wrappedGasMeter.CountDescriptor("OWASM_PREPARE_FEE"))
	require.Equal(1, wrappedGasMeter.CountDescriptor("OWASM_EXECUTE_FEE"))
}

func (s *AppTestSuite) TestPrepareRequestBaseOwasmFeePanic() {
	require := s.Require()
	app := s.app
	ctx := s.app.BaseApp.NewUncachedContext(false, cmtproto.Header{})
	k := app.OracleKeeper

	params := k.GetParams(ctx)
	params.BaseOwasmGas = 100000
	params.PerValidatorRequestGas = 0
	err := k.SetParams(ctx, params)
	require.NoError(err)
	m := types.NewMsgRequestData(
		1,
		basicCalldata,
		1,
		1,
		basicClientID,
		bandtesting.Coins100000000uband,
		bandtesting.TestDefaultPrepareGas,
		bandtesting.TestDefaultExecuteGas,
		bandtesting.Alice.Address,
	)
	ctx = ctx.WithGasMeter(storetypes.NewGasMeter(90000))
	require.PanicsWithValue(
		storetypes.ErrorOutOfGas{Descriptor: "BASE_OWASM_FEE"},
		func() { _, _ = k.PrepareRequest(ctx, m, bandtesting.FeePayer.Address, nil) },
	)
	ctx = ctx.WithGasMeter(storetypes.NewGasMeter(1000000))
	id, err := k.PrepareRequest(ctx, m, bandtesting.FeePayer.Address, nil)
	require.Equal(types.RequestID(1), id)
	require.NoError(err)
}

func (s *AppTestSuite) TestPrepareRequestPerValidatorRequestFeePanic() {
	require := s.Require()
	app := s.app
	ctx := s.app.BaseApp.NewUncachedContext(false, cmtproto.Header{})
	k := app.OracleKeeper

	params := k.GetParams(ctx)
	params.BaseOwasmGas = 100000
	params.PerValidatorRequestGas = 50000
	err := k.SetParams(ctx, params)
	require.NoError(err)
	m := types.NewMsgRequestData(
		1,
		basicCalldata,
		2,
		1,
		basicClientID,
		bandtesting.Coins100000000uband,
		bandtesting.TestDefaultPrepareGas,
		bandtesting.TestDefaultExecuteGas,
		bandtesting.Alice.Address,
	)
	ctx = ctx.WithGasMeter(storetypes.NewGasMeter(90000))
	require.PanicsWithValue(
		storetypes.ErrorOutOfGas{Descriptor: "PER_VALIDATOR_REQUEST_FEE"},
		func() { _, _ = k.PrepareRequest(ctx, m, bandtesting.FeePayer.Address, nil) },
	)
	m = types.NewMsgRequestData(
		1,
		basicCalldata,
		1,
		1,
		basicClientID,
		bandtesting.Coins100000000uband,
		bandtesting.TestDefaultPrepareGas,
		bandtesting.TestDefaultExecuteGas,
		bandtesting.Alice.Address,
	)
	ctx = ctx.WithGasMeter(storetypes.NewGasMeter(1000000))
	id, err := k.PrepareRequest(ctx, m, bandtesting.FeePayer.Address, nil)
	require.Equal(types.RequestID(1), id)
	require.NoError(err)
}

func (s *AppTestSuite) TestPrepareRequestEmptyCalldata() {
	require := s.Require()
	app := s.app
	ctx := s.app.BaseApp.NewUncachedContext(false, cmtproto.Header{})
	k := app.OracleKeeper
	// Send nil while oracle script expects calldata
	m := types.NewMsgRequestData(
		4,
		nil,
		1,
		1,
		basicClientID,
		bandtesting.Coins100000000uband,
		bandtesting.TestDefaultPrepareGas,
		bandtesting.TestDefaultExecuteGas,
		bandtesting.Alice.Address,
	)
	_, err := k.PrepareRequest(ctx, m, bandtesting.FeePayer.Address, nil)
	require.EqualError(err, "runtime error while executing the Wasm script: bad wasm execution")
}

func (s *AppTestSuite) TestPrepareRequestBadWasmExecutionFail() {
	require := s.Require()
	app := s.app
	ctx := s.app.BaseApp.NewUncachedContext(false, cmtproto.Header{})
	k := app.OracleKeeper

	m := types.NewMsgRequestData(
		2,
		basicCalldata,
		1,
		1,
		basicClientID,
		bandtesting.Coins100000000uband,
		bandtesting.TestDefaultPrepareGas,
		bandtesting.TestDefaultExecuteGas,
		bandtesting.Alice.Address,
	)
	_, err := k.PrepareRequest(ctx, m, bandtesting.FeePayer.Address, nil)
	require.EqualError(err, "OEI action to invoke is not available: bad wasm execution")
}

func (s *AppTestSuite) TestPrepareRequestWithEmptyRawRequest() {
	require := s.Require()
	app := s.app
	ctx := s.app.BaseApp.NewUncachedContext(false, cmtproto.Header{})
	k := app.OracleKeeper

	m := types.NewMsgRequestData(
		3,
		basicCalldata,
		1,
		1,
		basicClientID,
		bandtesting.Coins100000000uband,
		bandtesting.TestDefaultPrepareGas,
		bandtesting.TestDefaultExecuteGas,
		bandtesting.Alice.Address,
	)
	_, err := k.PrepareRequest(ctx, m, bandtesting.FeePayer.Address, nil)
	require.EqualError(err, "empty raw requests")
}

func (s *AppTestSuite) TestPrepareRequestUnknownDataSource() {
	require := s.Require()
	app := s.app
	ctx := s.app.BaseApp.NewUncachedContext(false, cmtproto.Header{})
	k := app.OracleKeeper

	m := types.NewMsgRequestData(4, obi.MustEncode(testdata.Wasm4Input{
		IDs:      []int64{1, 2, 99},
		Calldata: "test",
	}), 1, 1, basicClientID, bandtesting.Coins100000000uband, bandtesting.TestDefaultPrepareGas, bandtesting.TestDefaultExecuteGas, bandtesting.Alice.Address)
	_, err := k.PrepareRequest(ctx, m, bandtesting.FeePayer.Address, nil)
	require.EqualError(err, "id: 99: data source not found")
}

func (s *AppTestSuite) TestPrepareRequestInvalidDataSourceCount() {
	require := s.Require()
	app := s.app
	ctx := s.app.BaseApp.NewUncachedContext(false, cmtproto.Header{})
	k := app.OracleKeeper

	params := k.GetParams(ctx)
	params.MaxRawRequestCount = 3
	err := k.SetParams(ctx, params)
	require.NoError(err)
	m := types.NewMsgRequestData(4, obi.MustEncode(testdata.Wasm4Input{
		IDs:      []int64{1, 2, 3, 4},
		Calldata: "test",
	}), 1, 1, basicClientID, bandtesting.Coins100000000uband, bandtesting.TestDefaultPrepareGas, bandtesting.TestDefaultExecuteGas, bandtesting.Alice.Address)
	_, err = k.PrepareRequest(ctx, m, bandtesting.FeePayer.Address, nil)
	require.ErrorIs(err, types.ErrBadWasmExecution)
	m = types.NewMsgRequestData(4, obi.MustEncode(testdata.Wasm4Input{
		IDs:      []int64{1, 2, 3},
		Calldata: "test",
	}), 1, 1, basicClientID, bandtesting.Coins100000000uband, bandtesting.TestDefaultPrepareGas, bandtesting.TestDefaultExecuteGas, bandtesting.Alice.Address)
	id, err := k.PrepareRequest(ctx, m, bandtesting.FeePayer.Address, nil)
	require.Equal(types.RequestID(1), id)
	require.NoError(err)
}

func (s *AppTestSuite) TestPrepareRequestTooMuchWasmGas() {
	require := s.Require()
	app := s.app
	ctx := s.app.BaseApp.NewUncachedContext(false, cmtproto.Header{})
	k := app.OracleKeeper

	m := types.NewMsgRequestData(
		5,
		basicCalldata,
		1,
		1,
		basicClientID,
		bandtesting.Coins100000000uband,
		bandtesting.TestDefaultPrepareGas,
		bandtesting.TestDefaultExecuteGas,
		bandtesting.Alice.Address,
	)
	id, err := k.PrepareRequest(ctx, m, bandtesting.FeePayer.Address, nil)
	require.Equal(types.RequestID(1), id)
	require.NoError(err)
	m = types.NewMsgRequestData(
		6,
		basicCalldata,
		1,
		1,
		basicClientID,
		bandtesting.Coins100000000uband,
		bandtesting.TestDefaultPrepareGas,
		bandtesting.TestDefaultExecuteGas,
		bandtesting.Alice.Address,
	)
	_, err = k.PrepareRequest(ctx, m, bandtesting.FeePayer.Address, nil)
	require.EqualError(err, "out-of-gas while executing the wasm script: bad wasm execution")
}

func (s *AppTestSuite) TestPrepareRequestTooLargeCalldata() {
	require := s.Require()
	app := s.app
	ctx := s.app.BaseApp.NewUncachedContext(false, cmtproto.Header{})
	k := app.OracleKeeper

	m := types.NewMsgRequestData(
		7,
		basicCalldata,
		1,
		1,
		basicClientID,
		bandtesting.Coins100000000uband,
		bandtesting.TestDefaultPrepareGas,
		bandtesting.TestDefaultExecuteGas,
		bandtesting.Alice.Address,
	)
	id, err := k.PrepareRequest(ctx, m, bandtesting.FeePayer.Address, nil)
	require.Equal(types.RequestID(1), id)
	require.NoError(err)
	m = types.NewMsgRequestData(
		8,
		basicCalldata,
		1,
		1,
		basicClientID,
		bandtesting.Coins100000000uband,
		bandtesting.TestDefaultPrepareGas,
		bandtesting.TestDefaultExecuteGas,
		bandtesting.Alice.Address,
	)
	_, err = k.PrepareRequest(ctx, m, bandtesting.FeePayer.Address, nil)
	require.EqualError(err, "span to write is too small: bad wasm execution")
}

func (s *AppTestSuite) TestResolveRequestOutOfGas() {
	require := s.Require()
	app := s.app
	ctx := s.app.BaseApp.NewUncachedContext(false, cmtproto.Header{})
	k := app.OracleKeeper

	ctx = ctx.WithBlockTime(bandtesting.ParseTime(1581589890))
	k.SetRequest(ctx, 42, types.NewRequest(
		// 1st Wasm - return "test"
		1,
		basicCalldata,
		[]sdk.ValAddress{bandtesting.Validators[0].ValAddress, bandtesting.Validators[1].ValAddress},
		1,
		42,
		bandtesting.ParseTime(1581589790),
		basicClientID,
		[]types.RawRequest{
			types.NewRawRequest(1, 1, []byte("test")),
		},
		nil,
		0,
	))
	k.SetReport(ctx, 42, types.NewReport(
		bandtesting.Validators[0].ValAddress, true, []types.RawReport{
			types.NewRawReport(1, 0, []byte("test")),
		},
	))
	k.ResolveRequest(ctx, 42)
	result := types.NewResult(
		basicClientID, 1, basicCalldata, 2, 1,
		42, 1, bandtesting.ParseTime(1581589790).Unix(),
		bandtesting.ParseTime(1581589890).Unix(), types.RESOLVE_STATUS_FAILURE, nil,
	)
	require.Equal(result, k.MustGetResult(ctx, 42))
}

func (s *AppTestSuite) TestResolveReadNilExternalData() {
	require := s.Require()
	app := s.app
	ctx := s.app.BaseApp.NewUncachedContext(false, cmtproto.Header{})
	k := app.OracleKeeper

	ctx = ctx.WithBlockTime(bandtesting.ParseTime(1581589890))
	k.SetRequest(ctx, 42, types.NewRequest(
		// 4th Wasm. Append all reports from all validators.
		4, obi.MustEncode(testdata.Wasm4Input{
			IDs:      []int64{1, 2},
			Calldata: string(basicCalldata),
		}), []sdk.ValAddress{bandtesting.Validators[0].ValAddress, bandtesting.Validators[1].ValAddress}, 1,
		42, bandtesting.ParseTime(1581589790), basicClientID, []types.RawRequest{
			types.NewRawRequest(0, 1, basicCalldata),
			types.NewRawRequest(1, 2, basicCalldata),
		}, nil, bandtesting.TestDefaultExecuteGas,
	))
	k.SetReport(ctx, 42, types.NewReport(
		bandtesting.Validators[0].ValAddress, true, []types.RawReport{
			types.NewRawReport(0, 0, nil),
			types.NewRawReport(1, 0, []byte("testd2v1")),
		},
	))
	k.SetReport(ctx, 42, types.NewReport(
		bandtesting.Validators[1].ValAddress, true, []types.RawReport{
			types.NewRawReport(0, 0, []byte("testd1v2")),
			types.NewRawReport(1, 0, nil),
		},
	))
	k.ResolveRequest(ctx, 42)
	result := types.NewResult(
		basicClientID, 4, obi.MustEncode(testdata.Wasm4Input{
			IDs:      []int64{1, 2},
			Calldata: string(basicCalldata),
		}), 2, 1,
		42, 2, bandtesting.ParseTime(1581589790).Unix(),
		bandtesting.ParseTime(1581589890).Unix(), types.RESOLVE_STATUS_SUCCESS,
		obi.MustEncode(testdata.Wasm4Output{Ret: "testd1v2testd2v1"}),
	)
	require.Equal(result, k.MustGetResult(ctx, 42))
	require.Equal(sdk.Events{sdk.NewEvent(
		types.EventTypeResolve,
		sdk.NewAttribute(types.AttributeKeyID, "42"),
		sdk.NewAttribute(types.AttributeKeyResolveStatus, "1"),
		sdk.NewAttribute(types.AttributeKeyResult, "0000001074657374643176327465737464327631"),
		sdk.NewAttribute(types.AttributeKeyGasUsed, "31168050000"),
	)}, ctx.EventManager().Events())
}

func (s *AppTestSuite) TestResolveRequestNoReturnData() {
	require := s.Require()
	app := s.app
	ctx := s.app.BaseApp.NewUncachedContext(false, cmtproto.Header{})
	k := app.OracleKeeper

	ctx = ctx.WithBlockTime(bandtesting.ParseTime(1581589890))
	k.SetRequest(ctx, 42, types.NewRequest(
		// 3rd Wasm - do nothing
		3,
		basicCalldata,
		[]sdk.ValAddress{bandtesting.Validators[0].ValAddress, bandtesting.Validators[1].ValAddress},
		1,
		42,
		bandtesting.ParseTime(1581589790),
		basicClientID,
		[]types.RawRequest{
			types.NewRawRequest(1, 1, []byte("test")),
		},
		nil,
		1,
	))
	k.SetReport(ctx, 42, types.NewReport(
		bandtesting.Validators[0].ValAddress, true, []types.RawReport{
			types.NewRawReport(1, 0, []byte("test")),
		},
	))
	k.ResolveRequest(ctx, 42)
	result := types.NewResult(
		basicClientID, 3, basicCalldata, 2, 1, 42, 1, bandtesting.ParseTime(1581589790).Unix(),
		bandtesting.ParseTime(1581589890).Unix(), types.RESOLVE_STATUS_FAILURE, nil,
	)
	require.Equal(result, k.MustGetResult(ctx, 42))
	require.Equal(sdk.Events{sdk.NewEvent(
		types.EventTypeResolve,
		sdk.NewAttribute(types.AttributeKeyID, "42"),
		sdk.NewAttribute(types.AttributeKeyResolveStatus, "2"),
		sdk.NewAttribute(types.AttributeKeyReason, "no return data"),
	)}, ctx.EventManager().Events())
}

func (s *AppTestSuite) TestResolveRequestWasmFailure() {
	require := s.Require()
	app := s.app
	ctx := s.app.BaseApp.NewUncachedContext(false, cmtproto.Header{})
	k := app.OracleKeeper

	ctx = ctx.WithBlockTime(bandtesting.ParseTime(1581589890))
	k.SetRequest(ctx, 42, types.NewRequest(
		// 6th Wasm - out-of-gas
		6,
		basicCalldata,
		[]sdk.ValAddress{bandtesting.Validators[0].ValAddress, bandtesting.Validators[1].ValAddress},
		1,
		42,
		bandtesting.ParseTime(1581589790),
		basicClientID,
		[]types.RawRequest{
			types.NewRawRequest(1, 1, []byte("test")),
		},
		nil,
		0,
	))
	k.SetReport(ctx, 42, types.NewReport(
		bandtesting.Validators[0].ValAddress, true, []types.RawReport{
			types.NewRawReport(1, 0, []byte("test")),
		},
	))
	k.ResolveRequest(ctx, 42)
	result := types.NewResult(
		basicClientID, 6, basicCalldata, 2, 1, 42, 1, bandtesting.ParseTime(1581589790).Unix(),
		bandtesting.ParseTime(1581589890).Unix(), types.RESOLVE_STATUS_FAILURE, nil,
	)
	require.Equal(result, k.MustGetResult(ctx, 42))
	require.Equal(sdk.Events{sdk.NewEvent(
		types.EventTypeResolve,
		sdk.NewAttribute(types.AttributeKeyID, "42"),
		sdk.NewAttribute(types.AttributeKeyResolveStatus, "2"),
		sdk.NewAttribute(types.AttributeKeyReason, "out-of-gas while executing the wasm script"),
	)}, ctx.EventManager().Events())
}

func (s *AppTestSuite) TestResolveRequestCallReturnDataSeveralTimes() {
	require := s.Require()
	app := s.app
	ctx := s.app.BaseApp.NewUncachedContext(false, cmtproto.Header{})
	k := app.OracleKeeper

	ctx = ctx.WithBlockTime(bandtesting.ParseTime(1581589890))
	k.SetRequest(ctx, 42, types.NewRequest(
		// 9th Wasm - set return data several times
		9,
		basicCalldata,
		[]sdk.ValAddress{bandtesting.Validators[0].ValAddress, bandtesting.Validators[1].ValAddress},
		1,
		42,
		bandtesting.ParseTime(1581589790),
		basicClientID,
		[]types.RawRequest{
			types.NewRawRequest(1, 1, []byte("test")),
		},
		nil,
		bandtesting.TestDefaultExecuteGas,
	))
	k.ResolveRequest(ctx, 42)

	result := types.NewResult(
		basicClientID, 9, basicCalldata, 2, 1, 42, 0, bandtesting.ParseTime(1581589790).Unix(),
		bandtesting.ParseTime(1581589890).Unix(), types.RESOLVE_STATUS_FAILURE, nil,
	)
	require.Equal(result, k.MustGetResult(ctx, 42))

	require.Equal(sdk.Events{sdk.NewEvent(
		types.EventTypeResolve,
		sdk.NewAttribute(types.AttributeKeyID, "42"),
		sdk.NewAttribute(types.AttributeKeyResolveStatus, "2"),
		sdk.NewAttribute(types.AttributeKeyReason, "set return data is called more than once"),
	)}, ctx.EventManager().Events())
}
