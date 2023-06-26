package keeper

import (
	"encoding/hex"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

type Keeper struct {
	cdc         codec.BinaryCodec
	storeKey    storetypes.StoreKey
	paramSpace  paramtypes.Subspace
	authzKeeper types.AuthzKeeper
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	paramSpace paramtypes.Subspace,
	authzKeeper types.AuthzKeeper,
) Keeper {
	// set KeyTable if it has not already been set
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		cdc:         cdc,
		storeKey:    storeKey,
		paramSpace:  paramSpace,
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
	for _, msg := range types.GetMsgGrants() {
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
	group.GroupID = groupID
	group.CreateTime = ctx.BlockHeader().Time
	expiryTime := group.CreateTime.Add(k.RoundPeriod(ctx))
	group.ExpiryTime = &expiryTime
	k.SetGroup(ctx, group)

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

// SetGroup function set a group in the store.
func (k Keeper) SetGroup(ctx sdk.Context, group types.Group) {
	ctx.KVStore(k.storeKey).Set(types.GroupStoreKey(group.GroupID), k.cdc.MustMarshal(&group))
}

// GetGroupsIterator function gets an iterator all group.
func (k Keeper) GetGroupsIterator(ctx sdk.Context) sdk.Iterator {
	return sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.GroupStoreKeyPrefix)
}

// GetGroups function retrieves all group of the store.
func (k Keeper) GetGroups(ctx sdk.Context) []types.Group {
	var groups []types.Group
	iterator := k.GetGroupsIterator(ctx)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var group types.Group
		k.cdc.MustUnmarshal(iterator.Value(), &group)
		groups = append(groups, group)
	}
	return groups
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
		return nil, sdkerrors.Wrapf(types.ErrMemberNotFound, "failed to get members with groupID: %d", groupID)
	}
	return members, nil
}

// HandleRequestSign function initiates the signing process by requesting signatures from assigned members.
// It assigns participants randomly, computes necessary values, and emits appropriate events.
func (k Keeper) HandleRequestSign(ctx sdk.Context, groupID tss.GroupID, msg []byte) (tss.SigningID, error) {
	// Get group
	group, err := k.GetGroup(ctx, groupID)
	if err != nil {
		return 0, err
	}

	// Check group status
	if group.Status != types.GROUP_STATUS_ACTIVE {
		return 0, sdkerrors.Wrap(types.ErrGroupIsNotActive, "group status is not active")
	}

	// Get members in the group
	members, err := k.GetMembers(ctx, groupID)
	if err != nil {
		return 0, err
	}

	// Random assigning participants
	mids, err := k.GetRandomAssigningParticipants(
		ctx,
		k.GetSigningCount(ctx)+1,
		group.Size_,
		group.Threshold,
	)
	if err != nil {
		return 0, err
	}

	// Get public D and E for each assigned members
	assignedMembers, pubDs, pubEs, err := k.HandlePollDEForAssignedMembers(ctx, mids, members)
	if err != nil {
		return 0, err
	}

	// Compute commitment from mids, public D and public E
	commitment, err := tss.ComputeCommitment(mids, pubDs, pubEs)
	if err != nil {
		return 0, err
	}

	// Compute binding factor and public nonce of each assigned member
	var ownPubNonces tss.PublicKeys
	for i, member := range assignedMembers {
		// Compute and assign binding factor and public nonce
		assignedMembers[i].PubNonce, err = tss.ComputeOwnPubNonce(
			member.PubD,
			member.PubE,
			tss.ComputeOwnBindingFactor(member.MemberID, msg, commitment),
		)
		if err != nil {
			return 0, err
		}

		ownPubNonces = append(ownPubNonces, assignedMembers[i].PubNonce)
	}

	// Compute group public nonce for this signing
	groupPubNonce, err := tss.ComputeGroupPublicNonce(ownPubNonces...)
	if err != nil {
		return 0, err
	}

	// Create signing struct
	signing := types.Signing{
		GroupID:         groupID,
		Message:         msg,
		GroupPubNonce:   groupPubNonce,
		Commitment:      commitment,
		AssignedMembers: assignedMembers,
		Signature:       nil,
	}

	// Add signing
	signingID := k.AddSigning(ctx, signing)
	for _, mid := range mids {
		accMember, err := sdk.AccAddressFromBech32(members[mid-1].Address)
		if err != nil {
			return 0, sdkerrors.Wrapf(types.ErrInvalidAccAddressFormat, err.Error())
		}

		k.SetPendingSign(ctx, accMember, signingID)
	}

	event := sdk.NewEvent(
		types.EventTypeRequestSign,
		sdk.NewAttribute(types.AttributeKeyGroupID, fmt.Sprintf("%d", groupID)),
		sdk.NewAttribute(types.AttributeKeySigningID, fmt.Sprintf("%d", signingID)),
		sdk.NewAttribute(types.AttributeKeyMessage, hex.EncodeToString(msg)),
		sdk.NewAttribute(types.AttributeKeyCommitment, hex.EncodeToString(commitment)),
		sdk.NewAttribute(types.AttributeKeyGroupPubNonce, hex.EncodeToString(groupPubNonce)),
	)
	for _, member := range assignedMembers {
		event = event.AppendAttributes(
			sdk.NewAttribute(types.AttributeKeyMemberID, fmt.Sprintf("%d", member.MemberID)),
			sdk.NewAttribute(types.AttributeKeyMember, fmt.Sprintf("%s", member.Member)),
			sdk.NewAttribute(types.AttributeKeyOwnPubNonces, hex.EncodeToString(member.PubNonce)),
			sdk.NewAttribute(types.AttributeKeyPubD, hex.EncodeToString(member.PubD)),
			sdk.NewAttribute(types.AttributeKeyPubE, hex.EncodeToString(member.PubE)),
		)
	}
	ctx.EventManager().EmitEvent(event)

	return signingID, nil
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
