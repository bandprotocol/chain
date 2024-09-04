package feechecker

import (
	"math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"

	oraclekeeper "github.com/bandprotocol/chain/v3/x/oracle/keeper"
	"github.com/bandprotocol/chain/v3/x/oracle/types"
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

// getMinGasPrices will also return sorted dec coins
func getMinGasPrices(ctx sdk.Context) sdk.DecCoins {
	return ctx.MinGasPrices().Sort()
}

// CombinedGasPricesRequirement will combine the global min_gas_prices and min_gas_prices. Both globalMinGasPrices and minGasPrices must be valid
func CombinedGasPricesRequirement(globalMinGasPrices, minGasPrices sdk.DecCoins) sdk.DecCoins {
	// return globalMinGasPrices if minGasPrices has not been set
	if minGasPrices.Empty() {
		return globalMinGasPrices
	}
	// return minGasPrices if globalMinGasPrices is empty
	if globalMinGasPrices.Empty() {
		return minGasPrices
	}

	// if min_gas_price denom is in globalfee, and the amount is higher than globalfee, add min_gas_price to allGasPrices
	var allGasPrices sdk.DecCoins
	for _, gmgp := range globalMinGasPrices {
		// min_gas_price denom in global fee
		mgp := minGasPrices.AmountOf(gmgp.Denom)
		if mgp.GT(gmgp.Amount) {
			allGasPrices = append(allGasPrices, sdk.NewDecCoinFromDec(gmgp.Denom, mgp))
		} else {
			allGasPrices = append(allGasPrices, sdk.NewDecCoinFromDec(gmgp.Denom, gmgp.Amount))
		}
	}

	return allGasPrices.Sort()
}

func checkValidReportMsg(ctx sdk.Context, oracleKeeper *oraclekeeper.Keeper, r *types.MsgReportData) error {
	validator, err := sdk.ValAddressFromBech32(r.Validator)
	if err != nil {
		return err
	}
	return oracleKeeper.CheckValidReport(ctx, r.RequestID, validator, r.RawReports)
}

func checkExecMsgReportFromReporter(ctx sdk.Context, oracleKeeper *oraclekeeper.Keeper, msgExec *authz.MsgExec) bool {
	// If cannot get message, then pretend as non-free transaction
	msgs, err := msgExec.GetMessages()
	if err != nil {
		return false
	}

	grantee, err := sdk.AccAddressFromBech32(msgExec.Grantee)
	if err != nil {
		return false
	}

	for _, m := range msgs {
		r, ok := m.(*types.MsgReportData)
		// If this is not report msg, skip other msgs on this exec msg
		if !ok {
			return false
		}

		// Fail to parse validator, then reject this message
		validator, err := sdk.ValAddressFromBech32(r.Validator)
		if err != nil {
			return false
		}

		// If this grantee is not a reporter of validator, then reject this message
		if !oracleKeeper.IsReporter(ctx, validator, grantee) {
			return false
		}

		// Check if it's not valid report msg, discard this message
		if err := checkValidReportMsg(ctx, oracleKeeper, r); err != nil {
			return false
		}
	}

	// Return true if all sub exec msgs have not been rejected
	return true
}
