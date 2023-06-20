package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

// SetDEQueue sets the DEQueue for a given address in the context's KVStore.
func (k Keeper) SetDEQueue(ctx sdk.Context, address sdk.AccAddress, deQueue types.DEQueue) {
	ctx.KVStore(k.storeKey).Set(types.DEQueueKeyStoreKey(address), k.cdc.MustMarshal(&deQueue))
}

// GetDEQueue retrieves the DEQueue for a given address from the context's KVStore.
func (k Keeper) GetDEQueue(ctx sdk.Context, address sdk.AccAddress) types.DEQueue {
	var deQueue types.DEQueue
	k.cdc.MustUnmarshal(ctx.KVStore(k.storeKey).Get(types.DEQueueKeyStoreKey(address)), &deQueue)
	return deQueue
}

// GetDESize retrieves the current size of DE for a given address from the context's KVStore.
func (k Keeper) GetDESize(ctx sdk.Context, address sdk.AccAddress) uint64 {
	deQueue := k.GetDEQueue(ctx, address)

	if deQueue.Head <= deQueue.Tail {
		return deQueue.Tail - deQueue.Head
	}
	return k.MaxDESize(ctx) - (deQueue.Head - deQueue.Tail)
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
		return types.DE{}, sdkerrors.Wrapf(
			types.ErrDENotFound,
			"failed to get DE with address %s index %d",
			address.String(),
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

// NextQueueValue returns next value of head/tail for DE queue
func (k Keeper) NextQueueValue(ctx sdk.Context, val uint64) uint64 {
	nextVal := (val + 1) % k.MaxDESize(ctx)
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
			return sdkerrors.Wrap(types.ErrDEQueueFull, fmt.Sprintf("DE size exceeds %d", k.MaxDESize(ctx)))
		}
	}

	k.SetDEQueue(ctx, address, deQueue)

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
	k.SetDEQueue(ctx, address, deQueue)

	return de, nil
}

// HandlePollDEForAssignedMembers function handles the polling of Diffie-Hellman key exchange results (DE) for the assigned members.
// It takes a list of member IDs (mids) and member information (members) and returns the assigned members along with their DE public keys.
func (k Keeper) HandlePollDEForAssignedMembers(
	ctx sdk.Context,
	mids []tss.MemberID,
	members []types.Member,
) ([]types.AssignedMember, tss.PublicKeys, tss.PublicKeys, error) {
	var assignedMembers []types.AssignedMember
	var pubDs, pubEs tss.PublicKeys

	for _, mid := range mids {
		member := members[mid-1]
		accMember, err := sdk.AccAddressFromBech32(member.Address)
		if err != nil {
			return nil, nil, nil, sdkerrors.Wrapf(types.ErrInvalidAccAddressFormat, err.Error())
		}

		de, err := k.PollDE(ctx, accMember)
		if err != nil {
			return nil, nil, nil, err
		}

		pubDs = append(pubDs, de.PubD)
		pubEs = append(pubEs, de.PubE)

		assignedMembers = append(assignedMembers, types.AssignedMember{
			MemberID: mid,
			Member:   member.Address,
			PubD:     de.PubD,
			PubE:     de.PubE,
			PubNonce: nil,
		})
	}

	return assignedMembers, pubDs, pubEs, nil
}
