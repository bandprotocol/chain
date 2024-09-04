package feechecker

import (
	"math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/authz"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"

	"github.com/bandprotocol/chain/v3/x/globalfee/keeper"
	oraclekeeper "github.com/bandprotocol/chain/v3/x/oracle/keeper"
	oracletypes "github.com/bandprotocol/chain/v3/x/oracle/types"
)

type FeeChecker struct {
	OracleKeeper    *oraclekeeper.Keeper
	GlobalfeeKeeper *keeper.Keeper
	StakingKeeper   *stakingkeeper.Keeper
}

func NewFeeChecker(
	oracleKeeper *oraclekeeper.Keeper,
	globalfeeKeeper *keeper.Keeper,
	stakingKeeper *stakingkeeper.Keeper,
) FeeChecker {
	return FeeChecker{
		OracleKeeper:    oracleKeeper,
		GlobalfeeKeeper: globalfeeKeeper,
		StakingKeeper:   stakingKeeper,
	}
}

func (fc FeeChecker) CheckTxFeeWithMinGasPrices(
	ctx sdk.Context,
	tx sdk.Tx,
) (sdk.Coins, int64, error) {
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return nil, 0, sdkerrors.ErrTxDecode.Wrap("Tx must be a FeeTx")
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
		var allFees sdk.Coins
		if !allGasPrices.IsZero() {
			glDec := sdk.NewDec(int64(gas))
			for _, gp := range allGasPrices {
				if !gp.IsZero() {
					fee := gp.Amount.Mul(glDec)
					allFees = append(allFees, sdk.NewCoin(gp.Denom, fee.Ceil().RoundInt()))
				}
			}
		}

		if !allFees.IsZero() && !feeCoins.IsAnyGTE(allFees) {
			return nil, 0, sdkerrors.ErrInsufficientFee.Wrapf(
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

	globalMinGasPrices = fc.GlobalfeeKeeper.GetParams(ctx).MinimumGasPrices
	// global fee is empty set, set global fee to 0uband (bondDenom)
	if len(globalMinGasPrices) == 0 {
		globalMinGasPrices, err = fc.DefaultZeroGlobalFee(ctx)
	}

	return globalMinGasPrices.Sort(), err
}

func (fc FeeChecker) DefaultZeroGlobalFee(ctx sdk.Context) ([]sdk.DecCoin, error) {
	bondDenom := fc.GetBondDenom(ctx)
	if bondDenom == "" {
		return nil, sdkerrors.ErrNotFound.Wrap("empty staking bond denomination")
	}

	return []sdk.DecCoin{sdk.NewDecCoinFromDec(bondDenom, sdk.NewDec(0))}, nil
}

func (fc FeeChecker) GetBondDenom(ctx sdk.Context) string {
	return fc.StakingKeeper.BondDenom(ctx)
}
