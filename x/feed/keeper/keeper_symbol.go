package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/feed/types"
)

func (k Keeper) GetSymbolsIterator(ctx sdk.Context) sdk.Iterator {
	return sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.SymbolStoreKeyPrefix)
}

func (k Keeper) GetSymbols(ctx sdk.Context) (symbols []types.Symbol) {
	iterator := k.GetSymbolsIterator(ctx)
	defer iterator.Close()

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
