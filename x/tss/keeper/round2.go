package keeper

import (
	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/tss/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// SetRound2Data method sets the round2Data of a member in the store and increments the count of round2Data.
func (k Keeper) SetRound2Data(
	ctx sdk.Context,
	groupID tss.GroupID,
	round2Data types.Round2Data,
) {
	// Add count
	k.AddRound2DataCount(ctx, groupID)

	ctx.KVStore(k.storeKey).
		Set(types.Round2DataMemberStoreKey(groupID, round2Data.MemberID), k.cdc.MustMarshal(&round2Data))
}

// GetRound2Data method retrieves the round2Data of a member from the store.
func (k Keeper) GetRound2Data(ctx sdk.Context, groupID tss.GroupID, memberID tss.MemberID) (types.Round2Data, error) {
	bz := ctx.KVStore(k.storeKey).Get(types.Round2DataMemberStoreKey(groupID, memberID))
	if bz == nil {
		return types.Round2Data{}, sdkerrors.Wrapf(
			types.ErrRound2DataNotFound,
			"failed to get round2Data with groupID: %d, memberID: %d",
			groupID,
			memberID,
		)
	}
	var r2 types.Round2Data
	k.cdc.MustUnmarshal(bz, &r2)
	return r2, nil
}

// DeleteRound2Data method deletes the round2Data of a member from the store.
func (k Keeper) DeleteRound2Data(ctx sdk.Context, groupID tss.GroupID, memberID tss.MemberID) {
	ctx.KVStore(k.storeKey).Delete(types.Round2DataMemberStoreKey(groupID, memberID))
}

// SetRound2DataCount method sets the count of round2Data in the store.
func (k Keeper) SetRound2DataCount(ctx sdk.Context, groupID tss.GroupID, count uint64) {
	ctx.KVStore(k.storeKey).Set(types.Round2DataCountStoreKey(groupID), sdk.Uint64ToBigEndian(count))
}

// GetRound2DataCount method retrieves the count of round2Data from the store.
func (k Keeper) GetRound2DataCount(ctx sdk.Context, groupID tss.GroupID) uint64 {
	bz := ctx.KVStore(k.storeKey).Get(types.Round2DataCountStoreKey(groupID))
	return sdk.BigEndianToUint64(bz)
}

// AddRound2DataCount method increments the count of round2Data in the store.
func (k Keeper) AddRound2DataCount(ctx sdk.Context, groupID tss.GroupID) {
	count := k.GetRound2DataCount(ctx, groupID)
	k.SetRound2DataCount(ctx, groupID, count+1)
}

// DeleteRound2DataCount remove the round 2 data count data of a group from the store.
func (k Keeper) DeleteRound2DataCount(ctx sdk.Context, groupID tss.GroupID) {
	ctx.KVStore(k.storeKey).Delete(types.Round2DataCountStoreKey(groupID))
}

// GetRound2DataIterator function gets an iterator over all round1 data of a group.
func (k Keeper) GetRound2DataIterator(ctx sdk.Context, groupID tss.GroupID) sdk.Iterator {
	return sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.Round2DataStoreKey(groupID))
}

// GetAllRound2Data method retrieves all round2Data for a given group from the store.
func (k Keeper) GetAllRound2Data(ctx sdk.Context, groupID tss.GroupID) []types.Round2Data {
	var allRound2Data []types.Round2Data
	iterator := k.GetRound2DataIterator(ctx, groupID)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var round2Data types.Round2Data
		k.cdc.MustUnmarshal(iterator.Value(), &round2Data)
		allRound2Data = append(allRound2Data, round2Data)
	}
	return allRound2Data
}
