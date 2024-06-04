package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/bandprotocol/chain/v2/x/tss/types"
)

// SetDECount sets the number of existing DE for a given address.
func (k Keeper) SetDECount(ctx sdk.Context, address sdk.AccAddress, count uint64) {
	ctx.KVStore(k.storeKey).Set(types.DECountStoreKey(address), sdk.Uint64ToBigEndian(count))
}

// GetDECount retrieves the number of existing DE for a given address.
func (k Keeper) GetDECount(ctx sdk.Context, address sdk.AccAddress) uint64 {
	return sdk.BigEndianToUint64(ctx.KVStore(k.storeKey).Get(types.DECountStoreKey(address)))
}

// GetDEByAddressIterator function gets an iterator over the DEs of the given address.
func (k Keeper) GetDEByAddressIterator(ctx sdk.Context, addr sdk.AccAddress) sdk.Iterator {
	return sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.DEStoreKeyPerAddressPrefix(addr))
}

// HasDE checks if a DE object exists in the context's KVStore.
func (k Keeper) HasDE(ctx sdk.Context, address sdk.AccAddress, de types.DE) bool {
	return ctx.KVStore(k.storeKey).Has(types.DEStoreKey(address, de))
}

// SetDE sets a DE object in the context's KVStore for a given address.
func (k Keeper) SetDE(ctx sdk.Context, address sdk.AccAddress, de types.DE) {
	ctx.KVStore(k.storeKey).Set(types.DEStoreKey(address, de), []byte{1})
}

// GetFirstDE retrieves a DE object from the context's KVStore for a given address.
// Returns an error if DE is not found.
func (k Keeper) GetFirstDE(ctx sdk.Context, address sdk.AccAddress) (types.DE, error) {
	iterator := k.GetDEByAddressIterator(ctx, address)
	defer iterator.Close()

	if !iterator.Valid() {
		return types.DE{}, types.ErrDENotFound.Wrapf("failed to get DE with address %s", address)
	}

	_, de := types.ExtractValueFromDEStoreKey(iterator.Key())
	return de, nil
}

// DeleteDE removes a DE object from the context's KVStore for a given address and index.
func (k Keeper) DeleteDE(ctx sdk.Context, address sdk.AccAddress, de types.DE) {
	ctx.KVStore(k.storeKey).Delete(types.DEStoreKey(address, de))
}

// GetDEIterator function gets an iterator over all de from the context's KVStore
func (k Keeper) GetDEIterator(ctx sdk.Context) sdk.Iterator {
	return sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.DEStoreKeyPrefix)
}

// GetDEsGenesis retrieves all de from the context's KVStore.
func (k Keeper) GetDEsGenesis(ctx sdk.Context) []types.DEGenesis {
	var des []types.DEGenesis
	iterator := k.GetDEIterator(ctx)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		address, de := types.ExtractValueFromDEStoreKey(iterator.Key())
		des = append(des, types.DEGenesis{
			Address: address.String(),
			DE:      de,
		})
	}
	return des
}

// HandleSetDEs sets multiple DE objects for a given address in the context's KVStore,
// if the given DEs reach maximum limit, return err as DE will be over the limit.
func (k Keeper) HandleSetDEs(ctx sdk.Context, address sdk.AccAddress, des []types.DE) error {
	added := uint64(0)
	for _, de := range des {
		if k.HasDE(ctx, address, de) {
			continue
		}

		k.SetDE(ctx, address, de)
		added++
	}

	cnt := k.GetDECount(ctx, address)
	if cnt+added > k.GetParams(ctx).MaxDESize {
		return types.ErrDEReachMaximumLimit.Wrapf("DE size exceeds %d", k.GetParams(ctx).MaxDESize)
	}
	k.SetDECount(ctx, address, cnt+uint64(len(des)))
	return nil
}

// PollDE retrieves and removes the first DE object being retrieved from the iterator of
// the given address. Returns an error if the DE object could not be retrieved.
func (k Keeper) PollDE(ctx sdk.Context, address sdk.AccAddress) (types.DE, error) {
	de, err := k.GetFirstDE(ctx, address)
	if err != nil {
		return types.DE{}, err
	}

	cnt := k.GetDECount(ctx, address)
	k.SetDECount(ctx, address, cnt-1)
	k.DeleteDE(ctx, address, de)

	return de, nil
}

// PollDEs handles the polling of DE from the selected members. It takes a list of member IDs (mids)
// and members information (members) and returns the list of selected DEs ordered by selected members.
func (k Keeper) PollDEs(ctx sdk.Context, members []types.Member) ([]types.DE, error) {
	des := make([]types.DE, 0, len(members))

	for _, member := range members {
		// Convert the address from Bech32 format to AccAddress format
		accMember, err := sdk.AccAddressFromBech32(member.Address)
		if err != nil {
			return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid account address: %s", err)
		}

		de, err := k.PollDE(ctx, accMember)
		if err != nil {
			return nil, err
		}
		des = append(des, de)
	}

	return des, nil
}

// FilterMembersHaveDEs function retrieves all members that have DEs in the store.
func (k Keeper) FilterMembersHaveDE(ctx sdk.Context, members []types.Member) ([]types.Member, error) {
	var filtered []types.Member
	for _, member := range members {
		// Convert the address from Bech32 format to AccAddress format
		accMember, err := sdk.AccAddressFromBech32(member.Address)
		if err != nil {
			return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid account address: %s", err)
		}

		count := k.GetDECount(ctx, accMember)
		if count > 0 {
			filtered = append(filtered, member)
		}
	}
	return filtered, nil
}
