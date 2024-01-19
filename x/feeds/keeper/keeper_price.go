package keeper

import (
	"time"

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
	defer iterator.Close()

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

func (k Keeper) CalculatePrice(ctx sdk.Context, symbol types.Symbol, deactivate bool) (types.Price, error) {
	var prices []uint64
	var powers []uint64
	totalPower := uint64(0)
	blockTime := ctx.BlockTime()

	// TODO: confirm if it's sorted by power already
	k.stakingKeeper.IterateBondedValidatorsByPower(
		ctx,
		func(_ int64, val stakingtypes.ValidatorI) (stop bool) {
			address := val.GetOperator()
			power := val.GetTokens().Uint64()
			status := k.oracleKeeper.GetValidatorStatus(ctx, address)

			if status.IsActive {
				priceVal, err := k.GetPriceValidator(ctx, symbol.Symbol, address)
				// TODO:
				// - check if it make sense?
				if err != nil || priceVal.Timestamp < blockTime.Unix()-int64(symbol.Interval)*2 {
					// deactivate only who that has activated before interval * 2
					if status.Since.Add(time.Duration(symbol.Interval*2) * time.Second).Before(blockTime) {
						k.oracleKeeper.MissReport(ctx, address, blockTime)
					}
				} else {
					prices = append(prices, priceVal.Price)
					powers = append(powers, power)
					totalPower += power
				}
			}

			return false
		},
	)

	n := len(prices)
	if n == 0 {
		return types.Price{}, types.ErrNotEnoughPriceValidator
	}

	price := prices[0]
	currentPower := powers[0]
	for i := 1; i < n; i++ {
		if currentPower >= totalPower/2 {
			break
		}
		currentPower += powers[i]
		price = prices[i]
	}

	return types.Price{
		Symbol:    symbol.Symbol,
		Price:     price,
		Timestamp: ctx.BlockTime().Unix(),
	}, nil
}

// ==================================go
// Price validator
// ==================================

func (k Keeper) GetPriceValidatorsIterator(ctx sdk.Context, symbol string) sdk.Iterator {
	return sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.PriceValidatorsStoreKey(symbol))
}

func (k Keeper) GetPriceValidators(ctx sdk.Context, symbol string) (priceVals []types.PriceValidator) {
	iterator := k.GetPriceValidatorsIterator(ctx, symbol)
	defer iterator.Close()

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
		return types.PriceValidator{}, types.ErrPriceNotFound.Wrapf(
			"failed to get price validator for symbol: %s, validator: %s",
			symbol,
			val.String(),
		)
	}

	var priceVal types.PriceValidator
	k.cdc.MustUnmarshal(bz, &priceVal)

	return priceVal, nil
}

func (k Keeper) SetPriceValidators(ctx sdk.Context, priceVals []types.PriceValidator) {
	for _, priceVal := range priceVals {
		k.SetPriceValidator(ctx, priceVal)
	}
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
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		ctx.KVStore(k.storeKey).Delete(iterator.Key())
	}
}

func (k Keeper) DeletePriceValidator(ctx sdk.Context, symbol string, val sdk.ValAddress) {
	ctx.KVStore(k.storeKey).Delete(types.PriceValidatorStoreKey(symbol, val))
}
