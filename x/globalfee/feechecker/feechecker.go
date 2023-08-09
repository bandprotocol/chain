package feechecker

import (
	"errors"
	"math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/authz"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/bandprotocol/chain/v2/x/globalfee"
	"github.com/bandprotocol/chain/v2/x/globalfee/types"
	oraclekeeper "github.com/bandprotocol/chain/v2/x/oracle/keeper"
	oracletypes "github.com/bandprotocol/chain/v2/x/oracle/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
)

type FeeChecker struct {
	OracleKeeper  *oraclekeeper.Keeper
	GlobalMinFee  globalfee.ParamSource
	StakingKeeper *stakingkeeper.Keeper
}

func NewFeeChecker(
	oracleKeeper *oraclekeeper.Keeper,
	globalfeeSubspace paramtypes.Subspace,
	stakingKeeper *stakingkeeper.Keeper,
) FeeChecker {
	if !globalfeeSubspace.HasKeyTable() {
		panic("global fee paramspace was not set up via module")
	}

	return FeeChecker{
		OracleKeeper:  oracleKeeper,
		GlobalMinFee:  globalfeeSubspace,
		StakingKeeper: stakingKeeper,
	}
}

func (fc FeeChecker) CheckTxFeeWithMinGasPrices(
	ctx sdk.Context,
	tx sdk.Tx,
) (sdk.Coins, int64, error) {
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return nil, 0, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}

	feeCoins := feeTx.GetFee()
	gas := feeTx.GetGas()

	// Ensure that the provided fees meet minimum-gas-prices and globalFees,
	// if this is a CheckTx. This is only for local mempool purposes, and thus
	// is only ran on check tx.
	if ctx.IsCheckTx() {
		// Check if this is a valid report transaction, we allow gas price to be zero and set priority to the highest
		isValidReportTx := fc.CheckReportTx(ctx, tx)
		if isValidReportTx {
			return sdk.Coins{}, int64(math.MaxInt64), nil
		}

		minGasPrices := getMinGasPrices(ctx)
		globalMinGasPrices, err := fc.GetGlobalMinGasPrices(ctx)
		if err != nil {
			return nil, 0, err
		}

		allGasPrices := CombinedGasPricesRequirement(minGasPrices, globalMinGasPrices)

		// Calculate all fees from all gas prices
		gas := feeTx.GetGas()
		allFees := make(sdk.Coins, len(allGasPrices))
		if !minGasPrices.IsZero() {
			glDec := sdk.NewDec(int64(gas))
			for i, gp := range minGasPrices {
				fee := gp.Amount.Mul(glDec)
				allFees[i] = sdk.NewCoin(gp.Denom, fee.Ceil().RoundInt())
			}
		}

		if !allFees.IsZero() && !feeCoins.IsAnyGTE(allFees) {
			return nil, 0, sdkerrors.Wrapf(
				sdkerrors.ErrInsufficientFee,
				"insufficient fees; got: %s required: %s",
				feeCoins,
				allFees,
			)
		}
	}

	priority := getTxPriority(feeCoins, int64(gas), fc.GetBondDenom(ctx))
	return feeCoins, priority, nil
}

func (fc FeeChecker) CheckReportTx(ctx sdk.Context, tx sdk.Tx) bool {
	for _, msg := range tx.GetMsgs() {
		switch msg := msg.(type) {
		case *oracletypes.MsgReportData:
			if err := checkValidReportMsg(ctx, fc.OracleKeeper, msg); err != nil {
				return false
			}
		case *authz.MsgExec:
			if !checkExecMsgReportFromReporter(ctx, fc.OracleKeeper, msg) {
				return false
			}
		default:
			return false
		}
	}
	return true
}

func (fc FeeChecker) GetGlobalMinGasPrices(ctx sdk.Context) (sdk.DecCoins, error) {
	var (
		globalMinGasPrices sdk.DecCoins
		err                error
	)

	if fc.GlobalMinFee.Has(ctx, types.ParamStoreKeyMinGasPrices) {
		fc.GlobalMinFee.Get(ctx, types.ParamStoreKeyMinGasPrices, &globalMinGasPrices)
	}
	// global fee is empty set, set global fee to 0uband (bondDenom)
	if len(globalMinGasPrices) == 0 {
		globalMinGasPrices, err = fc.DefaultZeroGlobalFee(ctx)
	}

	return globalMinGasPrices.Sort(), err
}

func (fc FeeChecker) DefaultZeroGlobalFee(ctx sdk.Context) ([]sdk.DecCoin, error) {
	bondDenom := fc.GetBondDenom(ctx)
	if bondDenom == "" {
		return nil, errors.New("empty staking bond denomination")
	}

	return []sdk.DecCoin{sdk.NewDecCoinFromDec(bondDenom, sdk.NewDec(0))}, nil
}

func (fc FeeChecker) GetBondDenom(ctx sdk.Context) string {
	// 0.47 TODO: test bond denom is not nil
	return fc.StakingKeeper.BondDenom(ctx)
}
