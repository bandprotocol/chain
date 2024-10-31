package keeper

import (
	"encoding/hex"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/pkg/tss"
	"github.com/bandprotocol/chain/v3/x/tss/types"
)

// RequestSigning creates a signing request on the given content from the specific groupID.
func (k Keeper) RequestSigning(
	ctx sdk.Context,
	groupID tss.GroupID,
	originator types.Originator,
	content types.Content,
) (tss.SigningID, error) {
	// convert content to bytes
	if !k.contentRouter.HasRoute(content.OrderRoute()) {
		return 0, types.ErrNoSignatureOrderHandlerExists.Wrap(content.OrderRoute())
	}
	handler := k.contentRouter.GetRoute(content.OrderRoute())
	contentMsg, err := handler(ctx, content)
	if err != nil {
		return 0, err
	}

	// convert originator to bytes
	originatorBz, err := originator.Encode()
	if err != nil {
		return 0, types.ErrEncodeOriginatorFailed
	}

	// create signing object
	signingID, err := k.CreateSigning(ctx, groupID, originatorBz, contentMsg)
	if err != nil {
		return 0, err
	}

	// initiate new signing round
	if err = k.InitiateNewSigningRound(ctx, signingID); err != nil {
		return 0, err
	}

	return signingID, nil
}

// AssignMembersForSigning handles the assignment of members for a group signature process.
func (k Keeper) AssignMembersForSigning(
	ctx sdk.Context,
	groupID tss.GroupID,
	msg []byte,
	nonce []byte,
) (types.AssignedMembers, error) {
	// Random assigning members
	selectedMembers, err := k.GetRandomMembers(ctx, groupID, nonce)
	if err != nil {
		return types.AssignedMembers{}, err
	}

	// Handle assigned members by polling DE and retrieve assigned members information.
	des, err := k.DequeueDEs(ctx, selectedMembers)
	if err != nil {
		return types.AssignedMembers{}, err
	}

	var assignedMembers types.AssignedMembers
	for i, member := range selectedMembers {
		am := types.NewAssignedMember(member, des[i], nil, nil)
		assignedMembers = append(assignedMembers, am)
	}

	// Compute commitment from mids, public D, and public E
	commitment, err := tss.ComputeCommitment(
		types.Members(selectedMembers).GetIDs(),
		assignedMembers.PubDs(),
		assignedMembers.PubEs(),
	)
	if err != nil {
		return types.AssignedMembers{}, err
	}

	// Compute binding factor and public nonce of each assigned member
	for i, member := range assignedMembers {
		// Compute binding factor
		assignedMembers[i].BindingFactor, err = tss.ComputeOwnBindingFactor(member.MemberID, msg, commitment)
		if err != nil {
			return types.AssignedMembers{}, err
		}
		// Compute own public nonce
		assignedMembers[i].PubNonce, err = tss.ComputeOwnPubNonce(
			member.PubD,
			member.PubE,
			assignedMembers[i].BindingFactor,
		)
		if err != nil {
			return types.AssignedMembers{}, err
		}
	}

	return assignedMembers, nil
}

// CreateSigning creates a signing object from an originator and contentMsg.
func (k Keeper) CreateSigning(
	ctx sdk.Context,
	groupID tss.GroupID,
	originator []byte,
	contentMsg []byte,
) (tss.SigningID, error) {
	// get signing message
	nextSigningID := k.GetSigningCount(ctx) + 1
	message := types.EncodeSigning(ctx, nextSigningID, originator, contentMsg)

	// Check group status
	group, err := k.GetGroup(ctx, groupID)
	if err != nil {
		return 0, err
	}
	if group.Status != types.GROUP_STATUS_ACTIVE {
		return 0, types.ErrGroupIsNotActive.Wrap("group status is not active")
	}

	// set new signing object
	signing := types.NewSigning(
		tss.SigningID(nextSigningID),
		0,
		groupID,
		group.PubKey,
		originator,
		message,
		nil,
		nil,
		types.SIGNING_STATUS_WAITING,
		uint64(ctx.BlockHeight()),
		ctx.BlockTime(),
	)
	k.SetSigning(ctx, signing)
	k.SetSigningCount(ctx, nextSigningID)

	return tss.SigningID(nextSigningID), nil
}

// InitiateNewSigningRound updates the signing information (finding new assigned members, generate
// group public nonce, and set the expiration height) and return the new signing information.
func (k Keeper) InitiateNewSigningRound(ctx sdk.Context, signingID tss.SigningID) error {
	signing, err := k.GetSigning(ctx, signingID)
	if err != nil {
		return err
	}

	signing.CurrentAttempt += 1
	params := k.GetParams(ctx)
	if signing.CurrentAttempt > params.MaxSigningAttempt {
		return types.ErrMaxSigningAttemptReached.Wrapf("signingID %d", signingID)
	}

	// assigned members within the context of the group.
	nonce := append(
		sdk.Uint64ToBigEndian(uint64(signingID)),
		sdk.Uint64ToBigEndian(signing.CurrentAttempt)...,
	)
	assignedMembers, err := k.AssignMembersForSigning(ctx, signing.GroupID, signing.Message, nonce)
	if err != nil {
		return err
	}

	// Compute group public nonce for this signing
	groupPubNonce, err := tss.ComputeGroupPublicNonce(assignedMembers.PubNonces()...)
	if err != nil {
		return err
	}

	expiredHeight := uint64(ctx.BlockHeight()) + params.SigningPeriod
	signingAttempt := types.NewSigningAttempt(
		signingID,
		signing.CurrentAttempt,
		expiredHeight,
		assignedMembers,
	)
	k.SetSigningAttempt(ctx, signingAttempt)

	signing.GroupPubNonce = groupPubNonce
	signing.Status = types.SIGNING_STATUS_WAITING
	k.SetSigning(ctx, signing)
	k.AddSigningExpiration(ctx, signing.ID, signing.CurrentAttempt)

	event := sdk.NewEvent(
		types.EventTypeRequestSignature,
		sdk.NewAttribute(types.AttributeKeyGroupID, fmt.Sprintf("%d", signing.GroupID)),
		sdk.NewAttribute(types.AttributeKeySigningID, fmt.Sprintf("%d", signing.ID)),
		sdk.NewAttribute(types.AttributeKeyMessage, hex.EncodeToString(signing.Message)),
		sdk.NewAttribute(types.AttributeKeyGroupPubNonce, hex.EncodeToString(signing.GroupPubNonce)),
		sdk.NewAttribute(types.AttributeKeyAttempt, fmt.Sprintf("%d", signing.CurrentAttempt)),
	)
	for _, am := range signingAttempt.AssignedMembers {
		event = event.AppendAttributes(
			sdk.NewAttribute(types.AttributeKeyMemberID, fmt.Sprintf("%d", am.MemberID)),
			sdk.NewAttribute(types.AttributeKeyAddress, am.Address),
			sdk.NewAttribute(types.AttributeKeyBindingFactor, hex.EncodeToString(am.BindingFactor)),
			sdk.NewAttribute(types.AttributeKeyPubNonce, hex.EncodeToString(am.PubNonce)),
			sdk.NewAttribute(types.AttributeKeyPubD, hex.EncodeToString(am.PubD)),
			sdk.NewAttribute(types.AttributeKeyPubE, hex.EncodeToString(am.PubE)),
		)
	}
	ctx.EventManager().EmitEvent(event)
	return nil
}

// ==================================
// Signing store
// ==================================

// SetSigning sets the signing data of a given signing ID.
func (k Keeper) SetSigning(ctx sdk.Context, signing types.Signing) {
	ctx.KVStore(k.storeKey).Set(types.SigningStoreKey(signing.ID), k.cdc.MustMarshal(&signing))
}

// GetSigning retrieves the signing data of a given signing ID from the store.
func (k Keeper) GetSigning(ctx sdk.Context, signingID tss.SigningID) (types.Signing, error) {
	bz := ctx.KVStore(k.storeKey).Get(types.SigningStoreKey(signingID))
	if bz == nil {
		return types.Signing{}, types.ErrSigningNotFound.Wrapf(
			"failed to get Signing with ID: %d",
			signingID,
		)
	}
	var signing types.Signing
	k.cdc.MustUnmarshal(bz, &signing)
	return signing, nil
}

// MustGetSigning returns the signing of the given ID. Panics error if not exists.
func (k Keeper) MustGetSigning(ctx sdk.Context, signingID tss.SigningID) types.Signing {
	signing, err := k.GetSigning(ctx, signingID)
	if err != nil {
		panic(err)
	}
	return signing
}

// SetSigningCount sets the number of signing count to the given value.
func (k Keeper) SetSigningCount(ctx sdk.Context, count uint64) {
	ctx.KVStore(k.storeKey).Set(types.SigningCountStoreKey, sdk.Uint64ToBigEndian(count))
}

// GetSigningCount returns the current number of all signing ever existed.
func (k Keeper) GetSigningCount(ctx sdk.Context) uint64 {
	return sdk.BigEndianToUint64(ctx.KVStore(k.storeKey).Get(types.SigningCountStoreKey))
}

// ==================================
// SigningAttempt store
// ==================================

// SetSigningAttempt sets the signing attempt of a given signing ID.
func (k Keeper) SetSigningAttempt(ctx sdk.Context, sa types.SigningAttempt) {
	key := types.SigningAttemptStoreKey(sa.SigningID, sa.Attempt)
	ctx.KVStore(k.storeKey).Set(key, k.cdc.MustMarshal(&sa))
}

// GetSigningAttempt retrieves the signing attempt of a given ID and attempt from the store.
func (k Keeper) GetSigningAttempt(
	ctx sdk.Context,
	signingID tss.SigningID,
	attempt uint64,
) (types.SigningAttempt, error) {
	bz := ctx.KVStore(k.storeKey).Get(types.SigningAttemptStoreKey(signingID, attempt))
	if bz == nil {
		return types.SigningAttempt{}, types.ErrSigningAttemptNotFound.Wrapf(
			"signingID: %d and attempt: %d",
			signingID, attempt,
		)
	}

	var sa types.SigningAttempt
	k.cdc.MustUnmarshal(bz, &sa)
	return sa, nil
}

// MustGetSigningAttempt returns the signing attempt of the given ID and attempt.
// Panics error if not exists.
func (k Keeper) MustGetSigningAttempt(
	ctx sdk.Context,
	signingID tss.SigningID,
	attempt uint64,
) types.SigningAttempt {
	signing, err := k.GetSigningAttempt(ctx, signingID, attempt)
	if err != nil {
		panic(err)
	}
	return signing
}

// DeleteSigningAttempt delete signingAttempt of a given signing ID and attempt from the store.
func (k Keeper) DeleteSigningAttempt(ctx sdk.Context, signingID tss.SigningID, attempt uint64) {
	ctx.KVStore(k.storeKey).Delete(types.SigningAttemptStoreKey(signingID, attempt))
}

// MustGetCurrentAssignedMembers retrieves the assigned members of a specific signing ID from the store.
// It will panic if the signing ID or signingAttempt does not exist.
func (k Keeper) MustGetCurrentAssignedMembers(ctx sdk.Context, signingID tss.SigningID) []sdk.AccAddress {
	signing := k.MustGetSigning(ctx, signingID)
	signingAttempt := k.MustGetSigningAttempt(ctx, signingID, signing.CurrentAttempt)

	var memberAddrs []sdk.AccAddress
	for _, am := range signingAttempt.AssignedMembers {
		memberAddrs = append(memberAddrs, sdk.MustAccAddressFromBech32(am.Address))
	}

	return memberAddrs
}
