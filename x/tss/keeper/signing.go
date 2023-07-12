package keeper

import (
	"encoding/hex"
	"fmt"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	gogotypes "github.com/gogo/protobuf/types"

	"github.com/bandprotocol/chain/v2/pkg/bandrng"
	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

// SetSigningCount function sets the number of signing count to the given value.
func (k Keeper) SetSigningCount(ctx sdk.Context, count uint64) {
	ctx.KVStore(k.storeKey).Set(types.SigningCountStoreKey, sdk.Uint64ToBigEndian(count))
}

// GetSigningCount function returns the current number of all signing ever existed.
func (k Keeper) GetSigningCount(ctx sdk.Context) uint64 {
	return sdk.BigEndianToUint64(ctx.KVStore(k.storeKey).Get(types.SigningCountStoreKey))
}

// GetNextSigningID function increments the signing count and returns the current number of signing.
func (k Keeper) GetNextSigningID(ctx sdk.Context) tss.SigningID {
	signingNumber := k.GetSigningCount(ctx)
	k.SetSigningCount(ctx, signingNumber+1)
	return tss.SigningID(signingNumber + 1)
}

// SetSigning function sets the signing data for a given signing ID.
func (k Keeper) SetSigning(ctx sdk.Context, signing types.Signing) {
	ctx.KVStore(k.storeKey).Set(types.SigningStoreKey(signing.SigningID), k.cdc.MustMarshal(&signing))
}

// GetSigning function retrieves the signing data for a given signing ID from the store.
func (k Keeper) GetSigning(ctx sdk.Context, signingID tss.SigningID) (types.Signing, error) {
	bz := ctx.KVStore(k.storeKey).Get(types.SigningStoreKey(signingID))
	if bz == nil {
		return types.Signing{}, sdkerrors.Wrapf(
			types.ErrSigningNotFound,
			"failed to get Signing with ID: %d",
			signingID,
		)
	}
	var signing types.Signing
	k.cdc.MustUnmarshal(bz, &signing)
	return signing, nil
}

// AddSigning function adds the signing data to the store and returns the new signing ID.
func (k Keeper) AddSigning(ctx sdk.Context, signing types.Signing) tss.SigningID {
	signingID := k.GetNextSigningID(ctx)
	signing.SigningID = signingID
	signing.CreatedAt = ctx.BlockHeader().Time
	expiredTime := signing.CreatedAt.Add(k.SigningPeriod(ctx))
	signing.Expiration = &expiredTime
	k.SetSigning(ctx, signing)

	return signingID
}

// DeleteSigning function deletes the signing data for a given signing ID from the store.
func (k Keeper) DeleteSigning(ctx sdk.Context, signingID tss.SigningID) {
	ctx.KVStore(k.storeKey).Delete(types.SigningStoreKey(signingID))
}

// SetPendingSign function sets the pending sign flag for a specific address and signing ID.
func (k Keeper) SetPendingSign(ctx sdk.Context, address sdk.AccAddress, signingID tss.SigningID) {
	bz := k.cdc.MustMarshal(&gogotypes.BoolValue{Value: true})
	ctx.KVStore(k.storeKey).Set(types.PendingSignStoreKey(address, signingID), bz)
}

// GetPendingSign function retrieves the pending sign flag for a specific address and signing ID from the store.
func (k Keeper) GetPendingSign(ctx sdk.Context, address sdk.AccAddress, signingID tss.SigningID) bool {
	bz := ctx.KVStore(k.storeKey).Get(types.PendingSignStoreKey(address, signingID))
	var have gogotypes.BoolValue
	if bz == nil {
		return false
	}
	k.cdc.MustUnmarshal(bz, &have)

	return have.Value
}

// DeletePendingSign function deletes the pending sign flag for a specific address and signing ID from the store.
func (k Keeper) DeletePendingSign(ctx sdk.Context, address sdk.AccAddress, signingID tss.SigningID) {
	ctx.KVStore(k.storeKey).Delete(types.PendingSignStoreKey(address, signingID))
}

// GetPendingSignIterator function gets an iterator over all pending sign data.
func (k Keeper) GetPendingSignIterator(ctx sdk.Context, address sdk.AccAddress) sdk.Iterator {
	return sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.PendingSignsStoreKey(address))
}

// GetPendingSignIDs method retrieves all pending sign ids for a given address from the store.
func (k Keeper) GetPendingSignIDs(ctx sdk.Context, address sdk.AccAddress) []uint64 {
	var pendingSigns []uint64
	iterator := k.GetPendingSignIterator(ctx, address)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var have gogotypes.BoolValue
		k.cdc.MustUnmarshal(iterator.Value(), &have)
		if have.Value {
			pendingSigns = append(pendingSigns, types.SigningIDFromPendingSignStoreKey(iterator.Key()))
		}
	}
	return pendingSigns
}

// SetSigCount sets the count of signature data for a sign in the store.
func (k Keeper) SetSigCount(ctx sdk.Context, signingID tss.SigningID, count uint64) {
	ctx.KVStore(k.storeKey).Set(types.SigCountStoreKey(signingID), sdk.Uint64ToBigEndian(count))
}

// GetSigCount retrieves the count of signature data for a sign from the store.
func (k Keeper) GetSigCount(ctx sdk.Context, signingID tss.SigningID) uint64 {
	bz := ctx.KVStore(k.storeKey).Get(types.SigCountStoreKey(signingID))
	return sdk.BigEndianToUint64(bz)
}

// AddSigCount increments the count of signature data for a sign in the store.
func (k Keeper) AddSigCount(ctx sdk.Context, signingID tss.SigningID) {
	count := k.GetSigCount(ctx, signingID)
	k.SetSigCount(ctx, signingID, count+1)
}

// DeleteSigCount delete the signature count data of a sign from the store.
func (k Keeper) DeleteSigCount(ctx sdk.Context, signingID tss.SigningID) {
	ctx.KVStore(k.storeKey).Delete(types.SigCountStoreKey(signingID))
}

// SetPartialSig function sets the partial signature for a specific signing ID and member ID.
func (k Keeper) SetPartialSig(ctx sdk.Context, signingID tss.SigningID, memberID tss.MemberID, sig tss.Signature) {
	k.AddSigCount(ctx, signingID)
	ctx.KVStore(k.storeKey).Set(types.PartialSigMemberStoreKey(signingID, memberID), sig)
}

// GetPartialSig function retrieves the partial signature for a specific signing ID and member ID from the store.
func (k Keeper) GetPartialSig(ctx sdk.Context, signingID tss.SigningID, memberID tss.MemberID) (tss.Signature, error) {
	bz := ctx.KVStore(k.storeKey).Get(types.PartialSigMemberStoreKey(signingID, memberID))
	if bz == nil {
		return nil, sdkerrors.Wrapf(
			types.ErrPartialSigNotFound,
			"failed to get partial signature with signingID: %d memberID: %d",
			signingID,
			memberID,
		)
	}
	return bz, nil
}

// DeletePartialSig delete the partial sign data of a sign from the store.
func (k Keeper) DeletePartialSig(ctx sdk.Context, signingID tss.SigningID, memberID tss.MemberID) {
	ctx.KVStore(k.storeKey).Delete(types.PartialSigMemberStoreKey(signingID, memberID))
}

// GetPartialSigIterator function gets an iterator over all partial signature of the signing.
func (k Keeper) GetPartialSigIterator(ctx sdk.Context, signingID tss.SigningID) sdk.Iterator {
	return sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.PartialSigStoreKey(signingID))
}

// GetPartialSigs function retrieves all partial signatures for a specific signing ID from the store.
func (k Keeper) GetPartialSigs(ctx sdk.Context, signingID tss.SigningID) tss.Signatures {
	var pzs tss.Signatures
	iterator := k.GetPartialSigIterator(ctx, signingID)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		pzs = append(pzs, iterator.Value())
	}
	return pzs
}

// GetPartialSigsWithKey function retrieves all partial signatures for a specific signing ID from the store along with their corresponding member IDs.
func (k Keeper) GetPartialSigsWithKey(ctx sdk.Context, signingID tss.SigningID) []types.PartialSignature {
	var pzs []types.PartialSignature
	iterator := k.GetPartialSigIterator(ctx, signingID)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		pzs = append(pzs, types.PartialSignature{
			MemberID:  types.MemberIDFromPartialSignMemberStoreKey(iterator.Key()),
			Signature: iterator.Value(),
		})
	}
	return pzs
}

// GetRandomAssigningParticipants function generates a random selection of participants for a signing process.
// It selects 't' participants out of 'members size' participants using a deterministic random number generator (DRBG).
func (k Keeper) GetRandomAssigningParticipants(
	ctx sdk.Context,
	signingID uint64,
	members []types.Member,
	t uint64,
) ([]types.Member, error) {
	members_size := uint64(len(members))
	if t > members_size {
		return nil, sdkerrors.Wrapf(types.ErrUnexpectedThreshold, "t must less than or equal to size")
	}

	// Create a deterministic random number generator (DRBG) using the rolling seed, signingID, and chain ID.
	rng, err := bandrng.NewRng(
		k.rollingseedKeeper.GetRollingSeed(ctx),
		sdk.Uint64ToBigEndian(signingID),
		[]byte(ctx.ChainID()),
	)
	if err != nil {
		return nil, sdkerrors.Wrapf(types.ErrBadDrbgInitialization, err.Error())
	}

	var selected []types.Member
	for i := uint64(0); i < t; i++ {
		luckyNumber := rng.NextUint64() % members_size

		// Get the selected member.
		selected = append(selected, members[luckyNumber])

		// Remove the selected member from the list.
		members = append(members[:luckyNumber], members[luckyNumber+1:]...)

		members_size -= 1
	}

	// Sort selected members
	sort.Slice(selected, func(i, j int) bool { return selected[i].MemberID < selected[j].MemberID })

	return selected, nil
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

	// Get active members
	members, err := k.GetActiveMembers(ctx, groupID)
	if err != nil {
		return 0, err
	}

	// Random assigning participants
	selectedMembers, err := k.GetRandomAssigningParticipants(
		ctx,
		k.GetSigningCount(ctx)+1,
		members,
		group.Threshold,
	)
	if err != nil {
		return 0, err
	}

	// Handle assigned members by polling DE and retrieve assigned members information.
	assignedMembers, err := k.HandleAssignedMembersPollDE(ctx, selectedMembers)
	if err != nil {
		return 0, err
	}

	// Compute commitment from mids, public D and public E
	commitment, err := tss.ComputeCommitment(
		types.Members(selectedMembers).GetIDs(),
		assignedMembers.PubDs(),
		assignedMembers.PubEs(),
	)
	if err != nil {
		return 0, err
	}

	// Compute binding factor and public nonce of each assigned member
	for i, member := range assignedMembers {
		// Compute binding factor
		assignedMembers[i].BindingFactor, err = tss.ComputeOwnBindingFactor(member.MemberID, msg, commitment)
		if err != nil {
			return 0, err
		}
		// Compute own public nonce
		assignedMembers[i].PubNonce, err = tss.ComputeOwnPubNonce(
			member.PubD,
			member.PubE,
			assignedMembers[i].BindingFactor,
		)
		if err != nil {
			return 0, err
		}
	}

	// Compute group public nonce for this signing
	groupPubNonce, err := tss.ComputeGroupPublicNonce(assignedMembers.PubNonces()...)
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

	// Set pending sign
	for _, member := range selectedMembers {
		accMember, err := sdk.AccAddressFromBech32(member.Address)
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
		sdk.NewAttribute(types.AttributeKeyGroupPubNonce, hex.EncodeToString(groupPubNonce)),
	)
	for _, member := range assignedMembers {
		event = event.AppendAttributes(
			sdk.NewAttribute(types.AttributeKeyMemberID, fmt.Sprintf("%d", member.MemberID)),
			sdk.NewAttribute(types.AttributeKeyMember, fmt.Sprintf("%s", member.Member)),
			sdk.NewAttribute(types.AttributeKeyBindingFactor, hex.EncodeToString(member.BindingFactor)),
			sdk.NewAttribute(types.AttributeKeyPubNonce, hex.EncodeToString(member.PubNonce)),
			sdk.NewAttribute(types.AttributeKeyPubD, hex.EncodeToString(member.PubD)),
			sdk.NewAttribute(types.AttributeKeyPubE, hex.EncodeToString(member.PubE)),
		)
	}
	ctx.EventManager().EmitEvent(event)

	return signingID, nil
}
