package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/bandprotocol/chain/v2/x/oracle/types"
)

// IsReporter returns true iff the address is an authorized reporter for the given validator.
func (k Keeper) IsReporter(ctx sdk.Context, val sdk.ValAddress, addr sdk.AccAddress) bool {
	if val.Equals(sdk.ValAddress(addr)) { // A validator is always a reporter of himself.
		return true
	}
	return ctx.KVStore(k.storeKey).Has(types.ReporterStoreKey(val, addr))
}

// AddReporter adds the reporter address to the list of reporters of the given validator.
func (k Keeper) AddReporter(ctx sdk.Context, val sdk.ValAddress, addr sdk.AccAddress) error {
	if k.IsReporter(ctx, val, addr) {
		return sdkerrors.Wrapf(
			types.ErrReporterAlreadyExists, "val: %s, addr: %s", val.String(), addr.String())
	}
	ctx.KVStore(k.storeKey).Set(types.ReporterStoreKey(val, addr), []byte{1})
	return nil
}

// RemoveReporter removes the reporter address from the list of reporters of the given validator.
func (k Keeper) RemoveReporter(ctx sdk.Context, val sdk.ValAddress, addr sdk.AccAddress) error {
	if !k.IsReporter(ctx, val, addr) {
		return sdkerrors.Wrapf(
			types.ErrReporterNotFound, "val: %s, addr: %s", val.String(), addr.String())
	}
	ctx.KVStore(k.storeKey).Delete(types.ReporterStoreKey(val, addr))
	return nil
}

// GetReporters returns the reporter list of the given validator.
func (k Keeper) GetReporters(ctx sdk.Context, val sdk.ValAddress) (reporters []sdk.AccAddress) {
	// Appends self reporter of validator to the list
	selfReporter := sdk.AccAddress(val)
	reporters = append(reporters, selfReporter)

	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.ReportersOfValidatorPrefixKey(val))
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		key := iterator.Key()
		reporterAddress := sdk.AccAddress(key[1+len(val):])
		reporters = append(reporters, reporterAddress)
	}
	return reporters
}

func (k Keeper) GetAllReporters(ctx sdk.Context) []types.ReportersPerValidator {
	var reporterList []types.ReportersPerValidator
	k.stakingKeeper.IterateBondedValidatorsByPower(ctx, func(index int64, validator stakingtypes.ValidatorI) (stop bool) {
		valAddress := validator.GetOperator()
		reporters := k.GetReporters(ctx, valAddress)
		var reportersBech32 []string
		for _, reporter := range reporters {
			reportersBech32 = append(reportersBech32, reporter.String())
		}
		reporterList = append(reporterList, types.ReportersPerValidator{
			Validator: sdk.ValAddress(valAddress).String(),
			Reporters: reportersBech32,
		})
		return false
	})

	return reporterList
}
