package feechecker

import (
	"errors"
	"math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/bandprotocol/chain/v2/x/globalfee"
	"github.com/bandprotocol/chain/v2/x/globalfee/types"
	oraclekeeper "github.com/bandprotocol/chain/v2/x/oracle/keeper"
	oracletypes "github.com/bandprotocol/chain/v2/x/oracle/types"
)

type FeeChecker struct {
	OracleKeeper    *oraclekeeper.Keeper
	GlobalMinFee    globalfee.ParamSource
	StakingSubspace paramtypes.Subspace
}

func NewFeeChecker(
	oracleKeeper *oraclekeeper.Keeper,
	globalfeeSubspace, stakingSubspace paramtypes.Subspace,
) FeeChecker {
	if !globalfeeSubspace.HasKeyTable() {
		panic("global fee paramspace was not set up via module")
	}

	if !stakingSubspace.HasKeyTable() {
		panic("staking paramspace was not set up via module")
	}

	return FeeChecker{
		OracleKeeper:    oracleKeeper,
		GlobalMinFee:    globalfeeSubspace,
		StakingSubspace: stakingSubspace,
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
		isValidReportTx := fc.CheckReportTx(ctx, tx)
		if isValidReportTx {
			return sdk.Coins{}, int64(math.MaxInt64), nil
		}

		requiredFees := getMinGasPrice(ctx, feeTx)
		requiredGlobalFees, err := fc.GetGlobalFee(ctx, feeTx)
		if err != nil {
			return nil, 0, err
		}

		allFees := CombinedFeeRequirement(requiredGlobalFees, requiredFees)

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
	isValidReportTx := true

	for _, msg := range tx.GetMsgs() {
		// Check direct report msg
		if dr, ok := msg.(*oracletypes.MsgReportData); ok {
			// Check if it's not valid report msg, discard this transaction
			if err := checkValidReportMsg(ctx, fc.OracleKeeper, dr); err != nil {
				return false
			}
		} else {
			isValid := checkExecMsgReportFromReporter(ctx, fc.OracleKeeper, msg)
			isValidReportTx = isValidReportTx && isValid
		}
	}

	return isValidReportTx
}

func (fc FeeChecker) GetGlobalFee(ctx sdk.Context, feeTx sdk.FeeTx) (sdk.Coins, error) {
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
	requiredGlobalFees := make(sdk.Coins, len(globalMinGasPrices))
	// Determine the required fees by multiplying each required minimum gas
	// price by the gas limit, where fee = ceil(minGasPrice * gasLimit).
	glDec := sdk.NewDec(int64(feeTx.GetGas()))
	for i, gp := range globalMinGasPrices {
		fee := gp.Amount.Mul(glDec)
		requiredGlobalFees[i] = sdk.NewCoin(gp.Denom, fee.Ceil().RoundInt())
	}

	return requiredGlobalFees.Sort(), err
}

func (fc FeeChecker) DefaultZeroGlobalFee(ctx sdk.Context) ([]sdk.DecCoin, error) {
	bondDenom := fc.GetBondDenom(ctx)
	if bondDenom == "" {
		return nil, errors.New("empty staking bond denomination")
	}

	return []sdk.DecCoin{sdk.NewDecCoinFromDec(bondDenom, sdk.NewDec(0))}, nil
}

func (fc FeeChecker) GetBondDenom(ctx sdk.Context) string {
	var bondDenom string
	if fc.StakingSubspace.Has(ctx, stakingtypes.KeyBondDenom) {
		fc.StakingSubspace.Get(ctx, stakingtypes.KeyBondDenom, &bondDenom)
	}

	return bondDenom
}
