package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/x/oracle/types"
)

// IsReporter returns true iff the address is an authorized reporter for the given validator.
func (k Keeper) IsReporter(ctx sdk.Context, val sdk.ValAddress, addr sdk.AccAddress) bool {
	if val.Equals(sdk.ValAddress(addr)) { // A validator is always a reporter of himself.
		return true
	}
	return ctx.KVStore(k.storeKey).Has(types.ReporterStoreKey(val, addr))
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
