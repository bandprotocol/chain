package keeper_test

import (
	"strings"

	"go.uber.org/mock/gomock"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/bandprotocol/chain/v3/pkg/obi"
	bandtesting "github.com/bandprotocol/chain/v3/testing"
	"github.com/bandprotocol/chain/v3/testing/testdata"
	"github.com/bandprotocol/chain/v3/x/oracle/keeper"
	oracletestutil "github.com/bandprotocol/chain/v3/x/oracle/testutil"
	"github.com/bandprotocol/chain/v3/x/oracle/types"
)

func (suite *KeeperTestSuite) mockIterateBondedValidatorsByPower() {
	suite.stakingKeeper.EXPECT().
		IterateBondedValidatorsByPower(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx sdk.Context, fn func(index int64, validator stakingtypes.ValidatorI) bool) {
			vals := []stakingtypes.Validator{
				{
					OperatorAddress: validators[0].Address.String(),
					Tokens:          math.NewInt(100000000),
				},
				{
					OperatorAddress: validators[1].Address.String(),
					Tokens:          math.NewInt(1000000),
				},
				{
					OperatorAddress: validators[2].Address.String(),
					Tokens:          math.NewInt(99999999),
				},
			}

			for i, val := range vals {
				stop := fn(int64(i), val)
				if stop {
					break
				}
			}
		}).Return(nil).AnyTimes()
}

func (suite *KeeperTestSuite) TestGetRandomValidatorsSuccessActivateAll() {
	ctx := suite.ctx
	k := suite.oracleKeeper
	require := suite.Require()
	suite.activeAllValidators()
	suite.mockIterateBondedValidatorsByPower()

	// Getting 3 validators using ROLLING_SEED_1_WITH_LONG_ENOUGH_ENTROPY
	suite.rollingseedKeeper.
		EXPECT().
		GetRollingSeed(gomock.Any()).
		Return([]byte("ROLLING_SEED_1_WITH_LONG_ENOUGH_ENTROPY"))
	vals, err := k.GetRandomValidators(ctx, 3, 1)
	require.NoError(err)
	require.Equal(
		[]sdk.ValAddress{
			validators[2].Address,
			validators[0].Address,
			validators[1].Address,
		},
		vals,
	)

	// Getting 3 validators using ROLLING_SEED_A_WITH_LONG_ENOUGH_ENTROPY
	suite.rollingseedKeeper.
		EXPECT().
		GetRollingSeed(gomock.Any()).
		Return([]byte("ROLLING_SEED_A_WITH_LONG_ENOUGH_ENTROPY"))
	vals, err = k.GetRandomValidators(ctx, 3, 1)
	require.NoError(err)
	require.Equal(
		[]sdk.ValAddress{
			validators[0].Address,
			validators[2].Address,
			validators[1].Address,
		},
		vals,
	)

	// Getting 3 validators using ROLLING_SEED_1_WITH_LONG_ENOUGH_ENTROPY again should return the same result as the first one.
	suite.rollingseedKeeper.
		EXPECT().
		GetRollingSeed(gomock.Any()).
		Return([]byte("ROLLING_SEED_1_WITH_LONG_ENOUGH_ENTROPY"))
	vals, err = k.GetRandomValidators(ctx, 3, 1)
	require.NoError(err)
	require.Equal(
		[]sdk.ValAddress{
			validators[2].Address,
			validators[0].Address,
			validators[1].Address,
		},
		vals,
	)

	// Getting 3 validators using ROLLING_SEED_1_WITH_LONG_ENOUGH_ENTROPY but for a different request ID.
	suite.rollingseedKeeper.
		EXPECT().
		GetRollingSeed(gomock.Any()).
		Return([]byte("ROLLING_SEED_1_WITH_LONG_ENOUGH_ENTROPY"))
	vals, err = k.GetRandomValidators(ctx, 3, 42)
	require.NoError(err)
	require.Equal(
		[]sdk.ValAddress{
			validators[0].Address,
			validators[2].Address,
			validators[1].Address,
		},
		vals,
	)
}

func (suite *KeeperTestSuite) TestGetRandomValidatorsTooBigSize() {
	ctx := suite.ctx
	k := suite.oracleKeeper
	require := suite.Require()
	suite.activeAllValidators()
	suite.mockIterateBondedValidatorsByPower()

	suite.rollingseedKeeper.
		EXPECT().
		GetRollingSeed(gomock.Any()).
		Return([]byte("ROLLING_SEED_1_WITH_LONG_ENOUGH_ENTROPY")).
		AnyTimes()

	_, err := k.GetRandomValidators(ctx, 1, 1)
	require.NoError(err)
	_, err = k.GetRandomValidators(ctx, 2, 1)
	require.NoError(err)
	_, err = k.GetRandomValidators(ctx, 3, 1)
	require.NoError(err)
	_, err = k.GetRandomValidators(ctx, 4, 1)
	require.ErrorIs(err, types.ErrInsufficientValidators)
	_, err = k.GetRandomValidators(ctx, 9999, 1)
	require.ErrorIs(err, types.ErrInsufficientValidators)
}

func (suite *KeeperTestSuite) TestGetRandomValidatorsNotEnoughEntropy() {
	ctx := suite.ctx
	k := suite.oracleKeeper
	require := suite.Require()
	suite.activeAllValidators()
	suite.mockIterateBondedValidatorsByPower()

	suite.rollingseedKeeper.
		EXPECT().
		GetRollingSeed(gomock.Any()).
		Return([]byte(""))

	_, err := k.GetRandomValidators(ctx, 3, 1)
	require.ErrorIs(err, types.ErrBadDrbgInitialization)
}

func (suite *KeeperTestSuite) TestPrepareRequestSuccessBasic() {
	suite.activeAllValidators()
	suite.mockIterateBondedValidatorsByPower()
	ctx := suite.ctx
	k := suite.oracleKeeper
	require := suite.Require()

	addSimpleDataSourceAndOracleScript(ctx, k, suite.fileDir)

	ctx = ctx.WithBlockTime(bandtesting.ParseTime(1581589790)).WithBlockHeight(42)
	wrappedGasMeter := bandtesting.NewGasMeterWrapper(ctx.GasMeter())
	ctx = ctx.WithGasMeter(wrappedGasMeter)

	// OracleScript#1: Prepare asks for DS#1,2,3 with ExtID#1,2,3 and calldata "test"
	msg := types.NewMsgRequestData(
		1,
		basicCalldata,
		1,
		1,
		basicClientID,
		bandtesting.Coins100band,
		testDefaultPrepareGas,
		testDefaultExecuteGas,
		alice,
		0,
	)

	// Define expected mock keeper
	suite.rollingseedKeeper.
		EXPECT().
		GetRollingSeed(gomock.Any()).
		Return([]byte("ROLLING_SEED_A_WITH_LONG_ENOUGH_ENTROPY"))

	suite.bankKeeper.EXPECT().SendCoins(gomock.Any(), alice, treasury, bandtesting.Coins1band)
	suite.bankKeeper.EXPECT().SendCoins(gomock.Any(), alice, treasury, bandtesting.Coins1band)
	suite.bankKeeper.EXPECT().SendCoins(gomock.Any(), alice, treasury, bandtesting.Coins1band)

	id, err := k.PrepareRequest(ctx, msg, alice, nil)
	require.NoError(err)
	require.Equal(types.RequestID(1), id)

	require.Equal(types.NewRequest(
		1,
		basicCalldata,
		[]sdk.ValAddress{validators[0].Address},
		1,
		42,
		bandtesting.ParseTime(1581589790),
		basicClientID,
		[]types.RawRequest{
			types.NewRawRequest(1, 1, []byte("test")),
			types.NewRawRequest(2, 2, []byte("test")),
			types.NewRawRequest(3, 3, []byte("test")),
		},
		nil,
		testDefaultExecuteGas,
		0,
		alice.String(),
		sdk.NewCoins(sdk.NewInt64Coin("uband", 97000000)),
	), k.MustGetRequest(ctx, 1))

	// assert gas consumption
	params := k.GetParams(ctx)
	require.Equal(2, wrappedGasMeter.CountRecord(params.BaseOwasmGas, "BASE_OWASM_FEE"))
	require.Equal(1, wrappedGasMeter.CountRecord(testDefaultPrepareGas, "OWASM_PREPARE_FEE"))
	require.Equal(1, wrappedGasMeter.CountRecord(testDefaultExecuteGas, "OWASM_EXECUTE_FEE"))
}

func (suite *KeeperTestSuite) TestPrepareRequestInvalidCalldataSize() {
	ctx := suite.ctx
	k := suite.oracleKeeper
	require := suite.Require()
	suite.activeAllValidators()

	m := types.NewMsgRequestData(
		1,
		[]byte(strings.Repeat("x", 2000)),
		1,
		1,
		basicClientID,
		bandtesting.Coins100band,
		bandtesting.TestDefaultPrepareGas,
		bandtesting.TestDefaultExecuteGas,
		bandtesting.Alice.Address,
		0,
	)
	_, err := k.PrepareRequest(ctx, m, bandtesting.FeePayer.Address, nil)
	require.EqualError(err, "got: 2000, max: 512: too large calldata")
}

func (suite *KeeperTestSuite) TestPrepareRequestOracleScriptNotFound() {
	suite.activeAllValidators()
	suite.mockIterateBondedValidatorsByPower()
	ctx := suite.ctx
	k := suite.oracleKeeper
	require := suite.Require()

	suite.rollingseedKeeper.
		EXPECT().
		GetRollingSeed(gomock.Any()).
		Return([]byte("ROLLING_SEED_1_WITH_LONG_ENOUGH_ENTROPY"))

	m := types.NewMsgRequestData(
		999,
		basicCalldata,
		1,
		1,
		basicClientID,
		bandtesting.Coins100band,
		bandtesting.TestDefaultPrepareGas,
		bandtesting.TestDefaultExecuteGas,
		bandtesting.Alice.Address,
		0,
	)
	_, err := k.PrepareRequest(ctx, m, bandtesting.FeePayer.Address, nil)
	require.EqualError(err, "id: 999: oracle script not found")
}

func (suite *KeeperTestSuite) TestPrepareRequestNotEnoughMaxFee() {
	require := suite.Require()
	suite.mockIterateBondedValidatorsByPower()
	suite.activeAllValidators()
	ctx := suite.ctx
	k := suite.oracleKeeper

	// Define expected mock keeper
	suite.rollingseedKeeper.
		EXPECT().
		GetRollingSeed(gomock.Any()).
		Return([]byte("ROLLING_SEED_1_WITH_LONG_ENOUGH_ENTROPY")).
		AnyTimes()

	suite.bankKeeper.EXPECT().
		SendCoins(gomock.Any(), bandtesting.FeePayer.Address, bandtesting.Treasury.Address, bandtesting.Coins1band).
		Return(nil).AnyTimes()

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
		0,
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
		0,
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
		0,
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
		0,
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
		0,
	)
	id, err := k.PrepareRequest(ctx, m, bandtesting.FeePayer.Address, nil)
	require.NoError(err)
	require.Equal(types.RequestID(1), id)
}

func (suite *KeeperTestSuite) TestPrepareRequestNotEnoughFund() {
	require := suite.Require()
	suite.mockIterateBondedValidatorsByPower()
	suite.activeAllValidators()
	ctx := suite.ctx
	k := suite.oracleKeeper

	// Define expected mock keeper
	suite.rollingseedKeeper.
		EXPECT().
		GetRollingSeed(gomock.Any()).
		Return([]byte("ROLLING_SEED_1_WITH_LONG_ENOUGH_ENTROPY"))

	suite.bankKeeper.EXPECT().
		SendCoins(gomock.Any(), bandtesting.Alice.Address, bandtesting.Treasury.Address, bandtesting.Coins1band).
		Return(errorsmod.Wrapf(
			sdkerrors.ErrInsufficientFunds,
			"spendable balance %s is smaller than %s",
			"0uband", "1band",
		))

	ctx = ctx.WithBlockTime(bandtesting.ParseTime(1581589790)).WithBlockHeight(42)
	// OracleScript#1: Prepare asks for DS#1,2,3 with ExtID#1,2,3 and calldata "test"
	m := types.NewMsgRequestData(
		1,
		basicCalldata,
		1,
		1,
		basicClientID,
		bandtesting.Coins100band,
		bandtesting.TestDefaultPrepareGas,
		bandtesting.TestDefaultExecuteGas,
		bandtesting.Alice.Address,
		0,
	)
	_, err := k.PrepareRequest(ctx, m, bandtesting.Alice.Address, nil)
	require.EqualError(err, "spendable balance 0uband is smaller than 1band: insufficient funds")
}

func (suite *KeeperTestSuite) TestPrepareRequestNotEnoughPrepareGas() {
	require := suite.Require()
	suite.mockIterateBondedValidatorsByPower()
	suite.activeAllValidators()
	ctx := suite.ctx
	k := suite.oracleKeeper

	// Define expected mock keeper
	suite.rollingseedKeeper.
		EXPECT().
		GetRollingSeed(gomock.Any()).
		Return([]byte("ROLLING_SEED_1_WITH_LONG_ENOUGH_ENTROPY"))

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
		0,
	)
	_, err := k.PrepareRequest(ctx, m, bandtesting.FeePayer.Address, nil)
	require.ErrorIs(err, types.ErrBadWasmExecution)
	require.Contains(err.Error(), "out-of-gas")

	params := k.GetParams(ctx)
	require.Equal(1, wrappedGasMeter.CountRecord(params.BaseOwasmGas, "BASE_OWASM_FEE"))
	require.Equal(0, wrappedGasMeter.CountRecord(100, "OWASM_PREPARE_FEE"))
	require.Equal(0, wrappedGasMeter.CountDescriptor("OWASM_EXECUTE_FEE"))
}

func (suite *KeeperTestSuite) TestPrepareRequestInvalidAskCountFail() {
	require := suite.Require()
	suite.mockIterateBondedValidatorsByPower()
	suite.activeAllValidators()
	ctx := suite.ctx
	k := suite.oracleKeeper

	// Define expected mock keeper
	suite.rollingseedKeeper.
		EXPECT().
		GetRollingSeed(gomock.Any()).
		Return([]byte("ROLLING_SEED_1_WITH_LONG_ENOUGH_ENTROPY"))

	suite.bankKeeper.EXPECT().
		SendCoins(gomock.Any(), bandtesting.FeePayer.Address, bandtesting.Treasury.Address, bandtesting.Coins1band).
		Return(nil).AnyTimes()

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
		bandtesting.Coins100band,
		bandtesting.TestDefaultPrepareGas,
		bandtesting.TestDefaultExecuteGas,
		bandtesting.Alice.Address,
		0,
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
		bandtesting.Coins100band,
		bandtesting.TestDefaultPrepareGas,
		bandtesting.TestDefaultExecuteGas,
		bandtesting.Alice.Address,
		0,
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
		bandtesting.Coins100band,
		bandtesting.TestDefaultPrepareGas,
		bandtesting.TestDefaultExecuteGas,
		bandtesting.Alice.Address,
		0,
	)
	id, err := k.PrepareRequest(ctx, m, bandtesting.FeePayer.Address, nil)
	require.Equal(types.RequestID(1), id)
	require.NoError(err)
	require.Equal(2, wrappedGasMeter.CountDescriptor("BASE_OWASM_FEE"))
	require.Equal(1, wrappedGasMeter.CountDescriptor("OWASM_PREPARE_FEE"))
	require.Equal(1, wrappedGasMeter.CountDescriptor("OWASM_EXECUTE_FEE"))
}

func (suite *KeeperTestSuite) TestPrepareRequestBaseOwasmFeePanic() {
	require := suite.Require()
	suite.mockIterateBondedValidatorsByPower()
	suite.activeAllValidators()
	ctx := suite.ctx
	k := suite.oracleKeeper

	// Define expected mock keeper
	suite.rollingseedKeeper.
		EXPECT().
		GetRollingSeed(gomock.Any()).
		Return([]byte("ROLLING_SEED_1_WITH_LONG_ENOUGH_ENTROPY")).
		AnyTimes()

	suite.bankKeeper.EXPECT().
		SendCoins(gomock.Any(), bandtesting.FeePayer.Address, bandtesting.Treasury.Address, bandtesting.Coins1band).
		Return(nil).AnyTimes()

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
		bandtesting.Coins100band,
		bandtesting.TestDefaultPrepareGas,
		bandtesting.TestDefaultExecuteGas,
		bandtesting.Alice.Address,
		0,
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

func (suite *KeeperTestSuite) TestPrepareRequestPerValidatorRequestFeePanic() {
	require := suite.Require()
	suite.mockIterateBondedValidatorsByPower()
	suite.activeAllValidators()
	ctx := suite.ctx
	k := suite.oracleKeeper

	// Define expected mock keeper
	suite.rollingseedKeeper.
		EXPECT().
		GetRollingSeed(gomock.Any()).
		Return([]byte("ROLLING_SEED_1_WITH_LONG_ENOUGH_ENTROPY")).
		AnyTimes()

	suite.bankKeeper.EXPECT().
		SendCoins(gomock.Any(), bandtesting.FeePayer.Address, bandtesting.Treasury.Address, bandtesting.Coins1band).
		Return(nil).AnyTimes()

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
		bandtesting.Coins100band,
		bandtesting.TestDefaultPrepareGas,
		bandtesting.TestDefaultExecuteGas,
		bandtesting.Alice.Address,
		0,
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
		bandtesting.Coins100band,
		bandtesting.TestDefaultPrepareGas,
		bandtesting.TestDefaultExecuteGas,
		bandtesting.Alice.Address,
		0,
	)
	ctx = ctx.WithGasMeter(storetypes.NewGasMeter(1000000))
	id, err := k.PrepareRequest(ctx, m, bandtesting.FeePayer.Address, nil)
	require.Equal(types.RequestID(1), id)
	require.NoError(err)
}

func (suite *KeeperTestSuite) TestPrepareRequestEmptyCalldata() {
	require := suite.Require()
	suite.mockIterateBondedValidatorsByPower()
	suite.activeAllValidators()
	ctx := suite.ctx
	k := suite.oracleKeeper

	// Define expected mock keeper
	suite.rollingseedKeeper.
		EXPECT().
		GetRollingSeed(gomock.Any()).
		Return([]byte("ROLLING_SEED_1_WITH_LONG_ENOUGH_ENTROPY")).
		AnyTimes()

	// Send nil while oracle script expects calldata
	m := types.NewMsgRequestData(
		4,
		nil,
		1,
		1,
		basicClientID,
		bandtesting.Coins100band,
		bandtesting.TestDefaultPrepareGas,
		bandtesting.TestDefaultExecuteGas,
		bandtesting.Alice.Address,
		0,
	)
	_, err := k.PrepareRequest(ctx, m, bandtesting.FeePayer.Address, nil)
	require.EqualError(err, "runtime error while executing the Wasm script: bad wasm execution")
}

func (suite *KeeperTestSuite) TestPrepareRequestBadWasmExecutionFail() {
	require := suite.Require()
	suite.mockIterateBondedValidatorsByPower()
	suite.activeAllValidators()
	ctx := suite.ctx
	k := suite.oracleKeeper

	// Define expected mock keeper
	suite.rollingseedKeeper.
		EXPECT().
		GetRollingSeed(gomock.Any()).
		Return([]byte("ROLLING_SEED_1_WITH_LONG_ENOUGH_ENTROPY")).
		AnyTimes()

	m := types.NewMsgRequestData(
		2,
		basicCalldata,
		1,
		1,
		basicClientID,
		bandtesting.Coins100band,
		bandtesting.TestDefaultPrepareGas,
		bandtesting.TestDefaultExecuteGas,
		bandtesting.Alice.Address,
		0,
	)
	_, err := k.PrepareRequest(ctx, m, bandtesting.FeePayer.Address, nil)
	require.EqualError(err, "OEI action to invoke is not available: bad wasm execution")
}

func (suite *KeeperTestSuite) TestPrepareRequestWithEmptyRawRequest() {
	require := suite.Require()
	suite.mockIterateBondedValidatorsByPower()
	suite.activeAllValidators()
	ctx := suite.ctx
	k := suite.oracleKeeper

	// Define expected mock keeper
	suite.rollingseedKeeper.
		EXPECT().
		GetRollingSeed(gomock.Any()).
		Return([]byte("ROLLING_SEED_1_WITH_LONG_ENOUGH_ENTROPY")).
		AnyTimes()

	m := types.NewMsgRequestData(
		3,
		basicCalldata,
		1,
		1,
		basicClientID,
		bandtesting.Coins100band,
		bandtesting.TestDefaultPrepareGas,
		bandtesting.TestDefaultExecuteGas,
		bandtesting.Alice.Address,
		0,
	)
	_, err := k.PrepareRequest(ctx, m, bandtesting.FeePayer.Address, nil)
	require.EqualError(err, "empty raw requests")
}

func (suite *KeeperTestSuite) TestPrepareRequestUnknownDataSource() {
	require := suite.Require()
	suite.mockIterateBondedValidatorsByPower()
	suite.activeAllValidators()
	ctx := suite.ctx
	k := suite.oracleKeeper

	// Define expected mock keeper
	suite.rollingseedKeeper.
		EXPECT().
		GetRollingSeed(gomock.Any()).
		Return([]byte("ROLLING_SEED_1_WITH_LONG_ENOUGH_ENTROPY")).
		AnyTimes()

	suite.bankKeeper.EXPECT().
		SendCoins(gomock.Any(), bandtesting.FeePayer.Address, bandtesting.Treasury.Address, bandtesting.Coins1band).
		Return(nil).AnyTimes()

	m := types.NewMsgRequestData(4, obi.MustEncode(testdata.Wasm4Input{
		IDs:      []int64{1, 2, 99},
		Calldata: "test",
	}), 1, 1, basicClientID, bandtesting.Coins100band, bandtesting.TestDefaultPrepareGas, bandtesting.TestDefaultExecuteGas, bandtesting.Alice.Address, 0)
	_, err := k.PrepareRequest(ctx, m, bandtesting.FeePayer.Address, nil)
	require.EqualError(err, "id: 99: data source not found")
}

func (suite *KeeperTestSuite) TestPrepareRequestInvalidDataSourceCount() {
	require := suite.Require()
	suite.mockIterateBondedValidatorsByPower()
	suite.activeAllValidators()
	ctx := suite.ctx
	k := suite.oracleKeeper

	// Define expected mock keeper
	suite.rollingseedKeeper.
		EXPECT().
		GetRollingSeed(gomock.Any()).
		Return([]byte("ROLLING_SEED_1_WITH_LONG_ENOUGH_ENTROPY")).
		AnyTimes()

	suite.bankKeeper.EXPECT().
		SendCoins(gomock.Any(), bandtesting.FeePayer.Address, bandtesting.Treasury.Address, bandtesting.Coins1band).
		Return(nil).AnyTimes()

	params := k.GetParams(ctx)
	params.MaxRawRequestCount = 3
	err := k.SetParams(ctx, params)
	require.NoError(err)
	m := types.NewMsgRequestData(4, obi.MustEncode(testdata.Wasm4Input{
		IDs:      []int64{1, 2, 3, 4},
		Calldata: "test",
	}), 1, 1, basicClientID, bandtesting.Coins100band, bandtesting.TestDefaultPrepareGas, bandtesting.TestDefaultExecuteGas, bandtesting.Alice.Address, 0)
	_, err = k.PrepareRequest(ctx, m, bandtesting.FeePayer.Address, nil)
	require.ErrorIs(err, types.ErrBadWasmExecution)
	m = types.NewMsgRequestData(4, obi.MustEncode(testdata.Wasm4Input{
		IDs:      []int64{1, 2, 3},
		Calldata: "test",
	}), 1, 1, basicClientID, bandtesting.Coins100band, bandtesting.TestDefaultPrepareGas, bandtesting.TestDefaultExecuteGas, bandtesting.Alice.Address, 0)
	id, err := k.PrepareRequest(ctx, m, bandtesting.FeePayer.Address, nil)
	require.Equal(types.RequestID(1), id)
	require.NoError(err)
}

func (suite *KeeperTestSuite) TestPrepareRequestTooMuchWasmGas() {
	require := suite.Require()
	suite.mockIterateBondedValidatorsByPower()
	suite.activeAllValidators()
	ctx := suite.ctx
	k := suite.oracleKeeper

	// Define expected mock keeper
	suite.rollingseedKeeper.
		EXPECT().
		GetRollingSeed(gomock.Any()).
		Return([]byte("ROLLING_SEED_1_WITH_LONG_ENOUGH_ENTROPY")).
		AnyTimes()

	suite.bankKeeper.EXPECT().
		SendCoins(gomock.Any(), bandtesting.FeePayer.Address, bandtesting.Treasury.Address, bandtesting.Coins1band).
		Return(nil).AnyTimes()

	m := types.NewMsgRequestData(
		5,
		basicCalldata,
		1,
		1,
		basicClientID,
		bandtesting.Coins100band,
		bandtesting.TestDefaultPrepareGas,
		bandtesting.TestDefaultExecuteGas,
		bandtesting.Alice.Address,
		0,
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
		bandtesting.Coins100band,
		bandtesting.TestDefaultPrepareGas,
		bandtesting.TestDefaultExecuteGas,
		bandtesting.Alice.Address,
		0,
	)
	_, err = k.PrepareRequest(ctx, m, bandtesting.FeePayer.Address, nil)
	require.EqualError(err, "out-of-gas while executing the wasm script: bad wasm execution")
}

func (suite *KeeperTestSuite) TestPrepareRequestTooLargeCalldata() {
	require := suite.Require()
	suite.mockIterateBondedValidatorsByPower()
	suite.activeAllValidators()
	ctx := suite.ctx
	k := suite.oracleKeeper

	// Define expected mock keeper
	suite.rollingseedKeeper.
		EXPECT().
		GetRollingSeed(gomock.Any()).
		Return([]byte("ROLLING_SEED_1_WITH_LONG_ENOUGH_ENTROPY")).
		AnyTimes()

	suite.bankKeeper.EXPECT().
		SendCoins(gomock.Any(), bandtesting.FeePayer.Address, bandtesting.Treasury.Address, bandtesting.Coins1band).
		Return(nil).AnyTimes()

	m := types.NewMsgRequestData(
		7,
		basicCalldata,
		1,
		1,
		basicClientID,
		bandtesting.Coins100band,
		bandtesting.TestDefaultPrepareGas,
		bandtesting.TestDefaultExecuteGas,
		bandtesting.Alice.Address,
		0,
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
		bandtesting.Coins100band,
		bandtesting.TestDefaultPrepareGas,
		bandtesting.TestDefaultExecuteGas,
		bandtesting.Alice.Address,
		0,
	)
	_, err = k.PrepareRequest(ctx, m, bandtesting.FeePayer.Address, nil)
	require.EqualError(err, "span to write is too small: bad wasm execution")
}

func (suite *KeeperTestSuite) TestResolveRequestOutOfGas() {
	require := suite.Require()
	suite.mockIterateBondedValidatorsByPower()
	suite.activeAllValidators()
	ctx := suite.ctx
	k := suite.oracleKeeper

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
		0,
		bandtesting.FeePayer.Address.String(),
		bandtesting.Coins100band,
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

func (suite *KeeperTestSuite) TestResolveReadNilExternalData() {
	require := suite.Require()
	suite.mockIterateBondedValidatorsByPower()
	suite.activeAllValidators()
	ctx := suite.ctx
	k := suite.oracleKeeper

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
		0,
		bandtesting.FeePayer.Address.String(),
		bandtesting.Coins100band,
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

func (suite *KeeperTestSuite) TestResolveRequestNoReturnData() {
	require := suite.Require()
	suite.mockIterateBondedValidatorsByPower()
	suite.activeAllValidators()
	ctx := suite.ctx
	k := suite.oracleKeeper

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
		0,
		bandtesting.FeePayer.Address.String(),
		bandtesting.Coins100band,
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

func (suite *KeeperTestSuite) TestResolveRequestWasmFailure() {
	require := suite.Require()
	suite.mockIterateBondedValidatorsByPower()
	suite.activeAllValidators()
	ctx := suite.ctx
	k := suite.oracleKeeper

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
		0,
		bandtesting.FeePayer.Address.String(),
		bandtesting.Coins100band,
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

func (suite *KeeperTestSuite) TestResolveRequestCallReturnDataSeveralTimes() {
	require := suite.Require()
	suite.mockIterateBondedValidatorsByPower()
	suite.activeAllValidators()
	ctx := suite.ctx
	k := suite.oracleKeeper

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
		0,
		bandtesting.FeePayer.Address.String(),
		bandtesting.Coins100band,
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

func (suite *KeeperTestSuite) TestResolveRequestSuccess() {
	ctx := suite.ctx
	k := suite.oracleKeeper
	require := suite.Require()

	addSimpleDataSourceAndOracleScript(ctx, k, suite.fileDir)

	ctx = ctx.WithBlockTime(bandtesting.ParseTime(1581589890))
	k.SetRequest(ctx, 42, types.NewRequest(
		// 1st Wasm - return "test"
		1,
		basicCalldata,
		[]sdk.ValAddress{validators[0].Address, validators[1].Address},
		1,
		42,
		bandtesting.ParseTime(1581589790),
		basicClientID,
		[]types.RawRequest{
			types.NewRawRequest(1, 1, []byte("test")),
		},
		nil,
		testDefaultExecuteGas,
		0,
		bandtesting.FeePayer.Address.String(),
		bandtesting.Coins100band,
	))
	k.SetReport(ctx, 42, types.NewReport(
		validators[0].Address, true, []types.RawReport{
			types.NewRawReport(1, 0, []byte("test")),
		},
	))
	k.ResolveRequest(ctx, 42)
	expectResult := types.NewResult(
		basicClientID, 1, basicCalldata, 2, 1,
		42, 1, bandtesting.ParseTime(1581589790).Unix(),
		bandtesting.ParseTime(1581589890).Unix(), types.RESOLVE_STATUS_SUCCESS, []byte("test"),
	)
	require.Equal(expectResult, k.MustGetResult(ctx, 42))
}

func (suite *KeeperTestSuite) TestResolveRequestSuccessComplex() {
	ctx := suite.ctx
	k := suite.oracleKeeper
	require := suite.Require()

	addSimpleDataSourceAndOracleScript(ctx, k, suite.fileDir)

	ctx = ctx.WithBlockTime(bandtesting.ParseTime(1581589890))
	k.SetRequest(ctx, 42, types.NewRequest(
		4, // 4th Wasm. Append all reports from all validators.
		obi.MustEncode(testdata.Wasm4Input{
			IDs:      []int64{1, 2},
			Calldata: string(basicCalldata),
		}),
		[]sdk.ValAddress{validators[0].Address, validators[1].Address},
		1,
		42,
		bandtesting.ParseTime(1581589790), basicClientID, []types.RawRequest{
			types.NewRawRequest(0, 1, basicCalldata),
			types.NewRawRequest(1, 2, basicCalldata),
		},
		nil,
		testDefaultExecuteGas,
		0,
		bandtesting.FeePayer.Address.String(),
		bandtesting.Coins100band,
	))
	k.SetReport(ctx, 42, types.NewReport(
		validators[0].Address, true, []types.RawReport{
			types.NewRawReport(0, 0, []byte("testd1v1")),
			types.NewRawReport(1, 0, []byte("testd2v1")),
		},
	))
	k.SetReport(ctx, 42, types.NewReport(
		validators[1].Address, true, []types.RawReport{
			types.NewRawReport(0, 0, []byte("testd1v2")),
			types.NewRawReport(1, 0, []byte("testd2v2")),
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
		obi.MustEncode(testdata.Wasm4Output{Ret: "testd1v1testd1v2testd2v1testd2v2"}),
	)
	require.Equal(result, k.MustGetResult(ctx, 42))
}

func rawRequestsFromFees(ctx sdk.Context, k keeper.Keeper, fees []sdk.Coins) []types.RawRequest {
	var rawRequests []types.RawRequest
	for _, f := range fees {
		id := k.AddDataSource(ctx, types.NewDataSource(
			bandtesting.Owner.Address,
			"mock ds",
			"there is no real code",
			"no file",
			f,
			bandtesting.Treasury.Address,
		))

		rawRequests = append(rawRequests, types.NewRawRequest(
			0, id, nil,
		))
	}

	return rawRequests
}

func (suite *KeeperTestSuite) TestCollectFeeEmptyFee() {
	ctx := suite.ctx
	k := suite.oracleKeeper
	require := suite.Require()

	raws := rawRequestsFromFees(ctx, k, []sdk.Coins{
		bandtesting.EmptyCoins,
		bandtesting.EmptyCoins,
		bandtesting.EmptyCoins,
		bandtesting.EmptyCoins,
		bandtesting.EmptyCoins,
	})

	coins, err := k.CollectFee(ctx, bandtesting.Alice.Address, bandtesting.EmptyCoins, 1, raws)
	require.NoError(err)
	require.Empty(coins)

	coins, err = k.CollectFee(ctx, bandtesting.Alice.Address, bandtesting.Coins100band, 1, raws)
	require.NoError(err)
	require.Empty(coins)

	coins, err = k.CollectFee(ctx, bandtesting.Alice.Address, bandtesting.EmptyCoins, 2, raws)
	require.NoError(err)
	require.Empty(coins)

	coins, err = k.CollectFee(ctx, bandtesting.Alice.Address, bandtesting.Coins100band, 2, raws)
	require.NoError(err)
	require.Empty(coins)
}

func (suite *KeeperTestSuite) TestCollectFeeBasicSuccess() {
	ctx := suite.ctx
	k := suite.oracleKeeper
	require := suite.Require()

	oracletestutil.ChainGoMockCalls(
		suite.bankKeeper.EXPECT().
			SendCoins(gomock.Any(), bandtesting.FeePayer.Address, bandtesting.Treasury.Address, bandtesting.Coins1band).
			Return(nil),
		suite.bankKeeper.EXPECT().
			SendCoins(gomock.Any(), bandtesting.FeePayer.Address, bandtesting.Treasury.Address, sdk.NewCoins(sdk.NewCoin("uband", math.NewInt(2000000)))).
			Return(nil),
	)

	raws := rawRequestsFromFees(ctx, k, []sdk.Coins{
		bandtesting.EmptyCoins,
		bandtesting.Coins1band,
		bandtesting.EmptyCoins,
		sdk.NewCoins(sdk.NewCoin("uband", math.NewInt(2000000))),
		bandtesting.EmptyCoins,
	})

	coins, err := k.CollectFee(ctx, bandtesting.FeePayer.Address, bandtesting.Coins100band, 1, raws)
	require.NoError(err)
	require.Equal(sdk.NewCoins(sdk.NewCoin("uband", math.NewInt(3000000))), coins)
}

func (suite *KeeperTestSuite) TestCollectFeeBasicSuccessWithOtherAskCount() {
	ctx := suite.ctx
	k := suite.oracleKeeper
	require := suite.Require()

	oracletestutil.ChainGoMockCalls(
		suite.bankKeeper.EXPECT().
			SendCoins(gomock.Any(), bandtesting.FeePayer.Address, bandtesting.Treasury.Address, sdk.NewCoins(sdk.NewCoin("uband", math.NewInt(4000000)))).
			Return(nil),
		suite.bankKeeper.EXPECT().
			SendCoins(gomock.Any(), bandtesting.FeePayer.Address, bandtesting.Treasury.Address, sdk.NewCoins(sdk.NewCoin("uband", math.NewInt(8000000)))).
			Return(nil),
	)

	raws := rawRequestsFromFees(ctx, k, []sdk.Coins{
		bandtesting.EmptyCoins,
		bandtesting.Coins1band,
		bandtesting.EmptyCoins,
		sdk.NewCoins(sdk.NewCoin("uband", math.NewInt(2000000))),
		bandtesting.EmptyCoins,
	})

	coins, err := k.CollectFee(ctx, bandtesting.FeePayer.Address, bandtesting.Coins100band, 4, raws)
	require.NoError(err)
	require.Equal(sdk.NewCoins(sdk.NewCoin("uband", math.NewInt(12000000))), coins)
}

func (suite *KeeperTestSuite) TestCollectFeeWithMixedAndFeeNotEnough() {
	ctx := suite.ctx
	k := suite.oracleKeeper
	require := suite.Require()

	suite.bankKeeper.EXPECT().
		SendCoins(gomock.Any(), bandtesting.FeePayer.Address, bandtesting.Treasury.Address, bandtesting.Coins1band).
		Return(nil)

	raws := rawRequestsFromFees(ctx, k, []sdk.Coins{
		bandtesting.EmptyCoins,
		bandtesting.Coins1band,
		bandtesting.EmptyCoins,
		sdk.NewCoins(sdk.NewCoin("uband", math.NewInt(2000000))),
		bandtesting.EmptyCoins,
	})

	coins, err := k.CollectFee(ctx, bandtesting.FeePayer.Address, bandtesting.EmptyCoins, 1, raws)
	require.ErrorIs(err, types.ErrNotEnoughFee)
	require.Nil(coins)

	coins, err = k.CollectFee(ctx, bandtesting.FeePayer.Address, bandtesting.Coins1band, 1, raws)
	require.ErrorIs(err, types.ErrNotEnoughFee)
	require.Nil(coins)
}

func (suite *KeeperTestSuite) TestCollectFeeWithEnoughFeeButInsufficientBalance() {
	ctx := suite.ctx
	k := suite.oracleKeeper
	require := suite.Require()

	oracletestutil.ChainGoMockCalls(
		suite.bankKeeper.EXPECT().
			SendCoins(gomock.Any(), bandtesting.Alice.Address, bandtesting.Treasury.Address, bandtesting.Coins1band).
			Return(nil),
		suite.bankKeeper.EXPECT().
			SendCoins(gomock.Any(), bandtesting.Alice.Address, bandtesting.Treasury.Address, sdk.NewCoins(sdk.NewCoin("uband", math.NewInt(2000000)))).
			Return(errorsmod.Wrapf(
				sdkerrors.ErrInsufficientFunds,
				"spendable balance %s is smaller than %s",
				"0uband", "2band",
			)),
	)

	raws := rawRequestsFromFees(ctx, k, []sdk.Coins{
		bandtesting.EmptyCoins,
		bandtesting.Coins1band,
		bandtesting.EmptyCoins,
		sdk.NewCoins(sdk.NewCoin("uband", math.NewInt(2000000))),
		bandtesting.EmptyCoins,
	})

	coins, err := k.CollectFee(ctx, bandtesting.Alice.Address, bandtesting.Coins100band, 1, raws)
	require.Nil(coins)
	// MAX is 100m but have only 1m in account
	// First ds collect 1m so there no balance enough for next ds but it doesn't touch limit
	require.EqualError(err, "spendable balance 0uband is smaller than 2band: insufficient funds")
}

func (suite *KeeperTestSuite) TestCollectFeeWithWithManyUnitSuccess() {
	ctx := suite.ctx
	k := suite.oracleKeeper
	require := suite.Require()

	oracletestutil.ChainGoMockCalls(
		suite.bankKeeper.EXPECT().
			SendCoins(gomock.Any(), bandtesting.FeePayer.Address, bandtesting.Treasury.Address, bandtesting.Coins1band).
			Return(nil),
		suite.bankKeeper.EXPECT().
			SendCoins(gomock.Any(), bandtesting.FeePayer.Address, bandtesting.Treasury.Address, sdk.NewCoins(sdk.NewCoin("uband", math.NewInt(2000000)), sdk.NewCoin("uabc", math.NewInt(1000000)))).
			Return(nil),
	)

	raws := rawRequestsFromFees(ctx, k, []sdk.Coins{
		bandtesting.EmptyCoins,
		bandtesting.Coins1band,
		bandtesting.EmptyCoins,
		sdk.NewCoins(sdk.NewCoin("uband", math.NewInt(2000000)), sdk.NewCoin("uabc", math.NewInt(1000000))),
		bandtesting.EmptyCoins,
	})

	coins, err := k.CollectFee(
		ctx,
		bandtesting.FeePayer.Address,
		sdk.NewCoins(sdk.NewCoin("uband", math.NewInt(1000000000)), sdk.NewCoin("uabc", math.NewInt(1000000))),
		1,
		raws,
	)
	require.NoError(err)

	// Coins sum is correct
	require.True(
		sdk.NewCoins(sdk.NewCoin("uband", math.NewInt(3000000)), sdk.NewCoin("uabc", math.NewInt(1000000))).
			Equal(coins),
	)
}
