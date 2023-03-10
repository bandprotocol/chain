package feechecker

import (
	"math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/authz"

	oraclekeeper "github.com/bandprotocol/chain/v2/x/oracle/keeper"
	"github.com/bandprotocol/chain/v2/x/oracle/types"
)

// getTxPriority returns priority of the provided fee based on gas prices of uband
func getTxPriority(fee sdk.Coins, gas int64, denom string) int64 {
	ok, c := fee.Find(denom)
	if !ok {
		return 0
	}

	// multiplied by 10000 first to support our current standard (0.0025) because priority is int64.
	// otherwise, if gas_price < 1, the priority will be 0.
	priority := int64(math.MaxInt64)
	gasPrice := c.Amount.MulRaw(10000).QuoRaw(gas)
	if gasPrice.IsInt64() {
		priority = gasPrice.Int64()
	}

	return priority
}

// getMinGasPrice will also return sorted coins
func getMinGasPrice(ctx sdk.Context, feeTx sdk.FeeTx) sdk.Coins {
	minGasPrices := ctx.MinGasPrices()
	gas := feeTx.GetGas()
	// special case: if minGasPrices=[], requiredFees=[]
	requiredFees := make(sdk.Coins, len(minGasPrices))
	// if not all coins are zero, check fee with min_gas_price
	if !minGasPrices.IsZero() {
		// Determine the required fees by multiplying each required minimum gas
		// price by the gas limit, where fee = ceil(minGasPrice * gasLimit).
		glDec := sdk.NewDec(int64(gas))
		for i, gp := range minGasPrices {
			fee := gp.Amount.Mul(glDec)
			requiredFees[i] = sdk.NewCoin(gp.Denom, fee.Ceil().RoundInt())
		}
	}

	return requiredFees.Sort()
}

// CombinedFeeRequirement will combine the global fee and min_gas_price. Both globalFees and minGasPrices must be valid, but CombinedFeeRequirement does not validate them, so it may return 0denom.
func CombinedFeeRequirement(globalFees, minGasPrices sdk.Coins) sdk.Coins {
	// empty min_gas_price
	if len(minGasPrices) == 0 {
		return globalFees
	}
	// empty global fee is not possible if we set default global fee
	if len(globalFees) == 0 && len(minGasPrices) != 0 {
		return minGasPrices
	}

	// if min_gas_price denom is in globalfee, and the amount is higher than globalfee, add min_gas_price to allFees
	var allFees sdk.Coins
	for _, fee := range globalFees {
		// min_gas_price denom in global fee
		ok, c := minGasPrices.Find(fee.Denom)
		if ok && c.Amount.GT(fee.Amount) {
			allFees = append(allFees, c)
		} else {
			allFees = append(allFees, fee)
		}
	}

	return allFees.Sort()
}

func checkValidReportMsg(ctx sdk.Context, oracleKeeper *oraclekeeper.Keeper, r *types.MsgReportData) error {
	validator, err := sdk.ValAddressFromBech32(r.Validator)
	if err != nil {
		return err
	}
	report := types.NewReport(validator, false, r.RawReports)
	return oracleKeeper.CheckValidReport(ctx, r.RequestID, report)
}

func checkExecMsgReportFromReporter(ctx sdk.Context, oracleKeeper *oraclekeeper.Keeper, msg sdk.Msg) (bool, error) {
	// Check is the MsgExec from reporter
	me, ok := msg.(*authz.MsgExec)
	if !ok {
		return false, nil
	}

	// If cannot get message, then pretend as non-free transaction
	msgs, err := me.GetMessages()
	if err != nil {
		return false, nil
	}

	grantee, err := sdk.AccAddressFromBech32(me.Grantee)
	if err != nil {
		return false, nil
	}

	allValidReportMsg := true
	for _, m := range msgs {
		r, ok := m.(*types.MsgReportData)
		// If this is not report msg, skip other msgs on this exec msg
		if !ok {
			allValidReportMsg = false
			break
		}

		// Fail to parse validator, then discard this transaction
		validator, err := sdk.ValAddressFromBech32(r.Validator)
		if err != nil {
			return false, err
		}

		// If this grantee is not a reporter of validator, then discard this transaction
		if !oracleKeeper.IsReporter(ctx, validator, grantee) {
			return false, sdkerrors.ErrUnauthorized.Wrap("authorization not found")
		}

		// Check if it's not valid report msg, discard this transaction
		if err := checkValidReportMsg(ctx, oracleKeeper, r); err != nil {
			return false, err
		}
	}
	// Return false if this exec msg has other non-report msg
	return allValidReportMsg, nil
}
