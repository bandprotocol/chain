package keeper

import (
	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/tss/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// GetMaliciousMembers retrieves the malicious members within a group identified by groupID.
func (k Keeper) GetMaliciousMembers(ctx sdk.Context, groupID tss.GroupID) ([]types.Member, error) {
	var maliciousMembers []types.Member
	members, err := k.GetMembers(ctx, groupID)
	if err != nil {
		return []types.Member{}, err
	}

	for _, m := range members {
		if m.IsMalicious {
			maliciousMembers = append(maliciousMembers, m)
		}
	}

	return maliciousMembers, nil
}

// HandleVerifyComplainSig verifies the complain signature for a given groupID and complain.
func (k Keeper) HandleVerifyComplain(
	ctx sdk.Context,
	groupID tss.GroupID,
	complain types.Complain,
) error {
	// Get round 1 data from member I
	round1I, err := k.GetRound1Data(ctx, groupID, complain.I)
	if err != nil {
		return err
	}

	// Get round 1 data from member J
	round1J, err := k.GetRound1Data(ctx, groupID, complain.J)
	if err != nil {
		return err
	}

	// Get round 2 data from member J
	round2J, err := k.GetRound2Data(ctx, groupID, complain.J)
	if err != nil {
		return err
	}

	// Find index J in encrypted secret shares
	indexJ := complain.I - 1
	if complain.J < complain.I {
		indexJ -= 1
	}

	// Verify the complain signature
	err = tss.VerifyComplain(
		round1I.OneTimePubKey,
		round1J.OneTimePubKey,
		complain.KeySym,
		complain.Sig,
		round2J.EncryptedSecretShares[indexJ],
		complain.I,
		round1J.CoefficientsCommit,
	)
	if err != nil {
		return sdkerrors.Wrapf(
			types.ErrComplainFailed,
			"failed to complain member: %d with groupID: %d; %s",
			complain.J,
			groupID,
			err,
		)
	}

	return nil
}

// HandleVerifyOwnPubKeySig verifies the own public key signature for a given groupID, memberID, and ownPubKeySig.
func (k Keeper) HandleVerifyOwnPubKeySig(
	ctx sdk.Context,
	groupID tss.GroupID,
	memberID tss.MemberID,
	ownPubKeySig tss.Signature,
) error {
	// Get member public key
	member, err := k.GetMember(ctx, groupID, memberID)
	if err != nil {
		return err
	}

	// Get dkg context
	dkgContext, err := k.GetDKGContext(ctx, groupID)
	if err != nil {
		return err
	}

	// Verify own public key sig
	err = tss.VerifyOwnPubKeySig(memberID, dkgContext, ownPubKeySig, member.PubKey)
	if err != nil {
		return sdkerrors.Wrapf(
			types.ErrConfirmFailed,
			"failed to verify own public key with memberID: %d; %s",
			memberID,
			err,
		)
	}

	return nil
}

// SetComplainsWithStatus sets the complains with status for a specific groupID and memberID in the store.
func (k Keeper) SetComplainsWithStatus(
	ctx sdk.Context,
	groupID tss.GroupID,
	complainsWithStatus types.ComplainsWithStatus,
) {
	// Add confirm complain count
	k.AddConfirmComplainCount(ctx, groupID)
	ctx.KVStore(k.storeKey).
		Set(types.ComplainsWithStatusMemberStoreKey(groupID, complainsWithStatus.MemberID), k.cdc.MustMarshal(&complainsWithStatus))
}

// GetComplainsWithStatus retrieves the complains with status for a specific groupID and memberID from the store.
func (k Keeper) GetComplainsWithStatus(
	ctx sdk.Context,
	groupID tss.GroupID,
	memberID tss.MemberID,
) (types.ComplainsWithStatus, error) {
	bz := ctx.KVStore(k.storeKey).Get(types.ComplainsWithStatusMemberStoreKey(groupID, memberID))
	if bz == nil {
		return types.ComplainsWithStatus{}, sdkerrors.Wrapf(
			types.ErrComplainsWithStatusNotFound,
			"failed to get complains with status with groupID %d memberID %d",
			groupID,
			memberID,
		)
	}
	var c types.ComplainsWithStatus
	k.cdc.MustUnmarshal(bz, &c)
	return c, nil
}

// GetComplainsWithStatusIterator function gets an iterator over all complains with status data of a group.
func (k Keeper) GetComplainsWithStatusIterator(ctx sdk.Context, groupID tss.GroupID) sdk.Iterator {
	return sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.ComplainsWithStatusStoreKey(groupID))
}

// GetAllComplainsWithStatus method retrieves all complains with status for a given group from the store.
func (k Keeper) GetAllComplainsWithStatus(ctx sdk.Context, groupID tss.GroupID) []types.ComplainsWithStatus {
	var cs []types.ComplainsWithStatus
	iterator := k.GetComplainsWithStatusIterator(ctx, groupID)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var c types.ComplainsWithStatus
		k.cdc.MustUnmarshal(iterator.Value(), &c)
		cs = append(cs, c)
	}
	return cs
}

// DeleteComplainsWithStatus method deletes the complain with status of a member from the store.
func (k Keeper) DeleteComplainsWithStatus(ctx sdk.Context, groupID tss.GroupID, memberID tss.MemberID) {
	ctx.KVStore(k.storeKey).Delete(types.ComplainsWithStatusMemberStoreKey(groupID, memberID))
}

// SetConfirm sets the confirm for a specific groupID and memberID in the store.
func (k Keeper) SetConfirm(
	ctx sdk.Context,
	groupID tss.GroupID,
	confirm types.Confirm,
) {
	// add confirm complain count
	k.AddConfirmComplainCount(ctx, groupID)
	ctx.KVStore(k.storeKey).
		Set(types.ConfirmMemberStoreKey(groupID, confirm.MemberID), k.cdc.MustMarshal(&confirm))
}

// GetConfirm retrieves the confirm for a specific groupID and memberID from the store.
func (k Keeper) GetConfirm(
	ctx sdk.Context,
	groupID tss.GroupID,
	memberID tss.MemberID,
) (types.Confirm, error) {
	bz := ctx.KVStore(k.storeKey).Get(types.ConfirmMemberStoreKey(groupID, memberID))
	if bz == nil {
		return types.Confirm{}, sdkerrors.Wrapf(
			types.ErrConfirmNotFound,
			"failed to get confirm with groupID %d memberID %d",
			groupID,
			memberID,
		)
	}
	var c types.Confirm
	k.cdc.MustUnmarshal(bz, &c)
	return c, nil
}

// GetConfirmIterator function gets an iterator over all confirm data of a group.
func (k Keeper) GetConfirmIterator(ctx sdk.Context, groupID tss.GroupID) sdk.Iterator {
	return sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.ConfirmStoreKey(groupID))
}

// GetConfirms method retrieves all confirm for a given group from the store.
func (k Keeper) GetConfirms(ctx sdk.Context, groupID tss.GroupID) []types.Confirm {
	var cs []types.Confirm
	iterator := k.GetConfirmIterator(ctx, groupID)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var c types.Confirm
		k.cdc.MustUnmarshal(iterator.Value(), &c)
		cs = append(cs, c)
	}
	return cs
}

// DeleteConfirm method deletes the confirm of a member from the store.
func (k Keeper) DeleteConfirm(ctx sdk.Context, groupID tss.GroupID, memberID tss.MemberID) {
	ctx.KVStore(k.storeKey).Delete(types.ConfirmMemberStoreKey(groupID, memberID))
}

// SetConfirmComplainCount sets the confirm complain count for a specific groupID in the store.
func (k Keeper) SetConfirmComplainCount(ctx sdk.Context, groupID tss.GroupID, count uint64) {
	ctx.KVStore(k.storeKey).Set(types.ConfirmComplainCountStoreKey(groupID), sdk.Uint64ToBigEndian(count))
}

// GetConfirmComplainCount retrieves the confirm complain count for a specific groupID from the store
func (k Keeper) GetConfirmComplainCount(ctx sdk.Context, groupID tss.GroupID) uint64 {
	bz := ctx.KVStore(k.storeKey).Get(types.ConfirmComplainCountStoreKey(groupID))
	return sdk.BigEndianToUint64(bz)
}

// AddConfirmComplainCount method increments the count of confirm and complain in the store.
func (k Keeper) AddConfirmComplainCount(ctx sdk.Context, groupID tss.GroupID) {
	count := k.GetConfirmComplainCount(ctx, groupID)
	k.SetConfirmComplainCount(ctx, groupID, count+1)
}

// DeleteConfirmComplainCount remove the confirm complain count data of a group from the store.
func (k Keeper) DeleteConfirmComplainCount(ctx sdk.Context, groupID tss.GroupID) {
	ctx.KVStore(k.storeKey).Delete(types.ConfirmComplainCountStoreKey(groupID))
}

// MarkMalicious change member status to malicious.
func (k Keeper) MarkMalicious(ctx sdk.Context, groupID tss.GroupID, memberID tss.MemberID) error {
	member, err := k.GetMember(ctx, groupID, memberID)
	if err != nil {
		return err
	}
	if member.IsMalicious {
		return nil
	}

	// update member status
	member.IsMalicious = true
	k.SetMember(ctx, groupID, memberID, member)
	return nil
}

// DeleteAllDKGInterimData deletes all DKG interim data for a given groupID and groupSize
func (k Keeper) DeleteAllDKGInterimData(
	ctx sdk.Context,
	groupID tss.GroupID,
	groupSize uint64,
	groupThreshold uint64,
) {
	// Delete DKG context
	k.DeleteDKGContext(ctx, groupID)

	for i := uint64(1); i <= groupSize; i++ {
		memberID := tss.MemberID(i)
		// Delete round 1 data
		k.DeleteRound1Data(ctx, groupID, memberID)
		// Delete round 2 data
		k.DeleteRound2Data(ctx, groupID, memberID)
		// Delete complain with status
		k.DeleteComplainsWithStatus(ctx, groupID, memberID)
		// Delete confirm
		k.DeleteConfirm(ctx, groupID, memberID)
	}

	for i := uint64(0); i < groupThreshold; i++ {
		// Delete accumulated commit
		k.DeleteAccumulatedCommit(ctx, groupID, i)
	}

	// Delete round 1 data count
	k.DeleteRound1DataCount(ctx, groupID)
	// Delete round 2 data count
	k.DeleteRound2DataCount(ctx, groupID)
	// Delete confirm complain count
	k.DeleteConfirmComplainCount(ctx, groupID)
}
