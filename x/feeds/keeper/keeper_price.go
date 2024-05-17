package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/bandprotocol/chain/v2/x/feeds/types"
)

// ==================================
// Price
// ==================================

// GetPricesIterator returns an iterator for prices store.
func (k Keeper) GetPricesIterator(ctx sdk.Context) sdk.Iterator {
	return sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.PriceStoreKeyPrefix)
}

// GetPrices returns a list of all prices.
func (k Keeper) GetPrices(ctx sdk.Context) (prices []types.Price) {
	iterator := k.GetPricesIterator(ctx)
	defer func(iterator sdk.Iterator) {
		_ = iterator.Close()
	}(iterator)

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
	k.DeleteValidatorPrices(ctx, signalID)
	ctx.KVStore(k.storeKey).Delete(types.PriceStoreKey(signalID))
}

// CalculatePrices calculates final prices for all supported feeds.
func (k Keeper) CalculatePrices(ctx sdk.Context) {
	supportedFeeds := k.GetSupportedFeeds(ctx)
	for _, feed := range supportedFeeds.Feeds {
		price, err := k.CalculatePrice(ctx, feed, supportedFeeds.LastUpdateTimestamp, supportedFeeds.LastUpdateBlock)
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

// CalculatePrice calculates final price from price-validator and punish validators those did not report.
func (k Keeper) CalculatePrice(
	ctx sdk.Context,
	feed types.Feed,
	lastUpdateTimestamp int64,
	lastUpdateBlock int64,
) (types.Price, error) {
	var priceFeedInfos []types.PriceFeedInfo
	blockTime := ctx.BlockTime()
	transitionTime := k.GetParams(ctx).TransitionTime

	k.stakingKeeper.IterateBondedValidatorsByPower(
		ctx,
		func(idx int64, val stakingtypes.ValidatorI) (stop bool) {
			address := val.GetOperator()
			power := val.GetTokens().Uint64()
			status := k.oracleKeeper.GetValidatorStatus(ctx, address)

			if status.IsActive {
				timePerBlock := int64(3) // assume block time will be 3 second.
				lastTime := lastUpdateTimestamp + transitionTime
				lastBlock := lastUpdateBlock + transitionTime/timePerBlock

				if status.Since.Unix()+transitionTime > lastTime {
					lastTime = status.Since.Unix() + transitionTime
				}

				priceVal, err := k.GetValidatorPrice(ctx, feed.SignalID, address)
				if err == nil {
					// if timestamp of price is in acception period, append it
					if priceVal.Timestamp >= blockTime.Unix()-feed.Interval {
						priceFeedInfos = append(
							priceFeedInfos, types.PriceFeedInfo{
								PriceStatus: priceVal.PriceStatus,
								Price:       priceVal.Price,
								Power:       power,
								Deviation:   0,
								Timestamp:   priceVal.Timestamp,
								Index:       idx,
							},
						)
					}

					if priceVal.Timestamp+feed.Interval > lastTime {
						lastTime = priceVal.Timestamp + feed.Interval
					}

					if priceVal.BlockHeight+feed.Interval/timePerBlock > lastBlock {
						lastBlock = priceVal.BlockHeight + feed.Interval/timePerBlock
					}
				}

				// deactivate if last action is too old.
				if lastTime < blockTime.Unix() && lastBlock < ctx.BlockHeight() {
					k.oracleKeeper.MissReport(ctx, address, blockTime)
				}
			}

			return false
		},
	)

	n := len(priceFeedInfos)
	if n == 0 {
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

// ==================================
// Validator Price
// ==================================

// GetValidatorPricesIterator returns an iterator for price-validators store.
func (k Keeper) GetValidatorPricesIterator(ctx sdk.Context, signalID string) sdk.Iterator {
	return sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.ValidatorPricesStoreKey(signalID))
}

// GetValidatorPrices gets a list of all price-validators.
func (k Keeper) GetValidatorPrices(ctx sdk.Context, signalID string) (priceVals []types.ValidatorPrice) {
	iterator := k.GetValidatorPricesIterator(ctx, signalID)
	defer func(iterator sdk.Iterator) {
		_ = iterator.Close()
	}(iterator)

	for ; iterator.Valid(); iterator.Next() {
		var priceVal types.ValidatorPrice
		k.cdc.MustUnmarshal(iterator.Value(), &priceVal)
		priceVals = append(priceVals, priceVal)
	}

	return priceVals
}

// GetValidatorPrice gets a price-validator by signal id.
func (k Keeper) GetValidatorPrice(ctx sdk.Context, signalID string, val sdk.ValAddress) (types.ValidatorPrice, error) {
	bz := ctx.KVStore(k.storeKey).Get(types.ValidatorPriceStoreKey(signalID, val))
	if bz == nil {
		return types.ValidatorPrice{}, types.ErrValidatorPriceNotFound.Wrapf(
			"failed to get validator price for signal id: %s, validator: %s",
			signalID,
			val.String(),
		)
	}

	var priceVal types.ValidatorPrice
	k.cdc.MustUnmarshal(bz, &priceVal)

	return priceVal, nil
}

// SetValidatorPrices sets multiple price-validator.
func (k Keeper) SetValidatorPrices(ctx sdk.Context, priceVals []types.ValidatorPrice) error {
	for _, priceVal := range priceVals {
		if err := k.SetValidatorPrice(ctx, priceVal); err != nil {
			return err
		}
	}
	return nil
}

// SetValidatorPrice sets a new price-validator or replace if price-validator with the same signal id and validator address existed.
func (k Keeper) SetValidatorPrice(ctx sdk.Context, priceVal types.ValidatorPrice) error {
	valAddress, err := sdk.ValAddressFromBech32(priceVal.Validator)
	if err != nil {
		return err
	}

	ctx.KVStore(k.storeKey).
		Set(types.ValidatorPriceStoreKey(priceVal.SignalID, valAddress), k.cdc.MustMarshal(&priceVal))

	return nil
}

// DeleteValidatorPrices deletes all price-validator of specified signal id.
func (k Keeper) DeleteValidatorPrices(ctx sdk.Context, signalID string) {
	iterator := k.GetValidatorPricesIterator(ctx, signalID)
	defer func(iterator sdk.Iterator) {
		_ = iterator.Close()
	}(iterator)

	for ; iterator.Valid(); iterator.Next() {
		ctx.KVStore(k.storeKey).Delete(iterator.Key())
	}
}

// DeleteValidatorPrice deletes a price-validators.
func (k Keeper) DeleteValidatorPrice(ctx sdk.Context, signalID string, val sdk.ValAddress) {
	ctx.KVStore(k.storeKey).Delete(types.ValidatorPriceStoreKey(signalID, val))
}
