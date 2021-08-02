package keeper_test

import (
	"encoding/hex"
	"strings"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	authtypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/stretchr/testify/require"

	"github.com/bandprotocol/chain/v2/pkg/obi"

	"github.com/bandprotocol/chain/v2/testing/testapp"
	"github.com/bandprotocol/chain/v2/x/oracle/keeper"
	"github.com/bandprotocol/chain/v2/x/oracle/types"
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
	require.ErrorIs(t, err, types.ErrInsufficientValidators)
	_, err = k.GetRandomValidators(ctx, 9999, 1)
	require.ErrorIs(t, err, types.ErrInsufficientValidators)
}

func TestGetRandomValidatorsWithActivate(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(false)
	k.SetRollingSeed(ctx, []byte("ROLLING_SEED_WITH_LONG_ENOUGH_ENTROPY"))
	// If no validators are active, you must not be able to get random validators
	_, err := k.GetRandomValidators(ctx, 1, 1)
	require.ErrorIs(t, err, types.ErrInsufficientValidators)
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
	require.ErrorIs(t, err, types.ErrInsufficientValidators)
	// After we deactivate 1 validator due to missing a report, we can only get at most 1 validator.
	k.MissReport(ctx, testapp.Validators[0].ValAddress, time.Now())
	vals, err = k.GetRandomValidators(ctx, 1, 1)
	require.NoError(t, err)
	require.Equal(t, []sdk.ValAddress{testapp.Validators[1].ValAddress}, vals)
	_, err = k.GetRandomValidators(ctx, 2, 1)
	require.ErrorIs(t, err, types.ErrInsufficientValidators)
}

func TestPrepareRequestSuccessBasic(t *testing.T) {
	app, ctx, k := testapp.CreateTestInput(true)
	ctx = ctx.WithBlockTime(testapp.ParseTime(1581589790)).WithBlockHeight(42)

	wrappedGasMeter := testapp.NewGasMeterWrapper(ctx.GasMeter())
	ctx = ctx.WithGasMeter(wrappedGasMeter)

	balancesRes, err := app.BankKeeper.AllBalances(
		sdk.WrapSDKContext(ctx),
		authtypes.NewQueryAllBalancesRequest(testapp.FeePayer.Address, &query.PageRequest{}),
	)
	feePayerBalances := balancesRes.Balances

	// OracleScript#1: Prepare asks for DS#1,2,3 with ExtID#1,2,3 and calldata "beeb"
	m := types.NewMsgRequestData(1, BasicCalldata, 1, 1, BasicClientID, testapp.Coins100000000uband, testapp.TestDefaultPrepareGas, testapp.TestDefaultExecuteGas, testapp.FeePayer.Address)
	id, err := k.PrepareRequest(ctx, m, testapp.FeePayer.Address, nil)
	require.Equal(t, types.RequestID(1), id)
	require.NoError(t, err)
	require.Equal(t, types.NewRequest(
		1, BasicCalldata, []sdk.ValAddress{testapp.Validators[0].ValAddress}, 1,
		42, testapp.ParseTime(1581589790), BasicClientID, []types.RawRequest{
			types.NewRawRequest(1, 1, []byte("beeb")),
			types.NewRawRequest(2, 2, []byte("beeb")),
			types.NewRawRequest(3, 3, []byte("beeb")),
		}, nil, testapp.TestDefaultExecuteGas,
	), k.MustGetRequest(ctx, 1))
	require.Equal(t, sdk.Events{
		sdk.NewEvent(
			authtypes.EventTypeTransfer,
			sdk.NewAttribute(authtypes.AttributeKeyRecipient, testapp.Treasury.Address.String()),
			sdk.NewAttribute(authtypes.AttributeKeySender, testapp.FeePayer.Address.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, testapp.Coins1000000uband.String()),
		), sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeySender, testapp.FeePayer.Address.String()),
		), sdk.NewEvent(
			authtypes.EventTypeTransfer,
			sdk.NewAttribute(authtypes.AttributeKeyRecipient, testapp.Treasury.Address.String()),
			sdk.NewAttribute(authtypes.AttributeKeySender, testapp.FeePayer.Address.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, testapp.Coins1000000uband.String()),
		), sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeySender, testapp.FeePayer.Address.String()),
		), sdk.NewEvent(
			authtypes.EventTypeTransfer,
			sdk.NewAttribute(authtypes.AttributeKeyRecipient, testapp.Treasury.Address.String()),
			sdk.NewAttribute(authtypes.AttributeKeySender, testapp.FeePayer.Address.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, testapp.Coins1000000uband.String()),
		), sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeySender, testapp.FeePayer.Address.String()),
		), sdk.NewEvent(
			types.EventTypeRequest,
			sdk.NewAttribute(types.AttributeKeyID, "1"),
			sdk.NewAttribute(types.AttributeKeyClientID, BasicClientID),
			sdk.NewAttribute(types.AttributeKeyOracleScriptID, "1"),
			sdk.NewAttribute(types.AttributeKeyCalldata, hex.EncodeToString(BasicCalldata)),
			sdk.NewAttribute(types.AttributeKeyAskCount, "1"),
			sdk.NewAttribute(types.AttributeKeyMinCount, "1"),
			sdk.NewAttribute(types.AttributeKeyGasUsed, "785"),
			sdk.NewAttribute(types.AttributeKeyTotalFees, "3000000uband"),
			sdk.NewAttribute(types.AttributeKeyValidator, testapp.Validators[0].ValAddress.String()),
		), sdk.NewEvent(
			types.EventTypeRawRequest,
			sdk.NewAttribute(types.AttributeKeyDataSourceID, "1"),
			sdk.NewAttribute(types.AttributeKeyDataSourceHash, testapp.DataSources[1].Filename),
			sdk.NewAttribute(types.AttributeKeyExternalID, "1"),
			sdk.NewAttribute(types.AttributeKeyCalldata, "beeb"),
			sdk.NewAttribute(types.AttributeKeyFee, "1000000uband"),
		), sdk.NewEvent(
			types.EventTypeRawRequest,
			sdk.NewAttribute(types.AttributeKeyDataSourceID, "2"),
			sdk.NewAttribute(types.AttributeKeyDataSourceHash, testapp.DataSources[2].Filename),
			sdk.NewAttribute(types.AttributeKeyExternalID, "2"),
			sdk.NewAttribute(types.AttributeKeyCalldata, "beeb"),
			sdk.NewAttribute(types.AttributeKeyFee, "1000000uband"),
		), sdk.NewEvent(
			types.EventTypeRawRequest,
			sdk.NewAttribute(types.AttributeKeyDataSourceID, "3"),
			sdk.NewAttribute(types.AttributeKeyDataSourceHash, testapp.DataSources[3].Filename),
			sdk.NewAttribute(types.AttributeKeyExternalID, "3"),
			sdk.NewAttribute(types.AttributeKeyCalldata, "beeb"),
			sdk.NewAttribute(types.AttributeKeyFee, "1000000uband"),
		)}, ctx.EventManager().Events())

	// assert gas consumption
	params := k.GetParams(ctx)
	require.Equal(t, 2, wrappedGasMeter.CountRecord(params.BaseOwasmGas, "BASE_OWASM_FEE"))
	require.Equal(t, 1, wrappedGasMeter.CountRecord(testapp.TestDefaultPrepareGas, "OWASM_PREPARE_FEE"))
	require.Equal(t, 1, wrappedGasMeter.CountRecord(testapp.TestDefaultExecuteGas, "OWASM_EXECUTE_FEE"))

	paid := sdk.NewCoins(sdk.NewInt64Coin("uband", 3000000))
	feePayerBalances = feePayerBalances.Sub(paid)
	testapp.CheckBalances(t, ctx, app.BankKeeper, testapp.FeePayer.Address, feePayerBalances)
	testapp.CheckBalances(t, ctx, app.BankKeeper, testapp.Treasury.Address, paid)
}

func TestPrepareRequestNotEnoughMaxFee(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true)
	ctx = ctx.WithBlockTime(testapp.ParseTime(1581589790)).WithBlockHeight(42)
	// OracleScript#1: Prepare asks for DS#1,2,3 with ExtID#1,2,3 and calldata "beeb"
	m := types.NewMsgRequestData(1, BasicCalldata, 1, 1, BasicClientID, testapp.EmptyCoins, testapp.TestDefaultPrepareGas, testapp.TestDefaultExecuteGas, testapp.FeePayer.Address)
	_, err := k.PrepareRequest(ctx, m, testapp.FeePayer.Address, nil)
	require.EqualError(t, err, "require: 1000000uband, max: 0uband: not enough fee")
	m = types.NewMsgRequestData(1, BasicCalldata, 1, 1, BasicClientID, sdk.NewCoins(sdk.NewInt64Coin("uband", 1000000)), testapp.TestDefaultPrepareGas, testapp.TestDefaultExecuteGas, testapp.FeePayer.Address)
	_, err = k.PrepareRequest(ctx, m, testapp.FeePayer.Address, nil)
	require.EqualError(t, err, "require: 2000000uband, max: 1000000uband: not enough fee")
	m = types.NewMsgRequestData(1, BasicCalldata, 1, 1, BasicClientID, sdk.NewCoins(sdk.NewInt64Coin("uband", 2000000)), testapp.TestDefaultPrepareGas, testapp.TestDefaultExecuteGas, testapp.FeePayer.Address)
	_, err = k.PrepareRequest(ctx, m, testapp.FeePayer.Address, nil)
	require.EqualError(t, err, "require: 3000000uband, max: 2000000uband: not enough fee")
	m = types.NewMsgRequestData(1, BasicCalldata, 1, 1, BasicClientID, sdk.NewCoins(sdk.NewInt64Coin("uband", 2999999)), testapp.TestDefaultPrepareGas, testapp.TestDefaultExecuteGas, testapp.FeePayer.Address)
	_, err = k.PrepareRequest(ctx, m, testapp.FeePayer.Address, nil)
	require.EqualError(t, err, "require: 3000000uband, max: 2999999uband: not enough fee")
	m = types.NewMsgRequestData(1, BasicCalldata, 1, 1, BasicClientID, sdk.NewCoins(sdk.NewInt64Coin("uband", 3000000)), testapp.TestDefaultPrepareGas, testapp.TestDefaultExecuteGas, testapp.FeePayer.Address)
	id, err := k.PrepareRequest(ctx, m, testapp.FeePayer.Address, nil)
	require.NoError(t, err)
	require.Equal(t, types.RequestID(1), id)
}

func TestPrepareRequestNotEnoughFund(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true)
	ctx = ctx.WithBlockTime(testapp.ParseTime(1581589790)).WithBlockHeight(42)
	// OracleScript#1: Prepare asks for DS#1,2,3 with ExtID#1,2,3 and calldata "beeb"
	m := types.NewMsgRequestData(1, BasicCalldata, 1, 1, BasicClientID, testapp.Coins100000000uband, testapp.TestDefaultPrepareGas, testapp.TestDefaultExecuteGas, testapp.Alice.Address)
	_, err := k.PrepareRequest(ctx, m, testapp.Alice.Address, nil)
	require.EqualError(t, err, "0uband is smaller than 1000000uband: insufficient funds")
}

func TestPrepareRequestInvalidCalldataSize(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true)
	m := types.NewMsgRequestData(1, []byte(strings.Repeat("x", 2000)), 1, 1, BasicClientID, testapp.Coins100000000uband, testapp.TestDefaultPrepareGas, testapp.TestDefaultExecuteGas, testapp.Alice.Address)
	_, err := k.PrepareRequest(ctx, m, testapp.FeePayer.Address, nil)
	require.EqualError(t, err, "got: 2000, max: 256: too large calldata")
}

func TestPrepareRequestNotEnoughPrepareGas(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true)
	ctx = ctx.WithBlockTime(testapp.ParseTime(1581589790)).WithBlockHeight(42)

	wrappedGasMeter := testapp.NewGasMeterWrapper(ctx.GasMeter())
	ctx = ctx.WithGasMeter(wrappedGasMeter)

	m := types.NewMsgRequestData(1, BasicCalldata, 1, 1, BasicClientID, testapp.EmptyCoins, 100, testapp.TestDefaultExecuteGas, testapp.Alice.Address)
	_, err := k.PrepareRequest(ctx, m, testapp.FeePayer.Address, nil)
	require.ErrorIs(t, err, types.ErrBadWasmExecution)
	require.Contains(t, err.Error(), "out-of-gas")

	params := k.GetParams(ctx)
	require.Equal(t, 1, wrappedGasMeter.CountRecord(params.BaseOwasmGas, "BASE_OWASM_FEE"))
	require.Equal(t, 1, wrappedGasMeter.CountRecord(100, "OWASM_PREPARE_FEE"))
	require.Equal(t, 0, wrappedGasMeter.CountDescriptor("OWASM_EXECUTE_FEE"))
}

func TestPrepareRequestInvalidAskCountFail(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true)
	params := k.GetParams(ctx)
	params.MaxAskCount = 5
	k.SetParams(ctx, params)

	wrappedGasMeter := testapp.NewGasMeterWrapper(ctx.GasMeter())
	ctx = ctx.WithGasMeter(wrappedGasMeter)

	m := types.NewMsgRequestData(1, BasicCalldata, 10, 1, BasicClientID, testapp.Coins100000000uband, testapp.TestDefaultPrepareGas, testapp.TestDefaultExecuteGas, testapp.Alice.Address)
	_, err := k.PrepareRequest(ctx, m, testapp.FeePayer.Address, nil)
	require.ErrorIs(t, err, types.ErrInvalidAskCount)

	require.Equal(t, 0, wrappedGasMeter.CountDescriptor("BASE_OWASM_FEE"))
	require.Equal(t, 0, wrappedGasMeter.CountDescriptor("OWASM_PREPARE_FEE"))
	require.Equal(t, 0, wrappedGasMeter.CountDescriptor("OWASM_EXECUTE_FEE"))

	m = types.NewMsgRequestData(1, BasicCalldata, 4, 1, BasicClientID, testapp.Coins100000000uband, testapp.TestDefaultPrepareGas, testapp.TestDefaultExecuteGas, testapp.Alice.Address)
	_, err = k.PrepareRequest(ctx, m, testapp.FeePayer.Address, nil)
	require.ErrorIs(t, err, types.ErrInsufficientValidators)

	require.Equal(t, 0, wrappedGasMeter.CountDescriptor("BASE_OWASM_FEE"))
	require.Equal(t, 0, wrappedGasMeter.CountDescriptor("OWASM_PREPARE_FEE"))
	require.Equal(t, 0, wrappedGasMeter.CountDescriptor("OWASM_EXECUTE_FEE"))

	m = types.NewMsgRequestData(1, BasicCalldata, 1, 1, BasicClientID, testapp.Coins100000000uband, testapp.TestDefaultPrepareGas, testapp.TestDefaultExecuteGas, testapp.Alice.Address)
	id, err := k.PrepareRequest(ctx, m, testapp.FeePayer.Address, nil)
	require.Equal(t, types.RequestID(1), id)
	require.NoError(t, err)
	require.Equal(t, 2, wrappedGasMeter.CountDescriptor("BASE_OWASM_FEE"))
	require.Equal(t, 1, wrappedGasMeter.CountDescriptor("OWASM_PREPARE_FEE"))
	require.Equal(t, 1, wrappedGasMeter.CountDescriptor("OWASM_EXECUTE_FEE"))
}

func TestPrepareRequestBaseOwasmFeePanic(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true)
	params := k.GetParams(ctx)
	params.BaseOwasmGas = 100000
	params.PerValidatorRequestGas = 0
	k.SetParams(ctx, params)
	m := types.NewMsgRequestData(1, BasicCalldata, 1, 1, BasicClientID, testapp.Coins100000000uband, testapp.TestDefaultPrepareGas, testapp.TestDefaultExecuteGas, testapp.Alice.Address)
	ctx = ctx.WithGasMeter(sdk.NewGasMeter(90000))
	require.PanicsWithValue(t, sdk.ErrorOutOfGas{Descriptor: "BASE_OWASM_FEE"}, func() { k.PrepareRequest(ctx, m, testapp.FeePayer.Address, nil) })
	ctx = ctx.WithGasMeter(sdk.NewGasMeter(1000000))
	id, err := k.PrepareRequest(ctx, m, testapp.FeePayer.Address, nil)
	require.Equal(t, types.RequestID(1), id)
	require.NoError(t, err)
}

func TestPrepareRequestPerValidatorRequestFeePanic(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true)
	params := k.GetParams(ctx)
	params.BaseOwasmGas = 100000
	params.PerValidatorRequestGas = 50000
	k.SetParams(ctx, params)
	m := types.NewMsgRequestData(1, BasicCalldata, 2, 1, BasicClientID, testapp.Coins100000000uband, testapp.TestDefaultPrepareGas, testapp.TestDefaultExecuteGas, testapp.Alice.Address)
	ctx = ctx.WithGasMeter(sdk.NewGasMeter(90000))
	require.PanicsWithValue(t, sdk.ErrorOutOfGas{Descriptor: "PER_VALIDATOR_REQUEST_FEE"}, func() { k.PrepareRequest(ctx, m, testapp.FeePayer.Address, nil) })
	m = types.NewMsgRequestData(1, BasicCalldata, 1, 1, BasicClientID, testapp.Coins100000000uband, testapp.TestDefaultPrepareGas, testapp.TestDefaultExecuteGas, testapp.Alice.Address)
	ctx = ctx.WithGasMeter(sdk.NewGasMeter(1000000))
	id, err := k.PrepareRequest(ctx, m, testapp.FeePayer.Address, nil)
	require.Equal(t, types.RequestID(1), id)
	require.NoError(t, err)
}

func TestPrepareRequestEmptyCalldata(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true) // Send nil while oracle script expects calldata
	m := types.NewMsgRequestData(4, nil, 1, 1, BasicClientID, testapp.Coins100000000uband, testapp.TestDefaultPrepareGas, testapp.TestDefaultExecuteGas, testapp.Alice.Address)
	_, err := k.PrepareRequest(ctx, m, testapp.FeePayer.Address, nil)
	require.EqualError(t, err, "runtime error while executing the Wasm script: bad wasm execution")
}

func TestPrepareRequestOracleScriptNotFound(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true)
	m := types.NewMsgRequestData(999, BasicCalldata, 1, 1, BasicClientID, testapp.Coins100000000uband, testapp.TestDefaultPrepareGas, testapp.TestDefaultExecuteGas, testapp.Alice.Address)
	_, err := k.PrepareRequest(ctx, m, testapp.FeePayer.Address, nil)
	require.EqualError(t, err, "id: 999: oracle script not found")
}

func TestPrepareRequestBadWasmExecutionFail(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true)
	m := types.NewMsgRequestData(2, BasicCalldata, 1, 1, BasicClientID, testapp.Coins100000000uband, testapp.TestDefaultPrepareGas, testapp.TestDefaultExecuteGas, testapp.Alice.Address)
	_, err := k.PrepareRequest(ctx, m, testapp.FeePayer.Address, nil)
	require.EqualError(t, err, "OEI action to invoke is not available: bad wasm execution")
}

func TestPrepareRequestWithEmptyRawRequest(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true)
	m := types.NewMsgRequestData(3, BasicCalldata, 1, 1, BasicClientID, testapp.Coins100000000uband, testapp.TestDefaultPrepareGas, testapp.TestDefaultExecuteGas, testapp.Alice.Address)
	_, err := k.PrepareRequest(ctx, m, testapp.FeePayer.Address, nil)
	require.EqualError(t, err, "empty raw requests")
}

func TestPrepareRequestUnknownDataSource(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true)
	m := types.NewMsgRequestData(4, obi.MustEncode(testapp.Wasm4Input{
		IDs:      []int64{1, 2, 99},
		Calldata: "beeb",
	}), 1, 1, BasicClientID, testapp.Coins100000000uband, testapp.TestDefaultPrepareGas, testapp.TestDefaultExecuteGas, testapp.Alice.Address)
	_, err := k.PrepareRequest(ctx, m, testapp.FeePayer.Address, nil)
	require.EqualError(t, err, "id: 99: data source not found")
}

func TestPrepareRequestInvalidDataSourceCount(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true)
	params := k.GetParams(ctx)
	params.MaxRawRequestCount = 3
	k.SetParams(ctx, params)
	m := types.NewMsgRequestData(4, obi.MustEncode(testapp.Wasm4Input{
		IDs:      []int64{1, 2, 3, 4},
		Calldata: "beeb",
	}), 1, 1, BasicClientID, testapp.Coins100000000uband, testapp.TestDefaultPrepareGas, testapp.TestDefaultExecuteGas, testapp.Alice.Address)
	_, err := k.PrepareRequest(ctx, m, testapp.FeePayer.Address, nil)
	require.ErrorIs(t, err, types.ErrBadWasmExecution)
	m = types.NewMsgRequestData(4, obi.MustEncode(testapp.Wasm4Input{
		IDs:      []int64{1, 2, 3},
		Calldata: "beeb",
	}), 1, 1, BasicClientID, testapp.Coins100000000uband, testapp.TestDefaultPrepareGas, testapp.TestDefaultExecuteGas, testapp.Alice.Address)
	id, err := k.PrepareRequest(ctx, m, testapp.FeePayer.Address, nil)
	require.Equal(t, types.RequestID(1), id)
	require.NoError(t, err)
}

func TestPrepareRequestTooMuchWasmGas(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true)

	m := types.NewMsgRequestData(5, BasicCalldata, 1, 1, BasicClientID, testapp.Coins100000000uband, testapp.TestDefaultPrepareGas, testapp.TestDefaultExecuteGas, testapp.Alice.Address)
	id, err := k.PrepareRequest(ctx, m, testapp.FeePayer.Address, nil)
	require.Equal(t, types.RequestID(1), id)
	require.NoError(t, err)
	m = types.NewMsgRequestData(6, BasicCalldata, 1, 1, BasicClientID, testapp.Coins100000000uband, testapp.TestDefaultPrepareGas, testapp.TestDefaultExecuteGas, testapp.Alice.Address)
	_, err = k.PrepareRequest(ctx, m, testapp.FeePayer.Address, nil)
	require.EqualError(t, err, "out-of-gas while executing the wasm script: bad wasm execution")
}

func TestPrepareRequestTooLargeCalldata(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true)

	m := types.NewMsgRequestData(7, BasicCalldata, 1, 1, BasicClientID, testapp.Coins100000000uband, testapp.TestDefaultPrepareGas, testapp.TestDefaultExecuteGas, testapp.Alice.Address)
	id, err := k.PrepareRequest(ctx, m, testapp.FeePayer.Address, nil)
	require.Equal(t, types.RequestID(1), id)
	require.NoError(t, err)
	m = types.NewMsgRequestData(8, BasicCalldata, 1, 1, BasicClientID, testapp.Coins100000000uband, testapp.TestDefaultPrepareGas, testapp.TestDefaultExecuteGas, testapp.Alice.Address)
	_, err = k.PrepareRequest(ctx, m, testapp.FeePayer.Address, nil)
	require.EqualError(t, err, "span to write is too small: bad wasm execution")
}

func TestResolveRequestSuccess(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true)
	ctx = ctx.WithBlockTime(testapp.ParseTime(1581589890))
	k.SetRequest(ctx, 42, types.NewRequest(
		// 1st Wasm - return "beeb"
		1, BasicCalldata, []sdk.ValAddress{testapp.Validators[0].ValAddress, testapp.Validators[1].ValAddress}, 1,
		42, testapp.ParseTime(1581589790), BasicClientID, []types.RawRequest{
			types.NewRawRequest(1, 1, []byte("beeb")),
		}, nil, testapp.TestDefaultExecuteGas,
	))
	k.SetReport(ctx, 42, types.NewReport(
		testapp.Validators[0].ValAddress, true, []types.RawReport{
			types.NewRawReport(1, 0, []byte("beeb")),
		},
	))
	k.ResolveRequest(ctx, 42)
	expectResult := types.NewResult(
		BasicClientID, 1, BasicCalldata, 2, 1,
		42, 1, testapp.ParseTime(1581589790).Unix(),
		testapp.ParseTime(1581589890).Unix(), types.RESOLVE_STATUS_SUCCESS, []byte("beeb"),
	)
	require.Equal(t, expectResult, k.MustGetResult(ctx, 42))
	require.Equal(t, sdk.Events{sdk.NewEvent(
		types.EventTypeResolve,
		sdk.NewAttribute(types.AttributeKeyID, "42"),
		sdk.NewAttribute(types.AttributeKeyResolveStatus, "1"),
		sdk.NewAttribute(types.AttributeKeyResult, "62656562"), // hex of "beeb"
		sdk.NewAttribute(types.AttributeKeyGasUsed, "516"),
	)}, ctx.EventManager().Events())
}

func TestResolveRequestSuccessComplex(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true)
	ctx = ctx.WithBlockTime(testapp.ParseTime(1581589890))
	k.SetRequest(ctx, 42, types.NewRequest(
		// 4th Wasm. Append all reports from all validators.
		4, obi.MustEncode(testapp.Wasm4Input{
			IDs:      []int64{1, 2},
			Calldata: string(BasicCalldata),
		}), []sdk.ValAddress{testapp.Validators[0].ValAddress, testapp.Validators[1].ValAddress}, 1,
		42, testapp.ParseTime(1581589790), BasicClientID, []types.RawRequest{
			types.NewRawRequest(0, 1, BasicCalldata),
			types.NewRawRequest(1, 2, BasicCalldata),
		}, nil, testapp.TestDefaultExecuteGas,
	))
	k.SetReport(ctx, 42, types.NewReport(
		testapp.Validators[0].ValAddress, true, []types.RawReport{
			types.NewRawReport(0, 0, []byte("beebd1v1")),
			types.NewRawReport(1, 0, []byte("beebd2v1")),
		},
	))
	k.SetReport(ctx, 42, types.NewReport(
		testapp.Validators[1].ValAddress, true, []types.RawReport{
			types.NewRawReport(0, 0, []byte("beebd1v2")),
			types.NewRawReport(1, 0, []byte("beebd2v2")),
		},
	))
	k.ResolveRequest(ctx, 42)
	result := types.NewResult(
		BasicClientID, 4, obi.MustEncode(testapp.Wasm4Input{
			IDs:      []int64{1, 2},
			Calldata: string(BasicCalldata),
		}), 2, 1,
		42, 2, testapp.ParseTime(1581589790).Unix(),
		testapp.ParseTime(1581589890).Unix(), types.RESOLVE_STATUS_SUCCESS,
		obi.MustEncode(testapp.Wasm4Output{Ret: "beebd1v1beebd1v2beebd2v1beebd2v2"}),
	)
	require.Equal(t, result, k.MustGetResult(ctx, 42))
	require.Equal(t, sdk.Events{sdk.NewEvent(
		types.EventTypeResolve,
		sdk.NewAttribute(types.AttributeKeyID, "42"),
		sdk.NewAttribute(types.AttributeKeyResolveStatus, "1"),
		sdk.NewAttribute(types.AttributeKeyResult, "000000206265656264317631626565626431763262656562643276316265656264327632"),
		sdk.NewAttribute(types.AttributeKeyGasUsed, "10274"),
	)}, ctx.EventManager().Events())
}

func TestResolveRequestOutOfGas(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true)
	ctx = ctx.WithBlockTime(testapp.ParseTime(1581589890))
	k.SetRequest(ctx, 42, types.NewRequest(
		// 1st Wasm - return "beeb"
		1, BasicCalldata, []sdk.ValAddress{testapp.Validators[0].ValAddress, testapp.Validators[1].ValAddress}, 1,
		42, testapp.ParseTime(1581589790), BasicClientID, []types.RawRequest{
			types.NewRawRequest(1, 1, []byte("beeb")),
		}, nil, 0,
	))
	k.SetReport(ctx, 42, types.NewReport(
		testapp.Validators[0].ValAddress, true, []types.RawReport{
			types.NewRawReport(1, 0, []byte("beeb")),
		},
	))
	k.ResolveRequest(ctx, 42)
	result := types.NewResult(
		BasicClientID, 1, BasicCalldata, 2, 1,
		42, 1, testapp.ParseTime(1581589790).Unix(),
		testapp.ParseTime(1581589890).Unix(), types.RESOLVE_STATUS_FAILURE, nil,
	)
	require.Equal(t, result, k.MustGetResult(ctx, 42))
}

func TestResolveReadNilExternalData(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true)
	ctx = ctx.WithBlockTime(testapp.ParseTime(1581589890))
	k.SetRequest(ctx, 42, types.NewRequest(
		// 4th Wasm. Append all reports from all validators.
		4, obi.MustEncode(testapp.Wasm4Input{
			IDs:      []int64{1, 2},
			Calldata: string(BasicCalldata),
		}), []sdk.ValAddress{testapp.Validators[0].ValAddress, testapp.Validators[1].ValAddress}, 1,
		42, testapp.ParseTime(1581589790), BasicClientID, []types.RawRequest{
			types.NewRawRequest(0, 1, BasicCalldata),
			types.NewRawRequest(1, 2, BasicCalldata),
		}, nil, testapp.TestDefaultExecuteGas,
	))
	k.SetReport(ctx, 42, types.NewReport(
		testapp.Validators[0].ValAddress, true, []types.RawReport{
			types.NewRawReport(0, 0, nil),
			types.NewRawReport(1, 0, []byte("beebd2v1")),
		},
	))
	k.SetReport(ctx, 42, types.NewReport(
		testapp.Validators[1].ValAddress, true, []types.RawReport{
			types.NewRawReport(0, 0, []byte("beebd1v2")),
			types.NewRawReport(1, 0, nil),
		},
	))
	k.ResolveRequest(ctx, 42)
	result := types.NewResult(
		BasicClientID, 4, obi.MustEncode(testapp.Wasm4Input{
			IDs:      []int64{1, 2},
			Calldata: string(BasicCalldata),
		}), 2, 1,
		42, 2, testapp.ParseTime(1581589790).Unix(),
		testapp.ParseTime(1581589890).Unix(), types.RESOLVE_STATUS_SUCCESS,
		obi.MustEncode(testapp.Wasm4Output{Ret: "beebd1v2beebd2v1"}),
	)
	require.Equal(t, result, k.MustGetResult(ctx, 42))
	require.Equal(t, sdk.Events{sdk.NewEvent(
		types.EventTypeResolve,
		sdk.NewAttribute(types.AttributeKeyID, "42"),
		sdk.NewAttribute(types.AttributeKeyResolveStatus, "1"),
		sdk.NewAttribute(types.AttributeKeyResult, "0000001062656562643176326265656264327631"),
		sdk.NewAttribute(types.AttributeKeyGasUsed, "9293"),
	)}, ctx.EventManager().Events())
}

func TestResolveRequestNoReturnData(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true)
	ctx = ctx.WithBlockTime(testapp.ParseTime(1581589890))
	k.SetRequest(ctx, 42, types.NewRequest(
		// 3rd Wasm - do nothing
		3, BasicCalldata, []sdk.ValAddress{testapp.Validators[0].ValAddress, testapp.Validators[1].ValAddress}, 1,
		42, testapp.ParseTime(1581589790), BasicClientID, []types.RawRequest{
			types.NewRawRequest(1, 1, []byte("beeb")),
		}, nil, 0,
	))
	k.SetReport(ctx, 42, types.NewReport(
		testapp.Validators[0].ValAddress, true, []types.RawReport{
			types.NewRawReport(1, 0, []byte("beeb")),
		},
	))
	k.ResolveRequest(ctx, 42)
	result := types.NewResult(
		BasicClientID, 3, BasicCalldata, 2, 1, 42, 1, testapp.ParseTime(1581589790).Unix(),
		testapp.ParseTime(1581589890).Unix(), types.RESOLVE_STATUS_FAILURE, nil,
	)
	require.Equal(t, result, k.MustGetResult(ctx, 42))
	require.Equal(t, sdk.Events{sdk.NewEvent(
		types.EventTypeResolve,
		sdk.NewAttribute(types.AttributeKeyID, "42"),
		sdk.NewAttribute(types.AttributeKeyResolveStatus, "2"),
		sdk.NewAttribute(types.AttributeKeyReason, "no return data"),
	)}, ctx.EventManager().Events())
}

func TestResolveRequestWasmFailure(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true)
	ctx = ctx.WithBlockTime(testapp.ParseTime(1581589890))
	k.SetRequest(ctx, 42, types.NewRequest(
		// 6th Wasm - out-of-gas
		6, BasicCalldata, []sdk.ValAddress{testapp.Validators[0].ValAddress, testapp.Validators[1].ValAddress}, 1,
		42, testapp.ParseTime(1581589790), BasicClientID, []types.RawRequest{
			types.NewRawRequest(1, 1, []byte("beeb")),
		}, nil, 0,
	))
	k.SetReport(ctx, 42, types.NewReport(
		testapp.Validators[0].ValAddress, true, []types.RawReport{
			types.NewRawReport(1, 0, []byte("beeb")),
		},
	))
	k.ResolveRequest(ctx, 42)
	result := types.NewResult(
		BasicClientID, 6, BasicCalldata, 2, 1, 42, 1, testapp.ParseTime(1581589790).Unix(),
		testapp.ParseTime(1581589890).Unix(), types.RESOLVE_STATUS_FAILURE, nil,
	)
	require.Equal(t, result, k.MustGetResult(ctx, 42))
	require.Equal(t, sdk.Events{sdk.NewEvent(
		types.EventTypeResolve,
		sdk.NewAttribute(types.AttributeKeyID, "42"),
		sdk.NewAttribute(types.AttributeKeyResolveStatus, "2"),
		sdk.NewAttribute(types.AttributeKeyReason, "out-of-gas while executing the wasm script"),
	)}, ctx.EventManager().Events())
}

func TestResolveRequestCallReturnDataSeveralTimes(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true)
	ctx = ctx.WithBlockTime(testapp.ParseTime(1581589890))
	k.SetRequest(ctx, 42, types.NewRequest(
		// 9th Wasm - set return data several times
		9, BasicCalldata, []sdk.ValAddress{testapp.Validators[0].ValAddress, testapp.Validators[1].ValAddress}, 1,
		42, testapp.ParseTime(1581589790), BasicClientID, []types.RawRequest{
			types.NewRawRequest(1, 1, []byte("beeb")),
		}, nil, testapp.TestDefaultExecuteGas,
	))
	k.ResolveRequest(ctx, 42)

	result := types.NewResult(
		BasicClientID, 9, BasicCalldata, 2, 1, 42, 0, testapp.ParseTime(1581589790).Unix(),
		testapp.ParseTime(1581589890).Unix(), types.RESOLVE_STATUS_FAILURE, nil,
	)
	require.Equal(t, result, k.MustGetResult(ctx, 42))

	require.Equal(t, sdk.Events{sdk.NewEvent(
		types.EventTypeResolve,
		sdk.NewAttribute(types.AttributeKeyID, "42"),
		sdk.NewAttribute(types.AttributeKeyResolveStatus, "2"),
		sdk.NewAttribute(types.AttributeKeyReason, "set return data is called more than once"),
	)}, ctx.EventManager().Events())
}

func rawRequestsFromFees(ctx sdk.Context, k keeper.Keeper, fees []sdk.Coins) []types.RawRequest {
	var rawRequests []types.RawRequest
	for _, f := range fees {
		id := k.AddDataSource(ctx, types.NewDataSource(
			testapp.Owner.Address,
			"mock ds",
			"there is no real code",
			"no file",
			f,
			testapp.Treasury.Address,
		))

		rawRequests = append(rawRequests, types.NewRawRequest(
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

	coins, err = k.CollectFee(ctx, testapp.Alice.Address, testapp.Coins100000000uband, 1, raws)
	require.NoError(t, err)
	require.Empty(t, coins)

	coins, err = k.CollectFee(ctx, testapp.Alice.Address, testapp.EmptyCoins, 2, raws)
	require.NoError(t, err)
	require.Empty(t, coins)

	coins, err = k.CollectFee(ctx, testapp.Alice.Address, testapp.Coins100000000uband, 2, raws)
	require.NoError(t, err)
	require.Empty(t, coins)
}

func TestCollectFeeBasicSuccess(t *testing.T) {
	app, ctx, k := testapp.CreateTestInput(true)

	raws := rawRequestsFromFees(ctx, k, []sdk.Coins{
		testapp.EmptyCoins,
		testapp.Coins1000000uband,
		testapp.EmptyCoins,
		sdk.NewCoins(sdk.NewCoin("uband", sdk.NewInt(2000000))),
		testapp.EmptyCoins,
	})

	balancesRes, err := app.BankKeeper.AllBalances(
		sdk.WrapSDKContext(ctx),
		authtypes.NewQueryAllBalancesRequest(testapp.FeePayer.Address, &query.PageRequest{}),
	)
	feePayerBalances := balancesRes.Balances
	feePayerBalances[0].Amount = feePayerBalances[0].Amount.Sub(sdk.NewInt(3000000))

	coins, err := k.CollectFee(ctx, testapp.FeePayer.Address, testapp.Coins100000000uband, 1, raws)
	require.NoError(t, err)
	require.Equal(t, sdk.NewCoins(sdk.NewCoin("uband", sdk.NewInt(3000000))), coins)

	testapp.CheckBalances(t, ctx, app.BankKeeper, testapp.FeePayer.Address, feePayerBalances)
	testapp.CheckBalances(t, ctx, app.BankKeeper, testapp.Treasury.Address, sdk.NewCoins(sdk.NewCoin("uband", sdk.NewInt(3000000))))
}

func TestCollectFeeBasicSuccessWithOtherAskCount(t *testing.T) {
	app, ctx, k := testapp.CreateTestInput(true)

	raws := rawRequestsFromFees(ctx, k, []sdk.Coins{
		testapp.EmptyCoins,
		testapp.Coins1000000uband,
		testapp.EmptyCoins,
		sdk.NewCoins(sdk.NewCoin("uband", sdk.NewInt(2000000))),
		testapp.EmptyCoins,
	})

	balancesRes, err := app.BankKeeper.AllBalances(
		sdk.WrapSDKContext(ctx),
		authtypes.NewQueryAllBalancesRequest(testapp.FeePayer.Address, &query.PageRequest{}),
	)
	feePayerBalances := balancesRes.Balances
	feePayerBalances[0].Amount = feePayerBalances[0].Amount.Sub(sdk.NewInt(12000000))

	coins, err := k.CollectFee(ctx, testapp.FeePayer.Address, testapp.Coins100000000uband, 4, raws)
	require.NoError(t, err)
	require.Equal(t, sdk.NewCoins(sdk.NewCoin("uband", sdk.NewInt(12000000))), coins)

	testapp.CheckBalances(t, ctx, app.BankKeeper, testapp.FeePayer.Address, feePayerBalances)
	testapp.CheckBalances(t, ctx, app.BankKeeper, testapp.Treasury.Address, sdk.NewCoins(sdk.NewCoin("uband", sdk.NewInt(12000000))))
}

func TestCollectFeeWithMixedAndFeeNotEnough(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true)

	raws := rawRequestsFromFees(ctx, k, []sdk.Coins{
		testapp.EmptyCoins,
		testapp.Coins1000000uband,
		testapp.EmptyCoins,
		sdk.NewCoins(sdk.NewCoin("uband", sdk.NewInt(2000000))),
		testapp.EmptyCoins,
	})

	coins, err := k.CollectFee(ctx, testapp.FeePayer.Address, testapp.EmptyCoins, 1, raws)
	require.ErrorIs(t, err, types.ErrNotEnoughFee)
	require.Nil(t, coins)

	coins, err = k.CollectFee(ctx, testapp.FeePayer.Address, testapp.Coins1000000uband, 1, raws)
	require.ErrorIs(t, err, types.ErrNotEnoughFee)
	require.Nil(t, coins)
}

func TestCollectFeeWithEnoughFeeButInsufficientBalance(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true)

	raws := rawRequestsFromFees(ctx, k, []sdk.Coins{
		testapp.EmptyCoins,
		testapp.Coins1000000uband,
		testapp.EmptyCoins,
		sdk.NewCoins(sdk.NewCoin("uband", sdk.NewInt(2000000))),
		testapp.EmptyCoins,
	})

	coins, err := k.CollectFee(ctx, testapp.Alice.Address, testapp.Coins100000000uband, 1, raws)
	require.Nil(t, coins)
	// MAX is 100m but have only 1m in account
	// First ds collect 1m so there no balance enough for next ds but it doesn't touch limit
	require.EqualError(t, err, "0uband is smaller than 2000000uband: insufficient funds")
}

func TestCollectFeeWithWithManyUnitSuccess(t *testing.T) {
	app, ctx, k := testapp.CreateTestInput(true)

	raws := rawRequestsFromFees(ctx, k, []sdk.Coins{
		testapp.EmptyCoins,
		testapp.Coins1000000uband,
		testapp.EmptyCoins,
		sdk.NewCoins(sdk.NewCoin("uband", sdk.NewInt(2000000)), sdk.NewCoin("uabc", sdk.NewInt(1000000))),
		testapp.EmptyCoins,
	})

	app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, sdk.NewCoins(sdk.NewCoin("uabc", sdk.NewInt(2000000))))

	// Carol have not enough uband but have enough uabc
	app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, testapp.FeePayer.Address, sdk.NewCoins(sdk.NewCoin("uabc", sdk.NewInt(2000000))))

	coins, err := k.CollectFee(ctx, testapp.FeePayer.Address, testapp.MustGetBalances(ctx, app.BankKeeper, testapp.FeePayer.Address), 1, raws)
	require.NoError(t, err)

	// Coins sum is correct
	require.True(t, sdk.NewCoins(sdk.NewCoin("uband", sdk.NewInt(3000000)), sdk.NewCoin("uabc", sdk.NewInt(1000000))).IsEqual(coins))

	// FeePayer balance
	// start: 100band, 0abc
	// top-up: 100band, 2abc
	// collect 3 band and 1 abc => 97band, 1abc
	testapp.CheckBalances(t, ctx, app.BankKeeper, testapp.FeePayer.Address, sdk.NewCoins(sdk.NewCoin("uband", sdk.NewInt(97000000)), sdk.NewCoin("uabc", sdk.NewInt(1000000))))

	// Treasury balance
	// start: 0band, 0abc
	// collect 3 band and 1 abc => 3band, 1abc
	testapp.CheckBalances(t, ctx, app.BankKeeper, testapp.Treasury.Address, sdk.NewCoins(sdk.NewCoin("uband", sdk.NewInt(3000000)), sdk.NewCoin("uabc", sdk.NewInt(1000000))))
}

func TestCollectFeeWithWithManyUnitFail(t *testing.T) {
	app, ctx, k := testapp.CreateTestInput(true)

	raws := rawRequestsFromFees(ctx, k, []sdk.Coins{
		testapp.EmptyCoins,
		testapp.Coins1000000uband,
		testapp.EmptyCoins,
		sdk.NewCoins(sdk.NewCoin("uband", sdk.NewInt(2000000)), sdk.NewCoin("uabc", sdk.NewInt(1000000))),
		testapp.EmptyCoins,
	})

	app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, sdk.NewCoins(sdk.NewCoin("uband", sdk.NewInt(10000000)), sdk.NewCoin("uabc", sdk.NewInt(2000000))))
	// Alice have no enough uband and don't have uabc so don't top up
	// Bob have enough uband and have some but not enough uabc so add some
	app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, testapp.Bob.Address, sdk.NewCoins(sdk.NewCoin("uband", sdk.NewInt(3000000))))
	app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, testapp.Bob.Address, sdk.NewCoins(sdk.NewCoin("uabc", sdk.NewInt(1))))
	// Carol have not enough uband but have enough uabc
	app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, testapp.Carol.Address, sdk.NewCoins(sdk.NewCoin("uabc", sdk.NewInt(1000000))))

	// Alice
	_, err := k.CollectFee(ctx, testapp.Alice.Address, testapp.MustGetBalances(ctx, app.BankKeeper, testapp.Alice.Address), 1, raws)
	require.EqualError(t, err, "require: 1000000uabc, max: 0uabc: not enough fee")

	// Bob
	_, err = k.CollectFee(ctx, testapp.Bob.Address, testapp.MustGetBalances(ctx, app.BankKeeper, testapp.Bob.Address), 1, raws)
	require.EqualError(t, err, "require: 1000000uabc, max: 1uabc: not enough fee")

	// Carol
	_, err = k.CollectFee(ctx, testapp.Carol.Address, testapp.MustGetBalances(ctx, app.BankKeeper, testapp.Carol.Address), 1, raws)
	require.EqualError(t, err, "require: 3000000uband, max: 1000000uband: not enough fee")
}
