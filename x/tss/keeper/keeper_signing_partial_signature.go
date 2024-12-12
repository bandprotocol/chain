package keeper

import (
	dbm "github.com/cosmos/cosmos-db"

	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/pkg/tss"
	"github.com/bandprotocol/chain/v3/x/tss/types"
)

// GetMembersNotSubmitSignature get assigned members that haven't signed a requested message.
func (k Keeper) GetMembersNotSubmitSignature(
	ctx sdk.Context,
	signingID tss.SigningID,
	attempt uint64,
) []sdk.AccAddress {
	signingAttempt := k.MustGetSigningAttempt(ctx, signingID, attempt)

	var memberAddrs []sdk.AccAddress
	for _, am := range signingAttempt.AssignedMembers {
		if !k.HasPartialSignature(ctx, signingID, attempt, am.MemberID) {
			memberAddrs = append(memberAddrs, sdk.MustAccAddressFromBech32(am.Address))
		}
	}

	return memberAddrs
}

// ==================================
// Partial signature Store
// ==================================

// AddPartialSignature adds the partial signature of a specific signing ID from the given member ID
// and increments the count of partial signature.
func (k Keeper) AddPartialSignature(
	ctx sdk.Context,
	signingID tss.SigningID,
	attempt uint64,
	memberID tss.MemberID,
	signature tss.Signature,
) {
	k.AddPartialSignatureCount(ctx, signingID, attempt)
	k.SetPartialSignature(ctx, signingID, attempt, memberID, signature)
}

// SetPartialSignature sets the partial signature of a specific signing ID and member ID.
func (k Keeper) SetPartialSignature(
	ctx sdk.Context,
	signingID tss.SigningID,
	attempt uint64,
	memberID tss.MemberID,
	signature tss.Signature,
) {
	ctx.KVStore(k.storeKey).Set(types.PartialSignatureStoreKey(signingID, attempt, memberID), signature)
}

// HasPartialSignature checks if the partial signature of a specific signing ID and member ID exists in the store.
func (k Keeper) HasPartialSignature(
	ctx sdk.Context,
	signingID tss.SigningID,
	attempt uint64,
	memberID tss.MemberID,
) bool {
	return ctx.KVStore(k.storeKey).Has(types.PartialSignatureStoreKey(signingID, attempt, memberID))
}

// GetPartialSignature retrieves the partial signature of a specific signing ID and member ID from the store.
func (k Keeper) GetPartialSignature(
	ctx sdk.Context,
	signingID tss.SigningID,
	attempt uint64,
	memberID tss.MemberID,
) (tss.Signature, error) {
	bz := ctx.KVStore(k.storeKey).Get(types.PartialSignatureStoreKey(signingID, attempt, memberID))
	if bz == nil {
		return nil, types.ErrPartialSignatureNotFound.Wrapf(
			"failed to get partial signature with signingID: %d memberID: %d",
			signingID,
			memberID,
		)
	}
	return bz, nil
}

// DeletePartialSignatures delete partial signatures data of a given signing and attempt from the store.
func (k Keeper) DeletePartialSignatures(ctx sdk.Context, signingID tss.SigningID, attempt uint64) {
	prefixKey := types.PartialSignaturesStoreKey(signingID, attempt)
	iterator := storetypes.KVStorePrefixIterator(ctx.KVStore(k.storeKey), prefixKey)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		ctx.KVStore(k.storeKey).Delete(iterator.Key())
	}
}

// GetPartialSignatureBySigningAttemptIterator gets an iterator over all partial signature
// of the signing at the specific attempts.
func (k Keeper) GetPartialSignatureBySigningAttemptIterator(
	ctx sdk.Context,
	signingID tss.SigningID,
	attempt uint64,
) dbm.Iterator {
	return storetypes.KVStorePrefixIterator(
		ctx.KVStore(k.storeKey),
		types.PartialSignaturesStoreKey(signingID, attempt),
	)
}

// GetPartialSignatures retrieves all partial signatures for a specific signing ID of
// the specific attempt from the store.
func (k Keeper) GetPartialSignatures(ctx sdk.Context, signingID tss.SigningID, attempt uint64) tss.Signatures {
	iterator := k.GetPartialSignatureBySigningAttemptIterator(ctx, signingID, attempt)
	defer iterator.Close()

	var pzs tss.Signatures
	for ; iterator.Valid(); iterator.Next() {
		pzs = append(pzs, iterator.Value())
	}

	return pzs
}

// GetPartialSignaturesWithKey retrieves all partial signatures for a specific signing ID
// from the store along with their corresponding member IDs.
func (k Keeper) GetPartialSignaturesWithKey(
	ctx sdk.Context,
	signingID tss.SigningID,
	attempt uint64,
) []types.PartialSignature {
	iterator := k.GetPartialSignatureBySigningAttemptIterator(ctx, signingID, attempt)
	defer iterator.Close()

	var partialSigs []types.PartialSignature
	for ; iterator.Valid(); iterator.Next() {
		memberID := types.MemberIDFromPartialSignatureStoreKey(iterator.Key())
		sig := iterator.Value()

		partialSig := types.NewPartialSignature(signingID, attempt, memberID, sig)
		partialSigs = append(partialSigs, partialSig)
	}

	return partialSigs
}

// ==================================
// Partial signature count Store
// ==================================

// SetPartialSignatureCount sets the count of partial signatures of a given signing ID in the store.
func (k Keeper) SetPartialSignatureCount(ctx sdk.Context, signingID tss.SigningID, attempt uint64, count uint64) {
	ctx.KVStore(k.storeKey).Set(types.PartialSignatureCountStoreKey(signingID, attempt), sdk.Uint64ToBigEndian(count))
}

// GetPartialSignatureCount retrieves the count of partial signatures of a given signing ID from the store.
func (k Keeper) GetPartialSignatureCount(ctx sdk.Context, signingID tss.SigningID, attempt uint64) uint64 {
	bz := ctx.KVStore(k.storeKey).Get(types.PartialSignatureCountStoreKey(signingID, attempt))
	return sdk.BigEndianToUint64(bz)
}

// AddPartialSignatureCount increments the count of partial signatures of a given signing ID in the store.
func (k Keeper) AddPartialSignatureCount(ctx sdk.Context, signingID tss.SigningID, attempt uint64) {
	count := k.GetPartialSignatureCount(ctx, signingID, attempt)
	k.SetPartialSignatureCount(ctx, signingID, attempt, count+1)
}

// DeletePartialSignatureCount delete the signature count data of a given signingID and attempt from the store.
func (k Keeper) DeletePartialSignatureCount(ctx sdk.Context, signingID tss.SigningID, attempt uint64) {
	ctx.KVStore(k.storeKey).Delete(types.PartialSignatureCountStoreKey(signingID, attempt))
}
