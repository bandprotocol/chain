package feechecker

import (
	"math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	oraclekeeper "github.com/bandprotocol/chain/v2/x/oracle/keeper"
	oracletypes "github.com/bandprotocol/chain/v2/x/oracle/types"
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

func checkValidMsgReport(ctx sdk.Context, oracleKeeper *oraclekeeper.Keeper, msg *oracletypes.MsgReportData) error {
	validator, err := sdk.ValAddressFromBech32(msg.Validator)
	if err != nil {
		return err
	}
	return oracleKeeper.CheckValidReport(ctx, msg.RequestID, validator, msg.RawReports)
}
