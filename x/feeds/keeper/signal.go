package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/feeds/types"
)

// CheckDelegatorDelegation checks whether the delegator has enough delegation for signals.
func (k Keeper) CheckDelegatorDelegation(
	ctx sdk.Context,
	delegator sdk.AccAddress,
	signals []types.Signal,
) error {
	sumPower := sumPower(signals)
	sumDelegation := k.stakingKeeper.GetDelegatorBonded(ctx, delegator).Uint64()
	if sumPower > sumDelegation {
		return types.ErrNotEnoughDelegation
	}
	return nil
}

// RemoveDelegatorSignal deletes previous signals from delegator and decrease symbol power by the previous signals.
func (k Keeper) RemoveDelegatorPreviousSignals(
	ctx sdk.Context,
	delegator sdk.AccAddress,
	symbolToIntervalDiff map[string]int64,
) (map[string]int64, error) {
	prevSignals := k.GetDelegatorSignals(ctx, delegator)
	for _, prevSignal := range prevSignals {
		symbol, err := k.GetSymbol(ctx, prevSignal.Symbol)
		if err != nil {
			return nil, err
		}
		// before changing in symbol, delete the SymbolByPower index
		k.DeleteSymbolByPowerIndex(ctx, symbol)

		symbol.Power -= prevSignal.Power
		prevInterval := symbol.Interval
		symbol.Interval = calculateInterval(int64(symbol.Power), k.GetParams(ctx))
		k.SetSymbol(ctx, symbol)

		// setting SymbolByPowerIndex every time setting symbol
		k.SetSymbolByPowerIndex(ctx, symbol)

		intervalDiff := (symbol.Interval - prevInterval) + symbolToIntervalDiff[symbol.Symbol]
		if intervalDiff == 0 {
			delete(symbolToIntervalDiff, symbol.Symbol)
		} else {
			symbolToIntervalDiff[symbol.Symbol] = intervalDiff
		}
	}
	// Add delete delegator signal in store
	return symbolToIntervalDiff, nil
}

// RegisterDelegatorSignals increases symbol power by the new signals.
func (k Keeper) RegisterDelegatorSignals(
	ctx sdk.Context,
	delegator sdk.AccAddress,
	signals []types.Signal,
	symbolToIntervalDiff map[string]int64,
) (map[string]int64, error) {
	k.SetDelegatorSignals(ctx, delegator, types.Signals{Signals: signals})
	for _, signal := range signals {
		symbol, err := k.GetSymbol(ctx, signal.Symbol)
		if err != nil {
			symbol = types.Symbol{
				Symbol:                      signal.Symbol,
				Power:                       0,
				Interval:                    0,
				LastIntervalUpdateTimestamp: 0,
			}
		}
		// before changing in symbol, delete the SymbolByPower index
		k.DeleteSymbolByPowerIndex(ctx, symbol)

		symbol.Power += signal.Power
		prevInterval := symbol.Interval
		symbol.Interval = calculateInterval(int64(symbol.Power), k.GetParams(ctx))
		k.SetSymbol(ctx, symbol)

		// setting SymbolByPowerIndex every time setting symbol
		k.SetSymbolByPowerIndex(ctx, symbol)

		// if the sum interval differences is zero then the interval is not changed
		intervalDiff := (symbol.Interval - prevInterval) + symbolToIntervalDiff[symbol.Symbol]
		if intervalDiff == 0 {
			delete(symbolToIntervalDiff, symbol.Symbol)
		} else {
			symbolToIntervalDiff[symbol.Symbol] = intervalDiff
		}
	}
	return symbolToIntervalDiff, nil
}

// UpdateSymbolIntervalTimestamp updates the interval timestamp for symbols where the interval has changed.
func (k Keeper) UpdateSymbolIntervalTimestamp(
	ctx sdk.Context,
	symbolToIntervalDiff map[string]int64,
) error {
	for symbolName := range symbolToIntervalDiff {
		symbol, err := k.GetSymbol(ctx, symbolName)
		if err != nil {
			return err
		}
		// before changing in symbol, delete the SymbolByPower index
		k.DeleteSymbolByPowerIndex(ctx, symbol)

		symbol.LastIntervalUpdateTimestamp = ctx.BlockTime().Unix()
		k.SetSymbol(ctx, symbol)

		// setting SymbolByPowerIndex every time setting symbol
		k.SetSymbolByPowerIndex(ctx, symbol)
	}
	return nil
}
