package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

type Keeper struct {
	cdc         codec.BinaryCodec
	storeKey    storetypes.StoreKey
	authzKeeper types.AuthzKeeper
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	authzKeeper types.AuthzKeeper,
) Keeper {
	return Keeper{
		cdc:         cdc,
		storeKey:    storeKey,
		authzKeeper: authzKeeper,
	}
}

// SetGroupCount function sets the number of group count to the given value.
func (k Keeper) SetGroupCount(ctx sdk.Context, count uint64) {
	ctx.KVStore(k.storeKey).Set(types.GroupCountStoreKey, sdk.Uint64ToBigEndian(count))
}

// GetGroupCount function returns the current number of all groups ever existed.
func (k Keeper) GetGroupCount(ctx sdk.Context) uint64 {
	return sdk.BigEndianToUint64(ctx.KVStore(k.storeKey).Get(types.GroupCountStoreKey))
}

// GetNextGroupID function increments the group count and returns the current number of groups.
func (k Keeper) GetNextGroupID(ctx sdk.Context) tss.GroupID {
	groupNumber := k.GetGroupCount(ctx)
	k.SetGroupCount(ctx, groupNumber+1)
	return tss.GroupID(groupNumber + 1)
}

// IsGrantee function checks if the granter granted permissions to the grantee.
func (k Keeper) IsGrantee(ctx sdk.Context, granter sdk.AccAddress, grantee sdk.AccAddress) bool {
	for _, msg := range types.MsgGrants {
		cap, _ := k.authzKeeper.GetAuthorization(
			ctx,
			grantee,
			granter,
			msg,
		)

		if cap == nil {
			return false
		}
	}

	return true
}

// CreateNewGroup function creates a new group in the store and returns the id of the group.
func (k Keeper) CreateNewGroup(ctx sdk.Context, group types.Group) tss.GroupID {
	groupID := k.GetNextGroupID(ctx)
	ctx.KVStore(k.storeKey).Set(types.GroupStoreKey(tss.GroupID(groupID)), k.cdc.MustMarshal(&group))
	return groupID
}

// GetGroup function retrieves a group from the store.
func (k Keeper) GetGroup(ctx sdk.Context, groupID tss.GroupID) (types.Group, error) {
	bz := ctx.KVStore(k.storeKey).Get(types.GroupStoreKey(groupID))
	if bz == nil {
		return types.Group{}, sdkerrors.Wrapf(types.ErrGroupNotFound, "failed to get group with groupID: %d", groupID)
	}

	group := types.Group{}
	k.cdc.MustUnmarshal(bz, &group)
	return group, nil
}

// UpdateGroup function updates a group in the store.
func (k Keeper) UpdateGroup(ctx sdk.Context, groupID tss.GroupID, group types.Group) {
	ctx.KVStore(k.storeKey).Set(types.GroupStoreKey(groupID), k.cdc.MustMarshal(&group))
}

// SetDKGContext function sets DKG context for a group in the store.
func (k Keeper) SetDKGContext(ctx sdk.Context, groupID tss.GroupID, dkgContext []byte) {
	ctx.KVStore(k.storeKey).Set(types.DKGContextStoreKey(groupID), dkgContext)
}

// GetDKGContext function retrieves DKG context of a group from the store.
func (k Keeper) GetDKGContext(ctx sdk.Context, groupID tss.GroupID) ([]byte, error) {
	bz := ctx.KVStore(k.storeKey).Get(types.DKGContextStoreKey(groupID))
	if bz == nil {
		return nil, sdkerrors.Wrapf(types.ErrDKGContextNotFound, "failed to get dkg-context with groupID: %d", groupID)
	}
	return bz, nil
}

// DeleteDKGContext removes the DKG context data of a group from the store.
func (k Keeper) DeleteDKGContext(ctx sdk.Context, groupID tss.GroupID) {
	ctx.KVStore(k.storeKey).Delete(types.DKGContextStoreKey(groupID))
}

// SetMember function sets a member of a group in the store.
func (k Keeper) SetMember(ctx sdk.Context, groupID tss.GroupID, memberID tss.MemberID, member types.Member) {
	ctx.KVStore(k.storeKey).Set(types.MemberOfGroupKey(groupID, memberID), k.cdc.MustMarshal(&member))
}

// GetMember function retrieves a member of a group from the store.
func (k Keeper) GetMember(ctx sdk.Context, groupID tss.GroupID, memberID tss.MemberID) (types.Member, error) {
	bz := ctx.KVStore(k.storeKey).Get(types.MemberOfGroupKey(groupID, memberID))
	if bz == nil {
		return types.Member{}, sdkerrors.Wrapf(
			types.ErrMemberNotFound,
			"failed to get member with groupID: %d and memberID: %d",
			groupID,
			memberID,
		)
	}

	member := types.Member{}
	k.cdc.MustUnmarshal(bz, &member)
	return member, nil
}

// GetMembersIterator function gets an iterator over all members of a group.
func (k Keeper) GetMembersIterator(ctx sdk.Context, groupID tss.GroupID) sdk.Iterator {
	return sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.MembersStoreKey(groupID))
}

// GetMembers function retrieves all members of a group from the store.
func (k Keeper) GetMembers(ctx sdk.Context, groupID tss.GroupID) ([]types.Member, error) {
	var members []types.Member
	iterator := k.GetMembersIterator(ctx, groupID)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var member types.Member
		k.cdc.MustUnmarshal(iterator.Value(), &member)
		members = append(members, member)
	}
	if len(members) == 0 {
		return nil, sdkerrors.Wrapf(types.ErrGroupNotFound, "failed to get members with groupID: %d", groupID)
	}
	return members, nil
}

// VerifyMember function verifies if a member is part of a group.
func (k Keeper) VerifyMember(ctx sdk.Context, groupID tss.GroupID, memberID tss.MemberID, memberAddress string) bool {
	member, err := k.GetMember(ctx, groupID, memberID)
	if err != nil || member.Member != memberAddress {
		return false
	}
	return true
}

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
			types.ErrRoundExpired,
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
func (k Keeper) HandleVerifyComplainSig(
	ctx sdk.Context,
	groupID tss.GroupID,
	complain types.Complain,
) error {
	// Get the member I from the store
	memberI, err := k.GetMember(ctx, groupID, complain.I)
	if err != nil {
		return err
	}

	// Get the member J from the store
	memberJ, err := k.GetMember(ctx, groupID, complain.J)
	if err != nil {
		return err
	}

	// Verify the complain signature
	err = tss.VerifyComplainSig(memberI.PubKey, memberJ.PubKey, complain.KeySym, complain.NonceSym, complain.Signature)
	if err != nil {
		return sdkerrors.Wrapf(
			types.ErrComplainFailed,
			"failed to complain member: %d with groupID: %d; %s",
			memberJ,
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

// HandleComputeGroupPublicKey computes the group public key for a given groupID.
func (k Keeper) HandleComputeGroupPublicKey(ctx sdk.Context, groupID tss.GroupID) (tss.PublicKey, error) {
	var rawA0Commits tss.Points
	allRound1Data := k.GetAllRound1Data(ctx, groupID)
	for _, r1 := range allRound1Data {
		rawA0Commits = append(rawA0Commits, (r1.CoefficientsCommit[0]))
	}

	groupPubKey, err := tss.ComputeGroupPublicKey(rawA0Commits...)
	if err != nil {
		return nil, sdkerrors.Wrapf(
			types.ErrConfirmFailed,
			"failed to compute group public key; %s",
			err,
		)
	}
	return groupPubKey, nil
}

// SetComplainsWithStatus sets the complains with status for a specific groupID and memberID in the store.
func (k Keeper) SetComplainsWithStatus(
	ctx sdk.Context,
	groupID tss.GroupID,
	complainsWithStatus types.ComplainsWithStatus,
) {
	// add confirm complain count
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
func (k Keeper) DeleteAllDKGInterimData(ctx sdk.Context, groupID tss.GroupID, groupSize uint64) {
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

	// Delete round 1 data count
	k.DeleteRound1DataCount(ctx, groupID)
	// Delete round 2 data count
	k.DeleteRound2DataCount(ctx, groupID)
	// Delete confirm complain count
	k.DeleteConfirmComplainCount(ctx, groupID)
}

func (k Keeper) SetDE(ctx sdk.Context, address sdk.AccAddress, index uint64, de types.DE) {
	ctx.KVStore(k.storeKey).Set(types.DEIndexStoreKey(address, index), k.cdc.MustMarshal(&de))
}

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

func (k Keeper) SetDEQueue(ctx sdk.Context, address sdk.AccAddress, deQueue types.DEQueue) {
	ctx.KVStore(k.storeKey).Set(types.DEQueueKeyStoreKey(address), k.cdc.MustMarshal(&deQueue))
}

func (k Keeper) GetDEQueue(ctx sdk.Context, address sdk.AccAddress) types.DEQueue {
	var deQueue types.DEQueue
	k.cdc.MustUnmarshal(ctx.KVStore(k.storeKey).Get(types.DEQueueKeyStoreKey(address)), &deQueue)
	return deQueue
}

func (k Keeper) DeleteDEQueue(ctx sdk.Context, address sdk.AccAddress, index uint64) {
	ctx.KVStore(k.storeKey).Delete(types.DEIndexStoreKey(address, index))
}

func (k Keeper) PollDEPairs(ctx sdk.Context, address sdk.AccAddress) (types.DE, error) {
	deQueue := k.GetDEQueue(ctx, address)
	de, err := k.GetDE(ctx, address, deQueue.Head)
	if err != nil {
		return types.DE{}, err
	}

	k.DeleteDEQueue(ctx, address, deQueue.Head)

	deQueue.Head += 1
	k.SetDEQueue(ctx, address, deQueue)

	return de, nil
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
