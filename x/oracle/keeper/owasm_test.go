package oraclekeeper_test

import (
	"encoding/hex"
	minttypes "github.com/GeoDB-Limited/odin-core/x/mint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"
	"testing"
	"time"

	"github.com/GeoDB-Limited/odin-core/pkg/obi"
	"github.com/GeoDB-Limited/odin-core/x/common/testapp"
	oraclekeeper "github.com/GeoDB-Limited/odin-core/x/oracle/keeper"
	oracletypes "github.com/GeoDB-Limited/odin-core/x/oracle/types"
)

func TestGetRandomValidatorsSuccessActivateAll(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true)
	// Getting 3 validators using ROLLING_SEED_1_WITH_LONG_ENOUGH_ENTROPY
	k.SetRollingSeed(ctx, []byte("ROLLING_SEED_1_WITH_LONG_ENOUGH_ENTROPY"))
	vals, err := k.GetRandomValidators(ctx, 3, 1)
	require.NoError(t, err)
	require.Equal(t, []sdk.ValAddress{testapp.Validators[2].ValAddress, testapp.Validators[0].ValAddress, testapp.Validators[1].ValAddress}, vals)
	// Getting 3 validators using ROLLING_SEED_A
	k.SetRollingSeed(ctx, []byte("ROLLING_SEED_A_WITH_LONG_ENOUGH_ENTROPY"))
	vals, err = k.GetRandomValidators(ctx, 3, 1)
	require.NoError(t, err)
	require.Equal(t, []sdk.ValAddress{testapp.Validators[0].ValAddress, testapp.Validators[2].ValAddress, testapp.Validators[1].ValAddress}, vals)
	// Getting 3 validators using ROLLING_SEED_1_WITH_LONG_ENOUGH_ENTROPY again should return the same result as the first one.
	k.SetRollingSeed(ctx, []byte("ROLLING_SEED_1_WITH_LONG_ENOUGH_ENTROPY"))
	vals, err = k.GetRandomValidators(ctx, 3, 1)
	require.NoError(t, err)
	require.Equal(t, []sdk.ValAddress{testapp.Validators[2].ValAddress, testapp.Validators[0].ValAddress, testapp.Validators[1].ValAddress}, vals)
	// Getting 3 validators using ROLLING_SEED_1_WITH_LONG_ENOUGH_ENTROPY but for a different request ID.
	k.SetRollingSeed(ctx, []byte("ROLLING_SEED_1_WITH_LONG_ENOUGH_ENTROPY"))
	vals, err = k.GetRandomValidators(ctx, 3, 42)
	require.NoError(t, err)
	require.Equal(t, []sdk.ValAddress{testapp.Validators[0].ValAddress, testapp.Validators[2].ValAddress, testapp.Validators[1].ValAddress}, vals)
}

func TestGetRandomValidatorsTooBigSize(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true)
	_, err := k.GetRandomValidators(ctx, 1, 1)
	require.NoError(t, err)
	_, err = k.GetRandomValidators(ctx, 2, 1)
	require.NoError(t, err)
	_, err = k.GetRandomValidators(ctx, 3, 1)
	require.NoError(t, err)
	_, err = k.GetRandomValidators(ctx, 4, 1)
	require.Error(t, err)
	_, err = k.GetRandomValidators(ctx, 9999, 1)
	require.Error(t, err)
}

func TestGetRandomValidatorsWithActivate(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(false)
	k.SetRollingSeed(ctx, []byte("ROLLING_SEED_WITH_LONG_ENOUGH_ENTROPY"))
	// If no validators are active, you must not be able to get random validators
	_, err := k.GetRandomValidators(ctx, 1, 1)
	require.Error(t, err)
	// If we activate 2 validators, we should be able to get at most 2 from the function.
	k.Activate(ctx, testapp.Validators[0].ValAddress)
	k.Activate(ctx, testapp.Validators[1].ValAddress)
	vals, err := k.GetRandomValidators(ctx, 1, 1)
	require.NoError(t, err)
	require.Equal(t, []sdk.ValAddress{testapp.Validators[0].ValAddress}, vals)
	vals, err = k.GetRandomValidators(ctx, 2, 1)
	require.NoError(t, err)
	require.Equal(t, []sdk.ValAddress{testapp.Validators[0].ValAddress, testapp.Validators[1].ValAddress}, vals)
	_, err = k.GetRandomValidators(ctx, 3, 1)
	require.Error(t, err)
	// After we deactivate 1 validator due to missing a report, we can only get at most 1 validator.
	k.MissReport(ctx, testapp.Validators[0].ValAddress, time.Now())
	vals, err = k.GetRandomValidators(ctx, 1, 1)
	require.NoError(t, err)
	require.Equal(t, []sdk.ValAddress{testapp.Validators[1].ValAddress}, vals)
	_, err = k.GetRandomValidators(ctx, 2, 1)
	require.Error(t, err)
}

func TestPrepareRequestSuccessBasic(t *testing.T) {
	app, ctx, k := testapp.CreateTestInput(true)
	ctx = ctx.WithBlockTime(testapp.ParseTime(1581589790)).WithBlockHeight(42)

	wrappedGasMeter := testapp.NewGasMeterWrapper(ctx.GasMeter())
	ctx = ctx.WithGasMeter(wrappedGasMeter)

	// OracleScript#1: Prepare asks for DS#1,2,3 with ExtID#1,2,3 and calldata "beeb"
	m := oracletypes.NewMsgRequestData(1, BasicCalldata, 1, 1, BasicClientID, testapp.Coins10000000000loki, oracletypes.DefaultPrepareGas, oracletypes.DefaultExecuteGas, testapp.FeePayer.Address)
	id, err := k.PrepareRequest(ctx, m, testapp.FeePayer.Address, nil)
	require.NoError(t, err)
	require.Equal(t, oracletypes.RequestID(1), id)

	require.Equal(t, sdk.Events{
		sdk.NewEvent(
			banktypes.EventTypeTransfer,
			sdk.NewAttribute(banktypes.AttributeKeyRecipient, app.AccountKeeper.GetModuleAddress(oracletypes.ModuleName).String()),
			sdk.NewAttribute(banktypes.AttributeKeySender, testapp.FeePayer.Address.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, testapp.Coins1000000loki[0].String())),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(banktypes.AttributeKeySender, testapp.FeePayer.Address.String()),
		),
		sdk.NewEvent(
			banktypes.EventTypeTransfer,
			sdk.NewAttribute(banktypes.AttributeKeyRecipient, app.AccountKeeper.GetModuleAddress(oracletypes.ModuleName).String()),
			sdk.NewAttribute(banktypes.AttributeKeySender, testapp.FeePayer.Address.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, testapp.Coins1000000loki[0].String())),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(banktypes.AttributeKeySender, testapp.FeePayer.Address.String()),
		),
		sdk.NewEvent(
			banktypes.EventTypeTransfer,
			sdk.NewAttribute(banktypes.AttributeKeyRecipient, app.AccountKeeper.GetModuleAddress(oracletypes.ModuleName).String()),
			sdk.NewAttribute(banktypes.AttributeKeySender, testapp.FeePayer.Address.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, testapp.Coins1000000loki[0].String())),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(banktypes.AttributeKeySender, testapp.FeePayer.Address.String()),
		),
		sdk.NewEvent(
			oracletypes.EventTypeRequest,
			sdk.NewAttribute(oracletypes.AttributeKeyID, "1"),
			sdk.NewAttribute(oracletypes.AttributeKeyClientID, BasicClientID),
			sdk.NewAttribute(oracletypes.AttributeKeyOracleScriptID, "1"),
			sdk.NewAttribute(oracletypes.AttributeKeyCalldata, hex.EncodeToString(BasicCalldata)),
			sdk.NewAttribute(oracletypes.AttributeKeyAskCount, "1"),
			sdk.NewAttribute(oracletypes.AttributeKeyMinCount, "1"),
			sdk.NewAttribute(oracletypes.AttributeKeyGasUsed, "3089"), // TODO: might change
			sdk.NewAttribute(oracletypes.AttributeKeyValidator, testapp.Validators[0].ValAddress.String()),
		), sdk.NewEvent(
			oracletypes.EventTypeRawRequest,
			sdk.NewAttribute(oracletypes.AttributeKeyDataSourceID, "1"),
			sdk.NewAttribute(oracletypes.AttributeKeyDataSourceHash, testapp.DataSources[1].Filename),
			sdk.NewAttribute(oracletypes.AttributeKeyExternalID, "1"),
			sdk.NewAttribute(oracletypes.AttributeKeyCalldata, "beeb"),
		), sdk.NewEvent(
			oracletypes.EventTypeRawRequest,
			sdk.NewAttribute(oracletypes.AttributeKeyDataSourceID, "2"),
			sdk.NewAttribute(oracletypes.AttributeKeyDataSourceHash, testapp.DataSources[2].Filename),
			sdk.NewAttribute(oracletypes.AttributeKeyExternalID, "2"),
			sdk.NewAttribute(oracletypes.AttributeKeyCalldata, "beeb"),
		), sdk.NewEvent(
			oracletypes.EventTypeRawRequest,
			sdk.NewAttribute(oracletypes.AttributeKeyDataSourceID, "3"),
			sdk.NewAttribute(oracletypes.AttributeKeyDataSourceHash, testapp.DataSources[3].Filename),
			sdk.NewAttribute(oracletypes.AttributeKeyExternalID, "3"),
			sdk.NewAttribute(oracletypes.AttributeKeyCalldata, "beeb"),
		)}, ctx.EventManager().Events())

	// assert gas consumation
	params := k.GetParams(ctx)
	require.Equal(t, 2, wrappedGasMeter.CountRecord(params.BaseOwasmGas, "BASE_OWASM_FEE"))
	require.Equal(t, 1, wrappedGasMeter.CountRecord(oracletypes.DefaultPrepareGas, "OWASM_PREPARE_FEE"))
	require.Equal(t, 1, wrappedGasMeter.CountRecord(oracletypes.DefaultExecuteGas, "OWASM_EXECUTE_FEE"))
}

func TestPrepareRequestSuccessBasicNotEnoughMaxFee(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true)
	ctx = ctx.WithBlockTime(testapp.ParseTime(1581589790)).WithBlockHeight(42)
	// OracleScript#1: Prepare asks for DS#1,2,3 with ExtID#1,2,3 and calldata "beeb"
	m := oracletypes.NewMsgRequestData(1, BasicCalldata, 1, 1, BasicClientID, testapp.EmptyCoins, oracletypes.DefaultPrepareGas, oracletypes.DefaultExecuteGas, testapp.FeePayer.Address)
	_, err := k.PrepareRequest(ctx, m, testapp.FeePayer.Address, nil)
	require.EqualError(t, err, "require: 1000000loki, max: 0loki: not enough fee")
	m = oracletypes.NewMsgRequestData(1, BasicCalldata, 1, 1, BasicClientID, sdk.NewCoins(sdk.NewInt64Coin("loki", 1000000)), oracletypes.DefaultPrepareGas, oracletypes.DefaultExecuteGas, testapp.FeePayer.Address)
	_, err = k.PrepareRequest(ctx, m, testapp.FeePayer.Address, nil)
	require.EqualError(t, err, "require: 2000000loki, max: 1000000loki: not enough fee")
	m = oracletypes.NewMsgRequestData(1, BasicCalldata, 1, 1, BasicClientID, sdk.NewCoins(sdk.NewInt64Coin("loki", 2000000)), oracletypes.DefaultPrepareGas, oracletypes.DefaultExecuteGas, testapp.FeePayer.Address)
	_, err = k.PrepareRequest(ctx, m, testapp.FeePayer.Address, nil)
	require.EqualError(t, err, "require: 3000000loki, max: 2000000loki: not enough fee")
	m = oracletypes.NewMsgRequestData(1, BasicCalldata, 1, 1, BasicClientID, sdk.NewCoins(sdk.NewInt64Coin("loki", 2999999)), oracletypes.DefaultPrepareGas, oracletypes.DefaultExecuteGas, testapp.FeePayer.Address)
	_, err = k.PrepareRequest(ctx, m, testapp.FeePayer.Address, nil)
	require.EqualError(t, err, "require: 3000000loki, max: 2999999loki: not enough fee")
	m = oracletypes.NewMsgRequestData(1, BasicCalldata, 1, 1, BasicClientID, sdk.NewCoins(sdk.NewInt64Coin("loki", 3000000)), oracletypes.DefaultPrepareGas, oracletypes.DefaultExecuteGas, testapp.FeePayer.Address)
	id, err := k.PrepareRequest(ctx, m, testapp.FeePayer.Address, nil)
	require.NoError(t, err)
	require.Equal(t, oracletypes.RequestID(1), id)
}

func TestPrepareRequestSuccessBasicNotEnoughFund(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true)
	ctx = ctx.WithBlockTime(testapp.ParseTime(1581589790)).WithBlockHeight(42)
	// OracleScript#1: Prepare asks for DS#1,2,3 with ExtID#1,2,3 and calldata "beeb"
	m := oracletypes.NewMsgRequestData(1, BasicCalldata, 1, 1, BasicClientID, testapp.Coins100000000loki, oracletypes.DefaultPrepareGas, oracletypes.DefaultExecuteGas, testapp.Alice.Address)
	_, err := k.PrepareRequest(ctx, m, testapp.Alice.Address, nil)
	require.EqualError(t, err, "0loki is smaller than 1000000loki: insufficient funds")
}

func TestPrepareRequestNotEnoughPrepareGas(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true)
	ctx = ctx.WithBlockTime(testapp.ParseTime(1581589790)).WithBlockHeight(42)

	wrappedGasMeter := testapp.NewGasMeterWrapper(ctx.GasMeter())
	ctx = ctx.WithGasMeter(wrappedGasMeter)

	m := oracletypes.NewMsgRequestData(1, BasicCalldata, 1, 1, BasicClientID, testapp.EmptyCoins, 100, oracletypes.DefaultExecuteGas, testapp.Alice.Address)
	_, err := k.PrepareRequest(ctx, m, testapp.FeePayer.Address, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "out-of-gas")

	params := k.GetParams(ctx)
	require.Equal(t, 1, wrappedGasMeter.CountRecord(params.BaseOwasmGas, "BASE_OWASM_FEE"))
	require.Equal(t, 1, wrappedGasMeter.CountRecord(100, "OWASM_PREPARE_FEE"))
	require.Equal(t, 0, wrappedGasMeter.CountDescriptor("OWASM_EXECUTE_FEE"))
}

func TestPrepareRequestInvalidAskCountFail(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true)
	k.SetParamUint64(ctx, oracletypes.KeyMaxAskCount, 5)

	wrappedGasMeter := testapp.NewGasMeterWrapper(ctx.GasMeter())
	ctx = ctx.WithGasMeter(wrappedGasMeter)

	m := oracletypes.NewMsgRequestData(1, BasicCalldata, 10, 1, BasicClientID, testapp.Coins100000000loki, oracletypes.DefaultPrepareGas, oracletypes.DefaultExecuteGas, testapp.Alice.Address)
	_, err := k.PrepareRequest(ctx, m, testapp.FeePayer.Address, nil)
	// require.EqualError(t, err, "invalid ask count: got: 10, max: 5")

	require.Equal(t, 0, wrappedGasMeter.CountDescriptor("BASE_OWASM_FEE"))
	require.Equal(t, 0, wrappedGasMeter.CountDescriptor("OWASM_PREPARE_FEE"))
	require.Equal(t, 0, wrappedGasMeter.CountDescriptor("OWASM_EXECUTE_FEE"))

	m = oracletypes.NewMsgRequestData(1, BasicCalldata, 4, 1, BasicClientID, testapp.Coins100000000loki, oracletypes.DefaultPrepareGas, oracletypes.DefaultExecuteGas, testapp.Alice.Address)
	_, err = k.PrepareRequest(ctx, m, testapp.FeePayer.Address, nil)
	// require.EqualError(t, err, "insufficent available validators: 3 < 4")

	require.Equal(t, 0, wrappedGasMeter.CountDescriptor("BASE_OWASM_FEE"))
	require.Equal(t, 0, wrappedGasMeter.CountDescriptor("OWASM_PREPARE_FEE"))
	require.Equal(t, 0, wrappedGasMeter.CountDescriptor("OWASM_EXECUTE_FEE"))

	m = oracletypes.NewMsgRequestData(1, BasicCalldata, 1, 1, BasicClientID, testapp.Coins10000000000loki, oracletypes.DefaultPrepareGas, oracletypes.DefaultExecuteGas, testapp.Alice.Address)
	id, err := k.PrepareRequest(ctx, m, testapp.FeePayer.Address, nil)
	require.NoError(t, err)
	require.Equal(t, oracletypes.RequestID(1), id)
	require.Equal(t, 2, wrappedGasMeter.CountDescriptor("BASE_OWASM_FEE"))
	require.Equal(t, 1, wrappedGasMeter.CountDescriptor("OWASM_PREPARE_FEE"))
	require.Equal(t, 1, wrappedGasMeter.CountDescriptor("OWASM_EXECUTE_FEE"))
}

func TestPrepareRequestBaseOwasmFeePanic(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true)
	k.SetParamUint64(ctx, oracletypes.KeyBaseOwasmGas, 100000) // Set BaseRequestGas to 100000
	k.SetParamUint64(ctx, oracletypes.KeyPerValidatorRequestGas, 0)
	m := oracletypes.NewMsgRequestData(1, BasicCalldata, 1, 1, BasicClientID, testapp.Coins100000000loki, oracletypes.DefaultPrepareGas, oracletypes.DefaultExecuteGas, testapp.Alice.Address)
	ctx = ctx.WithGasMeter(sdk.NewGasMeter(90000))
	require.PanicsWithValue(t, sdk.ErrorOutOfGas{Descriptor: "BASE_OWASM_FEE"}, func() { k.PrepareRequest(ctx, m, testapp.FeePayer.Address, nil) })
	ctx = ctx.WithGasMeter(sdk.NewGasMeter(1000000))
	id, err := k.PrepareRequest(ctx, m, testapp.FeePayer.Address, nil)
	require.NoError(t, err)
	require.Equal(t, oracletypes.RequestID(1), id)
}

func TestPrepareRequestPerValidatorRequestFeePanic(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true)
	k.SetParamUint64(ctx, oracletypes.KeyBaseOwasmGas, 100000)
	k.SetParamUint64(ctx, oracletypes.KeyPerValidatorRequestGas, 50000) // Set perValidatorRequestGas to 50000
	m := oracletypes.NewMsgRequestData(1, BasicCalldata, 2, 1, BasicClientID, testapp.Coins100000000loki, oracletypes.DefaultPrepareGas, oracletypes.DefaultExecuteGas, testapp.Alice.Address)
	ctx = ctx.WithGasMeter(sdk.NewGasMeter(90000))
	require.PanicsWithValue(t, sdk.ErrorOutOfGas{Descriptor: "PER_VALIDATOR_REQUEST_FEE"}, func() { k.PrepareRequest(ctx, m, testapp.FeePayer.Address, nil) })
	m = oracletypes.NewMsgRequestData(1, BasicCalldata, 1, 1, BasicClientID, testapp.Coins100000000loki, oracletypes.DefaultPrepareGas, oracletypes.DefaultExecuteGas, testapp.Alice.Address)
	ctx = ctx.WithGasMeter(sdk.NewGasMeter(1000000))
	id, err := k.PrepareRequest(ctx, m, testapp.FeePayer.Address, nil)
	require.NoError(t, err)
	require.Equal(t, oracletypes.RequestID(1), id)
}

func TestPrepareRequestEmptyCalldata(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true) // Send nil while oracle script expects calldata
	m := oracletypes.NewMsgRequestData(4, nil, 1, 1, BasicClientID, testapp.Coins100000000loki, oracletypes.DefaultPrepareGas, oracletypes.DefaultExecuteGas, testapp.Alice.Address)
	_, err := k.PrepareRequest(ctx, m, testapp.FeePayer.Address, nil)
	require.EqualError(t, err, "runtime error while executing the Wasm script: bad wasm execution")
}

func TestPrepareRequestOracleScriptNotFound(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true)
	m := oracletypes.NewMsgRequestData(999, BasicCalldata, 1, 1, BasicClientID, testapp.Coins100000000loki, oracletypes.DefaultPrepareGas, oracletypes.DefaultExecuteGas, testapp.Alice.Address)
	_, err := k.PrepareRequest(ctx, m, testapp.FeePayer.Address, nil)
	require.EqualError(t, err, "id: 999: oracle script not found")
}

func TestPrepareRequestBadWasmExecutionFail(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true)
	m := oracletypes.NewMsgRequestData(2, BasicCalldata, 1, 1, BasicClientID, testapp.Coins100000000loki, oracletypes.DefaultPrepareGas, oracletypes.DefaultExecuteGas, testapp.Alice.Address)
	_, err := k.PrepareRequest(ctx, m, testapp.FeePayer.Address, nil)
	require.EqualError(t, err, "OEI action to invoke is not available: bad wasm execution")
}

func TestPrepareRequestWithEmptyRawRequest(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true)
	m := oracletypes.NewMsgRequestData(3, BasicCalldata, 1, 1, BasicClientID, testapp.Coins100000000loki, oracletypes.DefaultPrepareGas, oracletypes.DefaultExecuteGas, testapp.Alice.Address)
	_, err := k.PrepareRequest(ctx, m, testapp.FeePayer.Address, nil)
	require.EqualError(t, err, "empty raw requests")
}

func TestPrepareRequestUnknownDataSource(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true)
	m := oracletypes.NewMsgRequestData(4, obi.MustEncode(testapp.Wasm4Input{
		IDs:      []int64{1, 2, 99},
		Calldata: "beeb",
	}), 1, 1, BasicClientID, testapp.Coins100000000loki, oracletypes.DefaultPrepareGas, oracletypes.DefaultExecuteGas, testapp.Alice.Address)
	require.PanicsWithErrorf(
		t,
		"id: 99: data source not found",
		func() { k.PrepareRequest(ctx, m, testapp.FeePayer.Address, nil) },
		"Expected data source 99 not found",
	)
}

func TestPrepareRequestInvalidDataSourceCount(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true)
	k.SetParamUint64(ctx, oracletypes.KeyMaxRawRequestCount, 3)
	m := oracletypes.NewMsgRequestData(4, obi.MustEncode(testapp.Wasm4Input{
		IDs:      []int64{1, 2, 3, 4},
		Calldata: "beeb",
	}), 1, 1, BasicClientID, testapp.Coins100000000loki, oracletypes.DefaultPrepareGas, oracletypes.DefaultExecuteGas, testapp.Alice.Address)
	_, err := k.PrepareRequest(ctx, m, testapp.FeePayer.Address, nil)
	require.EqualError(t, err, "too many external data requests: bad wasm execution")
	m = oracletypes.NewMsgRequestData(4, obi.MustEncode(testapp.Wasm4Input{
		IDs:      []int64{1, 2, 3},
		Calldata: "beeb",
	}), 1, 1, BasicClientID, testapp.Coins100000000loki, oracletypes.DefaultPrepareGas, oracletypes.DefaultExecuteGas, testapp.Alice.Address)
	id, err := k.PrepareRequest(ctx, m, testapp.FeePayer.Address, nil)
	require.NoError(t, err)
	require.Equal(t, oracletypes.RequestID(1), id)
}

func TestPrepareRequestTooMuchWasmGas(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true)

	m := oracletypes.NewMsgRequestData(5, BasicCalldata, 1, 1, BasicClientID, testapp.Coins100000000loki, oracletypes.DefaultPrepareGas, oracletypes.DefaultExecuteGas, testapp.Alice.Address)
	id, err := k.PrepareRequest(ctx, m, testapp.FeePayer.Address, nil)
	require.Equal(t, oracletypes.RequestID(1), id)
	require.NoError(t, err)
	m = oracletypes.NewMsgRequestData(6, BasicCalldata, 1, 1, BasicClientID, testapp.Coins100000000loki, oracletypes.DefaultPrepareGas, oracletypes.DefaultExecuteGas, testapp.Alice.Address)
	_, err = k.PrepareRequest(ctx, m, testapp.FeePayer.Address, nil)
	require.EqualError(t, err, "out-of-gas while executing the wasm script: bad wasm execution")
}

func TestPrepareRequestTooLargeCalldata(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true)

	m := oracletypes.NewMsgRequestData(7, BasicCalldata, 1, 1, BasicClientID, testapp.Coins100000000loki, oracletypes.DefaultPrepareGas, oracletypes.DefaultExecuteGas, testapp.Alice.Address)
	id, err := k.PrepareRequest(ctx, m, testapp.FeePayer.Address, nil)
	require.Equal(t, oracletypes.RequestID(1), id)
	require.NoError(t, err)
	m = oracletypes.NewMsgRequestData(8, BasicCalldata, 1, 1, BasicClientID, testapp.Coins100000000loki, oracletypes.DefaultPrepareGas, oracletypes.DefaultExecuteGas, testapp.Alice.Address)
	_, err = k.PrepareRequest(ctx, m, testapp.FeePayer.Address, nil)
	require.EqualError(t, err, "span to write is too small: bad wasm execution")
}

func TestResolveRequestSuccess(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true)
	ctx = ctx.WithBlockTime(testapp.ParseTime(1581589890))
	k.SetRequest(ctx, 42, oracletypes.NewRequest(
		// 1st Wasm - return "beeb"
		1, BasicCalldata, []sdk.ValAddress{testapp.Validators[0].ValAddress, testapp.Validators[1].ValAddress}, 1,
		42, testapp.ParseTime(1581589790), BasicClientID, []oracletypes.RawRequest{
			oracletypes.NewRawRequest(1, 1, []byte("beeb")),
		}, nil, oracletypes.DefaultExecuteGas,
	))
	k.SetReport(ctx, 42, oracletypes.NewReport(
		testapp.Validators[0].ValAddress, true, []oracletypes.RawReport{
			oracletypes.NewRawReport(1, 0, []byte("beeb")),
		},
	))
	k.ResolveRequest(ctx, 42)
	expectResult := oracletypes.NewResult(
		BasicClientID, 1, BasicCalldata, 2, 1,
		42, 1, testapp.ParseTime(1581589790).Unix(),
		testapp.ParseTime(1581589890).Unix(), oracletypes.RESOLVE_STATUS_SUCCESS, []byte("beeb"),
	)
	require.Equal(t, expectResult, k.MustGetResult(ctx, 42))
	require.Equal(t, sdk.Events{sdk.NewEvent(
		oracletypes.EventTypeResolve,
		sdk.NewAttribute(oracletypes.AttributeKeyID, "42"),
		sdk.NewAttribute(oracletypes.AttributeKeyResolveStatus, "1"),
		sdk.NewAttribute(oracletypes.AttributeKeyResult, "62656562"), // hex of "beeb"
		sdk.NewAttribute(oracletypes.AttributeKeyGasUsed, "1028"),    // TODO might change
	)}, ctx.EventManager().Events())
}

func TestResolveRequestSuccessComplex(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true)
	ctx = ctx.WithBlockTime(testapp.ParseTime(1581589890))
	k.SetRequest(ctx, 42, oracletypes.NewRequest(
		// 4th Wasm. Append all reports from all validators.
		4, obi.MustEncode(testapp.Wasm4Input{
			IDs:      []int64{1, 2},
			Calldata: string(BasicCalldata),
		}), []sdk.ValAddress{testapp.Validators[0].ValAddress, testapp.Validators[1].ValAddress}, 1,
		42, testapp.ParseTime(1581589790), BasicClientID, []oracletypes.RawRequest{
			oracletypes.NewRawRequest(0, 1, BasicCalldata),
			oracletypes.NewRawRequest(1, 2, BasicCalldata),
		}, nil, oracletypes.DefaultExecuteGas,
	))
	k.SetReport(ctx, 42, oracletypes.NewReport(
		testapp.Validators[0].ValAddress, true, []oracletypes.RawReport{
			oracletypes.NewRawReport(0, 0, []byte("beebd1v1")),
			oracletypes.NewRawReport(1, 0, []byte("beebd2v1")),
		},
	))
	k.SetReport(ctx, 42, oracletypes.NewReport(
		testapp.Validators[1].ValAddress, true, []oracletypes.RawReport{
			oracletypes.NewRawReport(0, 0, []byte("beebd1v2")),
			oracletypes.NewRawReport(1, 0, []byte("beebd2v2")),
		},
	))
	k.ResolveRequest(ctx, 42)
	result := oracletypes.NewResult(
		BasicClientID, 4, obi.MustEncode(testapp.Wasm4Input{
			IDs:      []int64{1, 2},
			Calldata: string(BasicCalldata),
		}), 2, 1,
		42, 2, testapp.ParseTime(1581589790).Unix(),
		testapp.ParseTime(1581589890).Unix(), oracletypes.RESOLVE_STATUS_SUCCESS,
		obi.MustEncode(testapp.Wasm4Output{Ret: "beebd1v1beebd1v2beebd2v1beebd2v2"}),
	)
	require.Equal(t, result, k.MustGetResult(ctx, 42))
	require.Equal(t, sdk.Events{sdk.NewEvent(
		oracletypes.EventTypeResolve,
		sdk.NewAttribute(oracletypes.AttributeKeyID, "42"),
		sdk.NewAttribute(oracletypes.AttributeKeyResolveStatus, "1"),
		sdk.NewAttribute(oracletypes.AttributeKeyResult, "000000206265656264317631626565626431763262656562643276316265656264327632"),
		sdk.NewAttribute(oracletypes.AttributeKeyGasUsed, "13634"), // todo might change
	)}, ctx.EventManager().Events())
}

func TestResolveRequestOutOfGas(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true)
	ctx = ctx.WithBlockTime(testapp.ParseTime(1581589890))
	k.SetRequest(ctx, 42, oracletypes.NewRequest(
		// 1st Wasm - return "beeb"
		1, BasicCalldata, []sdk.ValAddress{testapp.Validators[0].ValAddress, testapp.Validators[1].ValAddress}, 1,
		42, testapp.ParseTime(1581589790), BasicClientID, []oracletypes.RawRequest{
			oracletypes.NewRawRequest(1, 1, []byte("beeb")),
		}, nil, 0,
	))
	k.SetReport(ctx, 42, oracletypes.NewReport(
		testapp.Validators[0].ValAddress, true, []oracletypes.RawReport{
			oracletypes.NewRawReport(1, 0, []byte("beeb")),
		},
	))
	k.ResolveRequest(ctx, 42)
	result := oracletypes.NewResult(
		BasicClientID, 1, BasicCalldata, 2, 1,
		42, 1, testapp.ParseTime(1581589790).Unix(),
		testapp.ParseTime(1581589890).Unix(), oracletypes.RESOLVE_STATUS_FAILURE, []byte{},
	)
	require.Equal(t, result, k.MustGetResult(ctx, 42))
}

func TestResolveReadNilExternalData(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true)
	ctx = ctx.WithBlockTime(testapp.ParseTime(1581589890))
	k.SetRequest(ctx, 42, oracletypes.NewRequest(
		// 4th Wasm. Append all reports from all validators.
		4, obi.MustEncode(testapp.Wasm4Input{
			IDs:      []int64{1, 2},
			Calldata: string(BasicCalldata),
		}), []sdk.ValAddress{testapp.Validators[0].ValAddress, testapp.Validators[1].ValAddress}, 1,
		42, testapp.ParseTime(1581589790), BasicClientID, []oracletypes.RawRequest{
			oracletypes.NewRawRequest(0, 1, BasicCalldata),
			oracletypes.NewRawRequest(1, 2, BasicCalldata),
		}, nil, oracletypes.DefaultExecuteGas,
	))
	k.SetReport(ctx, 42, oracletypes.NewReport(
		testapp.Validators[0].ValAddress, true, []oracletypes.RawReport{
			oracletypes.NewRawReport(0, 0, nil),
			oracletypes.NewRawReport(1, 0, []byte("beebd2v1")),
		},
	))
	k.SetReport(ctx, 42, oracletypes.NewReport(
		testapp.Validators[1].ValAddress, true, []oracletypes.RawReport{
			oracletypes.NewRawReport(0, 0, []byte("beebd1v2")),
			oracletypes.NewRawReport(1, 0, nil),
		},
	))
	k.ResolveRequest(ctx, 42)
	result := oracletypes.NewResult(
		BasicClientID, 4, obi.MustEncode(testapp.Wasm4Input{
			IDs:      []int64{1, 2},
			Calldata: string(BasicCalldata),
		}), 2, 1,
		42, 2, testapp.ParseTime(1581589790).Unix(),
		testapp.ParseTime(1581589890).Unix(), oracletypes.RESOLVE_STATUS_SUCCESS,
		obi.MustEncode(testapp.Wasm4Output{Ret: "beebd1v2beebd2v1"}),
	)
	require.Equal(t, result, k.MustGetResult(ctx, 42))
	require.Equal(t, sdk.Events{sdk.NewEvent(
		oracletypes.EventTypeResolve,
		sdk.NewAttribute(oracletypes.AttributeKeyID, "42"),
		sdk.NewAttribute(oracletypes.AttributeKeyResolveStatus, "1"),
		sdk.NewAttribute(oracletypes.AttributeKeyResult, "0000001062656562643176326265656264327631"),
		sdk.NewAttribute(oracletypes.AttributeKeyGasUsed, "12653"),
	)}, ctx.EventManager().Events())
}

func TestResolveRequestNoReturnData(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true)
	ctx = ctx.WithBlockTime(testapp.ParseTime(1581589890))
	k.SetRequest(ctx, 42, oracletypes.NewRequest(
		// 3rd Wasm - do nothing
		3, BasicCalldata, []sdk.ValAddress{testapp.Validators[0].ValAddress, testapp.Validators[1].ValAddress}, 1,
		42, testapp.ParseTime(1581589790), BasicClientID, []oracletypes.RawRequest{
			oracletypes.NewRawRequest(1, 1, []byte("beeb")),
		}, nil, 0,
	))
	k.SetReport(ctx, 42, oracletypes.NewReport(
		testapp.Validators[0].ValAddress, true, []oracletypes.RawReport{
			oracletypes.NewRawReport(1, 0, []byte("beeb")),
		},
	))
	k.ResolveRequest(ctx, 42)
	result := oracletypes.NewResult(
		BasicClientID, 3, BasicCalldata, 2, 1, 42, 1, testapp.ParseTime(1581589790).Unix(),
		testapp.ParseTime(1581589890).Unix(), oracletypes.RESOLVE_STATUS_FAILURE, []byte{},
	)
	require.Equal(t, result, k.MustGetResult(ctx, 42))
	require.Equal(t, sdk.Events{sdk.NewEvent(
		oracletypes.EventTypeResolve,
		sdk.NewAttribute(oracletypes.AttributeKeyID, "42"),
		sdk.NewAttribute(oracletypes.AttributeKeyResolveStatus, "2"),
		sdk.NewAttribute(oracletypes.AttributeKeyReason, "no return data"),
	)}, ctx.EventManager().Events())
}

func TestResolveRequestWasmFailure(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true)
	ctx = ctx.WithBlockTime(testapp.ParseTime(1581589890))
	k.SetRequest(ctx, 42, oracletypes.NewRequest(
		// 6th Wasm - out-of-gas
		6, BasicCalldata, []sdk.ValAddress{testapp.Validators[0].ValAddress, testapp.Validators[1].ValAddress}, 1,
		42, testapp.ParseTime(1581589790), BasicClientID, []oracletypes.RawRequest{
			oracletypes.NewRawRequest(1, 1, []byte("beeb")),
		}, nil, 0,
	))
	k.SetReport(ctx, 42, oracletypes.NewReport(
		testapp.Validators[0].ValAddress, true, []oracletypes.RawReport{
			oracletypes.NewRawReport(1, 0, []byte("beeb")),
		},
	))
	k.ResolveRequest(ctx, 42)
	result := oracletypes.NewResult(
		BasicClientID, 6, BasicCalldata, 2, 1, 42, 1, testapp.ParseTime(1581589790).Unix(),
		testapp.ParseTime(1581589890).Unix(), oracletypes.RESOLVE_STATUS_FAILURE, []byte{},
	)
	require.Equal(t, result, k.MustGetResult(ctx, 42))
	require.Equal(t, sdk.Events{sdk.NewEvent(
		oracletypes.EventTypeResolve,
		sdk.NewAttribute(oracletypes.AttributeKeyID, "42"),
		sdk.NewAttribute(oracletypes.AttributeKeyResolveStatus, "2"),
		sdk.NewAttribute(oracletypes.AttributeKeyReason, "out-of-gas while executing the wasm script"),
	)}, ctx.EventManager().Events())
}

func TestResolveRequestCallReturnDataSeveralTimes(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true)
	ctx = ctx.WithBlockTime(testapp.ParseTime(1581589890))
	k.SetRequest(ctx, 42, oracletypes.NewRequest(
		// 9th Wasm - set return data several times
		9, BasicCalldata, []sdk.ValAddress{testapp.Validators[0].ValAddress, testapp.Validators[1].ValAddress}, 1,
		42, testapp.ParseTime(1581589790), BasicClientID, []oracletypes.RawRequest{
			oracletypes.NewRawRequest(1, 1, []byte("beeb")),
		}, nil, oracletypes.DefaultExecuteGas,
	))
	k.ResolveRequest(ctx, 42)

	result := oracletypes.NewResult(
		BasicClientID, 9, BasicCalldata, 2, 1, 42, 0, testapp.ParseTime(1581589790).Unix(),
		testapp.ParseTime(1581589890).Unix(), oracletypes.RESOLVE_STATUS_FAILURE, []byte{},
	)
	require.Equal(t, result, k.MustGetResult(ctx, 42))

	require.Equal(t, sdk.Events{sdk.NewEvent(
		oracletypes.EventTypeResolve,
		sdk.NewAttribute(oracletypes.AttributeKeyID, "42"),
		sdk.NewAttribute(oracletypes.AttributeKeyResolveStatus, "2"),
		sdk.NewAttribute(oracletypes.AttributeKeyReason, "set return data is called more than once"),
	)}, ctx.EventManager().Events())
}

func rawRequestsFromFees(ctx sdk.Context, k oraclekeeper.Keeper, fees []sdk.Coins) []oracletypes.RawRequest {
	var rawRequests []oracletypes.RawRequest
	for _, f := range fees {
		id := k.AddDataSource(ctx, oracletypes.NewDataSource(
			testapp.Owner.Address,
			"mock ds",
			"there is no real code",
			"no file",
			f,
		))

		rawRequests = append(rawRequests, oracletypes.NewRawRequest(
			0, id, nil,
		))
	}

	return rawRequests
}

func TestCollectFeeEmptyFee(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true)

	raws := rawRequestsFromFees(ctx, k, []sdk.Coins{
		testapp.EmptyCoins,
		testapp.EmptyCoins,
		testapp.EmptyCoins,
		testapp.EmptyCoins,
		testapp.EmptyCoins,
	})

	coins, err := k.CollectFee(ctx, testapp.Alice.Address, testapp.EmptyCoins, 1, raws)
	require.NoError(t, err)
	require.Empty(t, coins)

	coins, err = k.CollectFee(ctx, testapp.Alice.Address, testapp.Coins100000000loki, 1, raws)
	require.NoError(t, err)
	require.Empty(t, coins)

	coins, err = k.CollectFee(ctx, testapp.Alice.Address, testapp.EmptyCoins, 2, raws)
	require.NoError(t, err)
	require.Empty(t, coins)

	coins, err = k.CollectFee(ctx, testapp.Alice.Address, testapp.Coins100000000loki, 2, raws)
	require.NoError(t, err)
	require.Empty(t, coins)
}

func TestCollectFeeBasicSuccess(t *testing.T) {
	app, ctx, k := testapp.CreateTestInput(true)

	raws := rawRequestsFromFees(ctx, k, []sdk.Coins{
		testapp.EmptyCoins,
		testapp.Coins1000000loki,
		testapp.EmptyCoins,
		sdk.NewCoins(sdk.NewCoin("loki", sdk.NewInt(2000000))),
		testapp.EmptyCoins,
	})

	balancesRes, err := app.BankKeeper.AllBalances(
		sdk.WrapSDKContext(ctx),
		banktypes.NewQueryAllBalancesRequest(testapp.FeePayer.Address, &query.PageRequest{}),
	)
	feePayerBalances := balancesRes.Balances
	feePayerBalances[0].Amount = feePayerBalances[0].Amount.Sub(sdk.NewInt(3000000))

	coins, err := k.CollectFee(ctx, testapp.FeePayer.Address, testapp.Coins10000000000loki, 1, raws)
	require.NoError(t, err)
	require.Equal(t, sdk.NewCoins(sdk.NewCoin("loki", sdk.NewInt(3000000))), coins)

	testapp.CheckBalances(t, ctx, app.BankKeeper, testapp.FeePayer.Address, feePayerBalances)
	testapp.CheckBalances(t, ctx, app.BankKeeper, app.AccountKeeper.GetModuleAddress(oracletypes.ModuleName), sdk.NewCoins(sdk.NewCoin("loki", sdk.NewInt(3000000))))
}

func TestCollectFeeBasicSuccessWithOtherAskCount(t *testing.T) {
	app, ctx, k := testapp.CreateTestInput(true)

	raws := rawRequestsFromFees(ctx, k, []sdk.Coins{
		testapp.EmptyCoins,
		testapp.Coins1000000loki,
		testapp.EmptyCoins,
		sdk.NewCoins(sdk.NewCoin("loki", sdk.NewInt(2000000))),
		testapp.EmptyCoins,
	})

	balancesRes, err := app.BankKeeper.AllBalances(
		sdk.WrapSDKContext(ctx),
		banktypes.NewQueryAllBalancesRequest(testapp.FeePayer.Address, &query.PageRequest{}),
	)
	feePayerBalances := balancesRes.Balances
	feePayerBalances[0].Amount = feePayerBalances[0].Amount.Sub(sdk.NewInt(12000000))

	coins, err := k.CollectFee(ctx, testapp.FeePayer.Address, testapp.Coins100000000loki, 4, raws)
	require.NoError(t, err)
	require.Equal(t, sdk.NewCoins(sdk.NewCoin("loki", sdk.NewInt(12000000))), coins)

	testapp.CheckBalances(t, ctx, app.BankKeeper, testapp.FeePayer.Address, feePayerBalances)
	testapp.CheckBalances(t, ctx, app.BankKeeper, app.AccountKeeper.GetModuleAddress(oracletypes.ModuleName), sdk.NewCoins(sdk.NewCoin("loki", sdk.NewInt(12000000))))
}

func TestCollectFeeWithMixedAndFeeNotEnough(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true)

	raws := rawRequestsFromFees(ctx, k, []sdk.Coins{
		testapp.EmptyCoins,
		testapp.Coins100000000loki,
		testapp.EmptyCoins,
		sdk.NewCoins(sdk.NewCoin("loki", sdk.NewInt(2000000))),
		testapp.EmptyCoins,
	})

	coins, err := k.CollectFee(ctx, testapp.FeePayer.Address, testapp.EmptyCoins, 1, raws)
	require.Error(t, err)
	require.Nil(t, coins)

	coins, err = k.CollectFee(ctx, testapp.FeePayer.Address, testapp.Coins100000000loki, 1, raws)
	require.Error(t, err)
	require.Nil(t, coins)
}

func TestCollectFeeWithEnoughFeeButInsufficientBalance(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true)

	raws := rawRequestsFromFees(ctx, k, []sdk.Coins{
		testapp.EmptyCoins,
		testapp.Coins100000000loki,
		testapp.EmptyCoins,
		sdk.NewCoins(sdk.NewCoin("loki", sdk.NewInt(2000000))),
		testapp.EmptyCoins,
	})

	coins, err := k.CollectFee(ctx, testapp.Alice.Address, testapp.Coins100000000loki, 1, raws)
	require.Nil(t, coins)
	// MAX is 100m but have only 1m in account
	// First ds collect 1m so there no balance enough for next ds but it doesn't touch limit
	require.EqualError(t, err, "1000000loki is smaller than 100000000loki: insufficient funds")
}

func TestCollectFeeWithWithManyUnitSuccess(t *testing.T) {
	app, ctx, k := testapp.CreateTestInput(true, true)

	raws := rawRequestsFromFees(ctx, k, []sdk.Coins{
		testapp.EmptyCoins,
		testapp.Coins1000000loki,
		testapp.EmptyCoins,
		sdk.NewCoins(sdk.NewCoin("loki", sdk.NewInt(2000000)), sdk.NewCoin("minigeo", sdk.NewInt(1000000))),
		testapp.EmptyCoins,
	})

	newGeo := sdk.NewCoins(sdk.NewInt64Coin("minigeo", 2000000))
	app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, newGeo)

	// Carol have not enough loki but have enough minigeo
	app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, testapp.FeePayer.Address, newGeo)

	coins, err := k.CollectFee(ctx, testapp.FeePayer.Address, testapp.MustGetBalances(ctx, app.BankKeeper, testapp.FeePayer.Address), 1, raws)
	require.NoError(t, err)

	// Coins sum is correct
	require.True(t, sdk.NewCoins(sdk.NewCoin("loki", sdk.NewInt(3000000)), sdk.NewCoin("minigeo", sdk.NewInt(1000000))).IsEqual(coins))

	// FeePayer balance
	// start: 100band, 0abc
	// top-up: 100band, 2abc
	// collect 3 band and 1 abc => 97band, 1abc
	testapp.CheckBalances(t, ctx, app.BankKeeper, testapp.FeePayer.Address, sdk.NewCoins(sdk.NewCoin("loki", sdk.NewInt(97000000)), sdk.NewCoin("minigeo", sdk.NewInt(1000000))))

	// Treasury balance
	// start: 0band, 0abc
	// collect 3 band and 1 abc => 3band, 1abc
	testapp.CheckBalances(t, ctx, app.BankKeeper, app.AccountKeeper.GetModuleAddress(oracletypes.ModuleName), sdk.NewCoins(sdk.NewCoin("loki", sdk.NewInt(103000000)), sdk.NewCoin("minigeo", sdk.NewInt(1000000))))
}

func TestCollectFeeWithWithManyUnitFail(t *testing.T) {
	app, ctx, k := testapp.CreateTestInput(true, true)

	raws := rawRequestsFromFees(ctx, k, []sdk.Coins{
		testapp.EmptyCoins,
		testapp.Coins1000000loki,
		testapp.EmptyCoins,
		sdk.NewCoins(sdk.NewCoin("loki", sdk.NewInt(2000000)), sdk.NewCoin("minigeo", sdk.NewInt(1000000))),
		testapp.EmptyCoins,
	})

	err := app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, sdk.NewCoins(sdk.NewCoin("loki", sdk.NewInt(10000000)), sdk.NewCoin("minigeo", sdk.NewInt(2000000))))
	require.NoError(t, err)
	// Peter have not enough loki and don't have minigeo so don't top up
	// Bob have enough loki and have some but not enough minigeo so add some
	err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, testapp.Bob.Address, sdk.NewCoins(sdk.NewCoin("loki", sdk.NewInt(3000000))))
	require.NoError(t, err)

	err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, testapp.Bob.Address, sdk.NewCoins(sdk.NewCoin("minigeo", sdk.NewInt(1))))
	require.NoError(t, err)

	// Carol have not enough loki but have enough minigeo
	err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, testapp.Carol.Address, sdk.NewCoins(sdk.NewCoin("minigeo", sdk.NewInt(1000000))))
	require.NoError(t, err)

	// Alice
	_, err = k.CollectFee(ctx, testapp.Peter.Address, testapp.MustGetBalances(ctx, app.BankKeeper, testapp.Peter.Address), 1, raws)
	require.EqualError(t, err, "require: 3000000loki, max: 1000000loki: not enough fee")

	// Bob
	_, err = k.CollectFee(ctx, testapp.Bob.Address, testapp.MustGetBalances(ctx, app.BankKeeper, testapp.Bob.Address), 1, raws)
	require.EqualError(t, err, "require: 1000000minigeo, max: 1minigeo: not enough fee")

	// Carol
	_, err = k.CollectFee(ctx, testapp.Carol.Address, testapp.MustGetBalances(ctx, app.BankKeeper, testapp.Carol.Address), 1, raws)
	require.EqualError(t, err, "require: 3000000loki, max: 1000000loki: not enough fee")
}
