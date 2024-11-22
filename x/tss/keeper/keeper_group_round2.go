package keeper

import (
	dbm "github.com/cosmos/cosmos-db"

	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/pkg/tss"
	"github.com/bandprotocol/chain/v3/x/tss/types"
)

// =====================================
// Round2Info store
// =====================================

// AddRound2Info adds the round2Info of a member in the store and increments the count of round2Info.
func (k Keeper) AddRound2Info(ctx sdk.Context, groupID tss.GroupID, round2Info types.Round2Info) {
	k.SetRound2Info(ctx, groupID, round2Info)

	count := k.GetRound2InfoCount(ctx, groupID)
	k.SetRound2InfoCount(ctx, groupID, count+1)
}

// SetRound2Info sets the round2Info of a member in the store and increments the count of round2Info.
func (k Keeper) SetRound2Info(ctx sdk.Context, groupID tss.GroupID, round2Info types.Round2Info) {
	ctx.KVStore(k.storeKey).
		Set(types.Round2InfoStoreKey(groupID, round2Info.MemberID), k.cdc.MustMarshal(&round2Info))
}

// HasRound2Info checks if the round2Info of a member exists in the store.
func (k Keeper) HasRound2Info(ctx sdk.Context, groupID tss.GroupID, memberID tss.MemberID) bool {
	return ctx.KVStore(k.storeKey).Has(types.Round2InfoStoreKey(groupID, memberID))
}

// GetRound2Info retrieves the round2Info of a member from the store.
func (k Keeper) GetRound2Info(ctx sdk.Context, groupID tss.GroupID, memberID tss.MemberID) (types.Round2Info, error) {
	bz := ctx.KVStore(k.storeKey).Get(types.Round2InfoStoreKey(groupID, memberID))
	if bz == nil {
		return types.Round2Info{}, types.ErrRound2InfoNotFound.Wrapf(
			"failed to get round2Info with groupID: %d, memberID: %d",
			groupID,
			memberID,
		)
	}

	var r2 types.Round2Info
	k.cdc.MustUnmarshal(bz, &r2)
	return r2, nil
}

// DeleteRound2Infos removes all round2Info associated with a specific group ID from the store.
func (k Keeper) DeleteRound2Infos(ctx sdk.Context, groupID tss.GroupID) {
	iterator := k.GetRound2InfoIterator(ctx, groupID)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		key := iterator.Key()
		ctx.KVStore(k.storeKey).Delete(key)
	}

	k.DeleteRound2InfoCount(ctx, groupID)
}

// SetRound2InfoCount sets the count of round2Info in the store.
func (k Keeper) SetRound2InfoCount(ctx sdk.Context, groupID tss.GroupID, count uint64) {
	ctx.KVStore(k.storeKey).Set(types.Round2InfoCountStoreKey(groupID), sdk.Uint64ToBigEndian(count))
}

// GetRound2InfoCount retrieves the count of round2Info from the store.
func (k Keeper) GetRound2InfoCount(ctx sdk.Context, groupID tss.GroupID) uint64 {
	bz := ctx.KVStore(k.storeKey).Get(types.Round2InfoCountStoreKey(groupID))
	return sdk.BigEndianToUint64(bz)
}

// DeleteRound2InfoCount remove the round2Info count data of a group from the store.
func (k Keeper) DeleteRound2InfoCount(ctx sdk.Context, groupID tss.GroupID) {
	ctx.KVStore(k.storeKey).Delete(types.Round2InfoCountStoreKey(groupID))
}

// GetRound2InfoIterator gets an iterator over all round2Info of a group.
func (k Keeper) GetRound2InfoIterator(ctx sdk.Context, groupID tss.GroupID) dbm.Iterator {
	return storetypes.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.Round2InfosStoreKey(groupID))
}

// GetRound2Infos retrieves all round2Info for a given group from the store.
func (k Keeper) GetRound2Infos(ctx sdk.Context, groupID tss.GroupID) []types.Round2Info {
	iterator := k.GetRound2InfoIterator(ctx, groupID)
	defer iterator.Close()

	var round2Infos []types.Round2Info
	for ; iterator.Valid(); iterator.Next() {
		var round2Info types.Round2Info
		k.cdc.MustUnmarshal(iterator.Value(), &round2Info)
		round2Infos = append(round2Infos, round2Info)
	}

	return round2Infos
}
