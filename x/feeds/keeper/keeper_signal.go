package keeper

import (
	"github.com/bandprotocol/chain/v2/x/feeds/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) GetDelegatorDelegationsSum(ctx sdk.Context, delegator sdk.AccAddress) (sum uint64) {
	delegations := k.stakingKeeper.GetDelegatorDelegations(ctx, delegator, 100)
	for _, del := range delegations {
		val, found := k.stakingKeeper.GetValidator(ctx, del.GetValidatorAddr())
		if found {
			sum = sum + val.TokensFromShares(del.Shares).TruncateInt().Uint64()
		}
	}
	return
}

func (k Keeper) GetDelegatorSignal(ctx sdk.Context, delegator sdk.AccAddress) (types.Signal, bool) {
	bz := ctx.KVStore(k.storeKey).Get(types.DelegatorSignalStoreKey(delegator))
	if bz == nil {
		return types.Signal{}, false
	}

	var s types.Signal
	k.cdc.MustUnmarshal(bz, &s)

	return s, true
}

func (k Keeper) SetDelegatorSignal(ctx sdk.Context, delegator sdk.AccAddress, signal types.Signal) {
	ctx.KVStore(k.storeKey).Set(types.DelegatorSignalStoreKey(delegator), k.cdc.MustMarshal(&signal))
}

func (k Keeper) GetSymbolPower(ctx sdk.Context, symbol string) uint64 {
	bz := ctx.KVStore(k.storeKey).Get(types.SymbolPowerStoreKey(symbol))
	return sdk.BigEndianToUint64(bz)
}

func (k Keeper) SetSymbolPower(ctx sdk.Context, symbol string, power uint64) {
	ctx.KVStore(k.storeKey).Set(types.SymbolPowerStoreKey(symbol), sdk.Uint64ToBigEndian(power))
}
