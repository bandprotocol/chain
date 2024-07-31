package keeper

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/bandprotocol/chain/v2/x/feeds/types"
)

// GetPricesIterator returns an iterator for prices store.
func (k Keeper) GetPricesIterator(ctx sdk.Context) sdk.Iterator {
	return sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.PriceStoreKeyPrefix)
}

// GetPrices returns a list of all prices.
func (k Keeper) GetPrices(ctx sdk.Context) (prices []types.Price) {
	iterator := k.GetPricesIterator(ctx)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var price types.Price
		k.cdc.MustUnmarshal(iterator.Value(), &price)
		prices = append(prices, price)
	}

	return prices
}

// GetPrice returns a price by signal id.
func (k Keeper) GetPrice(ctx sdk.Context, signalID string) (types.Price, error) {
	bz := ctx.KVStore(k.storeKey).Get(types.PriceStoreKey(signalID))
	if bz == nil {
		return types.Price{}, types.ErrPriceNotFound.Wrapf("failed to get price for signal id: %s", signalID)
	}

	var price types.Price
	k.cdc.MustUnmarshal(bz, &price)

	return price, nil
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

// DeletePrice deletes a price by signal id.
func (k Keeper) DeletePrice(ctx sdk.Context, signalID string) {
	ctx.KVStore(k.storeKey).Delete(types.PriceStoreKey(signalID))
}

// CalculatePrices calculates final prices for all supported feeds.
func (k Keeper) CalculatePrices(ctx sdk.Context) {
	currentFeeds := k.GetCurrentFeeds(ctx)

	var validatorsByPower []types.ValidatorInfo
	k.stakingKeeper.IterateBondedValidatorsByPower(
		ctx,
		func(idx int64, val stakingtypes.ValidatorI) (stop bool) {
			status := k.oracleKeeper.GetValidatorStatus(ctx, val.GetOperator())
			if !status.IsActive {
				return false
			}
			validatorInfo := types.ValidatorInfo{
				Index:   idx,
				Address: val.GetOperator(),
				Power:   val.GetTokens().Uint64(),
				Status:  status,
			}
			validatorsByPower = append(validatorsByPower, validatorInfo)
			return false
		})

	allValidatorPrices := make(map[string]map[string]types.ValidatorPrice)
	for _, val := range validatorsByPower {
		valPricesList, err := k.GetValidatorPriceList(ctx, val.Address)
		if err != nil {
			continue
		}

		valPricesMap := make(map[string]types.ValidatorPrice)
		for _, valPrice := range valPricesList.ValidatorPrices {
			if valPrice.SignalID != "" {
				valPricesMap[valPrice.SignalID] = valPrice
			}
		}

		allValidatorPrices[val.Address.String()] = valPricesMap
	}

	for _, feed := range currentFeeds.Feeds {
		var priceFeedInfos []types.PriceFeedInfo
		for _, valInfo := range validatorsByPower {
			valPrice := allValidatorPrices[valInfo.Address.String()][feed.SignalID]

			missReport, havePrice := CheckMissReport(
				feed,
				currentFeeds.LastUpdateTimestamp,
				currentFeeds.LastUpdateBlock,
				valPrice,
				valInfo,
				ctx.BlockTime(),
				ctx.BlockHeight(),
				k.GetParams(ctx).GracePeriod,
			)
			if missReport {
				k.oracleKeeper.MissReport(ctx, valInfo.Address, ctx.BlockTime())
			}

			if havePrice {
				priceFeedInfos = append(
					priceFeedInfos, types.PriceFeedInfo{
						PriceStatus: valPrice.PriceStatus,
						Price:       valPrice.Price,
						Power:       valInfo.Power,
						Timestamp:   valPrice.Timestamp,
						Index:       valInfo.Index,
					},
				)
			}
		}

		price, err := k.CalculatePrice(ctx, feed, priceFeedInfos)
		if err != nil {
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeCalculatePriceFailed,
					sdk.NewAttribute(types.AttributeKeySignalID, feed.SignalID),
					sdk.NewAttribute(types.AttributeKeyErrorMessage, err.Error()),
				),
			)
			continue
		}

		k.SetPrice(ctx, price)

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeUpdatePrice,
				sdk.NewAttribute(types.AttributeKeySignalID, price.SignalID),
				sdk.NewAttribute(types.AttributeKeyPrice, fmt.Sprintf("%d", price.Price)),
				sdk.NewAttribute(types.AttributeKeyTimestamp, fmt.Sprintf("%d", price.Timestamp)),
			),
		)
	}
}

// CalculatePrice calculates the final price from validator prices and punishes validators who did not report.
func (k Keeper) CalculatePrice(
	ctx sdk.Context,
	feed types.Feed,
	priceFeedInfos []types.PriceFeedInfo,
) (types.Price, error) {
	if len(priceFeedInfos) == 0 {
		return types.Price{}, types.ErrNotEnoughValidatorPrice
	}

	totalPower, availablePower, _, unsupportedPower := types.CalculatePricesPowers(priceFeedInfos)

	// If more than half of the total have unsupported price status, it returns an unsupported price status.
	if unsupportedPower > totalPower/2 {
		return types.Price{
			PriceStatus: types.PriceStatusUnsupported,
			SignalID:    feed.SignalID,
			Price:       0,
			Timestamp:   ctx.BlockTime().Unix(),
		}, nil
	}

	// If less than half of total have available price status, it returns an unavailable price status.
	if availablePower < totalPower/2 {
		return types.Price{
			PriceStatus: types.PriceStatusUnavailable,
			SignalID:    feed.SignalID,
			Price:       0,
			Timestamp:   ctx.BlockTime().Unix(),
		}, nil
	}

	price, err := types.CalculateMedianPriceFeedInfo(
		types.FilterPriceFeedInfos(priceFeedInfos, types.PriceStatusAvailable),
	)
	if err != nil {
		return types.Price{}, err
	}

	return types.Price{
		PriceStatus: types.PriceStatusAvailable,
		SignalID:    feed.SignalID,
		Price:       price,
		Timestamp:   ctx.BlockTime().Unix(),
	}, nil
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
) (missReport bool, havePrice bool) {
	// During the grace period, if the block time exceeds MaximumGuaranteeBlockTime, it will be capped at MaximumGuaranteeBlockTime.
	// This means that in cases of slow block time, the validator will not be deactivated
	// as long as the block height does not exceed the equivalent of assumed MaximumGuaranteeBlockTime of block time.
	lastTime := lastUpdateTimestamp + gracePeriod
	lastBlock := lastUpdateBlock + gracePeriod/types.MaximumGuaranteeBlockTime

	if valInfo.Status.Since.Unix()+gracePeriod > lastTime {
		lastTime = valInfo.Status.Since.Unix() + gracePeriod
	}

	if valPrice.SignalID != "" {
		// Append valid price feed info if within the acceptance period
		if valPrice.Timestamp >= blockTime.Unix()-feed.Interval {
			havePrice = true
		}

		if valPrice.Timestamp+feed.Interval > lastTime {
			lastTime = valPrice.Timestamp + feed.Interval
		}

		if valPrice.BlockHeight+feed.Interval/types.MaximumGuaranteeBlockTime > lastBlock {
			lastBlock = valPrice.BlockHeight + feed.Interval/types.MaximumGuaranteeBlockTime
		}
	}

	// Determine if the last action is too old, indicating a missed report
	missReport = lastTime < blockTime.Unix() && lastBlock < blockHeight
	return
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

// SetValidatorPrices sets validator prices list.
func (k Keeper) SetValidatorPriceList(
	ctx sdk.Context,
	valAddress sdk.ValAddress,
	valPrices []types.ValidatorPrice,
) error {
	valPricesList := types.ValidatorPriceList{
		Validator:       valAddress.String(),
		ValidatorPrices: valPrices,
	}

	ctx.KVStore(k.storeKey).Set(types.ValidatorPriceListStoreKey(valAddress), k.cdc.MustMarshal(&valPricesList))

	return nil
}
