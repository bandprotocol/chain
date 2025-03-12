package keeper

import (
	"fmt"
	"strconv"
	"time"

	dbm "github.com/cosmos/cosmos-db"

	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/bandprotocol/chain/v3/x/feeds/types"
)

// GetPricesIterator returns an iterator for prices store.
func (k Keeper) GetPricesIterator(ctx sdk.Context) dbm.Iterator {
	return storetypes.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.PriceStoreKeyPrefix)
}

// GetAllPrices returns a list of all prices.
func (k Keeper) GetAllPrices(ctx sdk.Context) (prices []types.Price) {
	iterator := k.GetPricesIterator(ctx)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var price types.Price
		k.cdc.MustUnmarshal(iterator.Value(), &price)
		prices = append(prices, price)
	}

	return
}

// GetPrices returns a list of prices by signal ids.
func (k Keeper) GetPrices(ctx sdk.Context, signalIDs []string) []types.Price {
	prices := make([]types.Price, 0, len(signalIDs))
	for _, signalID := range signalIDs {
		price := k.GetPrice(ctx, signalID)
		prices = append(prices, price)
	}

	return prices
}

// GetPrice returns a price by signal id.
func (k Keeper) GetPrice(ctx sdk.Context, signalID string) types.Price {
	bz := ctx.KVStore(k.storeKey).Get(types.PriceStoreKey(signalID))
	if bz == nil {
		return types.Price{
			SignalID:  signalID,
			Status:    types.PRICE_STATUS_NOT_IN_CURRENT_FEEDS,
			Price:     0,
			Timestamp: ctx.BlockTime().Unix(),
		}
	}

	var price types.Price
	k.cdc.MustUnmarshal(bz, &price)

	return price
}

// SetPrice sets multiple prices.
func (k Keeper) SetPrices(ctx sdk.Context, prices []types.Price) {
	for _, price := range prices {
		k.SetPrice(ctx, price)
	}
}

// SetPRice sets a new price to the prices store or replace if price with the same signal id existed.
func (k Keeper) SetPrice(ctx sdk.Context, price types.Price) {
	ctx.KVStore(k.storeKey).Set(types.PriceStoreKey(price.SignalID), k.cdc.MustMarshal(&price))
}

// DeleteAllPrices deletes all prices.
func (k Keeper) DeleteAllPrices(ctx sdk.Context) {
	iterator := k.GetPricesIterator(ctx)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		key := iterator.Key()
		ctx.KVStore(k.storeKey).Delete(key)
	}
}

// GetValidatorPriceList gets a validator price by validator address.
func (k Keeper) GetValidatorPriceList(ctx sdk.Context, val sdk.ValAddress) (types.ValidatorPriceList, error) {
	bz := ctx.KVStore(k.storeKey).Get(types.ValidatorPriceListStoreKey(val))
	if bz == nil {
		return types.ValidatorPriceList{}, types.ErrValidatorPriceNotFound.Wrapf(
			"failed to get validator prices list for validator: %s",
			val.String(),
		)
	}

	var valPricesList types.ValidatorPriceList
	k.cdc.MustUnmarshal(bz, &valPricesList)

	return valPricesList, nil
}

// SetValidatorPriceList sets validator prices list.
func (k Keeper) SetValidatorPriceList(
	ctx sdk.Context,
	valAddress sdk.ValAddress,
	valPrices []types.ValidatorPrice,
) error {
	valPricesList := types.NewValidatorPriceList(
		valAddress,
		valPrices,
	)

	ctx.KVStore(k.storeKey).Set(types.ValidatorPriceListStoreKey(valAddress), k.cdc.MustMarshal(&valPricesList))

	return nil
}

// CalculatePrices calculates final prices for all supported feeds.
func (k Keeper) CalculatePrices(ctx sdk.Context) error {
	// get the current feeds
	currentFeeds := k.GetCurrentFeeds(ctx)

	var validatorsByPower []types.ValidatorInfo
	// iterate over bonded validators sorted by power
	err := k.stakingKeeper.IterateBondedValidatorsByPower(
		ctx,
		func(_ int64, val stakingtypes.ValidatorI) (stop bool) {
			operator, err := sdk.ValAddressFromBech32(val.GetOperator())
			if err != nil {
				return false
			}
			// get the status of the validator
			status := k.oracleKeeper.GetValidatorStatus(ctx, operator)
			if !status.IsActive {
				return false
			}
			// collect validator information
			validatorInfo := types.NewValidatorInfo(operator, val.GetTokens().Uint64(), status)
			validatorsByPower = append(validatorsByPower, validatorInfo)
			return false
		})
	if err != nil {
		return err
	}

	// collect all validator prices
	allValidatorPrices := make(map[string]map[string]types.ValidatorPrice)
	for _, val := range validatorsByPower {
		valPricesList, err := k.GetValidatorPriceList(ctx, val.Address)
		if err != nil {
			continue
		}

		valPricesMap := make(map[string]types.ValidatorPrice)
		for _, valPrice := range valPricesList.ValidatorPrices {
			if valPrice.SignalPriceStatus != types.SIGNAL_PRICE_STATUS_UNSPECIFIED {
				valPricesMap[valPrice.SignalID] = valPrice
			}
		}

		allValidatorPrices[val.Address.String()] = valPricesMap
	}

	params := k.GetParams(ctx)

	gracePeriod := params.GracePeriod
	tbt, err := k.stakingKeeper.TotalBondedTokens(ctx)
	if err != nil {
		return err
	}
	totalBondedToken := sdkmath.LegacyNewDecFromInt(tbt)
	priceQuorum, err := sdkmath.LegacyNewDecFromStr(params.PriceQuorum)
	if err != nil {
		return err
	}
	powerQuorum := totalBondedToken.Mul(priceQuorum).TruncateInt()
	// calculate prices for each feed
	for _, feed := range currentFeeds.Feeds {
		var validatorPriceInfos []types.ValidatorPriceInfo
		for _, valInfo := range validatorsByPower {
			valPrice := allValidatorPrices[valInfo.Address.String()][feed.SignalID]

			// check for miss report
			missReport, log := CheckMissReport(
				feed,
				currentFeeds.LastUpdateTimestamp,
				currentFeeds.LastUpdateBlock,
				valPrice,
				valInfo,
				ctx.BlockTime(),
				ctx.BlockHeight(),
				gracePeriod,
			)
			if missReport {
				fmt.Println(log)
				k.oracleKeeper.MissReport(ctx, valInfo.Address, ctx.BlockTime())
			}

			// check if the price is available
			havePrice := checkHavePrice(feed, valPrice, ctx.BlockTime())
			if havePrice {
				validatorPriceInfos = append(
					validatorPriceInfos, types.NewValidatorPriceInfo(
						valPrice.SignalPriceStatus,
						sdkmath.NewIntFromUint64(valInfo.Power),
						valPrice.Price,
						valPrice.Timestamp,
					),
				)
			}
		}

		// calculate the final price for the feed
		price, err := k.CalculatePrice(ctx, feed, validatorPriceInfos, powerQuorum)
		if err != nil {
			return err
		}

		// set the calculated price in the store
		k.SetPrice(ctx, price)

		// emit event for updated price
		emitEventUpdatePrice(ctx, price)
	}

	return nil
}

// CalculatePrice calculates the final price from validator prices and punishes validators who did not report.
func (k Keeper) CalculatePrice(
	ctx sdk.Context,
	feed types.Feed,
	validatorPriceInfos []types.ValidatorPriceInfo,
	powerQuorum sdkmath.Int,
) (types.Price, error) {
	totalPower, availablePower, _, unsupportedPower := types.CalculatePricesPowers(validatorPriceInfos)

	// If more than half of the total have unsupported price status, it returns an unknown signal id price status.
	if unsupportedPower.MulRaw(2).GT(totalPower) {
		return types.NewPrice(
			types.PRICE_STATUS_UNKNOWN_SIGNAL_ID,
			feed.SignalID,
			0,
			ctx.BlockTime().Unix(),
		), nil
	}

	// If the total power is less than price quorum percentage of the total bonded token
	// or less than half of total have available price status, it will not be calculated.
	if totalPower.LT(powerQuorum) || availablePower.MulRaw(2).LT(totalPower) {
		// else, it returns an price not ready price status.
		return types.NewPrice(
			types.PRICE_STATUS_NOT_READY,
			feed.SignalID,
			0,
			ctx.BlockTime().Unix(),
		), nil
	}

	price, err := types.MedianValidatorPriceInfos(validatorPriceInfos)
	if err != nil {
		// should not happen
		return types.Price{}, err
	}

	return types.NewPrice(
		types.PRICE_STATUS_AVAILABLE,
		feed.SignalID,
		price,
		ctx.BlockTime().Unix(),
	), nil
}

// CheckMissReport checks if a validator has missed a report based on the given parameters.
// And returns a boolean indication whether the validator has price feed.
func CheckMissReport(
	feed types.Feed,
	lastUpdateTimestamp int64,
	lastUpdateBlock int64,
	valPrice types.ValidatorPrice,
	valInfo types.ValidatorInfo,
	blockTime time.Time,
	blockHeight int64,
	gracePeriod int64,
) (bool, string) {
	var log string
	var missReport bool
	// Calculate the deadline time and block height for the validator to report.
	// During the grace period, if the block time exceeds ExpectedBlockTime, it will be capped at ExpectedBlockTime.
	// This prevents validator deactivation due to slower block times, as long as the block height remains within the threshold.
	deadlineTime := lastUpdateTimestamp + gracePeriod
	deadlineBlock := lastUpdateBlock + gracePeriod/types.ExpectedBlockTime

	log += "-----------------------------------" + "\n"
	log += "feed.Interval: " + strconv.FormatInt(feed.Interval, 10) + "\n"
	log += "lastUpdateTimestamp: " + strconv.FormatInt(lastUpdateTimestamp, 10) + "\n"
	log += "lastUpdateBlock: " + strconv.FormatInt(lastUpdateBlock, 10) + "\n"
	log += "valPrice.Timestamp: " + strconv.FormatInt(valPrice.Timestamp, 10) + "\n"
	log += "valInfo.Status.Since: " + strconv.FormatInt(valInfo.Status.Since.Unix(), 10) + "\n"
	log += "blockTime: " + strconv.FormatInt(blockTime.Unix(), 10) + "\n"
	log += "blockHeight: " + strconv.FormatInt(blockHeight, 10) + "\n"
	log += "gracePeriod: " + strconv.FormatInt(gracePeriod, 10) + "\n"
	log += "deadlineTime: " + strconv.FormatInt(deadlineTime, 10) + "\n"
	log += "deadlineBlock: " + strconv.FormatInt(deadlineBlock, 10) + "\n"
	log += "-----------------------------------" + "\n"

	// Extend deadline if the validator just became active.
	if valInfo.Status.Since.Unix()+gracePeriod > deadlineTime {
		deadlineTime = valInfo.Status.Since.Unix() + gracePeriod

		log += "valInfo.Status.Since.Unix()+gracePeriod > deadlineTime" + "\n"
		log += "deadlineTime: " + strconv.FormatInt(deadlineTime, 10) + "\n"
	}

	// Extend deadline if the validator has a valid price within the feed interval.
	if valPrice.SignalPriceStatus != types.SIGNAL_PRICE_STATUS_UNSPECIFIED {
		log += "valPrice.SignalPriceStatus != types.SIGNAL_PRICE_STATUS_UNSPECIFIED" + "\n"
		// Extend deadline time based on the price timestamp.
		if valPrice.Timestamp+feed.Interval > deadlineTime {
			deadlineTime = valPrice.Timestamp + feed.Interval

			log += "valPrice.Timestamp+feed.Interval > deadlineTime" + "\n"
			log += "deadlineTime: " + strconv.FormatInt(deadlineTime, 10) + "\n"
		}

		// Extend deadline block based on the price block height.
		if valPrice.BlockHeight+feed.Interval/types.ExpectedBlockTime > deadlineBlock {
			deadlineBlock = valPrice.BlockHeight + feed.Interval/types.ExpectedBlockTime

			log += "valPrice.BlockHeight+feed.Interval/types.ExpectedBlockTime > deadlineBlock" + "\n"
			log += "deadlineBlock: " + strconv.FormatInt(deadlineBlock, 10) + "\n"
		}
	}

	log += "-----------------------------------" + "\n"

	// Determine if the validator has missed the report based on the deadline time and block height.
	missReport = deadlineTime < blockTime.Unix() && deadlineBlock < blockHeight

	log += "missReport: " + strconv.FormatBool(missReport) + "\n"
	log += "deadlineTime" + strconv.FormatInt(deadlineTime, 10) + "\n"
	log += "blockTime.Unix(): " + strconv.FormatInt(blockTime.Unix(), 10) + "\n"
	log += "deadlineBlock: " + strconv.FormatInt(deadlineBlock, 10) + "\n"
	log += "blockHeight: " + strconv.FormatInt(blockHeight, 10) + "\n"
	log += "deadlineTime < blockTime.Unix(): " + strconv.FormatBool(deadlineTime < blockTime.Unix()) + "\n"
	log += "deadlineBlock < blockHeight: " + strconv.FormatBool(deadlineBlock < blockHeight) + "\n"
	log += "-----------------------------------" + "\n"
	return missReport, log
}

// checkHavePrice checks if a validator has a price feed within interval range.
func checkHavePrice(
	feed types.Feed,
	valPrice types.ValidatorPrice,
	blockTime time.Time,
) bool {
	if valPrice.SignalPriceStatus != types.SIGNAL_PRICE_STATUS_UNSPECIFIED &&
		valPrice.Timestamp >= blockTime.Unix()-feed.Interval {
		return true
	}

	return false
}

// ValidateValidatorRequiredToSend validates validator is required for price submission.
func (k Keeper) ValidateValidatorRequiredToSend(
	ctx sdk.Context,
	val sdk.ValAddress,
) error {
	isValid := k.IsBondedValidator(ctx, val)
	if !isValid {
		return types.ErrNotBondedValidator
	}

	status := k.oracleKeeper.GetValidatorStatus(ctx, val)
	if !status.IsActive {
		return types.ErrOracleStatusNotActive.Wrapf("val: %s", val.String())
	}

	return nil
}
