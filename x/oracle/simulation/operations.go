package simulation

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/bandprotocol/chain/v2/testing/testdata"
	"github.com/bandprotocol/chain/v2/x/oracle/keeper"
	"github.com/bandprotocol/chain/v2/x/oracle/types"
)

// Simulation operation weights constants
const (
	OpWeightMsgRequestData        = "op_weight_msg_request_data"
	OpWeightMsgReportData         = "op_weight_msg_report_data"
	OpWeightMsgCreateDataSource   = "op_weight_msg_create_data_source"
	OpWeightMsgEditDataSource     = "op_weight_msg_edit_data_source"
	OpWeightMsgCreateOracleScript = "op_weight_msg_create_oracle_script"
	OpWeightMsgEditOracleScript   = "op_weight_msg_edit_oracle_script"
	OpWeightMsgActivate           = "op_weight_msg_activate"

	DefaultWeightMsgRequestData        int = 100
	DefaultWeightMsgReportData         int = 100
	DefaultWeightMsgCreateDataSource   int = 100
	DefaultWeightMsgEditDataSource     int = 100
	DefaultWeightMsgCreateOracleScript int = 100
	DefaultWeightMsgEditOracleScript   int = 100
	DefaultWeightMsgActivate           int = 100
)

type BankKeeper interface {
	simulation.BankKeeper
	IsSendEnabledCoin(ctx sdk.Context, coin sdk.Coin) bool
}

func WeightedOperations(
	appParams simtypes.AppParams,
	cdc codec.JSONCodec,
	ak types.AccountKeeper,
	bk simulation.BankKeeper,
	sk types.StakingKeeper,
	k keeper.Keeper,
) simulation.WeightedOperations {
	var (
		weightMsgRequestData        int
		weightMsgReportData         int
		weightMsgCreateDataSource   int
		weightMsgEditDataSource     int
		weightMsgCreateOracleScript int
		weightMsgEditOracleScript   int
		weightMsgActivate           int
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgRequestData, &weightMsgRequestData, nil,
		func(_ *rand.Rand) {
			weightMsgRequestData = DefaultWeightMsgRequestData
		},
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgReportData, &weightMsgReportData, nil,
		func(_ *rand.Rand) {
			weightMsgReportData = DefaultWeightMsgReportData
		},
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgCreateDataSource, &weightMsgCreateDataSource, nil,
		func(_ *rand.Rand) {
			weightMsgCreateDataSource = DefaultWeightMsgCreateDataSource
		},
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgEditDataSource, &weightMsgEditDataSource, nil,
		func(_ *rand.Rand) {
			weightMsgEditDataSource = DefaultWeightMsgEditDataSource
		},
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgCreateOracleScript, &weightMsgCreateOracleScript, nil,
		func(_ *rand.Rand) {
			weightMsgCreateOracleScript = DefaultWeightMsgCreateOracleScript
		},
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgEditOracleScript, &weightMsgEditOracleScript, nil,
		func(_ *rand.Rand) {
			weightMsgEditOracleScript = DefaultWeightMsgEditOracleScript
		},
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgActivate, &weightMsgActivate, nil,
		func(_ *rand.Rand) {
			weightMsgActivate = DefaultWeightMsgActivate
		},
	)

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgRequestData,
			SimulateMsgRequestData(ak, bk, sk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgReportData,
			SimulateMsgReportData(ak, bk, sk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgCreateDataSource,
			SimulateMsgCreateDataSource(ak, bk, sk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgEditDataSource,
			SimulateMsgEditDataSource(ak, bk, sk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgCreateOracleScript,
			SimulateMsgCreateOracleScript(ak, bk, sk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgEditOracleScript,
			SimulateMsgEditOracleScript(ak, bk, sk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgActivate,
			SimulateMsgActivate(ak, bk, sk, k),
		),
	}
}

// SimulateMsgRequestData generates a MsgRequestData with random values
func SimulateMsgRequestData(
	ak types.AccountKeeper,
	bk simulation.BankKeeper,
	sk types.StakingKeeper,
	keeper keeper.Keeper,
) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)

		oCount := keeper.GetOracleScriptCount(ctx)
		oid := types.OracleScriptID(0)
		for i := uint64(1); i <= oCount; i++ {
			os, _ := keeper.GetOracleScript(ctx, types.OracleScriptID(i))
			_, ok := simtypes.FindAccount(accs, sdk.MustAccAddressFromBech32(os.Owner))
			if ok {
				oid = types.OracleScriptID(i)
				break
			}
		}
		if oid == 0 {
			return simtypes.NoOpMsg(
				types.ModuleName,
				types.MsgRequestData{}.Type(),
				"no oracle script available",
			), nil, nil
		}

		did := keeper.GetDataSourceCount(ctx)
		if did < 3 {
			return simtypes.NoOpMsg(
				types.ModuleName,
				types.MsgRequestData{}.Type(),
				"data sources are not enough",
			), nil, nil
		}

		maxAskCount := 0
		sk.IterateBondedValidatorsByPower(ctx,
			func(idx int64, val stakingtypes.ValidatorI) (stop bool) {
				if keeper.GetValidatorStatus(ctx, val.GetOperator()).IsActive {
					maxAskCount++
				}

				return false
			},
		)
		if maxAskCount == 0 {
			return simtypes.NoOpMsg(
				types.ModuleName,
				types.MsgRequestData{}.Type(),
				"active validators are not enough",
			), nil, nil
		}

		if maxAskCount > 10 {
			maxAskCount = 10
		}
		askCount := simtypes.RandIntBetween(r, 1, maxAskCount+1)

		msg := types.MsgRequestData{
			Sender:         simAccount.Address.String(),
			OracleScriptID: types.OracleScriptID(oid),
			Calldata:       []byte(simtypes.RandStringOfLength(r, 100)),
			AskCount:       uint64(askCount),
			MinCount:       uint64(simtypes.RandIntBetween(r, 1, askCount+1)),
			ClientID:       simtypes.RandStringOfLength(r, 100),
			FeeLimit:       sdk.NewCoins(sdk.NewInt64Coin("uband", 0)),
			PrepareGas:     uint64(simtypes.RandIntBetween(r, 100000, 200000)),
			ExecuteGas:     uint64(simtypes.RandIntBetween(r, 100000, 200000)),
		}

		txCtx := BuildOperationInput(r, app, ctx, &msg, simAccount, ak, bk, sk, nil)

		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}

// SimulateMsgReportData generates a MsgReportData with random values
func SimulateMsgReportData(
	ak types.AccountKeeper,
	bk simulation.BankKeeper,
	sk types.StakingKeeper,
	keeper keeper.Keeper,
) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		var simAccount simtypes.Account

		rCount := keeper.GetRequestCount(ctx)
		rid := types.RequestID(0)
		for i := uint64(1); i <= rCount; i++ {
			req, _ := keeper.GetRequest(ctx, types.RequestID(i))

			for _, val := range req.RequestedValidators {
				valAddr, _ := sdk.ValAddressFromBech32(val)
				acc, ok := simtypes.FindAccount(accs, sdk.AccAddress(valAddr))

				if ok && !keeper.HasReport(ctx, types.RequestID(i), valAddr) {
					simAccount = acc
					rid = types.RequestID(i)
					break
				}
			}

			if rid != 0 {
				break
			}
		}

		if rid == 0 {
			return simtypes.NoOpMsg(
				types.ModuleName,
				types.MsgReportData{}.Type(),
				"no request available",
			), nil, nil
		}

		var rawReports []types.RawReport
		for i := 1; i <= 3; i++ {
			rawReports = append(rawReports, types.RawReport{
				ExternalID: types.ExternalID(i),
				ExitCode:   uint32(simtypes.RandIntBetween(r, 0, 255)),
				Data:       []byte(simtypes.RandStringOfLength(r, 100)),
			})
		}

		msg := types.MsgReportData{
			RequestID:  types.RequestID(rid),
			RawReports: rawReports,
			Validator:  sdk.ValAddress(simAccount.Address).String(),
		}

		txCtx := BuildOperationInput(r, app, ctx, &msg, simAccount, ak, bk, sk, nil)

		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}

// SimulateMsgCreateDataSource generates a MsgCreateDataSource with random values
func SimulateMsgCreateDataSource(
	ak types.AccountKeeper,
	bk simulation.BankKeeper,
	sk types.StakingKeeper,
	keeper keeper.Keeper,
) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		ownerAccount, _ := simtypes.RandomAcc(r, accs)
		treaAccount, _ := simtypes.RandomAcc(r, accs)

		msg := types.MsgCreateDataSource{
			Sender:      simAccount.Address.String(),
			Name:        simtypes.RandStringOfLength(r, 10),
			Description: simtypes.RandStringOfLength(r, 100),
			Executable:  []byte(simtypes.RandStringOfLength(r, 100)),
			Fee:         sdk.NewCoins(sdk.NewInt64Coin("uband", 0)),
			Treasury:    treaAccount.Address.String(),
			Owner:       ownerAccount.Address.String(),
		}

		txCtx := BuildOperationInput(r, app, ctx, &msg, simAccount, ak, bk, sk, nil)

		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}

// SimulateMsgEditDataSource generates a MsgEditDataSource with random values
func SimulateMsgEditDataSource(
	ak types.AccountKeeper,
	bk simulation.BankKeeper,
	sk types.StakingKeeper,
	keeper keeper.Keeper,
) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		ownerAccount, _ := simtypes.RandomAcc(r, accs)
		treaAccount, _ := simtypes.RandomAcc(r, accs)

		dCount := keeper.GetDataSourceCount(ctx)
		did := types.DataSourceID(0)
		for i := uint64(1); i <= dCount; i++ {
			os, _ := keeper.GetDataSource(ctx, types.DataSourceID(i))
			_, ok := simtypes.FindAccount(accs, sdk.MustAccAddressFromBech32(os.Owner))
			if ok {
				did = types.DataSourceID(i)
				break
			}
		}

		if did == 0 {
			return simtypes.NoOpMsg(
				types.ModuleName,
				types.MsgEditDataSource{}.Type(),
				"no data source available",
			), nil, nil
		}

		ds, _ := keeper.GetDataSource(ctx, types.DataSourceID(did))
		simAccount, ok := simtypes.FindAccount(accs, sdk.MustAccAddressFromBech32(ds.Owner))
		if !ok {
			return simtypes.NoOpMsg(
				types.ModuleName,
				types.MsgEditDataSource{}.Type(),
				"unknown owner",
			), nil, nil
		}

		msg := types.MsgEditDataSource{
			Sender:       simAccount.Address.String(),
			DataSourceID: types.DataSourceID(did),
			Name:         simtypes.RandStringOfLength(r, 10),
			Description:  simtypes.RandStringOfLength(r, 100),
			Executable:   []byte(simtypes.RandStringOfLength(r, 100)),
			Fee:          sdk.NewCoins(sdk.NewInt64Coin("uband", 0)),
			Treasury:     treaAccount.Address.String(),
			Owner:        ownerAccount.Address.String(),
		}

		txCtx := BuildOperationInput(r, app, ctx, &msg, simAccount, ak, bk, sk, nil)

		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}

// SimulateMsgCreateOracleScript generates a MsgCreateOracleScript with random values
func SimulateMsgCreateOracleScript(
	ak types.AccountKeeper,
	bk simulation.BankKeeper,
	sk types.StakingKeeper,
	keeper keeper.Keeper,
) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		ownerAccount, _ := simtypes.RandomAcc(r, accs)

		msg := types.MsgCreateOracleScript{
			Sender:        simAccount.Address.String(),
			Name:          simtypes.RandStringOfLength(r, 10),
			Description:   simtypes.RandStringOfLength(r, 100),
			Schema:        simtypes.RandStringOfLength(r, 100),
			SourceCodeURL: simtypes.RandStringOfLength(r, 100),
			Code:          testdata.Wasm1,
			Owner:         ownerAccount.Address.String(),
		}

		txCtx := BuildOperationInput(r, app, ctx, &msg, simAccount, ak, bk, sk, nil)

		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}

// SimulateMsgEditOracleScript generates a MsgEditOracleScript with random values
func SimulateMsgEditOracleScript(
	ak types.AccountKeeper,
	bk simulation.BankKeeper,
	sk types.StakingKeeper,
	keeper keeper.Keeper,
) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		var simAccount simtypes.Account

		oCount := keeper.GetOracleScriptCount(ctx)
		oid := types.OracleScriptID(0)
		for i := uint64(1); i <= oCount; i++ {
			os, _ := keeper.GetOracleScript(ctx, types.OracleScriptID(i))
			acc, ok := simtypes.FindAccount(accs, sdk.MustAccAddressFromBech32(os.Owner))
			if ok {
				simAccount = acc
				oid = types.OracleScriptID(i)
				break
			}
		}

		if oid == 0 {
			return simtypes.NoOpMsg(
				types.ModuleName,
				types.MsgEditOracleScript{}.Type(),
				"no oracle script available",
			), nil, nil
		}

		msg := types.MsgEditOracleScript{
			Sender:         simAccount.Address.String(),
			OracleScriptID: types.OracleScriptID(oid),
			Name:           simtypes.RandStringOfLength(r, 10),
			Description:    simtypes.RandStringOfLength(r, 100),
			Schema:         simtypes.RandStringOfLength(r, 100),
			SourceCodeURL:  simtypes.RandStringOfLength(r, 100),
			Code:           testdata.Wasm1,
			Owner:          simAccount.Address.String(),
		}

		txCtx := BuildOperationInput(r, app, ctx, &msg, simAccount, ak, bk, sk, nil)

		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}

// SimulateMsgActivate generates a MsgActivate with random values
func SimulateMsgActivate(
	ak types.AccountKeeper,
	bk simulation.BankKeeper,
	sk types.StakingKeeper,
	keeper keeper.Keeper,
) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)

		if keeper.GetValidatorStatus(ctx, sdk.ValAddress(simAccount.Address)).IsActive {
			return simtypes.NoOpMsg(
				types.ModuleName,
				types.MsgActivate{}.Type(),
				"already activate",
			), nil, nil
		}

		msg := types.MsgActivate{
			Validator: sdk.ValAddress(simAccount.Address).String(),
		}

		txCtx := BuildOperationInput(r, app, ctx, &msg, simAccount, ak, bk, sk, nil)

		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}

// BuildOperationInput helper to build object
func BuildOperationInput(
	r *rand.Rand,
	app *baseapp.BaseApp,
	ctx sdk.Context,
	msg interface {
		sdk.Msg
		Type() string
	},
	simAccount simtypes.Account,
	ak types.AccountKeeper,
	bk simulation.BankKeeper,
	sk types.StakingKeeper,
	deposit sdk.Coins,
) simulation.OperationInput {
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	txConfig := tx.NewTxConfig(codec.NewProtoCodec(interfaceRegistry), tx.DefaultSignModes)
	return simulation.OperationInput{
		R:               r,
		App:             app,
		TxGen:           txConfig,
		Cdc:             nil,
		Msg:             msg,
		MsgType:         msg.Type(),
		Context:         ctx,
		SimAccount:      simAccount,
		AccountKeeper:   ak,
		Bankkeeper:      bk,
		ModuleName:      types.ModuleName,
		CoinsSpentInMsg: deposit,
	}
}
