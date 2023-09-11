package keeper

import (
	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

// AddRound1Info adds the round1Info of a member in the store and increments the count of round1Info.
func (k Keeper) AddRound1Info(ctx sdk.Context, groupID tss.GroupID, round1Info types.Round1Info) {
	k.AddRound1InfoCount(ctx, groupID)
	k.SetRound1Info(ctx, groupID, round1Info)
}

// SetRound1Info sets round 1 info for a member of a group.
func (k Keeper) SetRound1Info(ctx sdk.Context, groupID tss.GroupID, round1Info types.Round1Info) {
	ctx.KVStore(k.storeKey).
		Set(types.Round1InfoMemberStoreKey(groupID, round1Info.MemberID), k.cdc.MustMarshal(&round1Info))
}

// GetRound1Info retrieves round 1 info of a member from the store.
func (k Keeper) GetRound1Info(ctx sdk.Context, groupID tss.GroupID, memberID tss.MemberID) (types.Round1Info, error) {
	bz := ctx.KVStore(k.storeKey).Get(types.Round1InfoMemberStoreKey(groupID, memberID))
	if bz == nil {
		return types.Round1Info{}, errors.Wrapf(
			types.ErrRound1InfoNotFound,
			"failed to get round 1 info with groupID: %d and memberID %d",
			groupID,
			memberID,
		)
	}
	var r1 types.Round1Info
	k.cdc.MustUnmarshal(bz, &r1)
	return r1, nil
}

// GetRound1InfoIterator gets an iterator over all round 1 info of a group.
func (k Keeper) GetRound1InfoIterator(ctx sdk.Context, groupID tss.GroupID) sdk.Iterator {
	return sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.Round1InfoStoreKey(groupID))
}

// GetRound1Infos retrieves round 1 infos for a group from the store.
func (k Keeper) GetRound1Infos(ctx sdk.Context, groupID tss.GroupID) []types.Round1Info {
	var round1Infos []types.Round1Info
	iterator := k.GetRound1InfoIterator(ctx, groupID)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var round1Info types.Round1Info
		k.cdc.MustUnmarshal(iterator.Value(), &round1Info)
		round1Infos = append(round1Infos, round1Info)
	}
	return round1Infos
}

// DeleteRound1Info removes the round 1 info of a group member from the store.
func (k Keeper) DeleteRound1Info(ctx sdk.Context, groupID tss.GroupID, memberID tss.MemberID) {
	ctx.KVStore(k.storeKey).Delete(types.Round1InfoMemberStoreKey(groupID, memberID))
}

// DeleteRound1Infos removes all round 1 info associated with a specific group ID from the store.
func (k Keeper) DeleteRound1Infos(ctx sdk.Context, groupID tss.GroupID) {
	iterator := k.GetRound1InfoIterator(ctx, groupID)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		key := iterator.Key()
		ctx.KVStore(k.storeKey).Delete(key)
	}
}

// SetRound1InfoCount sets the count of round 1 info for a group in the store.
func (k Keeper) SetRound1InfoCount(ctx sdk.Context, groupID tss.GroupID, count uint64) {
	ctx.KVStore(k.storeKey).Set(types.Round1InfoCountStoreKey(groupID), sdk.Uint64ToBigEndian(count))
}

// GetRound1InfoCount retrieves the count of round 1 info for a group from the store.
func (k Keeper) GetRound1InfoCount(ctx sdk.Context, groupID tss.GroupID) uint64 {
	bz := ctx.KVStore(k.storeKey).Get(types.Round1InfoCountStoreKey(groupID))
	return sdk.BigEndianToUint64(bz)
}

// AddRound1InfoCount increments the count of round 1 info for a group in the store.
func (k Keeper) AddRound1InfoCount(ctx sdk.Context, groupID tss.GroupID) {
	count := k.GetRound1InfoCount(ctx, groupID)
	k.SetRound1InfoCount(ctx, groupID, count+1)
}

// DeleteRound1InfoCount remove the round 1 info count data of a group from the store.
func (k Keeper) DeleteRound1InfoCount(ctx sdk.Context, groupID tss.GroupID) {
	ctx.KVStore(k.storeKey).Delete(types.Round1InfoCountStoreKey(groupID))
}

// GetAccumulatedCommitIterator gets an iterator over all accumulated commits of a group.
func (k Keeper) GetAccumulatedCommitIterator(ctx sdk.Context, groupID tss.GroupID) sdk.Iterator {
	return sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.AccumulatedCommitStoreKey(groupID))
}

// SetAccumulatedCommit sets accumulated commit for a index of a group.
func (k Keeper) SetAccumulatedCommit(ctx sdk.Context, groupID tss.GroupID, index uint64, commit tss.Point) {
	ctx.KVStore(k.storeKey).Set(types.AccumulatedCommitIndexStoreKey(groupID, index), commit)
}

// GetAccumulatedCommit retrieves accumulated commit of a index of the group from the store.
func (k Keeper) GetAccumulatedCommit(ctx sdk.Context, groupID tss.GroupID, index uint64) tss.Point {
	return ctx.KVStore(k.storeKey).Get(types.AccumulatedCommitIndexStoreKey(groupID, index))
}

// GetAllAccumulatedCommits retrieves all accumulated commits of a group from the store.
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

// DeleteAccumulatedCommits removes all accumulated commit associated with a specific group ID from the store.
func (k Keeper) DeleteAccumulatedCommits(ctx sdk.Context, groupID tss.GroupID) {
	iterator := k.GetAccumulatedCommitIterator(ctx, groupID)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		key := iterator.Key()
		ctx.KVStore(k.storeKey).Delete(key)
	}
}

// AddCommits adds each coefficient commit into the accumulated commit of its index.
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
