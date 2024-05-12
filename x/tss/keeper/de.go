package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/bandprotocol/chain/v2/x/tss/types"
)

// SetDEQueue sets the DEQueue for a given address in the context's KVStore.
func (k Keeper) SetDEQueue(ctx sdk.Context, deQueue types.DEQueue) {
	address := sdk.MustAccAddressFromBech32(deQueue.Address)
	ctx.KVStore(k.storeKey).Set(types.DEQueueKeyStoreKey(address), k.cdc.MustMarshal(&deQueue))
}

// GetDEQueue retrieves the DEQueue for a given address from the context's KVStore.
func (k Keeper) GetDEQueue(ctx sdk.Context, address sdk.AccAddress) types.DEQueue {
	bz := ctx.KVStore(k.storeKey).Get(types.DEQueueKeyStoreKey(address))
	if bz == nil {
		return types.DEQueue{
			Address: address.String(),
			Head:    0,
			Tail:    0,
		}
	}
	var deQueue types.DEQueue
	k.cdc.MustUnmarshal(bz, &deQueue)
	return deQueue
}

// GetDEQueueIterator function gets an iterator over all de queue from the context's KVStore.
func (k Keeper) GetDEQueueIterator(ctx sdk.Context) sdk.Iterator {
	return sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.DEQueueStoreKeyPrefix)
}

// GetDEQueues retrieves all DEQueues from the context's KVStore.
func (k Keeper) GetDEQueues(ctx sdk.Context) []types.DEQueue {
	var deQueues []types.DEQueue
	iterator := k.GetDEQueueIterator(ctx)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var deQueue types.DEQueue
		k.cdc.MustUnmarshal(iterator.Value(), &deQueue)
		deQueues = append(deQueues, deQueue)
	}
	return deQueues
}

// GetDECount retrieves the current count of DE for a given address from the context's KVStore.
func (k Keeper) GetDECount(ctx sdk.Context, address sdk.AccAddress) uint64 {
	deQueue := k.GetDEQueue(ctx, address)
	if deQueue.Head <= deQueue.Tail {
		return deQueue.Tail - deQueue.Head
	}

	return k.GetParams(ctx).MaxDESize - (deQueue.Head - deQueue.Tail)
}

// SetDE sets a DE object in the context's KVStore for a given address and index.
func (k Keeper) SetDE(ctx sdk.Context, address sdk.AccAddress, index uint64, de types.DE) {
	ctx.KVStore(k.storeKey).Set(types.DEIndexStoreKey(address, index), k.cdc.MustMarshal(&de))
}

// GetDE retrieves a DE object from the context's KVStore for a given address and index.
// Returns an error if DE is not found.
func (k Keeper) GetDE(ctx sdk.Context, address sdk.AccAddress, index uint64) (types.DE, error) {
	bz := ctx.KVStore(k.storeKey).Get(types.DEIndexStoreKey(address, index))
	if bz == nil {
		return types.DE{}, types.ErrDENotFound.Wrapf(
			"failed to get DE with address %s index %d",
			address,
			index,
		)
	}
	var de types.DE
	k.cdc.MustUnmarshal(bz, &de)
	return de, nil
}

// DeleteDE removes a DE object from the context's KVStore for a given address and index.
func (k Keeper) DeleteDE(ctx sdk.Context, address sdk.AccAddress, index uint64) {
	ctx.KVStore(k.storeKey).Delete(types.DEIndexStoreKey(address, index))
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
		var de types.DE
		k.cdc.MustUnmarshal(iterator.Value(), &de)
		address, index := types.AddressAndIndexFromDEStoreKey(iterator.Key())
		des = append(des, types.DEGenesis{
			Address: address.String(),
			Index:   index,
			DE:      de,
		})
	}
	return des
}

// NextQueueValue returns next value of head/tail for DE queue
func (k Keeper) NextQueueValue(ctx sdk.Context, val uint64) uint64 {
	nextVal := (val + 1) % k.GetParams(ctx).MaxDESize
	return nextVal
}

// HandleSetDEs sets multiple DE objects for a given address in the context's KVStore,
// if tail reaches to head, return err as DE is full
func (k Keeper) HandleSetDEs(ctx sdk.Context, address sdk.AccAddress, des []types.DE) error {
	deQueue := k.GetDEQueue(ctx, address)

	for _, de := range des {
		k.SetDE(ctx, address, deQueue.Tail, de)
		deQueue.Tail = k.NextQueueValue(ctx, deQueue.Tail)

		if deQueue.Tail == deQueue.Head {
			return types.ErrDEQueueFull.Wrap(fmt.Sprintf("DE size exceeds %d", k.GetParams(ctx).MaxDESize))
		}
	}

	k.SetDEQueue(ctx, deQueue)

	return nil
}

// PollDE retrieves and removes the DE object at the head of the DEQueue for a given address,
// then increments the head index in the DEQueue.
// Returns an error if the DE object could not be retrieved.
func (k Keeper) PollDE(ctx sdk.Context, address sdk.AccAddress) (types.DE, error) {
	deQueue := k.GetDEQueue(ctx, address)
	de, err := k.GetDE(ctx, address, deQueue.Head)
	if err != nil {
		return types.DE{}, err
	}

	k.DeleteDE(ctx, address, deQueue.Head)

	deQueue.Head = k.NextQueueValue(ctx, deQueue.Head)
	k.SetDEQueue(ctx, deQueue)

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
