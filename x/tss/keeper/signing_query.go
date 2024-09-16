package keeper

import (
	"bytes"

	tmbytes "github.com/cometbft/cometbft/libs/bytes"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

// =====================================
// Query signing-related information
// =====================================

// GetSigningResult returns the signing result of the given tss signingID.
func (k Keeper) GetSigningResult(ctx sdk.Context, signingID tss.SigningID) (*types.SigningResult, error) {
	signing, err := k.GetSigning(ctx, signingID)
	if err != nil {
		return nil, err
	}

	var currentSigningAttempt *types.SigningAttempt
	sa, err := k.GetSigningAttempt(ctx, signingID, signing.CurrentAttempt)
	if err == nil {
		currentSigningAttempt = &sa
	}

	partialSigs := k.GetPartialSignaturesWithKey(ctx, signingID, signing.CurrentAttempt)

	var evmSignature *types.EVMSignature
	if signing.Signature != nil {
		rAddress, err := signing.Signature.R().Address()
		if err != nil {
			return nil, err
		}

		evmSignature = &types.EVMSignature{
			RAddress:  rAddress,
			Signature: tmbytes.HexBytes(signing.Signature.S()),
		}
	}

	return &types.SigningResult{
		Signing:                   signing,
		CurrentSigningAttempt:     currentSigningAttempt,
		EVMSignature:              evmSignature,
		ReceivedPartialSignatures: partialSigs,
	}, nil
}

// =====================================
// Pending Signings
// =====================================

// GetPendingSignings retrieves the pending signing objects associated with the given account address.
func (k Keeper) GetPendingSignings(ctx sdk.Context, address sdk.AccAddress) []tss.SigningID {
	filterFunc := func(am types.AssignedMember) bool {
		return am.Address == address.String()
	}

	return k.getPendingSigningByFilterFunc(ctx, filterFunc)
}

// GetPendingSigningsByPubKey retrieves the pending signing objects associated with the given tss public key.
func (k Keeper) GetPendingSigningsByPubKey(ctx sdk.Context, pubKey tss.Point) []tss.SigningID {
	filterFunc := func(am types.AssignedMember) bool {
		return bytes.Equal(am.PubKey, pubKey)
	}

	return k.getPendingSigningByFilterFunc(ctx, filterFunc)
}

// getPendingSigningByFilterFunc retrieves the pending signing objects associated with the given filter function.
func (k Keeper) getPendingSigningByFilterFunc(
	ctx sdk.Context,
	filterFunc func(m types.AssignedMember) bool,
) []tss.SigningID {
	signingExpirations := k.GetSigningExpirations(ctx)

	checked := make(map[tss.SigningID]struct{})
	var signingIDs []tss.SigningID

	for _, se := range signingExpirations {
		signingID := se.SigningID

		// if the signingID is already checked, skip it
		if _, ok := checked[signingID]; ok {
			continue
		}
		checked[signingID] = struct{}{}

		// Check if the signing is still pending
		signing := k.MustGetSigning(ctx, signingID)
		if signing.Status != types.SIGNING_STATUS_WAITING {
			continue
		}

		// Check if address is assigned for signing
		// Add the signing to the pendingSignings if there is no partial sig of the member yet.
		// Shouldn't get any error from GetSigningAttempt since the signing is still pending.
		attempt := signing.CurrentAttempt
		signingAttempt, err := k.GetSigningAttempt(ctx, signingID, attempt)
		if err != nil {
			continue
		}

		for _, am := range signingAttempt.AssignedMembers {
			if filterFunc(am) && !k.HasPartialSignature(ctx, signingID, attempt, am.MemberID) {
				return append(signingIDs, signingID)
			}
		}
	}

	return signingIDs
}
