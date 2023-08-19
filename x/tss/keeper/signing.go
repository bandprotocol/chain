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

// DeleteAssignedMembers deletes the assigned members for a given signing ID from the store.
func (k Keeper) DeleteAssignedMembers(ctx sdk.Context, signingID tss.SigningID) {
	signing := k.MustGetSigning(ctx, signingID)
	signing.AssignedMembers = nil
	k.SetSigning(ctx, signing)
}

// GetPendingSignings retrieves the pending signing objects associated with the given account address.
func (k Keeper) GetPendingSignings(ctx sdk.Context, address sdk.AccAddress) []uint64 {
	// Get the ID of the last expired signing
	lastExpired := k.GetLastExpiredSigningID(ctx)

	// Get the total signing count
	signingCount := k.GetSigningCount(ctx)

	var pendingSignings []uint64
	for sid := lastExpired + 1; uint64(sid) <= signingCount; sid++ {
		// Retrieve the signing object
		signing := k.MustGetSigning(ctx, sid)

		// Ignore if it's successful already
		if signing.Status == types.SIGNING_STATUS_SUCCESS {
			continue
		}

		// Check if address is assigned for signing
		for _, am := range signing.AssignedMembers {
			if am.Member == address.String() {
				// Add the signing to the pendingSignings if there is no partial sig of the member yet.
				if _, err := k.GetPartialSig(ctx, sid, am.MemberID); err != nil {
					pendingSignings = append(pendingSignings, uint64(signing.SigningID))
				}
			}
		}
	}

	return pendingSignings
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

// AddPartialSig adds the partial signature for a specific signing ID and member ID and increments the count of signature data.
func (k Keeper) AddPartialSig(ctx sdk.Context, signingID tss.SigningID, memberID tss.MemberID, sig tss.Signature) {
	k.AddSigCount(ctx, signingID)
	k.SetPartialSig(ctx, signingID, memberID, sig)
}

// SetPartialSig sets the partial signature for a specific signing ID and member ID.
func (k Keeper) SetPartialSig(ctx sdk.Context, signingID tss.SigningID, memberID tss.MemberID, sig tss.Signature) {
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

// DeletePartialSigs delete all partial signatures data of a signing from the store.
func (k Keeper) DeletePartialSigs(ctx sdk.Context, signingID tss.SigningID) {
	iterator := k.GetPartialSigIterator(ctx, signingID)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		key := iterator.Key()
		ctx.KVStore(k.storeKey).Delete(key)
	}
}

// DeletePartialSig delete a partial signature of a signing from the store.
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
func (k Keeper) HandleRequestSign(
	ctx sdk.Context,
	groupID tss.GroupID,
	content types.Content,
	feePayer sdk.AccAddress,
	feeLimit sdk.Coins,
) (tss.SigningID, error) {
	if !k.router.HasRoute(content.RequestSignatureRoute()) {
		return 0, sdkerrors.Wrap(types.ErrNoRequestSignatureHandlerExists, content.RequestSignatureRoute())
	}

	// Get group
	group, err := k.GetGroup(ctx, groupID)
	if err != nil {
		return 0, err
	}

	// Check group status
	if group.Status != types.GROUP_STATUS_ACTIVE {
		return 0, sdkerrors.Wrap(types.ErrGroupIsNotActive, "group status is not active")
	}

	handler := k.router.GetRoute(content.RequestSignatureRoute())
	msg, err := handler(ctx, content)
	if err != nil {
		return 0, sdkerrors.Wrap(types.ErrInvalidRequestSignatureContent, err.Error())
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
		AssignedMembers: assignedMembers,
		Signature:       nil,
		Fee:             sdk.NewCoins(),
		Requester:       feePayer.String(),
		Status:          types.SIGNING_STATUS_WAITING,
	}

	// Collect fees if the function is not invoked by an authority
	if feePayer.String() != k.authority {
		// set group fee
		signing.Fee = group.Fee

		// If found any coins that exceed limit then return error
		feeCoins := group.Fee.MulInt(sdk.NewInt(int64(len(assignedMembers))))
		for _, fc := range feeCoins {
			limitAmt := feeLimit.AmountOf(fc.Denom)
			if fc.Amount.GT(limitAmt) {
				return 0, sdkerrors.Wrapf(
					types.ErrNotEnoughFee,
					"require: %s, limit: %s%s",
					fc.String(),
					limitAmt.String(),
					fc.Denom,
				)
			}
		}

		// Send coin to module account
		if !group.Fee.IsZero() {
			err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, feePayer, types.ModuleName, feeCoins)
			if err != nil {
				return 0, err
			}
		}
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

// AddPendingProcessSigning adds a new pending process signing to the store.
func (k Keeper) AddPendingProcessSigning(ctx sdk.Context, signingID tss.SigningID) {
	pss := k.GetPendingProcessSignings(ctx)
	pss = append(pss, signingID)
	k.SetPendingProcessSignings(ctx, types.PendingProcessSignings{
		SigningIDs: pss,
	})
}

// SetPendingProcessSignings sets the given pending process signings in the store.
func (k Keeper) SetPendingProcessSignings(ctx sdk.Context, pgs types.PendingProcessSignings) {
	ctx.KVStore(k.storeKey).Set(types.PendingSigningsStoreKey, k.cdc.MustMarshal(&pgs))
}

// GetPendingProcessSignings retrieves the list of pending process signings from the store.
// It returns an empty list if the key does not exist in the store.
func (k Keeper) GetPendingProcessSignings(ctx sdk.Context) []tss.SigningID {
	bz := ctx.KVStore(k.storeKey).Get(types.PendingSigningsStoreKey)
	if len(bz) == 0 {
		// Return an empty list if the key does not exist in the store.
		return []tss.SigningID{}
	}
	pss := types.PendingProcessSignings{}
	k.cdc.MustUnmarshal(bz, &pss)
	return pss.SigningIDs
}

// HandleExpiredSignings cleans up expired signings and removes them from the store.
func (k Keeper) HandleExpiredSignings(ctx sdk.Context) {
	// Get the current signing ID to start processing from
	currentSigningID := k.GetLastExpiredSigningID(ctx) + 1

	// Get the last signing ID in the store
	lastSigningID := tss.SigningID(k.GetSigningCount(ctx))

	// Process each signing starting from currentSigningID
	for ; currentSigningID <= lastSigningID; currentSigningID++ {
		// Get the signing
		signing := k.MustGetSigning(ctx, currentSigningID)

		// Check if the signing is still within the expiration period
		if signing.CreatedHeight+k.GetParams(ctx).SigningPeriod > ctx.BlockHeight() {
			break
		}

		// Set the signing status to expired
		if signing.Status != types.SIGNING_STATUS_FALLEN && signing.Status != types.SIGNING_STATUS_SUCCESS {
			k.RefundFee(ctx, signing)

			mids := types.AssignedMembers(signing.AssignedMembers).MemberIDs()
			pzs := k.GetPartialSigsWithKey(ctx, signing.SigningID)
			// Iterate through each member ID in the assigned members list.
			for _, mid := range mids {
				// Check if the member's partial signature is found in the list of partial signatures.
				found := slices.ContainsFunc(pzs, func(pz types.PartialSignature) bool { return pz.MemberID == mid })

				// If the partial signature is not found, deactivate the member
				if !found {
					member := k.MustGetMember(ctx, signing.GroupID, mid)
					accAddress := sdk.MustAccAddressFromBech32(member.Address)
					k.SetInactive(ctx, accAddress)
				}
			}

			signing.Status = types.SIGNING_STATUS_EXPIRED
			k.SetSigning(ctx, signing)
		}

		// Remove assigned members from the signing
		k.DeleteAssignedMembers(ctx, signing.SigningID)

		// Remove all partial signatures from the store
		k.DeletePartialSigs(ctx, signing.SigningID)

		// Set the last expired signing ID to the current signing ID
		k.SetLastExpiredSigningID(ctx, currentSigningID)
	}
}

func (k Keeper) HandleProcessSigning(ctx sdk.Context, signingID tss.SigningID) {
	signing := k.MustGetSigning(ctx, signingID)
	group := k.MustGetGroup(ctx, signing.GroupID)
	pzs := k.GetPartialSigs(ctx, signingID)

	sig, err := tss.CombineSignatures(pzs...)
	if err != nil {
		k.handleFailedSigning(ctx, signing, err.Error())
	}

	err = tss.VerifyGroupSigningSig(group.PubKey, signing.Message, sig)
	if err != nil {
		k.handleFailedSigning(ctx, signing, err.Error())
	}

	// Set signing with signature
	signing.Signature = sig
	// Set signing status
	signing.Status = types.SIGNING_STATUS_SUCCESS
	k.SetSigning(ctx, signing)

	for _, am := range signing.AssignedMembers {
		address := sdk.MustAccAddressFromBech32(am.Member)
		// Error is not possible
		err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, address, signing.Fee)
		if err != nil {
			panic(err)
		}
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSigningSuccess,
			sdk.NewAttribute(types.AttributeKeySigningID, fmt.Sprintf("%d", signingID)),
			sdk.NewAttribute(types.AttributeKeyGroupID, fmt.Sprintf("%d", signing.GroupID)),
			sdk.NewAttribute(types.AttributeKeySignature, hex.EncodeToString(sig)),
		),
	)
}

func (k Keeper) handleFailedSigning(ctx sdk.Context, signing types.Signing, reason string) {
	// Set signing status
	signing.Status = types.SIGNING_STATUS_FALLEN
	k.SetSigning(ctx, signing)

	k.RefundFee(ctx, signing)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSigningFailed,
			sdk.NewAttribute(types.AttributeKeySigningID, fmt.Sprintf("%d", signing.SigningID)),
			sdk.NewAttribute(types.AttributeKeyGroupID, fmt.Sprintf("%d", signing.GroupID)),
			sdk.NewAttribute(types.AttributeKeyReason, reason),
		),
	)
}

func (k Keeper) RefundFee(ctx sdk.Context, signing types.Signing) {
	if !signing.Fee.IsZero() {
		address := sdk.MustAccAddressFromBech32(signing.Requester)
		feeCoins := signing.Fee.MulInt(sdk.NewInt(int64(len(signing.AssignedMembers))))
		// Error is not possible
		err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, address, feeCoins)
		if err != nil {
			panic(err)
		}
	}
}
