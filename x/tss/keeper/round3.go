package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

// HandleVerifyComplaint verifies the complaint signature for a given groupID and complaint.
func (k Keeper) HandleVerifyComplaint(
	ctx sdk.Context,
	groupID tss.GroupID,
	complaint types.Complaint,
) error {
	// Get round 1 info from member Complainant
	round1Complainant, err := k.GetRound1Info(ctx, groupID, complaint.Complainant)
	if err != nil {
		return err
	}

	// Get round 1 info from member Respondent
	round1Respondent, err := k.GetRound1Info(ctx, groupID, complaint.Respondent)
	if err != nil {
		return err
	}

	// Get round 2 info from member Respondent
	round2Respondent, err := k.GetRound2Info(ctx, groupID, complaint.Respondent)
	if err != nil {
		return err
	}

	// Find complainant index in respondent encrypted secret shares
	complainantIndex := types.FindMemberSlot(complaint.Respondent, complaint.Complainant)

	// Return error if the slot exceeds length of shares
	if int(complainantIndex) >= len(round2Respondent.EncryptedSecretShares) {
		return errorsmod.Wrapf(
			types.ErrComplainFailed,
			"No encrypted secret share from MemberID(%d) to MemberID(%d)",
			complaint.Respondent,
			complaint.Complainant,
		)
	}

	// Verify the complaint signature
	err = tss.VerifyComplaint(
		round1Complainant.OneTimePubKey,
		round1Respondent.OneTimePubKey,
		complaint.KeySym,
		complaint.Signature,
		round2Respondent.EncryptedSecretShares[complainantIndex],
		complaint.Complainant,
		round1Respondent.CoefficientCommits,
	)
	if err != nil {
		return errorsmod.Wrapf(
			types.ErrComplainFailed,
			"failed to complaint member: %d with groupID: %d; %s",
			complaint.Respondent,
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
	err = tss.VerifyOwnPubKeySignature(memberID, dkgContext, ownPubKeySig, member.PubKey)
	if err != nil {
		return errorsmod.Wrapf(
			types.ErrConfirmFailed,
			"failed to verify own public key with memberID: %d; %s",
			memberID,
			err,
		)
	}

	return nil
}

// AddComplaintsWithStatus adds the complaints with status of a member in the store and increments the confirm and complain count.
func (k Keeper) AddComplaintsWithStatus(
	ctx sdk.Context,
	groupID tss.GroupID,
	complaintsWithStatus types.ComplaintsWithStatus,
) {
	k.AddConfirmComplaintCount(ctx, groupID)
	k.SetComplaintsWithStatus(ctx, groupID, complaintsWithStatus)
}

// SetComplaintsWithStatus sets the complaints with status for a specific groupID and memberID in the store.
func (k Keeper) SetComplaintsWithStatus(
	ctx sdk.Context,
	groupID tss.GroupID,
	complaintsWithStatus types.ComplaintsWithStatus,
) {
	ctx.KVStore(k.storeKey).
		Set(types.ComplainsWithStatusMemberStoreKey(groupID, complaintsWithStatus.MemberID), k.cdc.MustMarshal(&complaintsWithStatus))
}

// GetComplaintsWithStatus retrieves the complaints with status for a specific groupID and memberID from the store.
func (k Keeper) GetComplaintsWithStatus(
	ctx sdk.Context,
	groupID tss.GroupID,
	memberID tss.MemberID,
) (types.ComplaintsWithStatus, error) {
	bz := ctx.KVStore(k.storeKey).Get(types.ComplainsWithStatusMemberStoreKey(groupID, memberID))
	if bz == nil {
		return types.ComplaintsWithStatus{}, errorsmod.Wrapf(
			types.ErrComplaintsWithStatusNotFound,
			"failed to get complaints with status with groupID %d memberID %d",
			groupID,
			memberID,
		)
	}
	var c types.ComplaintsWithStatus
	k.cdc.MustUnmarshal(bz, &c)
	return c, nil
}

// GetComplainsWithStatusIterator gets an iterator over all complaints with status data of a group.
func (k Keeper) GetComplainsWithStatusIterator(ctx sdk.Context, groupID tss.GroupID) sdk.Iterator {
	return sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.ComplainsWithStatusStoreKey(groupID))
}

// GetAllComplainsWithStatus retrieves all complaints with status for a given group from the store.
func (k Keeper) GetAllComplainsWithStatus(ctx sdk.Context, groupID tss.GroupID) []types.ComplaintsWithStatus {
	var cs []types.ComplaintsWithStatus
	iterator := k.GetComplainsWithStatusIterator(ctx, groupID)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var c types.ComplaintsWithStatus
		k.cdc.MustUnmarshal(iterator.Value(), &c)
		cs = append(cs, c)
	}
	return cs
}

// DeleteComplainsWithStatus deletes the complaint with status of a member from the store.
func (k Keeper) DeleteComplainsWithStatus(ctx sdk.Context, groupID tss.GroupID, memberID tss.MemberID) {
	ctx.KVStore(k.storeKey).Delete(types.ComplainsWithStatusMemberStoreKey(groupID, memberID))
}

// DeleteAllComplainsWithStatus removes all complains with status associated with a specific group ID from the store.
func (k Keeper) DeleteAllComplainsWithStatus(ctx sdk.Context, groupID tss.GroupID) {
	iterator := k.GetComplainsWithStatusIterator(ctx, groupID)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		key := iterator.Key()
		ctx.KVStore(k.storeKey).Delete(key)
	}
}

// AddConfirm adds the confirm of a member in the store and increments the confirm and complain count.
func (k Keeper) AddConfirm(
	ctx sdk.Context,
	groupID tss.GroupID,
	confirm types.Confirm,
) {
	k.AddConfirmComplaintCount(ctx, groupID)
	k.SetConfirm(ctx, groupID, confirm)
}

// SetConfirm sets the confirm for a specific groupID and memberID in the store.
func (k Keeper) SetConfirm(
	ctx sdk.Context,
	groupID tss.GroupID,
	confirm types.Confirm,
) {
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
		return types.Confirm{}, errorsmod.Wrapf(
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

// GetConfirmIterator gets an iterator over all confirm data of a group.
func (k Keeper) GetConfirmIterator(ctx sdk.Context, groupID tss.GroupID) sdk.Iterator {
	return sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.ConfirmStoreKey(groupID))
}

// GetConfirms retrieves all confirm for a given group from the store.
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

// DeleteConfirm deletes the confirm of a member from the store.
func (k Keeper) DeleteConfirm(ctx sdk.Context, groupID tss.GroupID, memberID tss.MemberID) {
	ctx.KVStore(k.storeKey).Delete(types.ConfirmMemberStoreKey(groupID, memberID))
}

// DeleteConfirms removes all confirm with a specific group ID from the store.
func (k Keeper) DeleteConfirms(ctx sdk.Context, groupID tss.GroupID) {
	iterator := k.GetConfirmIterator(ctx, groupID)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		key := iterator.Key()
		ctx.KVStore(k.storeKey).Delete(key)
	}
}

// SetConfirmComplainCount sets the confirm complaint count for a specific groupID in the store.
func (k Keeper) SetConfirmComplainCount(ctx sdk.Context, groupID tss.GroupID, count uint64) {
	ctx.KVStore(k.storeKey).Set(types.ConfirmComplainCountStoreKey(groupID), sdk.Uint64ToBigEndian(count))
}

// GetConfirmComplainCount retrieves the confirm complaint count for a specific groupID from the store.
func (k Keeper) GetConfirmComplainCount(ctx sdk.Context, groupID tss.GroupID) uint64 {
	bz := ctx.KVStore(k.storeKey).Get(types.ConfirmComplainCountStoreKey(groupID))
	return sdk.BigEndianToUint64(bz)
}

// AddConfirmComplaintCount increments the count of confirm and complaint in the store.
func (k Keeper) AddConfirmComplaintCount(ctx sdk.Context, groupID tss.GroupID) {
	count := k.GetConfirmComplainCount(ctx, groupID)
	k.SetConfirmComplainCount(ctx, groupID, count+1)
}

// DeleteConfirmComplainCount remove the confirm complaint count data of a group from the store.
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
	k.SetMember(ctx, member)
	return nil
}

// DeleteAllDKGInterimData deletes all DKG interim data for a given groupID.
func (k Keeper) DeleteAllDKGInterimData(
	ctx sdk.Context,
	groupID tss.GroupID,
) {
	// Delete DKG context
	k.DeleteDKGContext(ctx, groupID)
	// Delete round 1 infos
	k.DeleteRound1Infos(ctx, groupID)
	// Delete round 1 info count
	k.DeleteRound1InfoCount(ctx, groupID)
	// Delete round 2 infos
	k.DeleteRound2Infos(ctx, groupID)
	// Delete round 2 info count
	k.DeleteRound2InfoCount(ctx, groupID)
	// Delete all complaint with status
	k.DeleteAllComplainsWithStatus(ctx, groupID)
	// Delete confirms
	k.DeleteConfirms(ctx, groupID)
	// Delete accumulated commits
	k.DeleteAccumulatedCommits(ctx, groupID)
	// Delete confirm complaint count
	k.DeleteConfirmComplainCount(ctx, groupID)
}
