package keeper

import (
	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/tss/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// SetRound1Data function sets round1 data for a member of a group.
func (k Keeper) SetRound1Data(ctx sdk.Context, groupID tss.GroupID, round1Data types.Round1Data) {
	// Add count
	k.AddRound1DataCount(ctx, groupID)
	ctx.KVStore(k.storeKey).
		Set(types.Round1DataMemberStoreKey(groupID, round1Data.MemberID), k.cdc.MustMarshal(&round1Data))
}

// GetRound1Data function retrieves round1 data of a member from the store.
func (k Keeper) GetRound1Data(ctx sdk.Context, groupID tss.GroupID, memberID tss.MemberID) (types.Round1Data, error) {
	bz := ctx.KVStore(k.storeKey).Get(types.Round1DataMemberStoreKey(groupID, memberID))
	if bz == nil {
		return types.Round1Data{}, sdkerrors.Wrapf(
			types.ErrRound1DataNotFound,
			"failed to get round1 data with groupID: %d and memberID %d",
			groupID,
			memberID,
		)
	}
	var r1 types.Round1Data
	k.cdc.MustUnmarshal(bz, &r1)
	return r1, nil
}

// GetRound1DataIterator function gets an iterator over all round1 data of a group.
func (k Keeper) GetRound1DataIterator(ctx sdk.Context, groupID tss.GroupID) sdk.Iterator {
	return sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.Round1DataStoreKey(groupID))
}

// GetAllRound1Data retrieves all round1 data for a group from the store.
func (k Keeper) GetAllRound1Data(ctx sdk.Context, groupID tss.GroupID) []types.Round1Data {
	var allRound1Data []types.Round1Data
	iterator := k.GetRound1DataIterator(ctx, groupID)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var round1Data types.Round1Data
		k.cdc.MustUnmarshal(iterator.Value(), &round1Data)
		allRound1Data = append(allRound1Data, round1Data)
	}
	return allRound1Data
}

// DeleteRound1Data removes the round1 data of a group member from the store.
func (k Keeper) DeleteRound1Data(ctx sdk.Context, groupID tss.GroupID, memberID tss.MemberID) {
	ctx.KVStore(k.storeKey).Delete(types.Round1DataMemberStoreKey(groupID, memberID))
}

// SetRound1DataCount sets the count of round1 data for a group in the store.
func (k Keeper) SetRound1DataCount(ctx sdk.Context, groupID tss.GroupID, count uint64) {
	ctx.KVStore(k.storeKey).Set(types.Round1DataCountStoreKey(groupID), sdk.Uint64ToBigEndian(count))
}

// GetRound1DataCount retrieves the count of round1 data for a group from the store.
func (k Keeper) GetRound1DataCount(ctx sdk.Context, groupID tss.GroupID) uint64 {
	bz := ctx.KVStore(k.storeKey).Get(types.Round1DataCountStoreKey(groupID))
	return sdk.BigEndianToUint64(bz)
}

// AddRound1DataCount increments the count of round1 data for a group in the store.
func (k Keeper) AddRound1DataCount(ctx sdk.Context, groupID tss.GroupID) {
	count := k.GetRound1DataCount(ctx, groupID)
	k.SetRound1DataCount(ctx, groupID, count+1)
}

// DeleteRound1DataCount remove the round 1 data count data of a group from the store.
func (k Keeper) DeleteRound1DataCount(ctx sdk.Context, groupID tss.GroupID) {
	ctx.KVStore(k.storeKey).Delete(types.Round1DataCountStoreKey(groupID))
}

// GetAccumulatedCommitIterator function gets an iterator over all accumulated commits of a group.
func (k Keeper) GetAccumulatedCommitIterator(ctx sdk.Context, groupID tss.GroupID) sdk.Iterator {
	return sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.AccumulatedCommitStoreKey(groupID))
}

// SetAccumulatedCommit function sets accumulated commit for a index of a group.
func (k Keeper) SetAccumulatedCommit(ctx sdk.Context, groupID tss.GroupID, index uint64, commit tss.Point) {
	ctx.KVStore(k.storeKey).Set(types.AccumulatedCommitIndexStoreKey(groupID, index), commit)
}

// GetAccumulatedCommit function retrieves accummulated commit of a index of the group from the store.
func (k Keeper) GetAccumulatedCommit(ctx sdk.Context, groupID tss.GroupID, index uint64) tss.Point {
	return ctx.KVStore(k.storeKey).Get(types.AccumulatedCommitIndexStoreKey(groupID, index))
}

// GetAllAccumulatedCommits function retrieves all accummulated commits of a group from the store.
func (k Keeper) GetAllAccumulatedCommits(ctx sdk.Context, groupID tss.GroupID) tss.Points {
	var commits tss.Points
	iterator := k.GetAccumulatedCommitIterator(ctx, groupID)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		commits = append(commits, iterator.Value())
	}
	return commits
}

// DeleteAccumulatedCommit removes a accumulated commit of a index of the group from the store.
func (k Keeper) DeleteAccumulatedCommit(ctx sdk.Context, groupID tss.GroupID, index uint64) {
	ctx.KVStore(k.storeKey).Delete(types.AccumulatedCommitIndexStoreKey(groupID, index))
}

// AddCommits function adds each coefficient commit into the accumulated commit of its index.
func (k Keeper) AddCommits(ctx sdk.Context, groupID tss.GroupID, commits tss.Points) error {
	// Add count
	for i, commit := range commits {
		points := []tss.Point{commit}

		accCommit := k.GetAccumulatedCommit(ctx, groupID, uint64(i))
		if accCommit != nil {
			points = append(points, accCommit)
		}

		total, err := tss.SumPoints(points...)
		if err != nil {
			return err
		}
		k.SetAccumulatedCommit(ctx, groupID, uint64(i), total)
	}

	return nil
}
