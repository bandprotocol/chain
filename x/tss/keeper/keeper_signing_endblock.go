package keeper

import (
	"encoding/hex"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/pkg/tss"
	"github.com/bandprotocol/chain/v3/x/tss/types"
)

// HandleSigningEndBlock handles the logic that related to signing process at the end of a block.
func (k Keeper) HandleSigningEndBlock(ctx sdk.Context) {
	// Aggregate pending signings.
	var retrySigningIDs []tss.SigningID
	sids := k.GetPendingProcessSignings(ctx)
	for _, sid := range sids {
		if err := k.AggregatePartialSignatures(ctx, sid); err != nil {
			retrySigningIDs = append(retrySigningIDs, sid)
		}
	}
	k.SetPendingProcessSignings(ctx, types.PendingProcessSignings{})

	// check expired signing
	timeoutSigningIDs := k.HandleExpiredSignings(ctx)

	// retry every failed and expired signings; rollback and handle failed signing
	// if any error occurred.
	retrySigningIDs = append(retrySigningIDs, timeoutSigningIDs...)
	for _, sid := range retrySigningIDs {
		// handle in case of panic
		defer func() {
			if r := recover(); r != nil {
				ctx.Logger().Error(fmt.Sprintf("Panic recovered: %v", r))
				k.HandleFailedSigning(ctx, sid, fmt.Errorf("panic recovered: %v", r).Error())
			}
		}()

		cacheCtx, writeFn := ctx.CacheContext()
		if err := k.InitiateNewSigningRound(cacheCtx, sid); err != nil {
			k.HandleFailedSigning(ctx, sid, err.Error())
		} else {
			writeFn()
		}
	}
}

// HandleExpiredSignings dequeues the first signing expiration from the store and returns
// list of signing IDs that should be retried.
func (k Keeper) HandleExpiredSignings(ctx sdk.Context) []tss.SigningID {
	signingExpirations := k.GetSigningExpirations(ctx)

	idx := 0
	var signingIDs []tss.SigningID
	for _, se := range signingExpirations {
		signingID, attempt := se.SigningID, se.SigningAttempt
		signing := k.MustGetSigning(ctx, signingID)

		sa := k.MustGetSigningAttempt(ctx, signingID, attempt)
		if sa.ExpiredHeight > uint64(ctx.BlockHeight()) {
			break
		}

		idx += 1

		// if there are missing partial signatures, the signing process is incomplete and regarded
		// as timeout; handle timeout hook and appends into a retry list.
		partialSigCount := k.GetPartialSignatureCount(ctx, signingID, attempt)
		if partialSigCount != uint64(len(sa.AssignedMembers)) {
			group := k.MustGetGroup(ctx, signing.GroupID)

			if cb, ok := k.cbRouter.GetRoute(group.ModuleOwner); ok {
				idleMembers := k.GetMembersNotSubmitSignature(ctx, signingID, signing.CurrentAttempt)
				cb.OnSigningTimeout(ctx, signing.ID, idleMembers)
			}

			signingIDs = append(signingIDs, signingID)
		}

		// delete interim signing data.
		k.DeleteInterimSigningData(ctx, signingID, attempt)
	}

	// remove processed signing expirations
	k.SetSigningExpirations(ctx, types.NewSigningExpirations(signingExpirations[idx:]))

	return signingIDs
}

// HandleFailedSigning handles the failed signing process by setting the signing status to fallen
// and emitting an event.
func (k Keeper) HandleFailedSigning(ctx sdk.Context, signingID tss.SigningID, reason string) {
	signing := k.MustGetSigning(ctx, signingID)

	// Set signing status
	signing.Status = types.SIGNING_STATUS_FALLEN
	k.SetSigning(ctx, signing)

	// Handle callback after signing failed; this shouldn't return any error.
	group := k.MustGetGroup(ctx, signing.GroupID)
	if cb, ok := k.cbRouter.GetRoute(group.ModuleOwner); ok {
		cb.OnSigningFailed(ctx, signing.ID)
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

// DeleteInterimSigningData deletes the interim signing data from the store.
func (k Keeper) DeleteInterimSigningData(ctx sdk.Context, signingID tss.SigningID, attempt uint64) {
	k.DeletePartialSignatures(ctx, signingID, attempt)
	k.DeletePartialSignatureCount(ctx, signingID, attempt)
	k.DeleteSigningAttempt(ctx, signingID, attempt)
}

// =====================================
// Process signing aggregation
// =====================================

// AddPendingProcessSigning adds a new pending process signing to the store.
func (k Keeper) AddPendingProcessSigning(ctx sdk.Context, signingID tss.SigningID) {
	pss := k.GetPendingProcessSignings(ctx)
	pss = append(pss, signingID)
	k.SetPendingProcessSignings(ctx, types.NewPendingProcessSignings(pss))
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
	var pss types.PendingProcessSignings
	k.cdc.MustUnmarshal(bz, &pss)
	return pss.SigningIDs
}

// AggregatePartialSignatures aggregates partial signatures and update the signing info if success.
func (k Keeper) AggregatePartialSignatures(ctx sdk.Context, signingID tss.SigningID) error {
	signing := k.MustGetSigning(ctx, signingID)
	partialSigs := k.GetPartialSignatures(ctx, signingID, signing.CurrentAttempt)

	sig, err := tss.CombineSignatures(partialSigs...)
	if err != nil {
		return types.ErrInvalidSignature.Wrapf("failed to combine partial signatures: %v", err)
	}

	if err = tss.VerifyGroupSigningSignature(signing.GroupPubKey, signing.Message, sig); err != nil {
		return types.ErrInvalidSignature.Wrapf("failed to verify group signature: %v", err)
	}

	// Set signing with signature and success status
	signing.Signature = sig
	signing.Status = types.SIGNING_STATUS_SUCCESS
	k.SetSigning(ctx, signing)

	// Handle callback after signing completed; this shouldn't return any error.
	group := k.MustGetGroup(ctx, signing.GroupID)
	if cb, ok := k.cbRouter.GetRoute(group.ModuleOwner); ok {
		assignedMemberAddrs := k.MustGetCurrentAssignedMembers(ctx, signingID)
		cb.OnSigningCompleted(ctx, signing.ID, assignedMemberAddrs)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSigningSuccess,
			sdk.NewAttribute(types.AttributeKeySigningID, fmt.Sprintf("%d", signingID)),
			sdk.NewAttribute(types.AttributeKeyGroupID, fmt.Sprintf("%d", signing.GroupID)),
			sdk.NewAttribute(types.AttributeKeySignature, hex.EncodeToString(sig)),
		),
	)

	return nil
}

// =====================================
// Signing expiration store
// =====================================

// SetSigningExpirations sets the expiration of a signing process.
func (k Keeper) SetSigningExpirations(ctx sdk.Context, signingExpires types.SigningExpirations) {
	ctx.KVStore(k.storeKey).Set(types.SigningExpirationsStoreKey, k.cdc.MustMarshal(&signingExpires))
}

// GetSigningExpirations retrieves the list of signing expiration process.
// It returns an empty list if the key does not exist in the store.
func (k Keeper) GetSigningExpirations(ctx sdk.Context) []types.SigningExpiration {
	bz := ctx.KVStore(k.storeKey).Get(types.SigningExpirationsStoreKey)
	if len(bz) == 0 {
		// Return an empty list if the key does not exist in the store.
		return []types.SigningExpiration{}
	}
	ses := types.SigningExpirations{}
	k.cdc.MustUnmarshal(bz, &ses)
	return ses.SigningExpirations
}

// AddPendingProcessSigning adds a new pending process signing to the store.
func (k Keeper) AddSigningExpiration(ctx sdk.Context, signingID tss.SigningID, attempt uint64) {
	signingExpirations := k.GetSigningExpirations(ctx)

	se := types.NewSigningExpiration(signingID, attempt)
	signingExpirations = append(signingExpirations, se)

	k.SetSigningExpirations(ctx, types.NewSigningExpirations(signingExpirations))
}
