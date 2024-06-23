package keeper

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"slices"
	"sort"

	tmbytes "github.com/cometbft/cometbft/libs/bytes"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/pkg/bandrng"
	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

// ==================================
// Signing Information Store
// ==================================

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

// DeleteSigning deletes the signing data of a given signing ID from the store.
func (k Keeper) DeleteSigning(ctx sdk.Context, signingID tss.SigningID) {
	ctx.KVStore(k.storeKey).Delete(types.SigningStoreKey(signingID))
}

// DeleteAssignedMembers deletes the assigned members of a given signing ID from the store.
func (k Keeper) DeleteAssignedMembers(ctx sdk.Context, signingID tss.SigningID) {
	signing := k.MustGetSigning(ctx, signingID)
	signing.AssignedMembers = nil
	k.SetSigning(ctx, signing)
}

// ==================================
// Partial Signature information Store
// ==================================

// SetPartialSignatureCount sets the count of partial signatures of a given signing ID in the store.
func (k Keeper) SetPartialSignatureCount(ctx sdk.Context, signingID tss.SigningID, count uint64) {
	ctx.KVStore(k.storeKey).Set(types.PartialSignatureCountStoreKey(signingID), sdk.Uint64ToBigEndian(count))
}

// GetPartialSignatureCount retrieves the count of partial signatures of a given signing ID from the store.
func (k Keeper) GetPartialSignatureCount(ctx sdk.Context, signingID tss.SigningID) uint64 {
	bz := ctx.KVStore(k.storeKey).Get(types.PartialSignatureCountStoreKey(signingID))
	return sdk.BigEndianToUint64(bz)
}

// AddPartialSignatureCount increments the count of partial signatures of a given signing ID in the store.
func (k Keeper) AddPartialSignatureCount(ctx sdk.Context, signingID tss.SigningID) {
	count := k.GetPartialSignatureCount(ctx, signingID)
	k.SetPartialSignatureCount(ctx, signingID, count+1)
}

// DeletePartialSignatureCount delete the signature count data of a sign from the store.
func (k Keeper) DeletePartialSignatureCount(ctx sdk.Context, signingID tss.SigningID) {
	ctx.KVStore(k.storeKey).Delete(types.PartialSignatureCountStoreKey(signingID))
}

// AddPartialSignature adds the partial signature of a specific signing ID from the given member ID
// and increments the count of partial signature.
func (k Keeper) AddPartialSignature(
	ctx sdk.Context,
	signingID tss.SigningID,
	memberID tss.MemberID,
	signature tss.Signature,
) {
	k.AddPartialSignatureCount(ctx, signingID)
	k.SetPartialSignature(ctx, signingID, memberID, signature)
}

// SetPartialSignature sets the partial signature of a specific signing ID and member ID.
func (k Keeper) SetPartialSignature(
	ctx sdk.Context,
	signingID tss.SigningID,
	memberID tss.MemberID,
	signature tss.Signature,
) {
	ctx.KVStore(k.storeKey).Set(types.PartialSignatureMemberStoreKey(signingID, memberID), signature)
}

// HasPartialSignature checks if the partial signature of a specific signing ID and member ID exists in the store.
func (k Keeper) HasPartialSignature(ctx sdk.Context, signingID tss.SigningID, memberID tss.MemberID) bool {
	return ctx.KVStore(k.storeKey).Has(types.PartialSignatureMemberStoreKey(signingID, memberID))
}

// GetPartialSignature retrieves the partial signature of a specific signing ID and member ID from the store.
func (k Keeper) GetPartialSignature(
	ctx sdk.Context,
	signingID tss.SigningID,
	memberID tss.MemberID,
) (tss.Signature, error) {
	bz := ctx.KVStore(k.storeKey).Get(types.PartialSignatureMemberStoreKey(signingID, memberID))
	if bz == nil {
		return nil, types.ErrPartialSignatureNotFound.Wrapf(
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

// GetPartialSignaturesWithKey retrieves all partial signatures for a specific signing ID
// from the store along with their corresponding member IDs.
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

// ==================================
// Create Signing
// ==================================

// GetRandomAssignedMembers select a random assigned members for a signing process.
// It selects 't' assigned members out of the available members in the given group using
// a deterministic random number generator (DRBG).
func (k Keeper) GetRandomAssignedMembers(
	ctx sdk.Context,
	groupID tss.GroupID,
	signingID uint64,
	t uint64,
) ([]types.Member, error) {
	// Get available members
	members, err := k.GetAvailableMembers(ctx, groupID)
	if err != nil {
		return nil, err
	}

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
	memberIdx := make([]int, members_size)
	for i := 0; i < int(members_size); i++ {
		memberIdx[i] = i
	}

	for i := uint64(0); i < t; i++ {
		randomNumber := rng.NextUint64() % (members_size - i)

		// Swap the selected member with the last member in the list
		memberId := memberIdx[randomNumber]
		memberIdx[randomNumber] = memberIdx[members_size-i-1]

		// Append the selected member to the selected list
		selected = append(selected, members[memberId])
	}

	// Sort selected members
	sort.Slice(selected, func(i, j int) bool { return selected[i].ID < selected[j].ID })

	return selected, nil
}

// AssignMembers handles the assignment of members for a group signature process.
func (k Keeper) AssignMembers(
	ctx sdk.Context,
	group types.Group,
	msg []byte,
) (types.AssignedMembers, error) {
	// Check group status
	if group.Status != types.GROUP_STATUS_ACTIVE {
		return types.AssignedMembers{}, types.ErrGroupIsNotActive.Wrap("group status is not active")
	}

	// Random assigning members
	selectedMembers, err := k.GetRandomAssignedMembers(
		ctx,
		group.ID,
		k.GetSigningCount(ctx)+1,
		group.Threshold,
	)
	if err != nil {
		return types.AssignedMembers{}, err
	}

	// Handle assigned members by polling DE and retrieve assigned members information.
	des, err := k.PollDEs(ctx, selectedMembers)
	if err != nil {
		return types.AssignedMembers{}, err
	}

	var assignedMembers types.AssignedMembers
	for i, member := range selectedMembers {
		assignedMembers = append(assignedMembers, types.AssignedMember{
			MemberID:      member.ID,
			Address:       member.Address,
			PubKey:        member.PubKey,
			PubD:          des[i].PubD,
			PubE:          des[i].PubE,
			BindingFactor: nil,
			PubNonce:      nil,
		})
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

// ConvertContentToBytes convert content to message bytes by the registered router.
func (k Keeper) ConvertContentToBytes(
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

// CreateSigning creates a new signing process and returns the result.
func (k Keeper) CreateSigning(
	ctx sdk.Context,
	group types.Group,
	content types.Content,
) (*types.Signing, error) {
	message, err := k.ConvertContentToBytes(ctx, content)
	if err != nil {
		return nil, err
	}

	// assigned members within the context of the group.
	assignedMembers, err := k.AssignMembers(ctx, group, message)
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
		message,
		groupPubNonce,
		nil,
		types.SIGNING_STATUS_WAITING,
	))

	signing, err := k.GetSigning(ctx, signingID)
	if err != nil {
		return nil, err
	}

	// emit an event.
	event := sdk.NewEvent(
		types.EventTypeRequestSignature,
		sdk.NewAttribute(types.AttributeKeyGroupID, fmt.Sprintf("%d", signing.GroupID)),
		sdk.NewAttribute(types.AttributeKeySigningID, fmt.Sprintf("%d", signing.ID)),
		sdk.NewAttribute(types.AttributeKeyMessage, hex.EncodeToString(message)),
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

	return &signing, nil
}

// =====================================
// Query signing-related information
// =====================================

// GetSigningResult returns the signing result of the given tss signingID.
func (k Keeper) GetSigningResult(ctx sdk.Context, signingID tss.SigningID) (*types.SigningResult, error) {
	tssSigning, err := k.GetSigning(ctx, signingID)
	if err != nil {
		return nil, err
	}

	partialSigs := k.GetPartialSignaturesWithKey(ctx, signingID)

	var evmSignature *types.EVMSignature
	if tssSigning.Signature != nil {
		rAddress, err := tssSigning.Signature.R().Address()
		if err != nil {
			return nil, err
		}

		evmSignature = &types.EVMSignature{
			RAddress:  rAddress,
			Signature: tmbytes.HexBytes(tssSigning.Signature.S()),
		}
	}

	return &types.SigningResult{
		Signing:                   tssSigning,
		EVMSignature:              evmSignature,
		ReceivedPartialSignatures: partialSigs,
	}, nil
}

// GetPenalizedMembersExpiredSigning get assigned members that haven't signed a request.
func (k Keeper) GetPenalizedMembersExpiredSigning(ctx sdk.Context, signing types.Signing) ([]sdk.AccAddress, error) {
	partialSigs := k.GetPartialSignaturesWithKey(ctx, signing.ID)
	var penalizedMembers []sdk.AccAddress

	mids := signing.AssignedMembers.MemberIDs()
	for _, mid := range mids {
		// Check if the member sends partial signature. If found, skip this member.
		found := slices.ContainsFunc(partialSigs, func(pz types.PartialSignature) bool { return pz.MemberID == mid })
		if found {
			continue
		}

		member := k.MustGetMember(ctx, signing.GroupID, mid)
		penalizedMembers = append(penalizedMembers, sdk.MustAccAddressFromBech32(member.Address))
	}

	return penalizedMembers, nil
}

// GetPendingSignings retrieves the pending signing objects associated with the given account address.
func (k Keeper) GetPendingSignings(ctx sdk.Context, address sdk.AccAddress) []uint64 {
	filterFunc := func(am types.AssignedMember) bool {
		return am.Address == address.String()
	}

	return k.getPendingSigningByFilterFunc(ctx, filterFunc)
}

// GetPendingSigningsByPubKey retrieves the pending signing objects associated with the given tss public key.
func (k Keeper) GetPendingSigningsByPubKey(ctx sdk.Context, pubKey tss.Point) []uint64 {
	filterFunc := func(am types.AssignedMember) bool {
		return bytes.Equal(am.PubKey, pubKey)
	}

	return k.getPendingSigningByFilterFunc(ctx, filterFunc)
}

// getPendingSigningByFilterFunc retrieves the pending signing objects associated with the given filter function.
func (k Keeper) getPendingSigningByFilterFunc(
	ctx sdk.Context,
	filterFunc func(m types.AssignedMember) bool,
) []uint64 {
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
			if filterFunc(am) {
				// Add the signing to the pendingSignings if there is no partial sig of the member yet.
				if _, err := k.GetPartialSignature(ctx, sid, am.MemberID); err != nil {
					pendingSignings = append(pendingSignings, uint64(signing.ID))
				}
			}
		}
	}

	return pendingSignings
}
