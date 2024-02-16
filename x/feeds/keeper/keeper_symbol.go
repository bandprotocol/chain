package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/feeds/types"
)

func (k Keeper) GetSymbolsIterator(ctx sdk.Context) sdk.Iterator {
	return sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.SymbolStoreKeyPrefix)
}

func (k Keeper) GetSymbols(ctx sdk.Context) (symbols []types.Symbol) {
	iterator := k.GetSymbolsIterator(ctx)
	defer func(iterator sdk.Iterator) {
		_ = iterator.Close()
	}(iterator)

	for ; iterator.Valid(); iterator.Next() {
		var symbol types.Symbol
		k.cdc.MustUnmarshal(iterator.Value(), &symbol)
		symbols = append(symbols, symbol)
	}

	return symbols
}

func (k Keeper) GetSymbol(ctx sdk.Context, symbol string) (types.Symbol, error) {
	bz := ctx.KVStore(k.storeKey).Get(types.SymbolStoreKey(symbol))
	if bz == nil {
		return types.Symbol{}, types.ErrPriceNotFound.Wrapf("failed to get symbol detail for symbol: %s", symbol)
	}

	var s types.Symbol
	k.cdc.MustUnmarshal(bz, &s)

	return s, nil
}

func (k Keeper) SetSymbols(ctx sdk.Context, symbols []types.Symbol) {
	for _, symbol := range symbols {
		k.SetSymbol(ctx, symbol)
	}
}

func (k Keeper) SetSymbol(ctx sdk.Context, symbol types.Symbol) {
	ctx.KVStore(k.storeKey).Set(types.SymbolStoreKey(symbol.Symbol), k.cdc.MustMarshal(&symbol))
}

func (k Keeper) DeleteSymbol(ctx sdk.Context, symbol string) {
	k.DeletePrice(ctx, symbol)
	ctx.KVStore(k.storeKey).Delete(types.SymbolStoreKey(symbol))
}

func (k Keeper) SetSymbolByPowerIndex(ctx sdk.Context, symbol types.Symbol) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetSymbolsByPowerIndexKey(symbol.Symbol, symbol.Power), k.cdc.MustMarshal(&symbol))
}

func (k Keeper) DeleteSymbolByPowerIndex(ctx sdk.Context, symbol types.Symbol) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetSymbolsByPowerIndexKey(symbol.Symbol, symbol.Power))
}

// GetSupportedSymbolsByPower gets the current group of bonded validators sorted by power-rank
func (k Keeper) GetSupportedSymbolsByPower(ctx sdk.Context) []types.Symbol {
	maxSymbols := k.GetParams(ctx).MaxSupportedSymbol
	symbols := make([]types.Symbol, maxSymbols)

	iterator := k.SymbolsPowerStoreIterator(ctx)
	defer func(iterator sdk.Iterator) {
		_ = iterator.Close()
	}(iterator)

	i := 0
	for ; iterator.Valid() && i < int(maxSymbols); iterator.Next() {
		var s types.Symbol
		bz := iterator.Value()
		k.cdc.MustUnmarshal(bz, &s)

		symbols[i] = s
		i++
	}

	return symbols[:i] // trim
}

func (k Keeper) SymbolsPowerStoreIterator(ctx sdk.Context) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return sdk.KVStoreReversePrefixIterator(store, types.SymbolsByPowerIndexKey)
}
