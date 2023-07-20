package keeper

import (
	"encoding/hex"
	"fmt"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"golang.org/x/exp/slices"

	"github.com/bandprotocol/chain/v2/pkg/bandrng"
	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

// SetSigningCount sets the number of signing count to the given value.
func (k Keeper) SetSigningCount(ctx sdk.Context, count uint64) {
	ctx.KVStore(k.storeKey).Set(types.SigningCountStoreKey, sdk.Uint64ToBigEndian(count))
}

// GetSigningCount returns the current number of all signing ever existed.
func (k Keeper) GetSigningCount(ctx sdk.Context) uint64 {
	return sdk.BigEndianToUint64(ctx.KVStore(k.storeKey).Get(types.SigningCountStoreKey))
}

// GetNextSigningID increments the signing count and returns the current number of signing.
func (k Keeper) GetNextSigningID(ctx sdk.Context) tss.SigningID {
	signingNumber := k.GetSigningCount(ctx)
	k.SetSigningCount(ctx, signingNumber+1)
	return tss.SigningID(signingNumber + 1)
}

// SetSigning sets the signing data for a given signing ID.
func (k Keeper) SetSigning(ctx sdk.Context, signing types.Signing) {
	ctx.KVStore(k.storeKey).Set(types.SigningStoreKey(signing.SigningID), k.cdc.MustMarshal(&signing))
}

// GetSigning retrieves the signing data for a given signing ID from the store.
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

// MustGetSigning returns the signing for the given ID. Panics error if not exists.
func (k Keeper) MustGetSigning(ctx sdk.Context, signingID tss.SigningID) types.Signing {
	signing, err := k.GetSigning(ctx, signingID)
	if err != nil {
		panic(err)
	}
	return signing
}

// AddSigning adds the signing data to the store and returns the new signing ID.
func (k Keeper) AddSigning(ctx sdk.Context, signing types.Signing) tss.SigningID {
	signingID := k.GetNextSigningID(ctx)
	signing.SigningID = signingID
	signing.CreatedHeight = ctx.BlockHeader().Height
	k.SetSigning(ctx, signing)

	return signingID
}

// DeleteSigning deletes the signing data for a given signing ID from the store.
func (k Keeper) DeleteSigning(ctx sdk.Context, signingID tss.SigningID) {
	ctx.KVStore(k.storeKey).Delete(types.SigningStoreKey(signingID))
}

// GetPendingSigns retrieves the pending signing objects associated with the given account address.
func (k Keeper) GetPendingSigns(ctx sdk.Context, address sdk.AccAddress) []types.Signing {
	// Get the ID of the last expired signing
	lastExpired := k.GetLastExpiredSigningID(ctx)

	// Get the total signing count
	signingCount := k.GetSigningCount(ctx)

	var pendingSigns []types.Signing
	for id := lastExpired + 1; uint64(id) <= signingCount; id++ {
		// Retrieve the signing object
		signing := k.MustGetSigning(ctx, id)

		// Iterate over the assigned members in the signing object
		for _, am := range signing.AssignedMembers {
			// Check if the member's address matches the given account address
			if am.Member == address.String() {
				// Add the signing to the pendingSigns slice
				pendingSigns = append(pendingSigns, signing)
			}
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

// SetPartialSig sets the partial signature for a specific signing ID and member ID.
func (k Keeper) SetPartialSig(ctx sdk.Context, signingID tss.SigningID, memberID tss.MemberID, sig tss.Signature) {
	k.AddSigCount(ctx, signingID)
	ctx.KVStore(k.storeKey).Set(types.PartialSigMemberStoreKey(signingID, memberID), sig)
}

// GetPartialSig retrieves the partial signature for a specific signing ID and member ID from the store.
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

// GetPartialSigIterator gets an iterator over all partial signature of the signing.
func (k Keeper) GetPartialSigIterator(ctx sdk.Context, signingID tss.SigningID) sdk.Iterator {
	return sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.PartialSigStoreKey(signingID))
}

// GetPartialSigs retrieves all partial signatures for a specific signing ID from the store.
func (k Keeper) GetPartialSigs(ctx sdk.Context, signingID tss.SigningID) tss.Signatures {
	var pzs tss.Signatures
	iterator := k.GetPartialSigIterator(ctx, signingID)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		pzs = append(pzs, iterator.Value())
	}
	return pzs
}

// GetPartialSigsWithKey retrieves all partial signatures for a specific signing ID from the store along with their corresponding member IDs.
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

// GetRandomAssigningParticipants generates a random selection of participants for a signing process.
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

// HandleRequestSign initiates the signing process by requesting signatures from assigned members.
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
		Status:          types.SIGNING_STATUS_WAITING,
	}

	// Add signing
	signingID := k.AddSigning(ctx, signing)

	event := sdk.NewEvent(
		types.EventTypeRequestSign,
		sdk.NewAttribute(types.AttributeKeyGroupID, fmt.Sprintf("%d", groupID)),
		sdk.NewAttribute(types.AttributeKeySigningID, fmt.Sprintf("%d", signingID)),
		sdk.NewAttribute(types.AttributeKeyMessage, hex.EncodeToString(msg)),
		sdk.NewAttribute(types.AttributeKeyGroupPubNonce, hex.EncodeToString(groupPubNonce)),
	)
	for _, am := range assignedMembers {
		event = event.AppendAttributes(
			sdk.NewAttribute(types.AttributeKeyMemberID, fmt.Sprintf("%d", am.MemberID)),
			sdk.NewAttribute(types.AttributeKeyMember, fmt.Sprintf("%s", am.Member)),
			sdk.NewAttribute(types.AttributeKeyBindingFactor, hex.EncodeToString(am.BindingFactor)),
			sdk.NewAttribute(types.AttributeKeyPubNonce, hex.EncodeToString(am.PubNonce)),
			sdk.NewAttribute(types.AttributeKeyPubD, hex.EncodeToString(am.PubD)),
			sdk.NewAttribute(types.AttributeKeyPubE, hex.EncodeToString(am.PubE)),
		)
	}
	ctx.EventManager().EmitEvent(event)

	return signingID, nil
}

// SetLastExpiredSigningID sets the last expired signing ID in the store.
func (k Keeper) SetLastExpiredSigningID(ctx sdk.Context, signingID tss.SigningID) {
	ctx.KVStore(k.storeKey).Set(types.LastExpiredSigningIDStoreKey, sdk.Uint64ToBigEndian(uint64(signingID)))
}

// GetLastExpiredSigningID retrieves the last expired signing ID from the store.
func (k Keeper) GetLastExpiredSigningID(ctx sdk.Context) tss.SigningID {
	bz := ctx.KVStore(k.storeKey).Get(types.LastExpiredSigningIDStoreKey)
	return tss.SigningID(sdk.BigEndianToUint64(bz))
}

// ProcessExpiredSignings cleans up expired signings and removes them from the store.
func (k Keeper) ProcessExpiredSignings(ctx sdk.Context) {
	// Get the current signing ID to start processing from
	currentSigningID := k.GetLastExpiredSigningID(ctx) + 1

	// Get the last signing ID in the store
	lastSigningID := tss.SigningID(k.GetSigningCount(ctx))

	// Process each signing starting from currentSigningID
	for ; currentSigningID <= lastSigningID; currentSigningID++ {
		// Get the signing
		signing := k.MustGetSigning(ctx, currentSigningID)

		// Check if the signing is still within the expiration period
		if signing.CreatedHeight+k.SigningPeriod(ctx) > ctx.BlockHeight() {
			break
		}

		mids := types.AssignedMembers(signing.AssignedMembers).MemberIDs()
		pzs := k.GetPartialSigsWithKey(ctx, signing.SigningID)
		for _, mid := range mids {
			// Check if the member's partial signature is found in the list of partial signatures.
			found := slices.ContainsFunc(pzs, func(pz types.PartialSignature) bool { return pz.MemberID == mid })

			// If the partial signature is not found, deactivate the member
			if !found {
				member := k.MustGetMember(ctx, signing.GroupID, mid)
				member.IsActive = false
				k.SetMember(ctx, signing.GroupID, member)
			}
		}

		// Set the signing status to expired
		signing.Status = types.SIGNING_STATUS_EXPIRED
		k.SetSigning(ctx, signing)

		// Set the last expired signing ID to the current signing ID
		k.SetLastExpiredSigningID(ctx, currentSigningID)
	}
}
