package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/bandprotocol/chain/v2/x/feeds/types"
)

// ==================================
// Price
// ==================================

func (k Keeper) GetPricesIterator(ctx sdk.Context) sdk.Iterator {
	return sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.PriceStoreKeyPrefix)
}

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

func (k Keeper) GetPrice(ctx sdk.Context, signalID string) (types.Price, error) {
	bz := ctx.KVStore(k.storeKey).Get(types.PriceStoreKey(signalID))
	if bz == nil {
		return types.Price{}, types.ErrPriceNotFound.Wrapf("failed to get price for signal id: %s", signalID)
	}

	var price types.Price
	k.cdc.MustUnmarshal(bz, &price)

	return price, nil
}

func (k Keeper) SetPrices(ctx sdk.Context, prices []types.Price) {
	for _, price := range prices {
		k.SetPrice(ctx, price)
	}
}

func (k Keeper) SetPrice(ctx sdk.Context, price types.Price) {
	ctx.KVStore(k.storeKey).Set(types.PriceStoreKey(price.SignalID), k.cdc.MustMarshal(&price))
}

func (k Keeper) DeletePrice(ctx sdk.Context, signalID string) {
	k.DeletePriceValidators(ctx, signalID)
	ctx.KVStore(k.storeKey).Delete(types.PriceStoreKey(signalID))
}

func (k Keeper) CalculatePrice(ctx sdk.Context, feed types.Feed) (types.Price, error) {
	var pfInfos []types.PriceFeedInfo
	blockTime := ctx.BlockTime()
	transitionTime := k.GetParams(ctx).TransitionTime

	k.stakingKeeper.IterateBondedValidatorsByPower(
		ctx,
		func(idx int64, val stakingtypes.ValidatorI) (stop bool) {
			address := val.GetOperator()
			power := val.GetTokens().Uint64()
			status := k.oracleKeeper.GetValidatorStatus(ctx, address)

			if status.IsActive {
				lastTime := status.Since.Unix()
				priceVal, err := k.GetPriceValidator(ctx, feed.SignalID, address)

				if err == nil {
					// if timestamp of price is in acception period, append it
					if priceVal.Timestamp >= blockTime.Unix()-feed.Interval {
						pfInfos = append(
							pfInfos, types.PriceFeedInfo{
								PriceOption: priceVal.PriceOption,
								Price:       priceVal.Price,
								Power:       power,
								Deviation:   0,
								Timestamp:   priceVal.Timestamp,
								Index:       idx,
							},
						)
					}

					// update last time of action
					if priceVal.Timestamp > lastTime {
						lastTime = priceVal.Timestamp
					}
				}

				if feed.LastIntervalUpdateTimestamp+transitionTime > lastTime {
					lastTime = feed.LastIntervalUpdateTimestamp + transitionTime
				}

				// deactivate if last time of action is too old
				if lastTime < blockTime.Unix()-feed.Interval {
					k.oracleKeeper.MissReport(ctx, address, blockTime)
				}
			}

			return false
		},
	)

	n := len(pfInfos)
	if n == 0 {
		return types.Price{}, types.ErrNotEnoughPriceValidator
	}

	// TODO: check final logic later
	// check if the price is available
	total, available, unavailable, unsupported := types.CalPricesPowers(pfInfos)
	if unsupported > total/2 {
		return types.Price{
			PriceOption: types.PriceOptionUnsupported,
			SignalID:    feed.SignalID,
			Price:       0,
			Timestamp:   ctx.BlockTime().Unix(),
		}, nil
	} else if unavailable > total/2 || available < total/2 {
		return types.Price{
			PriceOption: types.PriceOptionUnavailable,
			SignalID:    feed.SignalID,
			Price:       0,
			Timestamp:   ctx.BlockTime().Unix(),
		}, nil
	}

	price := types.CalculateMedianPriceFeedInfo(types.FilterPfInfos(pfInfos, types.PriceOptionAvailable))

	return types.Price{
		PriceOption: types.PriceOptionAvailable,
		SignalID:    feed.SignalID,
		Price:       price,
		Timestamp:   ctx.BlockTime().Unix(),
	}, nil
}

// ==================================
// Price validator
// ==================================

func (k Keeper) GetPriceValidatorsIterator(ctx sdk.Context, signalID string) sdk.Iterator {
	return sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.PriceValidatorsStoreKey(signalID))
}

func (k Keeper) GetPriceValidators(ctx sdk.Context, signalID string) (priceVals []types.PriceValidator) {
	iterator := k.GetPriceValidatorsIterator(ctx, signalID)
	defer func(iterator sdk.Iterator) {
		_ = iterator.Close()
	}(iterator)

	for ; iterator.Valid(); iterator.Next() {
		var priceVal types.PriceValidator
		k.cdc.MustUnmarshal(iterator.Value(), &priceVal)
		priceVals = append(priceVals, priceVal)
	}

	return priceVals
}

func (k Keeper) GetPriceValidator(ctx sdk.Context, signalID string, val sdk.ValAddress) (types.PriceValidator, error) {
	bz := ctx.KVStore(k.storeKey).Get(types.PriceValidatorStoreKey(signalID, val))
	if bz == nil {
		return types.PriceValidator{}, types.ErrPriceValidatorNotFound.Wrapf(
			"failed to get price validator for signal id: %s, validator: %s",
			signalID,
			val.String(),
		)
	}

	var priceVal types.PriceValidator
	k.cdc.MustUnmarshal(bz, &priceVal)

	return priceVal, nil
}

func (k Keeper) SetPriceValidators(ctx sdk.Context, priceVals []types.PriceValidator) error {
	for _, priceVal := range priceVals {
		if err := k.SetPriceValidator(ctx, priceVal); err != nil {
			return err
		}
	}
	return nil
}

func (k Keeper) SetPriceValidator(ctx sdk.Context, priceVal types.PriceValidator) error {
	valAddress, err := sdk.ValAddressFromBech32(priceVal.Validator)
	if err != nil {
		return err
	}

	ctx.KVStore(k.storeKey).
		Set(types.PriceValidatorStoreKey(priceVal.SignalID, valAddress), k.cdc.MustMarshal(&priceVal))

	return nil
}

func (k Keeper) DeletePriceValidators(ctx sdk.Context, signalID string) {
	iterator := k.GetPriceValidatorsIterator(ctx, signalID)

	defer func(iterator sdk.Iterator) {
		_ = iterator.Close()
	}(iterator)

	for ; iterator.Valid(); iterator.Next() {
		ctx.KVStore(k.storeKey).Delete(iterator.Key())
	}
}

func (k Keeper) DeletePriceValidator(ctx sdk.Context, signalID string, val sdk.ValAddress) {
	ctx.KVStore(k.storeKey).Delete(types.PriceValidatorStoreKey(signalID, val))
}
