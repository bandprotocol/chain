package keeper

import (
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/feeds/types"
)

// GetDelegatorSignals returns a list of all signals of a delegator.
func (k Keeper) GetDelegatorSignals(ctx sdk.Context, delegator sdk.AccAddress) []types.Signal {
	bz := ctx.KVStore(k.storeKey).Get(types.DelegatorSignalStoreKey(delegator))
	if bz == nil {
		return nil
	}

	var s types.DelegatorSignals
	k.cdc.MustUnmarshal(bz, &s)

	return s.Signals
}

// DeleteDelegatorSignals deletes all signals of a delegator.
func (k Keeper) DeleteDelegatorSignals(ctx sdk.Context, delegator sdk.AccAddress) {
	ctx.KVStore(k.storeKey).
		Delete(types.DelegatorSignalStoreKey(delegator))
}

// SetDelegatorSignals sets multiple signals of a delegator.
func (k Keeper) SetDelegatorSignals(ctx sdk.Context, signals types.DelegatorSignals) {
	ctx.KVStore(k.storeKey).
		Set(types.DelegatorSignalStoreKey(sdk.MustAccAddressFromBech32(signals.Delegator)), k.cdc.MustMarshal(&signals))
}

// GetDelegatorSignalsIterator returns an iterator of the delegator-signals store.
func (k Keeper) GetDelegatorSignalsIterator(ctx sdk.Context) sdk.Iterator {
	return sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.DelegatorSignalStoreKeyPrefix)
}

// GetAllDelegatorSignals returns a list of all delegator-signals.
func (k Keeper) GetAllDelegatorSignals(ctx sdk.Context) (delegatorSignalsList []types.DelegatorSignals) {
	iterator := k.GetDelegatorSignalsIterator(ctx)
	defer func(iterator sdk.Iterator) {
		_ = iterator.Close()
	}(iterator)

	for ; iterator.Valid(); iterator.Next() {
		var ds types.DelegatorSignals
		k.cdc.MustUnmarshal(iterator.Value(), &ds)
		delegatorSignalsList = append(delegatorSignalsList, ds)
	}

	return delegatorSignalsList
}

// SetAllDelegatorSignals sets multiple delegator-signals.
func (k Keeper) SetAllDelegatorSignals(ctx sdk.Context, delegatorSignalsList []types.DelegatorSignals) {
	for _, ds := range delegatorSignalsList {
		k.SetDelegatorSignals(ctx, ds)
	}
}

// SetSignalTotalPower sets signal-total-power to the store.
func (k Keeper) SetSignalTotalPower(ctx sdk.Context, signal types.Signal) {
	prevSignalTotalPower, err := k.GetSignalTotalPower(ctx, signal.ID)
	if err == nil {
		k.deleteSignalTotalPowerByPowerIndex(ctx, prevSignalTotalPower)
	}

	ctx.KVStore(k.storeKey).
		Set(types.SignalTotalPowerStoreKey(signal.ID), k.cdc.MustMarshal(&signal))
	k.setSignalTotalPowerByPowerIndex(ctx, signal)
	emitEventUpdateSignalTotalPower(ctx, signal)
}

// GetSignalTotalPower gets a signal-total-power from specified signal id.
func (k Keeper) GetSignalTotalPower(ctx sdk.Context, signalID string) (types.Signal, error) {
	bz := ctx.KVStore(k.storeKey).Get(types.SignalTotalPowerStoreKey(signalID))
	if bz == nil {
		return types.Signal{}, types.ErrSignalTotalPowerNotFound.Wrapf(
			"failed to get signal-total-power for signal id: %s",
			signalID,
		)
	}

	var s types.Signal
	k.cdc.MustUnmarshal(bz, &s)

	return s, nil
}

// SetSignalTotalPowers sets multiple signal-total-powers.
func (k Keeper) SetSignalTotalPowers(ctx sdk.Context, signalTotalPowersList []types.Signal) {
	for _, stp := range signalTotalPowersList {
		k.SetSignalTotalPower(ctx, stp)
	}
}

func (k Keeper) DeleteSignalTotalPower(ctx sdk.Context, signalTotalPower types.Signal) {
	k.DeletePrice(ctx, signalTotalPower.ID)
	k.deleteSignalTotalPowerByPowerIndex(ctx, signalTotalPower)
	ctx.KVStore(k.storeKey).Delete(types.SignalTotalPowerStoreKey(signalTotalPower.ID))
}

func (k Keeper) setSignalTotalPowerByPowerIndex(ctx sdk.Context, signalTotalPower types.Signal) {
	ctx.KVStore(k.storeKey).
		Set(types.SignalTotalPowerByPowerIndexKey(signalTotalPower.ID, signalTotalPower.Power), []byte(signalTotalPower.ID))
}

func (k Keeper) deleteSignalTotalPowerByPowerIndex(ctx sdk.Context, signalTotalPower types.Signal) {
	ctx.KVStore(k.storeKey).
		Delete(types.SignalTotalPowerByPowerIndexKey(signalTotalPower.ID, signalTotalPower.Power))
}

// GetSignalTotalPowersByPower gets the current signal-total-power sorted by power-rank.
func (k Keeper) GetSignalTotalPowersByPower(ctx sdk.Context, limit int64) []types.Signal {
	signalTotalPowers := make([]types.Signal, limit)

	iterator := k.SignalTotalPowersByPowerStoreIterator(ctx)
	defer func(iterator sdk.Iterator) {
		_ = iterator.Close()
	}(iterator)

	i := 0
	for ; iterator.Valid() && i < int(limit); iterator.Next() {
		bz := iterator.Value()
		signalID := string(bz)
		signalTotalPower, err := k.GetSignalTotalPower(ctx, signalID)
		if err != nil || signalTotalPower.Power == 0 {
			continue
		}

		signalTotalPowers[i] = signalTotalPower
		i++
	}

	return signalTotalPowers[:i] // trim
}

// SignalTotalPowersByPowerStoreIterator returns an iterator for signal-total-powers by power index store.
func (k Keeper) SignalTotalPowersByPowerStoreIterator(ctx sdk.Context) sdk.Iterator {
	return sdk.KVStoreReversePrefixIterator(ctx.KVStore(k.storeKey), types.SignalTotalPowerByPowerIndexKeyPrefix)
}

func (k Keeper) CalculateNewSignalTotalPowers(ctx sdk.Context) ([]types.Signal, error) {
	delegatorSignals := k.GetAllDelegatorSignals(ctx)
	signalIDToPower := make(map[string]int64)
	for _, ds := range delegatorSignals {
		for _, signal := range ds.Signals {
			if signal.ID == "" || signal.Power <= 0 {
				return []types.Signal{}, types.ErrInvalidSignal.Wrap(
					"signal id cannot be empty and its power must be positive",
				)
			}
			signalIDToPower[signal.ID] += signal.Power
		}
	}

	keys := make([]string, 0, len(signalIDToPower))
	for k := range signalIDToPower {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	signalTotalPowers := []types.Signal{}
	for _, signalID := range keys {
		signalTotalPowers = append(signalTotalPowers, types.Signal{
			ID:    signalID,
			Power: signalIDToPower[signalID],
		})
	}

	return signalTotalPowers, nil
}
