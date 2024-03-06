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

func (k Keeper) GetPrice(ctx sdk.Context, symbol string) (types.Price, error) {
	bz := ctx.KVStore(k.storeKey).Get(types.PriceStoreKey(symbol))
	if bz == nil {
		return types.Price{}, types.ErrPriceNotFound.Wrapf("failed to get price for symbol: %s", symbol)
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
	ctx.KVStore(k.storeKey).Set(types.PriceStoreKey(price.Symbol), k.cdc.MustMarshal(&price))
}

func (k Keeper) DeletePrice(ctx sdk.Context, symbol string) {
	k.DeletePriceValidators(ctx, symbol)
	ctx.KVStore(k.storeKey).Delete(types.PriceStoreKey(symbol))
}

func (k Keeper) CalculatePrice(ctx sdk.Context, symbol types.Symbol) (types.Price, error) {
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
				priceVal, err := k.GetPriceValidator(ctx, symbol.Symbol, address)

				if err == nil {
					// if timestamp of price is in acception period, append it
					if priceVal.Timestamp >= blockTime.Unix()-symbol.Interval {
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

				if symbol.LastIntervalUpdateTimestamp+transitionTime > lastTime {
					lastTime = symbol.LastIntervalUpdateTimestamp + transitionTime
				}

				// deactivate if last time of action is too old
				if lastTime < blockTime.Unix()-symbol.Interval {
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
			Symbol:      symbol.Symbol,
			Price:       0,
			Timestamp:   ctx.BlockTime().Unix(),
		}, nil
	} else if unavailable > total/2 || available < total/2 {
		return types.Price{
			PriceOption: types.PriceOptionUnavailable,
			Symbol:      symbol.Symbol,
			Price:       0,
			Timestamp:   ctx.BlockTime().Unix(),
		}, nil
	}

	price := types.CalculateMedianPriceFeedInfo(types.FilterPfInfos(pfInfos, types.PriceOptionAvailable))

	return types.Price{
		PriceOption: types.PriceOptionAvailable,
		Symbol:      symbol.Symbol,
		Price:       price,
		Timestamp:   ctx.BlockTime().Unix(),
	}, nil
}

// ==================================
// Price validator
// ==================================

func (k Keeper) GetPriceValidatorsIterator(ctx sdk.Context, symbol string) sdk.Iterator {
	return sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.PriceValidatorsStoreKey(symbol))
}

func (k Keeper) GetPriceValidators(ctx sdk.Context, symbol string) (priceVals []types.PriceValidator) {
	iterator := k.GetPriceValidatorsIterator(ctx, symbol)
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

func (k Keeper) GetPriceValidator(ctx sdk.Context, symbol string, val sdk.ValAddress) (types.PriceValidator, error) {
	bz := ctx.KVStore(k.storeKey).Get(types.PriceValidatorStoreKey(symbol, val))
	if bz == nil {
		return types.PriceValidator{}, types.ErrPriceValidatorNotFound.Wrapf(
			"failed to get price validator for symbol: %s, validator: %s",
			symbol,
			val.String(),
		)
	}

	var priceVal types.PriceValidator
	k.cdc.MustUnmarshal(bz, &priceVal)

	return priceVal, nil
}

func (k Keeper) SetPriceValidators(ctx sdk.Context, priceVals []types.PriceValidator) error {
	for _, priceVal := range priceVals {
		err := k.SetPriceValidator(ctx, priceVal)
		if err != nil {
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
		Set(types.PriceValidatorStoreKey(priceVal.Symbol, valAddress), k.cdc.MustMarshal(&priceVal))

	return nil
}

func (k Keeper) DeletePriceValidators(ctx sdk.Context, symbol string) {
	iterator := k.GetPriceValidatorsIterator(ctx, symbol)

	defer func(iterator sdk.Iterator) {
		_ = iterator.Close()
	}(iterator)

	for ; iterator.Valid(); iterator.Next() {
		ctx.KVStore(k.storeKey).Delete(iterator.Key())
	}
}

func (k Keeper) DeletePriceValidator(ctx sdk.Context, symbol string, val sdk.ValAddress) {
	ctx.KVStore(k.storeKey).Delete(types.PriceValidatorStoreKey(symbol, val))
}
