package fee

import (
	"math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	oraclekeeper "github.com/bandprotocol/chain/v2/x/oracle/keeper"
	"github.com/bandprotocol/chain/v2/x/oracle/types"
)

// TODO: move to be params of globalfee module
var (
	Denom                   = "uband"
	ConsensusMinFee sdk.Dec = sdk.NewDecWithPrec(25, 4)
)

type FeeChecker struct {
	oracleKeeper *oraclekeeper.Keeper
}

func NewFeeChecker(ork *oraclekeeper.Keeper) FeeChecker {
	return FeeChecker{
		oracleKeeper: ork,
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
		isValidReportTx, err := fc.checkReportTx(ctx, tx)
		if err != nil {
			return nil, 0, err
		}

		if isValidReportTx {
			return sdk.Coins{}, int64(math.MaxInt64), nil
		}

		requiredFees := getMinGasPrice(ctx, feeTx)
		requiredGlobalFees := fc.getGlobalFee(ctx, feeTx)

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

	priority := getTxPriority(feeCoins, int64(gas))
	return feeCoins, priority, nil
}

func (fc FeeChecker) checkReportTx(ctx sdk.Context, tx sdk.Tx) (bool, error) {
	isValidReportTx := true

	for _, msg := range tx.GetMsgs() {
		// Check direct report msg
		if dr, ok := msg.(*types.MsgReportData); ok {
			// Check if it's not valid report msg, discard this transaction
			if err := checkValidReportMsg(ctx, fc.oracleKeeper, dr); err != nil {
				return false, err
			}
		} else {
			isValid, err := checkExecMsgReportFromReporter(ctx, fc.oracleKeeper, msg)
			if err != nil {
				return false, err
			}

			isValidReportTx = isValidReportTx && isValid
		}
	}

	return isValidReportTx, nil
}

func (fc FeeChecker) getGlobalFee(ctx sdk.Context, feeTx sdk.FeeTx) sdk.Coins {
	gas := feeTx.GetGas()

	glDec := sdk.NewDec(int64(gas))
	fee := ConsensusMinFee.Mul(glDec)
	requiredGlobalFees := sdk.NewCoins(sdk.NewCoin(Denom, fee.Ceil().RoundInt()))

	return requiredGlobalFees.Sort()
}
