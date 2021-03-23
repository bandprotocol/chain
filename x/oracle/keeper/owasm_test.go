package keeper_test

import (
	"encoding/hex"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"

	"github.com/bandprotocol/chain/pkg/obi"
	"github.com/bandprotocol/chain/x/oracle/testapp"
	"github.com/bandprotocol/chain/x/oracle/types"
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
	_, ctx, k := testapp.CreateTestInput(true)
	ctx = ctx.WithBlockTime(testapp.ParseTime(1581589790)).WithBlockHeight(42)
	// OracleScript#1: Prepare asks for DS#1,2,3 with ExtID#1,2,3 and calldata "beeb"
	m := types.NewMsgRequestData(1, BasicCalldata, 1, 1, BasicClientID, testapp.Coins100000000uband, testapp.FeePayer.Address)
	id, err := k.PrepareRequest(ctx, m, testapp.FeePayer.Address, nil)
	require.Equal(t, types.RequestID(1), id)
	require.NoError(t, err)
	require.Equal(t, types.NewRequest(
		1, BasicCalldata, []sdk.ValAddress{testapp.Validators[0].ValAddress}, 1,
		42, testapp.ParseTime(1581589790), BasicClientID, []types.RawRequest{
			types.NewRawRequest(1, 1, []byte("beeb")),
			types.NewRawRequest(2, 2, []byte("beeb")),
			types.NewRawRequest(3, 3, []byte("beeb")),
		}, nil,
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
			sdk.NewAttribute(types.AttributeKeyValidator, testapp.Validators[0].ValAddress.String()),
		), sdk.NewEvent(
			types.EventTypeRawRequest,
			sdk.NewAttribute(types.AttributeKeyDataSourceID, "1"),
			sdk.NewAttribute(types.AttributeKeyDataSourceHash, testapp.DataSources[1].Filename),
			sdk.NewAttribute(types.AttributeKeyExternalID, "1"),
			sdk.NewAttribute(types.AttributeKeyCalldata, "beeb"),
		), sdk.NewEvent(
			types.EventTypeRawRequest,
			sdk.NewAttribute(types.AttributeKeyDataSourceID, "2"),
			sdk.NewAttribute(types.AttributeKeyDataSourceHash, testapp.DataSources[2].Filename),
			sdk.NewAttribute(types.AttributeKeyExternalID, "2"),
			sdk.NewAttribute(types.AttributeKeyCalldata, "beeb"),
		), sdk.NewEvent(
			types.EventTypeRawRequest,
			sdk.NewAttribute(types.AttributeKeyDataSourceID, "3"),
			sdk.NewAttribute(types.AttributeKeyDataSourceHash, testapp.DataSources[3].Filename),
			sdk.NewAttribute(types.AttributeKeyExternalID, "3"),
			sdk.NewAttribute(types.AttributeKeyCalldata, "beeb"),
		)}, ctx.EventManager().Events())
}

func TestPrepareRequestSuccessBasicNotEnoughMaxFee(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true)
	ctx = ctx.WithBlockTime(testapp.ParseTime(1581589790)).WithBlockHeight(42)
	// OracleScript#1: Prepare asks for DS#1,2,3 with ExtID#1,2,3 and calldata "beeb"
	m := types.NewMsgRequestData(1, BasicCalldata, 1, 1, BasicClientID, testapp.EmptyCoins, testapp.FeePayer.Address)
	_, err := k.PrepareRequest(ctx, m, testapp.FeePayer.Address, nil)
	require.EqualError(t, err, "require: 1000000, max: 0: not enough fee: not enough fee")
	m = types.NewMsgRequestData(1, BasicCalldata, 1, 1, BasicClientID, sdk.NewCoins(sdk.NewInt64Coin("uband", 1000000)), testapp.FeePayer.Address)
	_, err = k.PrepareRequest(ctx, m, testapp.FeePayer.Address, nil)
	require.EqualError(t, err, "require: 2000000, max: 1000000: not enough fee: not enough fee")
	m = types.NewMsgRequestData(1, BasicCalldata, 1, 1, BasicClientID, sdk.NewCoins(sdk.NewInt64Coin("uband", 2000000)), testapp.FeePayer.Address)
	_, err = k.PrepareRequest(ctx, m, testapp.FeePayer.Address, nil)
	require.EqualError(t, err, "require: 3000000, max: 2000000: not enough fee: not enough fee")
	m = types.NewMsgRequestData(1, BasicCalldata, 1, 1, BasicClientID, sdk.NewCoins(sdk.NewInt64Coin("uband", 2999999)), testapp.FeePayer.Address)
	_, err = k.PrepareRequest(ctx, m, testapp.FeePayer.Address, nil)
	require.EqualError(t, err, "require: 3000000, max: 2999999: not enough fee: not enough fee")
	m = types.NewMsgRequestData(1, BasicCalldata, 1, 1, BasicClientID, sdk.NewCoins(sdk.NewInt64Coin("uband", 3000000)), testapp.FeePayer.Address)
	id, err := k.PrepareRequest(ctx, m, testapp.FeePayer.Address, nil)
	require.NoError(t, err)
	require.Equal(t, types.RequestID(1), id)
}

func TestPrepareRequestSuccessBasicNotEnoughFund(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true)
	ctx = ctx.WithBlockTime(testapp.ParseTime(1581589790)).WithBlockHeight(42)
	// OracleScript#1: Prepare asks for DS#1,2,3 with ExtID#1,2,3 and calldata "beeb"
	m := types.NewMsgRequestData(1, BasicCalldata, 1, 1, BasicClientID, testapp.Coins100000000uband, testapp.Alice.Address)
	_, err := k.PrepareRequest(ctx, m, testapp.Alice.Address, nil)
	require.EqualError(t, err, "0uband is smaller than 1000000uband: insufficient funds: not enough fee")
}

func TestPrepareRequestInvalidAskCountFail(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true)
	k.SetParam(ctx, types.KeyMaxAskCount, 5)
	m := types.NewMsgRequestData(1, BasicCalldata, 10, 1, BasicClientID, testapp.Coins100000000uband, testapp.Alice.Address)
	_, err := k.PrepareRequest(ctx, m, testapp.FeePayer.Address, nil)
	// require.EqualError(t, err, "invalid ask count: got: 10, max: 5")
	m = types.NewMsgRequestData(1, BasicCalldata, 4, 1, BasicClientID, testapp.Coins100000000uband, testapp.Alice.Address)
	_, err = k.PrepareRequest(ctx, m, testapp.FeePayer.Address, nil)
	// require.EqualError(t, err, "insufficent available validators: 3 < 4")
	m = types.NewMsgRequestData(1, BasicCalldata, 1, 1, BasicClientID, testapp.Coins100000000uband, testapp.Alice.Address)
	id, err := k.PrepareRequest(ctx, m, testapp.FeePayer.Address, nil)
	require.Equal(t, types.RequestID(1), id)
	require.NoError(t, err)
}

func TestPrepareRequestBaseRequestFeePanic(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true)
	k.SetParam(ctx, types.KeyBaseRequestGas, 100000) // Set BaseRequestGas to 100000
	k.SetParam(ctx, types.KeyPerValidatorRequestGas, 0)
	m := types.NewMsgRequestData(1, BasicCalldata, 1, 1, BasicClientID, testapp.Coins100000000uband, testapp.Alice.Address)
	ctx = ctx.WithGasMeter(sdk.NewGasMeter(90000))
	require.PanicsWithValue(t, sdk.ErrorOutOfGas{Descriptor: "BASE_REQUEST_FEE"}, func() { k.PrepareRequest(ctx, m, testapp.FeePayer.Address, nil) })
	ctx = ctx.WithGasMeter(sdk.NewGasMeter(200000))
	id, err := k.PrepareRequest(ctx, m, testapp.FeePayer.Address, nil)
	require.Equal(t, types.RequestID(1), id)
	require.NoError(t, err)
}

func TestPrepareRequestPerValidatorRequestFeePanic(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true)
	k.SetParam(ctx, types.KeyBaseRequestGas, 100000)
	k.SetParam(ctx, types.KeyPerValidatorRequestGas, 50000) // Set erValidatorRequestGas to 50000
	m := types.NewMsgRequestData(1, BasicCalldata, 2, 1, BasicClientID, testapp.Coins100000000uband, testapp.Alice.Address)
	ctx = ctx.WithGasMeter(sdk.NewGasMeter(200000))
	require.PanicsWithValue(t, sdk.ErrorOutOfGas{Descriptor: "PER_VALIDATOR_REQUEST_FEE"}, func() { k.PrepareRequest(ctx, m, testapp.FeePayer.Address, nil) })
	m = types.NewMsgRequestData(1, BasicCalldata, 1, 1, BasicClientID, testapp.Coins100000000uband, testapp.Alice.Address)
	ctx = ctx.WithGasMeter(sdk.NewGasMeter(240000))
	id, err := k.PrepareRequest(ctx, m, testapp.FeePayer.Address, nil)
	require.Equal(t, types.RequestID(1), id)
	require.NoError(t, err)
}

func TestPrepareRequestEmptyCalldata(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true) // Send nil while oracle script expects calldata
	m := types.NewMsgRequestData(4, nil, 1, 1, BasicClientID, testapp.Coins100000000uband, testapp.Alice.Address)
	_, err := k.PrepareRequest(ctx, m, testapp.FeePayer.Address, nil)
	require.EqualError(t, err, "runtime error while executing the Wasm script: bad wasm execution")
}

func TestPrepareRequestOracleScriptNotFound(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true)
	m := types.NewMsgRequestData(999, BasicCalldata, 1, 1, BasicClientID, testapp.Coins100000000uband, testapp.Alice.Address)
	_, err := k.PrepareRequest(ctx, m, testapp.FeePayer.Address, nil)
	require.EqualError(t, err, "id: 999: oracle script not found")
}

func TestPrepareRequestBadWasmExecutionFail(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true)
	m := types.NewMsgRequestData(2, BasicCalldata, 1, 1, BasicClientID, testapp.Coins100000000uband, testapp.Alice.Address)
	_, err := k.PrepareRequest(ctx, m, testapp.FeePayer.Address, nil)
	require.EqualError(t, err, "OEI action to invoke is not available: bad wasm execution")
}

func TestPrepareRequestWithEmptyRawRequest(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true)
	m := types.NewMsgRequestData(3, BasicCalldata, 1, 1, BasicClientID, testapp.Coins100000000uband, testapp.Alice.Address)
	_, err := k.PrepareRequest(ctx, m, testapp.FeePayer.Address, nil)
	require.EqualError(t, err, "empty raw requests")
}

func TestPrepareRequestUnknownDataSource(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true)
	m := types.NewMsgRequestData(4, obi.MustEncode(testapp.Wasm4Input{
		IDs:      []int64{1, 2, 99},
		Calldata: "beeb",
	}), 1, 1, BasicClientID, testapp.Coins100000000uband, testapp.Alice.Address)
	_, err := k.PrepareRequest(ctx, m, testapp.FeePayer.Address, nil)
	require.EqualError(t, err, "id: 99: data source not found: not enough fee")
}

func TestPrepareRequestInvalidDataSourceCount(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true)
	k.SetParam(ctx, types.KeyMaxRawRequestCount, 3)
	m := types.NewMsgRequestData(4, obi.MustEncode(testapp.Wasm4Input{
		IDs:      []int64{1, 2, 3, 4},
		Calldata: "beeb",
	}), 1, 1, BasicClientID, testapp.Coins100000000uband, testapp.Alice.Address)
	_, err := k.PrepareRequest(ctx, m, testapp.FeePayer.Address, nil)
	// require.EqualError(t, err, "bad wasm execution: too many external data requests")
	m = types.NewMsgRequestData(4, obi.MustEncode(testapp.Wasm4Input{
		IDs:      []int64{1, 2, 3},
		Calldata: "beeb",
	}), 1, 1, BasicClientID, testapp.Coins100000000uband, testapp.Alice.Address)
	id, err := k.PrepareRequest(ctx, m, testapp.FeePayer.Address, nil)
	require.Equal(t, types.RequestID(1), id)
	require.NoError(t, err)
}

func TestPrepareRequestTooMuchWasmGas(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true)
	m := types.NewMsgRequestData(5, BasicCalldata, 1, 1, BasicClientID, testapp.Coins100000000uband, testapp.Alice.Address)
	id, err := k.PrepareRequest(ctx, m, testapp.FeePayer.Address, nil)
	require.Equal(t, types.RequestID(1), id)
	require.NoError(t, err)
	m = types.NewMsgRequestData(6, BasicCalldata, 1, 1, BasicClientID, testapp.Coins100000000uband, testapp.Alice.Address)
	_, err = k.PrepareRequest(ctx, m, testapp.FeePayer.Address, nil)
	require.EqualError(t, err, "out-of-gas while executing the wasm script: bad wasm execution")
}

func TestPrepareRequestTooLargeCalldata(t *testing.T) {
	_, ctx, k := testapp.CreateTestInput(true)
	m := types.NewMsgRequestData(7, BasicCalldata, 1, 1, BasicClientID, testapp.Coins100000000uband, testapp.Alice.Address)
	id, err := k.PrepareRequest(ctx, m, testapp.FeePayer.Address, nil)
	require.Equal(t, types.RequestID(1), id)
	require.NoError(t, err)
	m = types.NewMsgRequestData(8, BasicCalldata, 1, 1, BasicClientID, testapp.Coins100000000uband, testapp.Alice.Address)
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
		}, nil,
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
		sdk.NewAttribute(types.AttributeKeyGasUsed, "260"),
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
		}, nil,
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
		sdk.NewAttribute(types.AttributeKeyGasUsed, "8738"),
	)}, ctx.EventManager().Events())
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
		}, nil,
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
		sdk.NewAttribute(types.AttributeKeyGasUsed, "7757"),
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
		}, nil,
	))
	k.SetReport(ctx, 42, types.NewReport(
		testapp.Validators[0].ValAddress, true, []types.RawReport{
			types.NewRawReport(1, 0, []byte("beeb")),
		},
	))
	k.ResolveRequest(ctx, 42)
	result := types.NewResult(
		BasicClientID, 3, BasicCalldata, 2, 1, 42, 1, testapp.ParseTime(1581589790).Unix(),
		testapp.ParseTime(1581589890).Unix(), types.RESOLVE_STATUS_FAILURE, []byte{},
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
		}, nil,
	))
	k.SetReport(ctx, 42, types.NewReport(
		testapp.Validators[0].ValAddress, true, []types.RawReport{
			types.NewRawReport(1, 0, []byte("beeb")),
		},
	))
	k.ResolveRequest(ctx, 42)
	result := types.NewResult(
		BasicClientID, 6, BasicCalldata, 2, 1, 42, 1, testapp.ParseTime(1581589790).Unix(),
		testapp.ParseTime(1581589890).Unix(), types.RESOLVE_STATUS_FAILURE, []byte{},
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
		}, nil,
	))
	k.ResolveRequest(ctx, 42)

	result := types.NewResult(
		BasicClientID, 9, BasicCalldata, 2, 1, 42, 0, testapp.ParseTime(1581589790).Unix(),
		testapp.ParseTime(1581589890).Unix(), types.RESOLVE_STATUS_FAILURE, []byte{},
	)
	require.Equal(t, result, k.MustGetResult(ctx, 42))

	require.Equal(t, sdk.Events{sdk.NewEvent(
		types.EventTypeResolve,
		sdk.NewAttribute(types.AttributeKeyID, "42"),
		sdk.NewAttribute(types.AttributeKeyResolveStatus, "2"),
		sdk.NewAttribute(types.AttributeKeyReason, "set return data is called more than once"),
	)}, ctx.EventManager().Events())
}
