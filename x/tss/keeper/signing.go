package keeper

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"sort"

	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
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
	signingNumber := k.GetSigningCount(ctx) + 1
	k.SetSigningCount(ctx, signingNumber)
	return tss.SigningID(signingNumber)
}

// SetSigning sets the signing data for a given signing ID.
func (k Keeper) SetSigning(ctx sdk.Context, signing types.Signing) {
	ctx.KVStore(k.storeKey).Set(types.SigningStoreKey(signing.ID), k.cdc.MustMarshal(&signing))
}

// GetSigning retrieves the signing data for a given signing ID from the store.
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

// MustGetSigning returns the signing for the given ID. Panics error if not exists.
func (k Keeper) MustGetSigning(ctx sdk.Context, signingID tss.SigningID) types.Signing {
	signing, err := k.GetSigning(ctx, signingID)
	if err != nil {
		panic(err)
	}
	return signing
}

// GetSigningsIterator gets an iterator all group.
func (k Keeper) GetSigningsIterator(ctx sdk.Context) sdk.Iterator {
	return sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.SigningStoreKeyPrefix)
}

// GetSignings retrieves all signing of the store.
func (k Keeper) GetSignings(ctx sdk.Context) []types.Signing {
	var signings []types.Signing
	iterator := k.GetSigningsIterator(ctx)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var signing types.Signing
		k.cdc.MustUnmarshal(iterator.Value(), &signing)
		signings = append(signings, signing)
	}
	return signings
}

// AddSigning adds the signing data to the store and returns the new signing ID.
func (k Keeper) AddSigning(ctx sdk.Context, signing types.Signing) tss.SigningID {
	signing.ID = k.GetNextSigningID(ctx)
	signing.CreatedHeight = uint64(ctx.BlockHeight())
	k.SetSigning(ctx, signing)

	return signing.ID
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
			if am.Address == address.String() {
				// Add the signing to the pendingSignings if there is no partial sig of the member yet.
				if _, err := k.GetPartialSignature(ctx, sid, am.MemberID); err != nil {
					pendingSignings = append(pendingSignings, uint64(signing.ID))
				}
			}
		}
	}

	return pendingSignings
}

// GetPendingSigningsByPubKey retrieves the pending signing objects associated with the given tss public key.
func (k Keeper) GetPendingSigningsByPubKey(ctx sdk.Context, pubKey tss.Point) []uint64 {
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
			if bytes.Equal(am.PubKey, pubKey) {
				// Add the signing to the pendingSignings if there is no partial sig of the member yet.
				if _, err := k.GetPartialSignature(ctx, sid, am.MemberID); err != nil {
					pendingSignings = append(pendingSignings, uint64(signing.ID))
				}
			}
		}
	}

	return pendingSignings
}

// SetSignatureCount sets the count of signature data for a sign in the store.
func (k Keeper) SetSignatureCount(ctx sdk.Context, signingID tss.SigningID, count uint64) {
	ctx.KVStore(k.storeKey).Set(types.SigCountStoreKey(signingID), sdk.Uint64ToBigEndian(count))
}

// GetSignatureCount retrieves the count of signature data for a sign from the store.
func (k Keeper) GetSignatureCount(ctx sdk.Context, signingID tss.SigningID) uint64 {
	bz := ctx.KVStore(k.storeKey).Get(types.SigCountStoreKey(signingID))
	return sdk.BigEndianToUint64(bz)
}

// AddSignatureCount increments the count of signature data for a sign in the store.
func (k Keeper) AddSignatureCount(ctx sdk.Context, signingID tss.SigningID) {
	count := k.GetSignatureCount(ctx, signingID)
	k.SetSignatureCount(ctx, signingID, count+1)
}

// DeleteSignatureCount delete the signature count data of a sign from the store.
func (k Keeper) DeleteSignatureCount(ctx sdk.Context, signingID tss.SigningID) {
	ctx.KVStore(k.storeKey).Delete(types.SigCountStoreKey(signingID))
}

// AddPartialSignature adds the partial signature for a specific signing ID and member ID and increments the count of signature data.
func (k Keeper) AddPartialSignature(
	ctx sdk.Context,
	signingID tss.SigningID,
	memberID tss.MemberID,
	signature tss.Signature,
) {
	k.AddSignatureCount(ctx, signingID)
	k.SetPartialSignature(ctx, signingID, memberID, signature)
}

// SetPartialSignature sets the partial signature for a specific signing ID and member ID.
func (k Keeper) SetPartialSignature(
	ctx sdk.Context,
	signingID tss.SigningID,
	memberID tss.MemberID,
	signature tss.Signature,
) {
	ctx.KVStore(k.storeKey).Set(types.PartialSignatureMemberStoreKey(signingID, memberID), signature)
}

// GetPartialSignature retrieves the partial signature for a specific signing ID and member ID from the store.
func (k Keeper) GetPartialSignature(
	ctx sdk.Context,
	signingID tss.SigningID,
	memberID tss.MemberID,
) (tss.Signature, error) {
	bz := ctx.KVStore(k.storeKey).Get(types.PartialSignatureMemberStoreKey(signingID, memberID))
	if bz == nil {
		return nil, errors.Wrapf(
			types.ErrPartialSignatureNotFound,
			"failed to get partial signature with signingID: %d memberID: %d",
			signingID,
			memberID,
		)
	}
	return bz, nil
}

// DeletePartialSignatures delete all partial signatures data of a signing from the store.
func (k Keeper) DeletePartialSignatures(ctx sdk.Context, signingID tss.SigningID) {
	iterator := k.GetPartialSignatureIterator(ctx, signingID)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		ctx.KVStore(k.storeKey).Delete(iterator.Key())
	}
}

// DeletePartialSignature delete a partial signature of a signing from the store.
func (k Keeper) DeletePartialSignature(ctx sdk.Context, signingID tss.SigningID, memberID tss.MemberID) {
	ctx.KVStore(k.storeKey).Delete(types.PartialSignatureMemberStoreKey(signingID, memberID))
}

// GetPartialSignatureIterator gets an iterator over all partial signature of the signing.
func (k Keeper) GetPartialSignatureIterator(ctx sdk.Context, signingID tss.SigningID) sdk.Iterator {
	return sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.PartialSignatureStoreKey(signingID))
}

// GetPartialSignatures retrieves all partial signatures for a specific signing ID from the store.
func (k Keeper) GetPartialSignatures(ctx sdk.Context, signingID tss.SigningID) tss.Signatures {
	var pzs tss.Signatures
	iterator := k.GetPartialSignatureIterator(ctx, signingID)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		pzs = append(pzs, iterator.Value())
	}
	return pzs
}

// GetPartialSignaturesWithKey retrieves all partial signatures for a specific signing ID from the store along with their corresponding member IDs.
func (k Keeper) GetPartialSignaturesWithKey(ctx sdk.Context, signingID tss.SigningID) []types.PartialSignature {
	var pzs []types.PartialSignature
	iterator := k.GetPartialSignatureIterator(ctx, signingID)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		pzs = append(pzs, types.PartialSignature{
			MemberID:  types.MemberIDFromPartialSignatureMemberStoreKey(iterator.Key()),
			Signature: iterator.Value(),
		})
	}
	return pzs
}

// GetRandomAssignedMembers generates a random selection of assigned members for a signing process.
// It selects 't' assigned members out of 'members size' assigned members using a deterministic random number generator (DRBG).
func (k Keeper) GetRandomAssignedMembers(
	ctx sdk.Context,
	signingID uint64,
	members []types.Member,
	t uint64,
) ([]types.Member, error) {
	members_size := uint64(len(members))
	if t > members_size {
		return nil, types.ErrUnexpectedThreshold.Wrapf("t must less than or equal to size")
	}

	// Create a deterministic random number generator (DRBG) using the rolling seed, signingID, and chain ID.
	rng, err := bandrng.NewRng(
		k.rollingseedKeeper.GetRollingSeed(ctx),
		sdk.Uint64ToBigEndian(signingID),
		[]byte(ctx.ChainID()),
	)
	if err != nil {
		return nil, types.ErrBadDrbgInitialization.Wrapf(err.Error())
	}

	var selected []types.Member
	for i := uint64(0); i < t; i++ {
		randomNumber := rng.NextUint64() % members_size

		// Get the selected member.
		selected = append(selected, members[randomNumber])

		// Remove the selected member from the list.
		members = append(members[:randomNumber], members[randomNumber+1:]...)

		members_size -= 1
	}

	// Sort selected members
	sort.Slice(selected, func(i, j int) bool { return selected[i].ID < selected[j].ID })

	return selected, nil
}

// HandleAssignedMembers handles the assignment of members for a group signature process.
func (k Keeper) HandleAssignedMembers(
	ctx sdk.Context,
	group types.Group,
	msg []byte,
) (types.AssignedMembers, error) {
	// Check group status
	if group.Status != types.GROUP_STATUS_ACTIVE {
		return types.AssignedMembers{}, errors.Wrap(
			types.ErrGroupIsNotActive,
			"group status is not active",
		)
	}

	// Get active members
	members, err := k.GetActiveMembers(ctx, group.ID)
	if err != nil {
		return types.AssignedMembers{}, err
	}

	// Random assigning members
	selectedMembers, err := k.GetRandomAssignedMembers(
		ctx,
		k.GetSigningCount(ctx)+1,
		members,
		group.Threshold,
	)
	if err != nil {
		return types.AssignedMembers{}, err
	}

	// Handle assigned members by polling DE and retrieve assigned members information.
	assignedMembers, err := k.HandleAssignedMembersPollDE(ctx, selectedMembers)
	if err != nil {
		return types.AssignedMembers{}, err
	}

	// Compute commitment from mids, public D and public E
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

func (k Keeper) HandleSigningContent(
	ctx sdk.Context,
	content types.Content,
) ([]byte, error) {
	if !k.router.HasRoute(content.OrderRoute()) {
		return nil, types.ErrNoSignatureOrderHandlerExists.Wrap(content.OrderRoute())
	}

	// Retrieve the appropriate handler for the request signature route.
	handler := k.router.GetRoute(content.OrderRoute())

	// Execute the handler to process the request.
	msg, err := handler(ctx, content)
	if err != nil {
		return nil, err
	}

	return msg, nil
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
		if signing.CreatedHeight+k.GetParams(ctx).SigningPeriod > uint64(ctx.BlockHeight()) {
			break
		}

		// Set the signing status to expired
		if signing.Status != types.SIGNING_STATUS_FALLEN && signing.Status != types.SIGNING_STATUS_SUCCESS {
			// Handle hooks before setting signing to be expired
			k.Hooks().AfterSigningFailed(ctx, signing)

			signing.Status = types.SIGNING_STATUS_EXPIRED
			k.SetSigning(ctx, signing)

			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeExpiredSigning,
					sdk.NewAttribute(types.AttributeKeySigningID, fmt.Sprintf("%d", signing.ID)),
				),
			)
		}

		// Remove assigned members from the signing
		k.DeleteAssignedMembers(ctx, signing.ID)

		// Remove all partial signatures from the store
		k.DeletePartialSignatures(ctx, signing.ID)

		// Set the last expired signing ID to the current signing ID
		k.SetLastExpiredSigningID(ctx, currentSigningID)
	}
}

func (k Keeper) GetPenalizedMembersExpiredSigning(ctx sdk.Context, signing types.Signing) ([]sdk.AccAddress, error) {
	pzs := k.GetPartialSignaturesWithKey(ctx, signing.ID)
	var penalizedMembers []sdk.AccAddress

	mids := signing.AssignedMembers.MemberIDs()
	for _, mid := range mids {
		// Check if the member sends partial signature. If found, skip this member.
		found := slices.ContainsFunc(pzs, func(pz types.PartialSignature) bool { return pz.MemberID == mid })
		if found {
			continue
		}

		member := k.MustGetMember(ctx, signing.GroupID, mid)
		accAddress, err := sdk.AccAddressFromBech32(member.Address)
		if err != nil {
			return nil, err
		}
		penalizedMembers = append(penalizedMembers, accAddress)
	}

	return penalizedMembers, nil
}

func (k Keeper) HandleProcessSigning(ctx sdk.Context, signingID tss.SigningID) {
	signing := k.MustGetSigning(ctx, signingID)
	pzs := k.GetPartialSignatures(ctx, signingID)

	sig, err := tss.CombineSignatures(pzs...)
	if err != nil {
		k.handleFailedSigning(ctx, signing, err.Error())
	}

	err = tss.VerifyGroupSigningSignature(signing.GroupPubKey, signing.Message, sig)
	if err != nil {
		k.handleFailedSigning(ctx, signing, err.Error())
	}

	// Set signing with signature
	signing.Signature = sig
	// Set signing status
	signing.Status = types.SIGNING_STATUS_SUCCESS
	k.SetSigning(ctx, signing)

	// Handle hooks after signing completed.
	k.Hooks().AfterSigningCompleted(ctx, signing)

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

	// Handle hooks after signing failed
	k.Hooks().AfterSigningFailed(ctx, signing)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSigningFailed,
			sdk.NewAttribute(types.AttributeKeySigningID, fmt.Sprintf("%d", signing.ID)),
			sdk.NewAttribute(types.AttributeKeyGroupID, fmt.Sprintf("%d", signing.GroupID)),
			sdk.NewAttribute(types.AttributeKeyReason, reason),
		),
	)
}

// CreateSigning creates a new signing process and returns the result.
func (k Keeper) CreateSigning(ctx sdk.Context, input types.CreateSigningInput) (*types.CreateSigningResult, error) {
	group, err := k.GetActiveGroup(ctx, input.GroupID)
	if err != nil {
		return nil, err
	}

	// charged fee if necessary
	fee := sdk.NewCoins()
	if input.IsFeeCharged {
		fee = group.Fee

		// If found any coins that exceed limit then return error
		feeCoins := group.Fee.MulInt(sdk.NewInt(int64(group.Threshold)))
		for _, fc := range feeCoins {
			limitAmt := input.FeeLimit.AmountOf(fc.Denom)
			if fc.Amount.GT(limitAmt) {
				return nil, types.ErrNotEnoughFee.Wrapf(
					"require: %s, limit: %s%s",
					fc.String(),
					limitAmt.String(),
					fc.Denom,
				)
			}
		}
	}

	// Handle assigned members within the context of the group.
	assignedMembers, err := k.HandleAssignedMembers(ctx, group, input.Message)
	if err != nil {
		return nil, err
	}

	// Compute group public nonce for this signing
	groupPubNonce, err := tss.ComputeGroupPublicNonce(assignedMembers.PubNonces()...)
	if err != nil {
		return nil, err
	}

	// Add signing
	signingID := k.AddSigning(ctx, types.NewSigning(
		group.ID,
		group.PubKey,
		assignedMembers,
		input.Message,
		groupPubNonce,
		nil,
		fee,
		types.SIGNING_STATUS_WAITING,
		input.FeePayer.String(),
	))

	signing, err := k.GetSigning(ctx, signingID)
	if err != nil {
		return nil, err
	}

	// Handle hooks after signing initiated.
	if err := k.Hooks().AfterSigningCreated(ctx, signing); err != nil {
		return nil, err
	}

	k.emitCreateSigningEvent(ctx, input.Message, signing)

	return &types.CreateSigningResult{
		Signing: signing,
	}, nil
}

func (k Keeper) emitCreateSigningEvent(ctx sdk.Context, msg []byte, signing types.Signing) {
	event := sdk.NewEvent(
		types.EventTypeRequestSignature,
		sdk.NewAttribute(types.AttributeKeyGroupID, fmt.Sprintf("%d", signing.GroupID)),
		sdk.NewAttribute(types.AttributeKeySigningID, fmt.Sprintf("%d", signing.ID)),
		sdk.NewAttribute(types.AttributeKeyMessage, hex.EncodeToString(msg)),
		sdk.NewAttribute(types.AttributeKeyGroupPubNonce, hex.EncodeToString(signing.GroupPubNonce)),
	)
	for _, am := range signing.AssignedMembers {
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
}
