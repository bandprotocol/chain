package keeper_test

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"go.uber.org/mock/gomock"
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

func (suite *KeeperTestSuite) activeAllValidators() {
	ctx := suite.ctx
	k := suite.oracleKeeper

	for _, v := range validators {
		k.Activate(ctx, v.Address)
	}

}

func (suite *KeeperTestSuite) TestGetRandomValidatorsSuccessActivateAll() {
	ctx := suite.ctx
	k := suite.oracleKeeper
	require := suite.Require()

	suite.activeAllValidators()
	suite.mockIterateBondedValidatorsByPower()

	// Getting 3 validators using ROLLING_SEED_1_WITH_LONG_ENOUGH_ENTROPY
	k.SetRollingSeed(ctx, []byte("ROLLING_SEED_1_WITH_LONG_ENOUGH_ENTROPY"))
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
	// Getting 3 validators using ROLLING_SEED_A
	k.SetRollingSeed(ctx, []byte("ROLLING_SEED_A_WITH_LONG_ENOUGH_ENTROPY"))
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
	k.SetRollingSeed(ctx, []byte("ROLLING_SEED_1_WITH_LONG_ENOUGH_ENTROPY"))
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
	k.SetRollingSeed(ctx, []byte("ROLLING_SEED_1_WITH_LONG_ENOUGH_ENTROPY"))
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

// func (suite *KeeperTestSuite) TestGetRandomValidatorsTooBigSize() {
// 	ctx := suite.ctx
// 	k := suite.oracleKeeper
// 	require := suite.Require()

// 	_, err := k.GetRandomValidators(ctx, 1, 1)
// 	require.NoError(t, err)
// 	_, err = k.GetRandomValidators(ctx, 2, 1)
// 	require.NoError(t, err)
// 	_, err = k.GetRandomValidators(ctx, 3, 1)
// 	require.NoError(t, err)
// 	_, err = k.GetRandomValidators(ctx, 4, 1)
// 	require.ErrorIs(t, err, types.ErrInsufficientValidators)
// 	_, err = k.GetRandomValidators(ctx, 9999, 1)
// 	require.ErrorIs(t, err, types.ErrInsufficientValidators)
// }

// func (suite *KeeperTestSuite) TestGetRandomValidatorsWithActivate() {
// 	app, ctx := bandtesting.CreateTestApp(t, false)
// 	k := app.OracleKeeper

// 	k.SetRollingSeed(ctx, []byte("ROLLING_SEED_WITH_LONG_ENOUGH_ENTROPY"))
// 	// If no validators are active, you must not be able to get random validators
// 	_, err := k.GetRandomValidators(ctx, 1, 1)
// 	require.ErrorIs(t, err, types.ErrInsufficientValidators)
// 	// If we activate 2 validators, we should be able to get at most 2 from the function.
// 	err = k.Activate(ctx, bandtesting.Validators[0].ValAddress)
// 	require.NoError(t, err)
// 	err = k.Activate(ctx, bandtesting.Validators[1].ValAddress)
// 	require.NoError(t, err)
// 	vals, err := k.GetRandomValidators(ctx, 1, 1)
// 	require.NoError(t, err)
// 	require.Equal(t, []sdk.ValAddress{bandtesting.Validators[0].ValAddress}, vals)
// 	vals, err = k.GetRandomValidators(ctx, 2, 1)
// 	require.NoError(t, err)
// 	require.Equal(
// 		t,
// 		[]sdk.ValAddress{bandtesting.Validators[0].ValAddress, bandtesting.Validators[1].ValAddress},
// 		vals,
// 	)
// 	_, err = k.GetRandomValidators(ctx, 3, 1)
// 	require.ErrorIs(t, err, types.ErrInsufficientValidators)
// 	// After we deactivate 1 validator due to missing a report, we can only get at most 1 validator.
// 	k.MissReport(ctx, bandtesting.Validators[0].ValAddress, time.Now())
// 	vals, err = k.GetRandomValidators(ctx, 1, 1)
// 	require.NoError(t, err)
// 	require.Equal(t, []sdk.ValAddress{bandtesting.Validators[1].ValAddress}, vals)
// 	_, err = k.GetRandomValidators(ctx, 2, 1)
// 	require.ErrorIs(t, err, types.ErrInsufficientValidators)
// }

// func (suite *KeeperTestSuite) TestPrepareRequestSuccessBasic() {
// 	app, ctx := bandtesting.CreateTestApp(t, true)
// 	k := app.OracleKeeper

// 	ctx = ctx.WithBlockTime(bandtesting.ParseTime(1581589790)).WithBlockHeight(42)

// 	wrappedGasMeter := bandtesting.NewGasMeterWrapper(ctx.GasMeter())
// 	ctx = ctx.WithGasMeter(wrappedGasMeter)

// 	balancesRes, err := app.BankKeeper.AllBalances(
// 		ctx,
// 		authtypes.NewQueryAllBalancesRequest(bandtesting.FeePayer.Address, &query.PageRequest{}, false),
// 	)
// 	require.NoError(t, err)
// 	feePayerBalances := balancesRes.Balances

// 	// OracleScript#1: Prepare asks for DS#1,2,3 with ExtID#1,2,3 and calldata "beeb"
// 	m := types.NewMsgRequestData(
// 		1,
// 		BasicCalldata,
// 		1,
// 		1,
// 		BasicClientID,
// 		bandtesting.Coins100000000uband,
// 		bandtesting.TestDefaultPrepareGas,
// 		bandtesting.TestDefaultExecuteGas,
// 		bandtesting.FeePayer.Address,
// 	)
// 	id, err := k.PrepareRequest(ctx, m, bandtesting.FeePayer.Address, nil)
// 	require.Equal(t, types.RequestID(1), id)
// 	require.NoError(t, err)
// 	require.Equal(t, types.NewRequest(
// 		1, BasicCalldata, []sdk.ValAddress{bandtesting.Validators[0].ValAddress}, 1,
// 		42, bandtesting.ParseTime(1581589790), BasicClientID, []types.RawRequest{
// 			types.NewRawRequest(1, 1, []byte("beeb")),
// 			types.NewRawRequest(2, 2, []byte("beeb")),
// 			types.NewRawRequest(3, 3, []byte("beeb")),
// 		}, nil, bandtesting.TestDefaultExecuteGas,
// 	), k.MustGetRequest(ctx, 1))
// 	require.Equal(t, sdk.Events{
// 		sdk.NewEvent(
// 			authtypes.EventTypeCoinSpent,
// 			sdk.NewAttribute(authtypes.AttributeKeySpender, bandtesting.FeePayer.Address.String()),
// 			sdk.NewAttribute(sdk.AttributeKeyAmount, bandtesting.Coins1000000uband.String()),
// 		), sdk.NewEvent(
// 			authtypes.EventTypeCoinReceived,
// 			sdk.NewAttribute(authtypes.AttributeKeyReceiver, bandtesting.Treasury.Address.String()),
// 			sdk.NewAttribute(sdk.AttributeKeyAmount, bandtesting.Coins1000000uband.String()),
// 		), sdk.NewEvent(
// 			authtypes.EventTypeTransfer,
// 			sdk.NewAttribute(authtypes.AttributeKeyRecipient, bandtesting.Treasury.Address.String()),
// 			sdk.NewAttribute(authtypes.AttributeKeySender, bandtesting.FeePayer.Address.String()),
// 			sdk.NewAttribute(sdk.AttributeKeyAmount, bandtesting.Coins1000000uband.String()),
// 		), sdk.NewEvent(
// 			sdk.EventTypeMessage,
// 			sdk.NewAttribute(sdk.AttributeKeySender, bandtesting.FeePayer.Address.String()),
// 		), sdk.NewEvent(
// 			authtypes.EventTypeCoinSpent,
// 			sdk.NewAttribute(authtypes.AttributeKeySpender, bandtesting.FeePayer.Address.String()),
// 			sdk.NewAttribute(sdk.AttributeKeyAmount, bandtesting.Coins1000000uband.String()),
// 		), sdk.NewEvent(
// 			authtypes.EventTypeCoinReceived,
// 			sdk.NewAttribute(authtypes.AttributeKeyReceiver, bandtesting.Treasury.Address.String()),
// 			sdk.NewAttribute(sdk.AttributeKeyAmount, bandtesting.Coins1000000uband.String()),
// 		), sdk.NewEvent(
// 			authtypes.EventTypeTransfer,
// 			sdk.NewAttribute(authtypes.AttributeKeyRecipient, bandtesting.Treasury.Address.String()),
// 			sdk.NewAttribute(authtypes.AttributeKeySender, bandtesting.FeePayer.Address.String()),
// 			sdk.NewAttribute(sdk.AttributeKeyAmount, bandtesting.Coins1000000uband.String()),
// 		), sdk.NewEvent(
// 			sdk.EventTypeMessage,
// 			sdk.NewAttribute(sdk.AttributeKeySender, bandtesting.FeePayer.Address.String()),
// 		), sdk.NewEvent(
// 			authtypes.EventTypeCoinSpent,
// 			sdk.NewAttribute(authtypes.AttributeKeySpender, bandtesting.FeePayer.Address.String()),
// 			sdk.NewAttribute(sdk.AttributeKeyAmount, bandtesting.Coins1000000uband.String()),
// 		), sdk.NewEvent(
// 			authtypes.EventTypeCoinReceived,
// 			sdk.NewAttribute(authtypes.AttributeKeyReceiver, bandtesting.Treasury.Address.String()),
// 			sdk.NewAttribute(sdk.AttributeKeyAmount, bandtesting.Coins1000000uband.String()),
// 		), sdk.NewEvent(
// 			authtypes.EventTypeTransfer,
// 			sdk.NewAttribute(authtypes.AttributeKeyRecipient, bandtesting.Treasury.Address.String()),
// 			sdk.NewAttribute(authtypes.AttributeKeySender, bandtesting.FeePayer.Address.String()),
// 			sdk.NewAttribute(sdk.AttributeKeyAmount, bandtesting.Coins1000000uband.String()),
// 		), sdk.NewEvent(
// 			sdk.EventTypeMessage,
// 			sdk.NewAttribute(sdk.AttributeKeySender, bandtesting.FeePayer.Address.String()),
// 		), sdk.NewEvent(
// 			types.EventTypeRequest,
// 			sdk.NewAttribute(types.AttributeKeyID, "1"),
// 			sdk.NewAttribute(types.AttributeKeyClientID, BasicClientID),
// 			sdk.NewAttribute(types.AttributeKeyOracleScriptID, "1"),
// 			sdk.NewAttribute(types.AttributeKeyCalldata, hex.EncodeToString(BasicCalldata)),
// 			sdk.NewAttribute(types.AttributeKeyAskCount, "1"),
// 			sdk.NewAttribute(types.AttributeKeyMinCount, "1"),
// 			sdk.NewAttribute(types.AttributeKeyGasUsed, "5294700000"),
// 			sdk.NewAttribute(types.AttributeKeyTotalFees, "3000000uband"),
// 			sdk.NewAttribute(types.AttributeKeyValidator, bandtesting.Validators[0].ValAddress.String()),
// 		), sdk.NewEvent(
// 			types.EventTypeRawRequest,
// 			sdk.NewAttribute(types.AttributeKeyDataSourceID, "1"),
// 			sdk.NewAttribute(types.AttributeKeyDataSourceHash, bandtesting.DataSources[1].Filename),
// 			sdk.NewAttribute(types.AttributeKeyExternalID, "1"),
// 			sdk.NewAttribute(types.AttributeKeyCalldata, "beeb"),
// 			sdk.NewAttribute(types.AttributeKeyFee, "1000000uband"),
// 		), sdk.NewEvent(
// 			types.EventTypeRawRequest,
// 			sdk.NewAttribute(types.AttributeKeyDataSourceID, "2"),
// 			sdk.NewAttribute(types.AttributeKeyDataSourceHash, bandtesting.DataSources[2].Filename),
// 			sdk.NewAttribute(types.AttributeKeyExternalID, "2"),
// 			sdk.NewAttribute(types.AttributeKeyCalldata, "beeb"),
// 			sdk.NewAttribute(types.AttributeKeyFee, "1000000uband"),
// 		), sdk.NewEvent(
// 			types.EventTypeRawRequest,
// 			sdk.NewAttribute(types.AttributeKeyDataSourceID, "3"),
// 			sdk.NewAttribute(types.AttributeKeyDataSourceHash, bandtesting.DataSources[3].Filename),
// 			sdk.NewAttribute(types.AttributeKeyExternalID, "3"),
// 			sdk.NewAttribute(types.AttributeKeyCalldata, "beeb"),
// 			sdk.NewAttribute(types.AttributeKeyFee, "1000000uband"),
// 		),
// 	}, ctx.EventManager().Events())

// 	// assert gas consumption
// 	params := k.GetParams(ctx)
// 	require.Equal(t, 2, wrappedGasMeter.CountRecord(params.BaseOwasmGas, "BASE_OWASM_FEE"))
// 	require.Equal(t, 1, wrappedGasMeter.CountRecord(bandtesting.TestDefaultPrepareGas, "OWASM_PREPARE_FEE"))
// 	require.Equal(t, 1, wrappedGasMeter.CountRecord(bandtesting.TestDefaultExecuteGas, "OWASM_EXECUTE_FEE"))

// 	paid := sdk.NewCoins(sdk.NewInt64Coin("uband", 3000000))
// 	feePayerBalances = feePayerBalances.Sub(paid...)
// 	bandtesting.CheckBalances(t, ctx, app.BankKeeper, bandtesting.FeePayer.Address, feePayerBalances)
// 	bandtesting.CheckBalances(t, ctx, app.BankKeeper, bandtesting.Treasury.Address, paid)
// }

// func (suite *KeeperTestSuite) TestPrepareRequestNotEnoughMaxFee() {
// 	app, ctx := bandtesting.CreateTestApp(t, true)
// 	k := app.OracleKeeper

// 	ctx = ctx.WithBlockTime(bandtesting.ParseTime(1581589790)).WithBlockHeight(42)
// 	// OracleScript#1: Prepare asks for DS#1,2,3 with ExtID#1,2,3 and calldata "beeb"
// 	m := types.NewMsgRequestData(
// 		1,
// 		BasicCalldata,
// 		1,
// 		1,
// 		BasicClientID,
// 		bandtesting.EmptyCoins,
// 		bandtesting.TestDefaultPrepareGas,
// 		bandtesting.TestDefaultExecuteGas,
// 		bandtesting.FeePayer.Address,
// 	)
// 	_, err := k.PrepareRequest(ctx, m, bandtesting.FeePayer.Address, nil)
// 	require.EqualError(t, err, "require: 1000000uband, max: 0uband: not enough fee")
// 	m = types.NewMsgRequestData(
// 		1,
// 		BasicCalldata,
// 		1,
// 		1,
// 		BasicClientID,
// 		sdk.NewCoins(sdk.NewInt64Coin("uband", 1000000)),
// 		bandtesting.TestDefaultPrepareGas,
// 		bandtesting.TestDefaultExecuteGas,
// 		bandtesting.FeePayer.Address,
// 	)
// 	_, err = k.PrepareRequest(ctx, m, bandtesting.FeePayer.Address, nil)
// 	require.EqualError(t, err, "require: 2000000uband, max: 1000000uband: not enough fee")
// 	m = types.NewMsgRequestData(
// 		1,
// 		BasicCalldata,
// 		1,
// 		1,
// 		BasicClientID,
// 		sdk.NewCoins(sdk.NewInt64Coin("uband", 2000000)),
// 		bandtesting.TestDefaultPrepareGas,
// 		bandtesting.TestDefaultExecuteGas,
// 		bandtesting.FeePayer.Address,
// 	)
// 	_, err = k.PrepareRequest(ctx, m, bandtesting.FeePayer.Address, nil)
// 	require.EqualError(t, err, "require: 3000000uband, max: 2000000uband: not enough fee")
// 	m = types.NewMsgRequestData(
// 		1,
// 		BasicCalldata,
// 		1,
// 		1,
// 		BasicClientID,
// 		sdk.NewCoins(sdk.NewInt64Coin("uband", 2999999)),
// 		bandtesting.TestDefaultPrepareGas,
// 		bandtesting.TestDefaultExecuteGas,
// 		bandtesting.FeePayer.Address,
// 	)
// 	_, err = k.PrepareRequest(ctx, m, bandtesting.FeePayer.Address, nil)
// 	require.EqualError(t, err, "require: 3000000uband, max: 2999999uband: not enough fee")
// 	m = types.NewMsgRequestData(
// 		1,
// 		BasicCalldata,
// 		1,
// 		1,
// 		BasicClientID,
// 		sdk.NewCoins(sdk.NewInt64Coin("uband", 3000000)),
// 		bandtesting.TestDefaultPrepareGas,
// 		bandtesting.TestDefaultExecuteGas,
// 		bandtesting.FeePayer.Address,
// 	)
// 	id, err := k.PrepareRequest(ctx, m, bandtesting.FeePayer.Address, nil)
// 	require.NoError(t, err)
// 	require.Equal(t, types.RequestID(1), id)
// }

// func (suite *KeeperTestSuite) TestPrepareRequestNotEnoughFund() {
// 	app, ctx := bandtesting.CreateTestApp(t, true)
// 	k := app.OracleKeeper

// 	ctx = ctx.WithBlockTime(bandtesting.ParseTime(1581589790)).WithBlockHeight(42)
// 	// OracleScript#1: Prepare asks for DS#1,2,3 with ExtID#1,2,3 and calldata "beeb"
// 	m := types.NewMsgRequestData(
// 		1,
// 		BasicCalldata,
// 		1,
// 		1,
// 		BasicClientID,
// 		bandtesting.Coins100000000uband,
// 		bandtesting.TestDefaultPrepareGas,
// 		bandtesting.TestDefaultExecuteGas,
// 		bandtesting.Alice.Address,
// 	)
// 	_, err := k.PrepareRequest(ctx, m, bandtesting.Alice.Address, nil)
// 	require.EqualError(t, err, "spendable balance  is smaller than 1000000uband: insufficient funds")
// }

// func (suite *KeeperTestSuite) TestPrepareRequestInvalidCalldataSize() {
// 	app, ctx := bandtesting.CreateTestApp(t, true)
// 	k := app.OracleKeeper

// 	m := types.NewMsgRequestData(
// 		1,
// 		[]byte(strings.Repeat("x", 2000)),
// 		1,
// 		1,
// 		BasicClientID,
// 		bandtesting.Coins100000000uband,
// 		bandtesting.TestDefaultPrepareGas,
// 		bandtesting.TestDefaultExecuteGas,
// 		bandtesting.Alice.Address,
// 	)
// 	_, err := k.PrepareRequest(ctx, m, bandtesting.FeePayer.Address, nil)
// 	require.EqualError(t, err, "got: 2000, max: 512: too large calldata")
// }

// func (suite *KeeperTestSuite) TestPrepareRequestNotEnoughPrepareGas() {
// 	app, ctx := bandtesting.CreateTestApp(t, true)
// 	k := app.OracleKeeper

// 	ctx = ctx.WithBlockTime(bandtesting.ParseTime(1581589790)).WithBlockHeight(42)

// 	wrappedGasMeter := bandtesting.NewGasMeterWrapper(ctx.GasMeter())
// 	ctx = ctx.WithGasMeter(wrappedGasMeter)

// 	m := types.NewMsgRequestData(
// 		1,
// 		BasicCalldata,
// 		1,
// 		1,
// 		BasicClientID,
// 		bandtesting.EmptyCoins,
// 		1,
// 		bandtesting.TestDefaultExecuteGas,
// 		bandtesting.Alice.Address,
// 	)
// 	_, err := k.PrepareRequest(ctx, m, bandtesting.FeePayer.Address, nil)
// 	require.ErrorIs(t, err, types.ErrBadWasmExecution)
// 	require.Contains(t, err.Error(), "out-of-gas")

// 	params := k.GetParams(ctx)
// 	require.Equal(t, 1, wrappedGasMeter.CountRecord(params.BaseOwasmGas, "BASE_OWASM_FEE"))
// 	require.Equal(t, 0, wrappedGasMeter.CountRecord(100, "OWASM_PREPARE_FEE"))
// 	require.Equal(t, 0, wrappedGasMeter.CountDescriptor("OWASM_EXECUTE_FEE"))
// }

// func (suite *KeeperTestSuite) TestPrepareRequestInvalidAskCountFail() {
// 	app, ctx := bandtesting.CreateTestApp(t, true)
// 	k := app.OracleKeeper

// 	params := k.GetParams(ctx)
// 	params.MaxAskCount = 5
// 	err := k.SetParams(ctx, params)
// 	require.NoError(t, err)

// 	wrappedGasMeter := bandtesting.NewGasMeterWrapper(ctx.GasMeter())
// 	ctx = ctx.WithGasMeter(wrappedGasMeter)

// 	m := types.NewMsgRequestData(
// 		1,
// 		BasicCalldata,
// 		10,
// 		1,
// 		BasicClientID,
// 		bandtesting.Coins100000000uband,
// 		bandtesting.TestDefaultPrepareGas,
// 		bandtesting.TestDefaultExecuteGas,
// 		bandtesting.Alice.Address,
// 	)
// 	_, err = k.PrepareRequest(ctx, m, bandtesting.FeePayer.Address, nil)
// 	require.ErrorIs(t, err, types.ErrInvalidAskCount)

// 	require.Equal(t, 0, wrappedGasMeter.CountDescriptor("BASE_OWASM_FEE"))
// 	require.Equal(t, 0, wrappedGasMeter.CountDescriptor("OWASM_PREPARE_FEE"))
// 	require.Equal(t, 0, wrappedGasMeter.CountDescriptor("OWASM_EXECUTE_FEE"))

// 	m = types.NewMsgRequestData(
// 		1,
// 		BasicCalldata,
// 		4,
// 		1,
// 		BasicClientID,
// 		bandtesting.Coins100000000uband,
// 		bandtesting.TestDefaultPrepareGas,
// 		bandtesting.TestDefaultExecuteGas,
// 		bandtesting.Alice.Address,
// 	)
// 	_, err = k.PrepareRequest(ctx, m, bandtesting.FeePayer.Address, nil)
// 	require.ErrorIs(t, err, types.ErrInsufficientValidators)

// 	require.Equal(t, 0, wrappedGasMeter.CountDescriptor("BASE_OWASM_FEE"))
// 	require.Equal(t, 0, wrappedGasMeter.CountDescriptor("OWASM_PREPARE_FEE"))
// 	require.Equal(t, 0, wrappedGasMeter.CountDescriptor("OWASM_EXECUTE_FEE"))

// 	m = types.NewMsgRequestData(
// 		1,
// 		BasicCalldata,
// 		1,
// 		1,
// 		BasicClientID,
// 		bandtesting.Coins100000000uband,
// 		bandtesting.TestDefaultPrepareGas,
// 		bandtesting.TestDefaultExecuteGas,
// 		bandtesting.Alice.Address,
// 	)
// 	id, err := k.PrepareRequest(ctx, m, bandtesting.FeePayer.Address, nil)
// 	require.Equal(t, types.RequestID(1), id)
// 	require.NoError(t, err)
// 	require.Equal(t, 2, wrappedGasMeter.CountDescriptor("BASE_OWASM_FEE"))
// 	require.Equal(t, 1, wrappedGasMeter.CountDescriptor("OWASM_PREPARE_FEE"))
// 	require.Equal(t, 1, wrappedGasMeter.CountDescriptor("OWASM_EXECUTE_FEE"))
// }

// func (suite *KeeperTestSuite) TestPrepareRequestBaseOwasmFeePanic() {
// 	app, ctx := bandtesting.CreateTestApp(t, true)
// 	k := app.OracleKeeper

// 	params := k.GetParams(ctx)
// 	params.BaseOwasmGas = 100000
// 	params.PerValidatorRequestGas = 0
// 	err := k.SetParams(ctx, params)
// 	require.NoError(t, err)
// 	m := types.NewMsgRequestData(
// 		1,
// 		BasicCalldata,
// 		1,
// 		1,
// 		BasicClientID,
// 		bandtesting.Coins100000000uband,
// 		bandtesting.TestDefaultPrepareGas,
// 		bandtesting.TestDefaultExecuteGas,
// 		bandtesting.Alice.Address,
// 	)
// 	ctx = ctx.WithGasMeter(storetypes.NewGasMeter(90000))
// 	require.PanicsWithValue(
// 		t,
// 		storetypes.ErrorOutOfGas{Descriptor: "BASE_OWASM_FEE"},
// 		func() { _, _ = k.PrepareRequest(ctx, m, bandtesting.FeePayer.Address, nil) },
// 	)
// 	ctx = ctx.WithGasMeter(storetypes.NewGasMeter(1000000))
// 	id, err := k.PrepareRequest(ctx, m, bandtesting.FeePayer.Address, nil)
// 	require.Equal(t, types.RequestID(1), id)
// 	require.NoError(t, err)
// }

// func (suite *KeeperTestSuite) TestPrepareRequestPerValidatorRequestFeePanic() {
// 	app, ctx := bandtesting.CreateTestApp(t, true)
// 	k := app.OracleKeeper

// 	params := k.GetParams(ctx)
// 	params.BaseOwasmGas = 100000
// 	params.PerValidatorRequestGas = 50000
// 	err := k.SetParams(ctx, params)
// 	require.NoError(t, err)
// 	m := types.NewMsgRequestData(
// 		1,
// 		BasicCalldata,
// 		2,
// 		1,
// 		BasicClientID,
// 		bandtesting.Coins100000000uband,
// 		bandtesting.TestDefaultPrepareGas,
// 		bandtesting.TestDefaultExecuteGas,
// 		bandtesting.Alice.Address,
// 	)
// 	ctx = ctx.WithGasMeter(storetypes.NewGasMeter(90000))
// 	require.PanicsWithValue(
// 		t,
// 		storetypes.ErrorOutOfGas{Descriptor: "PER_VALIDATOR_REQUEST_FEE"},
// 		func() { _, _ = k.PrepareRequest(ctx, m, bandtesting.FeePayer.Address, nil) },
// 	)
// 	m = types.NewMsgRequestData(
// 		1,
// 		BasicCalldata,
// 		1,
// 		1,
// 		BasicClientID,
// 		bandtesting.Coins100000000uband,
// 		bandtesting.TestDefaultPrepareGas,
// 		bandtesting.TestDefaultExecuteGas,
// 		bandtesting.Alice.Address,
// 	)
// 	ctx = ctx.WithGasMeter(storetypes.NewGasMeter(1000000))
// 	id, err := k.PrepareRequest(ctx, m, bandtesting.FeePayer.Address, nil)
// 	require.Equal(t, types.RequestID(1), id)
// 	require.NoError(t, err)
// }

// func (suite *KeeperTestSuite) TestPrepareRequestEmptyCalldata() {
// 	app, ctx := bandtesting.CreateTestApp(t, true)
// 	k := app.OracleKeeper
// 	// Send nil while oracle script expects calldata
// 	m := types.NewMsgRequestData(
// 		4,
// 		nil,
// 		1,
// 		1,
// 		BasicClientID,
// 		bandtesting.Coins100000000uband,
// 		bandtesting.TestDefaultPrepareGas,
// 		bandtesting.TestDefaultExecuteGas,
// 		bandtesting.Alice.Address,
// 	)
// 	_, err := k.PrepareRequest(ctx, m, bandtesting.FeePayer.Address, nil)
// 	require.EqualError(t, err, "runtime error while executing the Wasm script: bad wasm execution")
// }

// func (suite *KeeperTestSuite) TestPrepareRequestOracleScriptNotFound() {
// 	app, ctx := bandtesting.CreateTestApp(t, true)
// 	k := app.OracleKeeper

// 	m := types.NewMsgRequestData(
// 		999,
// 		BasicCalldata,
// 		1,
// 		1,
// 		BasicClientID,
// 		bandtesting.Coins100000000uband,
// 		bandtesting.TestDefaultPrepareGas,
// 		bandtesting.TestDefaultExecuteGas,
// 		bandtesting.Alice.Address,
// 	)
// 	_, err := k.PrepareRequest(ctx, m, bandtesting.FeePayer.Address, nil)
// 	require.EqualError(t, err, "id: 999: oracle script not found")
// }

// func (suite *KeeperTestSuite) TestPrepareRequestBadWasmExecutionFail() {
// 	app, ctx := bandtesting.CreateTestApp(t, true)
// 	k := app.OracleKeeper

// 	m := types.NewMsgRequestData(
// 		2,
// 		BasicCalldata,
// 		1,
// 		1,
// 		BasicClientID,
// 		bandtesting.Coins100000000uband,
// 		bandtesting.TestDefaultPrepareGas,
// 		bandtesting.TestDefaultExecuteGas,
// 		bandtesting.Alice.Address,
// 	)
// 	_, err := k.PrepareRequest(ctx, m, bandtesting.FeePayer.Address, nil)
// 	require.EqualError(t, err, "OEI action to invoke is not available: bad wasm execution")
// }

// func (suite *KeeperTestSuite) TestPrepareRequestWithEmptyRawRequest() {
// 	app, ctx := bandtesting.CreateTestApp(t, true)
// 	k := app.OracleKeeper

// 	m := types.NewMsgRequestData(
// 		3,
// 		BasicCalldata,
// 		1,
// 		1,
// 		BasicClientID,
// 		bandtesting.Coins100000000uband,
// 		bandtesting.TestDefaultPrepareGas,
// 		bandtesting.TestDefaultExecuteGas,
// 		bandtesting.Alice.Address,
// 	)
// 	_, err := k.PrepareRequest(ctx, m, bandtesting.FeePayer.Address, nil)
// 	require.EqualError(t, err, "empty raw requests")
// }

// func (suite *KeeperTestSuite) TestPrepareRequestUnknownDataSource() {
// 	app, ctx := bandtesting.CreateTestApp(t, true)
// 	k := app.OracleKeeper

// 	m := types.NewMsgRequestData(4, obi.MustEncode(testdata.Wasm4Input{
// 		IDs:      []int64{1, 2, 99},
// 		Calldata: "beeb",
// 	}), 1, 1, BasicClientID, bandtesting.Coins100000000uband, bandtesting.TestDefaultPrepareGas, bandtesting.TestDefaultExecuteGas, bandtesting.Alice.Address)
// 	_, err := k.PrepareRequest(ctx, m, bandtesting.FeePayer.Address, nil)
// 	require.EqualError(t, err, "id: 99: data source not found")
// }

// func (suite *KeeperTestSuite) TestPrepareRequestInvalidDataSourceCount() {
// 	app, ctx := bandtesting.CreateTestApp(t, true)
// 	k := app.OracleKeeper

// 	params := k.GetParams(ctx)
// 	params.MaxRawRequestCount = 3
// 	err := k.SetParams(ctx, params)
// 	require.NoError(t, err)
// 	m := types.NewMsgRequestData(4, obi.MustEncode(testdata.Wasm4Input{
// 		IDs:      []int64{1, 2, 3, 4},
// 		Calldata: "beeb",
// 	}), 1, 1, BasicClientID, bandtesting.Coins100000000uband, bandtesting.TestDefaultPrepareGas, bandtesting.TestDefaultExecuteGas, bandtesting.Alice.Address)
// 	_, err = k.PrepareRequest(ctx, m, bandtesting.FeePayer.Address, nil)
// 	require.ErrorIs(t, err, types.ErrBadWasmExecution)
// 	m = types.NewMsgRequestData(4, obi.MustEncode(testdata.Wasm4Input{
// 		IDs:      []int64{1, 2, 3},
// 		Calldata: "beeb",
// 	}), 1, 1, BasicClientID, bandtesting.Coins100000000uband, bandtesting.TestDefaultPrepareGas, bandtesting.TestDefaultExecuteGas, bandtesting.Alice.Address)
// 	id, err := k.PrepareRequest(ctx, m, bandtesting.FeePayer.Address, nil)
// 	require.Equal(t, types.RequestID(1), id)
// 	require.NoError(t, err)
// }

// func (suite *KeeperTestSuite) TestPrepareRequestTooMuchWasmGas() {
// 	app, ctx := bandtesting.CreateTestApp(t, true)
// 	k := app.OracleKeeper

// 	m := types.NewMsgRequestData(
// 		5,
// 		BasicCalldata,
// 		1,
// 		1,
// 		BasicClientID,
// 		bandtesting.Coins100000000uband,
// 		bandtesting.TestDefaultPrepareGas,
// 		bandtesting.TestDefaultExecuteGas,
// 		bandtesting.Alice.Address,
// 	)
// 	id, err := k.PrepareRequest(ctx, m, bandtesting.FeePayer.Address, nil)
// 	require.Equal(t, types.RequestID(1), id)
// 	require.NoError(t, err)
// 	m = types.NewMsgRequestData(
// 		6,
// 		BasicCalldata,
// 		1,
// 		1,
// 		BasicClientID,
// 		bandtesting.Coins100000000uband,
// 		bandtesting.TestDefaultPrepareGas,
// 		bandtesting.TestDefaultExecuteGas,
// 		bandtesting.Alice.Address,
// 	)
// 	_, err = k.PrepareRequest(ctx, m, bandtesting.FeePayer.Address, nil)
// 	require.EqualError(t, err, "out-of-gas while executing the wasm script: bad wasm execution")
// }

// func (suite *KeeperTestSuite) TestPrepareRequestTooLargeCalldata() {
// 	app, ctx := bandtesting.CreateTestApp(t, true)
// 	k := app.OracleKeeper

// 	m := types.NewMsgRequestData(
// 		7,
// 		BasicCalldata,
// 		1,
// 		1,
// 		BasicClientID,
// 		bandtesting.Coins100000000uband,
// 		bandtesting.TestDefaultPrepareGas,
// 		bandtesting.TestDefaultExecuteGas,
// 		bandtesting.Alice.Address,
// 	)
// 	id, err := k.PrepareRequest(ctx, m, bandtesting.FeePayer.Address, nil)
// 	require.Equal(t, types.RequestID(1), id)
// 	require.NoError(t, err)
// 	m = types.NewMsgRequestData(
// 		8,
// 		BasicCalldata,
// 		1,
// 		1,
// 		BasicClientID,
// 		bandtesting.Coins100000000uband,
// 		bandtesting.TestDefaultPrepareGas,
// 		bandtesting.TestDefaultExecuteGas,
// 		bandtesting.Alice.Address,
// 	)
// 	_, err = k.PrepareRequest(ctx, m, bandtesting.FeePayer.Address, nil)
// 	require.EqualError(t, err, "span to write is too small: bad wasm execution")
// }

// func (suite *KeeperTestSuite) TestResolveRequestSuccess() {
// 	app, ctx := bandtesting.CreateTestApp(t, true)
// 	k := app.OracleKeeper

// 	ctx = ctx.WithBlockTime(bandtesting.ParseTime(1581589890))
// 	k.SetRequest(ctx, 42, types.NewRequest(
// 		// 1st Wasm - return "beeb"
// 		1,
// 		BasicCalldata,
// 		[]sdk.ValAddress{bandtesting.Validators[0].ValAddress, bandtesting.Validators[1].ValAddress},
// 		1,
// 		42,
// 		bandtesting.ParseTime(1581589790),
// 		BasicClientID,
// 		[]types.RawRequest{
// 			types.NewRawRequest(1, 1, []byte("beeb")),
// 		},
// 		nil,
// 		bandtesting.TestDefaultExecuteGas,
// 	))
// 	k.SetReport(ctx, 42, types.NewReport(
// 		bandtesting.Validators[0].ValAddress, true, []types.RawReport{
// 			types.NewRawReport(1, 0, []byte("beeb")),
// 		},
// 	))
// 	k.ResolveRequest(ctx, 42)
// 	expectResult := types.NewResult(
// 		BasicClientID, 1, BasicCalldata, 2, 1,
// 		42, 1, bandtesting.ParseTime(1581589790).Unix(),
// 		bandtesting.ParseTime(1581589890).Unix(), types.RESOLVE_STATUS_SUCCESS, []byte("beeb"),
// 	)
// 	require.Equal(t, expectResult, k.MustGetResult(ctx, 42))
// 	require.Equal(t, sdk.Events{sdk.NewEvent(
// 		types.EventTypeResolve,
// 		sdk.NewAttribute(types.AttributeKeyID, "42"),
// 		sdk.NewAttribute(types.AttributeKeyResolveStatus, "1"),
// 		sdk.NewAttribute(types.AttributeKeyResult, "62656562"), // hex of "beeb"
// 		sdk.NewAttribute(types.AttributeKeyGasUsed, "2485000000"),
// 	)}, ctx.EventManager().Events())
// }

// func (suite *KeeperTestSuite) TestResolveRequestSuccessComplex() {
// 	app, ctx := bandtesting.CreateTestApp(t, true)
// 	k := app.OracleKeeper

// 	ctx = ctx.WithBlockTime(bandtesting.ParseTime(1581589890))
// 	k.SetRequest(ctx, 42, types.NewRequest(
// 		// 4th Wasm. Append all reports from all validators.
// 		4, obi.MustEncode(testdata.Wasm4Input{
// 			IDs:      []int64{1, 2},
// 			Calldata: string(BasicCalldata),
// 		}), []sdk.ValAddress{bandtesting.Validators[0].ValAddress, bandtesting.Validators[1].ValAddress}, 1,
// 		42, bandtesting.ParseTime(1581589790), BasicClientID, []types.RawRequest{
// 			types.NewRawRequest(0, 1, BasicCalldata),
// 			types.NewRawRequest(1, 2, BasicCalldata),
// 		}, nil, bandtesting.TestDefaultExecuteGas,
// 	))
// 	k.SetReport(ctx, 42, types.NewReport(
// 		bandtesting.Validators[0].ValAddress, true, []types.RawReport{
// 			types.NewRawReport(0, 0, []byte("beebd1v1")),
// 			types.NewRawReport(1, 0, []byte("beebd2v1")),
// 		},
// 	))
// 	k.SetReport(ctx, 42, types.NewReport(
// 		bandtesting.Validators[1].ValAddress, true, []types.RawReport{
// 			types.NewRawReport(0, 0, []byte("beebd1v2")),
// 			types.NewRawReport(1, 0, []byte("beebd2v2")),
// 		},
// 	))
// 	k.ResolveRequest(ctx, 42)
// 	result := types.NewResult(
// 		BasicClientID, 4, obi.MustEncode(testdata.Wasm4Input{
// 			IDs:      []int64{1, 2},
// 			Calldata: string(BasicCalldata),
// 		}), 2, 1,
// 		42, 2, bandtesting.ParseTime(1581589790).Unix(),
// 		bandtesting.ParseTime(1581589890).Unix(), types.RESOLVE_STATUS_SUCCESS,
// 		obi.MustEncode(testdata.Wasm4Output{Ret: "beebd1v1beebd1v2beebd2v1beebd2v2"}),
// 	)
// 	require.Equal(t, result, k.MustGetResult(ctx, 42))
// 	require.Equal(t, sdk.Events{sdk.NewEvent(
// 		types.EventTypeResolve,
// 		sdk.NewAttribute(types.AttributeKeyID, "42"),
// 		sdk.NewAttribute(types.AttributeKeyResolveStatus, "1"),
// 		sdk.NewAttribute(
// 			types.AttributeKeyResult,
// 			"000000206265656264317631626565626431763262656562643276316265656264327632",
// 		),
// 		sdk.NewAttribute(types.AttributeKeyGasUsed, "32492250000"),
// 	)}, ctx.EventManager().Events())
// }

// func (suite *KeeperTestSuite) TestResolveRequestOutOfGas() {
// 	app, ctx := bandtesting.CreateTestApp(t, true)
// 	k := app.OracleKeeper

// 	ctx = ctx.WithBlockTime(bandtesting.ParseTime(1581589890))
// 	k.SetRequest(ctx, 42, types.NewRequest(
// 		// 1st Wasm - return "beeb"
// 		1,
// 		BasicCalldata,
// 		[]sdk.ValAddress{bandtesting.Validators[0].ValAddress, bandtesting.Validators[1].ValAddress},
// 		1,
// 		42,
// 		bandtesting.ParseTime(1581589790),
// 		BasicClientID,
// 		[]types.RawRequest{
// 			types.NewRawRequest(1, 1, []byte("beeb")),
// 		},
// 		nil,
// 		0,
// 	))
// 	k.SetReport(ctx, 42, types.NewReport(
// 		bandtesting.Validators[0].ValAddress, true, []types.RawReport{
// 			types.NewRawReport(1, 0, []byte("beeb")),
// 		},
// 	))
// 	k.ResolveRequest(ctx, 42)
// 	result := types.NewResult(
// 		BasicClientID, 1, BasicCalldata, 2, 1,
// 		42, 1, bandtesting.ParseTime(1581589790).Unix(),
// 		bandtesting.ParseTime(1581589890).Unix(), types.RESOLVE_STATUS_FAILURE, nil,
// 	)
// 	require.Equal(t, result, k.MustGetResult(ctx, 42))
// }

// func (suite *KeeperTestSuite) TestResolveReadNilExternalData() {
// 	app, ctx := bandtesting.CreateTestApp(t, true)
// 	k := app.OracleKeeper

// 	ctx = ctx.WithBlockTime(bandtesting.ParseTime(1581589890))
// 	k.SetRequest(ctx, 42, types.NewRequest(
// 		// 4th Wasm. Append all reports from all validators.
// 		4, obi.MustEncode(testdata.Wasm4Input{
// 			IDs:      []int64{1, 2},
// 			Calldata: string(BasicCalldata),
// 		}), []sdk.ValAddress{bandtesting.Validators[0].ValAddress, bandtesting.Validators[1].ValAddress}, 1,
// 		42, bandtesting.ParseTime(1581589790), BasicClientID, []types.RawRequest{
// 			types.NewRawRequest(0, 1, BasicCalldata),
// 			types.NewRawRequest(1, 2, BasicCalldata),
// 		}, nil, bandtesting.TestDefaultExecuteGas,
// 	))
// 	k.SetReport(ctx, 42, types.NewReport(
// 		bandtesting.Validators[0].ValAddress, true, []types.RawReport{
// 			types.NewRawReport(0, 0, nil),
// 			types.NewRawReport(1, 0, []byte("beebd2v1")),
// 		},
// 	))
// 	k.SetReport(ctx, 42, types.NewReport(
// 		bandtesting.Validators[1].ValAddress, true, []types.RawReport{
// 			types.NewRawReport(0, 0, []byte("beebd1v2")),
// 			types.NewRawReport(1, 0, nil),
// 		},
// 	))
// 	k.ResolveRequest(ctx, 42)
// 	result := types.NewResult(
// 		BasicClientID, 4, obi.MustEncode(testdata.Wasm4Input{
// 			IDs:      []int64{1, 2},
// 			Calldata: string(BasicCalldata),
// 		}), 2, 1,
// 		42, 2, bandtesting.ParseTime(1581589790).Unix(),
// 		bandtesting.ParseTime(1581589890).Unix(), types.RESOLVE_STATUS_SUCCESS,
// 		obi.MustEncode(testdata.Wasm4Output{Ret: "beebd1v2beebd2v1"}),
// 	)
// 	require.Equal(t, result, k.MustGetResult(ctx, 42))
// 	require.Equal(t, sdk.Events{sdk.NewEvent(
// 		types.EventTypeResolve,
// 		sdk.NewAttribute(types.AttributeKeyID, "42"),
// 		sdk.NewAttribute(types.AttributeKeyResolveStatus, "1"),
// 		sdk.NewAttribute(types.AttributeKeyResult, "0000001062656562643176326265656264327631"),
// 		sdk.NewAttribute(types.AttributeKeyGasUsed, "31168050000"),
// 	)}, ctx.EventManager().Events())
// }

// func (suite *KeeperTestSuite) TestResolveRequestNoReturnData() {
// 	app, ctx := bandtesting.CreateTestApp(t, true)
// 	k := app.OracleKeeper

// 	ctx = ctx.WithBlockTime(bandtesting.ParseTime(1581589890))
// 	k.SetRequest(ctx, 42, types.NewRequest(
// 		// 3rd Wasm - do nothing
// 		3,
// 		BasicCalldata,
// 		[]sdk.ValAddress{bandtesting.Validators[0].ValAddress, bandtesting.Validators[1].ValAddress},
// 		1,
// 		42,
// 		bandtesting.ParseTime(1581589790),
// 		BasicClientID,
// 		[]types.RawRequest{
// 			types.NewRawRequest(1, 1, []byte("beeb")),
// 		},
// 		nil,
// 		1,
// 	))
// 	k.SetReport(ctx, 42, types.NewReport(
// 		bandtesting.Validators[0].ValAddress, true, []types.RawReport{
// 			types.NewRawReport(1, 0, []byte("beeb")),
// 		},
// 	))
// 	k.ResolveRequest(ctx, 42)
// 	result := types.NewResult(
// 		BasicClientID, 3, BasicCalldata, 2, 1, 42, 1, bandtesting.ParseTime(1581589790).Unix(),
// 		bandtesting.ParseTime(1581589890).Unix(), types.RESOLVE_STATUS_FAILURE, nil,
// 	)
// 	require.Equal(t, result, k.MustGetResult(ctx, 42))
// 	require.Equal(t, sdk.Events{sdk.NewEvent(
// 		types.EventTypeResolve,
// 		sdk.NewAttribute(types.AttributeKeyID, "42"),
// 		sdk.NewAttribute(types.AttributeKeyResolveStatus, "2"),
// 		sdk.NewAttribute(types.AttributeKeyReason, "no return data"),
// 	)}, ctx.EventManager().Events())
// }

// func (suite *KeeperTestSuite) TestResolveRequestWasmFailure() {
// 	app, ctx := bandtesting.CreateTestApp(t, true)
// 	k := app.OracleKeeper

// 	ctx = ctx.WithBlockTime(bandtesting.ParseTime(1581589890))
// 	k.SetRequest(ctx, 42, types.NewRequest(
// 		// 6th Wasm - out-of-gas
// 		6,
// 		BasicCalldata,
// 		[]sdk.ValAddress{bandtesting.Validators[0].ValAddress, bandtesting.Validators[1].ValAddress},
// 		1,
// 		42,
// 		bandtesting.ParseTime(1581589790),
// 		BasicClientID,
// 		[]types.RawRequest{
// 			types.NewRawRequest(1, 1, []byte("beeb")),
// 		},
// 		nil,
// 		0,
// 	))
// 	k.SetReport(ctx, 42, types.NewReport(
// 		bandtesting.Validators[0].ValAddress, true, []types.RawReport{
// 			types.NewRawReport(1, 0, []byte("beeb")),
// 		},
// 	))
// 	k.ResolveRequest(ctx, 42)
// 	result := types.NewResult(
// 		BasicClientID, 6, BasicCalldata, 2, 1, 42, 1, bandtesting.ParseTime(1581589790).Unix(),
// 		bandtesting.ParseTime(1581589890).Unix(), types.RESOLVE_STATUS_FAILURE, nil,
// 	)
// 	require.Equal(t, result, k.MustGetResult(ctx, 42))
// 	require.Equal(t, sdk.Events{sdk.NewEvent(
// 		types.EventTypeResolve,
// 		sdk.NewAttribute(types.AttributeKeyID, "42"),
// 		sdk.NewAttribute(types.AttributeKeyResolveStatus, "2"),
// 		sdk.NewAttribute(types.AttributeKeyReason, "out-of-gas while executing the wasm script"),
// 	)}, ctx.EventManager().Events())
// }

// func (suite *KeeperTestSuite) TestResolveRequestCallReturnDataSeveralTimes() {
// 	app, ctx := bandtesting.CreateTestApp(t, true)
// 	k := app.OracleKeeper

// 	ctx = ctx.WithBlockTime(bandtesting.ParseTime(1581589890))
// 	k.SetRequest(ctx, 42, types.NewRequest(
// 		// 9th Wasm - set return data several times
// 		9,
// 		BasicCalldata,
// 		[]sdk.ValAddress{bandtesting.Validators[0].ValAddress, bandtesting.Validators[1].ValAddress},
// 		1,
// 		42,
// 		bandtesting.ParseTime(1581589790),
// 		BasicClientID,
// 		[]types.RawRequest{
// 			types.NewRawRequest(1, 1, []byte("beeb")),
// 		},
// 		nil,
// 		bandtesting.TestDefaultExecuteGas,
// 	))
// 	k.ResolveRequest(ctx, 42)

// 	result := types.NewResult(
// 		BasicClientID, 9, BasicCalldata, 2, 1, 42, 0, bandtesting.ParseTime(1581589790).Unix(),
// 		bandtesting.ParseTime(1581589890).Unix(), types.RESOLVE_STATUS_FAILURE, nil,
// 	)
// 	require.Equal(t, result, k.MustGetResult(ctx, 42))

// 	require.Equal(t, sdk.Events{sdk.NewEvent(
// 		types.EventTypeResolve,
// 		sdk.NewAttribute(types.AttributeKeyID, "42"),
// 		sdk.NewAttribute(types.AttributeKeyResolveStatus, "2"),
// 		sdk.NewAttribute(types.AttributeKeyReason, "set return data is called more than once"),
// 	)}, ctx.EventManager().Events())
// }

// func rawRequestsFromFees(ctx sdk.Context, k keeper.Keeper, fees []sdk.Coins) []types.RawRequest {
// 	var rawRequests []types.RawRequest
// 	for _, f := range fees {
// 		id := k.AddDataSource(ctx, types.NewDataSource(
// 			bandtesting.Owner.Address,
// 			"mock ds",
// 			"there is no real code",
// 			"no file",
// 			f,
// 			bandtesting.Treasury.Address,
// 		))

// 		rawRequests = append(rawRequests, types.NewRawRequest(
// 			0, id, nil,
// 		))
// 	}

// 	return rawRequests
// }

// func (suite *KeeperTestSuite) TestCollectFeeEmptyFee() {
// 	app, ctx := bandtesting.CreateTestApp(t, true)
// 	k := app.OracleKeeper

// 	raws := rawRequestsFromFees(ctx, k, []sdk.Coins{
// 		bandtesting.EmptyCoins,
// 		bandtesting.EmptyCoins,
// 		bandtesting.EmptyCoins,
// 		bandtesting.EmptyCoins,
// 		bandtesting.EmptyCoins,
// 	})

// 	coins, err := k.CollectFee(ctx, bandtesting.Alice.Address, bandtesting.EmptyCoins, 1, raws)
// 	require.NoError(t, err)
// 	require.Empty(t, coins)

// 	coins, err = k.CollectFee(ctx, bandtesting.Alice.Address, bandtesting.Coins100000000uband, 1, raws)
// 	require.NoError(t, err)
// 	require.Empty(t, coins)

// 	coins, err = k.CollectFee(ctx, bandtesting.Alice.Address, bandtesting.EmptyCoins, 2, raws)
// 	require.NoError(t, err)
// 	require.Empty(t, coins)

// 	coins, err = k.CollectFee(ctx, bandtesting.Alice.Address, bandtesting.Coins100000000uband, 2, raws)
// 	require.NoError(t, err)
// 	require.Empty(t, coins)
// }

// func (suite *KeeperTestSuite) TestCollectFeeBasicSuccess() {
// 	app, ctx := bandtesting.CreateTestApp(t, false)
// 	k := app.OracleKeeper

// 	raws := rawRequestsFromFees(ctx, k, []sdk.Coins{
// 		bandtesting.EmptyCoins,
// 		bandtesting.Coins1000000uband,
// 		bandtesting.EmptyCoins,
// 		sdk.NewCoins(sdk.NewCoin("uband", math.NewInt(2000000))),
// 		bandtesting.EmptyCoins,
// 	})

// 	balancesRes, err := app.BankKeeper.AllBalances(
// 		ctx,
// 		authtypes.NewQueryAllBalancesRequest(bandtesting.FeePayer.Address, &query.PageRequest{}, false),
// 	)
// 	require.NoError(t, err)
// 	feePayerBalances := balancesRes.Balances
// 	feePayerBalances[0].Amount = feePayerBalances[0].Amount.Sub(math.NewInt(3000000))

// 	coins, err := k.CollectFee(ctx, bandtesting.FeePayer.Address, bandtesting.Coins100000000uband, 1, raws)
// 	require.NoError(t, err)
// 	require.Equal(t, sdk.NewCoins(sdk.NewCoin("uband", math.NewInt(3000000))), coins)

// 	bandtesting.CheckBalances(t, ctx, app.BankKeeper, bandtesting.FeePayer.Address, feePayerBalances)
// 	bandtesting.CheckBalances(
// 		t,
// 		ctx,
// 		app.BankKeeper,
// 		bandtesting.Treasury.Address,
// 		sdk.NewCoins(sdk.NewCoin("uband", math.NewInt(3000000))),
// 	)
// }

// func (suite *KeeperTestSuite) TestCollectFeeBasicSuccessWithOtherAskCount() {
// 	app, ctx := bandtesting.CreateTestApp(t, false)
// 	k := app.OracleKeeper

// 	raws := rawRequestsFromFees(ctx, k, []sdk.Coins{
// 		bandtesting.EmptyCoins,
// 		bandtesting.Coins1000000uband,
// 		bandtesting.EmptyCoins,
// 		sdk.NewCoins(sdk.NewCoin("uband", math.NewInt(2000000))),
// 		bandtesting.EmptyCoins,
// 	})

// 	balancesRes, err := app.BankKeeper.AllBalances(
// 		ctx,
// 		authtypes.NewQueryAllBalancesRequest(bandtesting.FeePayer.Address, &query.PageRequest{}, false),
// 	)
// 	require.NoError(t, err)
// 	feePayerBalances := balancesRes.Balances
// 	feePayerBalances[0].Amount = feePayerBalances[0].Amount.Sub(math.NewInt(12000000))

// 	coins, err := k.CollectFee(ctx, bandtesting.FeePayer.Address, bandtesting.Coins100000000uband, 4, raws)
// 	require.NoError(t, err)
// 	require.Equal(t, sdk.NewCoins(sdk.NewCoin("uband", math.NewInt(12000000))), coins)

// 	bandtesting.CheckBalances(t, ctx, app.BankKeeper, bandtesting.FeePayer.Address, feePayerBalances)
// 	bandtesting.CheckBalances(
// 		t,
// 		ctx,
// 		app.BankKeeper,
// 		bandtesting.Treasury.Address,
// 		sdk.NewCoins(sdk.NewCoin("uband", math.NewInt(12000000))),
// 	)
// }

// func (suite *KeeperTestSuite) TestCollectFeeWithMixedAndFeeNotEnough() {
// 	app, ctx := bandtesting.CreateTestApp(t, true)
// 	k := app.OracleKeeper

// 	raws := rawRequestsFromFees(ctx, k, []sdk.Coins{
// 		bandtesting.EmptyCoins,
// 		bandtesting.Coins1000000uband,
// 		bandtesting.EmptyCoins,
// 		sdk.NewCoins(sdk.NewCoin("uband", math.NewInt(2000000))),
// 		bandtesting.EmptyCoins,
// 	})

// 	coins, err := k.CollectFee(ctx, bandtesting.FeePayer.Address, bandtesting.EmptyCoins, 1, raws)
// 	require.ErrorIs(t, err, types.ErrNotEnoughFee)
// 	require.Nil(t, coins)

// 	coins, err = k.CollectFee(ctx, bandtesting.FeePayer.Address, bandtesting.Coins1000000uband, 1, raws)
// 	require.ErrorIs(t, err, types.ErrNotEnoughFee)
// 	require.Nil(t, coins)
// }

// func (suite *KeeperTestSuite) TestCollectFeeWithEnoughFeeButInsufficientBalance() {
// 	app, ctx := bandtesting.CreateTestApp(t, true)
// 	k := app.OracleKeeper

// 	raws := rawRequestsFromFees(ctx, k, []sdk.Coins{
// 		bandtesting.EmptyCoins,
// 		bandtesting.Coins1000000uband,
// 		bandtesting.EmptyCoins,
// 		sdk.NewCoins(sdk.NewCoin("uband", math.NewInt(2000000))),
// 		bandtesting.EmptyCoins,
// 	})

// 	coins, err := k.CollectFee(ctx, bandtesting.Alice.Address, bandtesting.Coins100000000uband, 1, raws)
// 	require.Nil(t, coins)
// 	// MAX is 100m but have only 1m in account
// 	// First ds collect 1m so there no balance enough for next ds but it doesn't touch limit
// 	require.EqualError(t, err, "spendable balance  is smaller than 2000000uband: insufficient funds")
// }

// func (suite *KeeperTestSuite) TestCollectFeeWithWithManyUnitSuccess() {
// 	app, ctx := bandtesting.CreateTestApp(t, false)
// 	k := app.OracleKeeper

// 	raws := rawRequestsFromFees(ctx, k, []sdk.Coins{
// 		bandtesting.EmptyCoins,
// 		bandtesting.Coins1000000uband,
// 		bandtesting.EmptyCoins,
// 		sdk.NewCoins(sdk.NewCoin("uband", math.NewInt(2000000)), sdk.NewCoin("uabc", math.NewInt(1000000))),
// 		bandtesting.EmptyCoins,
// 	})

// 	err := app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, sdk.NewCoins(sdk.NewCoin("uabc", math.NewInt(2000000))))
// 	require.NoError(t, err)

// 	// Carol have not enough uband but have enough uabc
// 	err = app.BankKeeper.SendCoinsFromModuleToAccount(
// 		ctx,
// 		minttypes.ModuleName,
// 		bandtesting.FeePayer.Address,
// 		sdk.NewCoins(sdk.NewCoin("uabc", math.NewInt(2000000))),
// 	)
// 	require.NoError(t, err)

// 	coins, err := k.CollectFee(
// 		ctx,
// 		bandtesting.FeePayer.Address,
// 		bandtesting.MustGetBalances(ctx, app.BankKeeper, bandtesting.FeePayer.Address),
// 		1,
// 		raws,
// 	)
// 	require.NoError(t, err)

// 	// Coins sum is correct
// 	require.True(
// 		t,
// 		sdk.NewCoins(sdk.NewCoin("uband", math.NewInt(3000000)), sdk.NewCoin("uabc", math.NewInt(1000000))).
// 			Equal(coins),
// 	)

// 	// FeePayer balance
// 	// start: 100band, 0abc
// 	// top-up: 100band, 2abc
// 	// collect 3 band and 1 abc => 97band, 1abc
// 	bandtesting.CheckBalances(
// 		t,
// 		ctx,
// 		app.BankKeeper,
// 		bandtesting.FeePayer.Address,
// 		sdk.NewCoins(sdk.NewCoin("uband", math.NewInt(97000000)), sdk.NewCoin("uabc", math.NewInt(1000000))),
// 	)

// 	// Treasury balance
// 	// start: 0band, 0abc
// 	// collect 3 band and 1 abc => 3band, 1abc
// 	bandtesting.CheckBalances(
// 		t,
// 		ctx,
// 		app.BankKeeper,
// 		bandtesting.Treasury.Address,
// 		sdk.NewCoins(sdk.NewCoin("uband", math.NewInt(3000000)), sdk.NewCoin("uabc", math.NewInt(1000000))),
// 	)
// }

// func (suite *KeeperTestSuite) TestCollectFeeWithWithManyUnitFail() {
// 	app, ctx := bandtesting.CreateTestApp(t, false)
// 	k := app.OracleKeeper

// 	raws := rawRequestsFromFees(ctx, k, []sdk.Coins{
// 		bandtesting.EmptyCoins,
// 		bandtesting.Coins1000000uband,
// 		bandtesting.EmptyCoins,
// 		sdk.NewCoins(sdk.NewCoin("uband", math.NewInt(2000000)), sdk.NewCoin("uabc", math.NewInt(1000000))),
// 		bandtesting.EmptyCoins,
// 	})

// 	err := app.BankKeeper.MintCoins(
// 		ctx,
// 		minttypes.ModuleName,
// 		sdk.NewCoins(sdk.NewCoin("uband", math.NewInt(10000000)), sdk.NewCoin("uabc", math.NewInt(2000000))),
// 	)
// 	require.NoError(t, err)

// 	// Alice have no enough uband and don't have uabc so don't top up
// 	// Bob have enough uband and have some but not enough uabc so add some
// 	err = app.BankKeeper.SendCoinsFromModuleToAccount(
// 		ctx,
// 		minttypes.ModuleName,
// 		bandtesting.Bob.Address,
// 		sdk.NewCoins(sdk.NewCoin("uband", math.NewInt(3000000))),
// 	)
// 	require.NoError(t, err)

// 	err = app.BankKeeper.SendCoinsFromModuleToAccount(
// 		ctx,
// 		minttypes.ModuleName,
// 		bandtesting.Bob.Address,
// 		sdk.NewCoins(sdk.NewCoin("uabc", math.NewInt(1))),
// 	)
// 	require.NoError(t, err)

// 	// Carol have not enough uband but have enough uabc
// 	err = app.BankKeeper.SendCoinsFromModuleToAccount(
// 		ctx,
// 		minttypes.ModuleName,
// 		bandtesting.Carol.Address,
// 		sdk.NewCoins(sdk.NewCoin("uabc", math.NewInt(1000000))),
// 	)
// 	require.NoError(t, err)

// 	// Alice
// 	_, err = k.CollectFee(
// 		ctx,
// 		bandtesting.Alice.Address,
// 		bandtesting.MustGetBalances(ctx, app.BankKeeper, bandtesting.Alice.Address),
// 		1,
// 		raws,
// 	)
// 	require.EqualError(t, err, "require: 1000000uabc, max: 0uabc: not enough fee")

// 	// Bob
// 	_, err = k.CollectFee(
// 		ctx,
// 		bandtesting.Bob.Address,
// 		bandtesting.MustGetBalances(ctx, app.BankKeeper, bandtesting.Bob.Address),
// 		1,
// 		raws,
// 	)
// 	require.EqualError(t, err, "require: 1000000uabc, max: 1uabc: not enough fee")

// 	// Carol
// 	_, err = k.CollectFee(
// 		ctx,
// 		bandtesting.Carol.Address,
// 		bandtesting.MustGetBalances(ctx, app.BankKeeper, bandtesting.Carol.Address),
// 		1,
// 		raws,
// 	)
// 	require.EqualError(t, err, "require: 3000000uband, max: 1000000uband: not enough fee")
// }
