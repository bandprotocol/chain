package keeper

import (
	"encoding/hex"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

// =====================================
// Process fully-signed signing
// =====================================

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

// HandleProcessSigning combine and verify group signature. It will be triggered at endblock.
func (k Keeper) HandleProcessSigning(ctx sdk.Context, signingID tss.SigningID) {
	signing := k.MustGetSigning(ctx, signingID)
	partialSigs := k.GetPartialSignatures(ctx, signingID)

	sig, err := tss.CombineSignatures(partialSigs...)
	if err != nil {
		k.handleFailedSigning(ctx, signing, err.Error())
		return
	}

	if err = tss.VerifyGroupSigningSignature(signing.GroupPubKey, signing.Message, sig); err != nil {
		k.handleFailedSigning(ctx, signing, err.Error())
		return
	}

	// Set signing with signature and success status
	signing.Signature = sig
	signing.Status = types.SIGNING_STATUS_SUCCESS
	k.SetSigning(ctx, signing)

	// Handle hooks after signing completed; this shouldn't return any error.
	if err := k.Hooks().AfterSigningCompleted(ctx, signing); err != nil {
		panic(err)
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

// handleFailedSigning handles the failed signing process by setting the signing status to fallen
// and emitting an event.
func (k Keeper) handleFailedSigning(ctx sdk.Context, signing types.Signing, reason string) {
	// Set signing status
	signing.Status = types.SIGNING_STATUS_FALLEN
	k.SetSigning(ctx, signing)

	// Handle hooks after signing failed; this shouldn't return any error.
	if err := k.Hooks().AfterSigningFailed(ctx, signing); err != nil {
		panic(err)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSigningFailed,
			sdk.NewAttribute(types.AttributeKeySigningID, fmt.Sprintf("%d", signing.ID)),
			sdk.NewAttribute(types.AttributeKeyGroupID, fmt.Sprintf("%d", signing.GroupID)),
			sdk.NewAttribute(types.AttributeKeyReason, reason),
		),
	)
}

// =====================================
// Process expired signature
// =====================================

// SetLastExpiredSigningID sets the last expired signing ID in the store.
func (k Keeper) SetLastExpiredSigningID(ctx sdk.Context, signingID tss.SigningID) {
	ctx.KVStore(k.storeKey).Set(types.LastExpiredSigningIDStoreKey, sdk.Uint64ToBigEndian(uint64(signingID)))
}

// GetLastExpiredSigningID retrieves the last expired signing ID from the store.
func (k Keeper) GetLastExpiredSigningID(ctx sdk.Context) tss.SigningID {
	bz := ctx.KVStore(k.storeKey).Get(types.LastExpiredSigningIDStoreKey)
	return tss.SigningID(sdk.BigEndianToUint64(bz))
}

// HandleExpiredSignings cleans up expired signings and removes assigned members and their partial
// signature information from the store. It will be triggered at endblock.
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
			// Handle hooks before setting signing to be expired; this shouldn't return any error.
			if err := k.Hooks().BeforeSetSigningExpired(ctx, signing); err != nil {
				panic(err)
			}

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
		k.DeletePartialSignatureCount(ctx, signing.ID)

		// Set the last expired signing ID to the current signing ID
		k.SetLastExpiredSigningID(ctx, currentSigningID)
	}
}
