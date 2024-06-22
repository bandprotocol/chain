package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/kv"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/bandtss/types"
	tsstypes "github.com/bandprotocol/chain/v2/x/tss/types"
)

// SetSigningCount sets the number of bandtss signing count to the given value.
func (k Keeper) SetSigningCount(ctx sdk.Context, count uint64) {
	ctx.KVStore(k.storeKey).Set(types.SigningCountStoreKey, sdk.Uint64ToBigEndian(count))
}

// GetSigningCount returns the current number of all bandtss signing ever existed.
func (k Keeper) GetSigningCount(ctx sdk.Context) uint64 {
	return sdk.BigEndianToUint64(ctx.KVStore(k.storeKey).Get(types.SigningCountStoreKey))
}

// GetNextSigningID increments the bandtss signing count and returns the current number of bandtss signing.
func (k Keeper) GetNextSigningID(ctx sdk.Context) types.SigningID {
	signingNumber := k.GetSigningCount(ctx) + 1
	k.SetSigningCount(ctx, signingNumber)
	return types.SigningID(signingNumber)
}

// SetSigning sets a signing of the bandtss module.
func (k Keeper) SetSigning(ctx sdk.Context, signing types.Signing) {
	ctx.KVStore(k.storeKey).Set(types.SigningInfoStoreKey(signing.ID), k.cdc.MustMarshal(&signing))
}

// AddSigning adds the signing data to the store and returns the new Signing ID.
func (k Keeper) AddSigning(ctx sdk.Context, signing types.Signing) types.SigningID {
	signing.ID = k.GetNextSigningID(ctx)
	k.SetSigning(ctx, signing)

	if signing.CurrentGroupSigningID != 0 {
		k.SetSigningIDMapping(ctx, signing.CurrentGroupSigningID, signing.ID)
	}
	if signing.ReplacingGroupSigningID != 0 {
		k.SetSigningIDMapping(ctx, signing.ReplacingGroupSigningID, signing.ID)
	}

	return signing.ID
}

// GetSigning retrieves a bandtss signing info.
func (k Keeper) GetSigning(ctx sdk.Context, id types.SigningID) (types.Signing, error) {
	bz := ctx.KVStore(k.storeKey).Get(types.SigningInfoStoreKey(id))
	if bz == nil {
		return types.Signing{}, types.ErrSigningNotFound.Wrapf("signingID: %d", id)
	}

	signing := types.Signing{}
	k.cdc.MustUnmarshal(bz, &signing)
	return signing, nil
}

// MustGetSigning retrieves a bandtss signing. Panics error if not exists.
func (k Keeper) MustGetSigning(ctx sdk.Context, id types.SigningID) types.Signing {
	req, err := k.GetSigning(ctx, id)
	if err != nil {
		panic(err)
	}
	return req
}

// GetSigningIterator gets an iterator all bandtss signing.
func (k Keeper) GetSigningIterator(ctx sdk.Context) sdk.Iterator {
	return sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.SigningInfoStoreKeyPrefix)
}

// GetSignings retrieves all signing of the store.
func (k Keeper) GetSignings(ctx sdk.Context) []types.Signing {
	var reqs []types.Signing
	iterator := k.GetSigningIterator(ctx)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var req types.Signing
		k.cdc.MustUnmarshal(iterator.Value(), &req)
		reqs = append(reqs, req)
	}
	return reqs
}

// SetSigningIDMapping sets a mapping between tss.signingID and bandtss signing id.
func (k Keeper) SetSigningIDMapping(ctx sdk.Context, signingID tss.SigningID, signingInfoID types.SigningID) {
	ctx.KVStore(k.storeKey).Set(
		types.SigningIDMappingStoreKey(signingID),
		sdk.Uint64ToBigEndian(uint64(signingInfoID)),
	)
}

// GetSigningIDMapping gets a signing id from the given tss signingID
func (k Keeper) GetSigningIDMapping(ctx sdk.Context, signingID tss.SigningID) types.SigningID {
	bz := ctx.KVStore(k.storeKey).Get(types.SigningIDMappingStoreKey(signingID))
	return types.SigningID(sdk.BigEndianToUint64(bz))
}

// GetSigningIDMappingIterator gets an iterator all signingIDMapping.
func (k Keeper) GetSigningRequestIDMappingIterator(ctx sdk.Context) sdk.Iterator {
	return sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.SigningIDMappingStoreKeyPrefix)
}

// GetSigningIDMappings retrieves all signingID mapping items of the store.
func (k Keeper) GetSigningIDMappings(ctx sdk.Context) []types.SigningIDMappingGenesis {
	var mappings []types.SigningIDMappingGenesis
	iterator := k.GetSigningRequestIDMappingIterator(ctx)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		mappings = append(mappings, types.SigningIDMappingGenesis{
			SigningID:        decodeSigningMappingKeyToSigningID(iterator.Key()),
			BandtssSigningID: types.SigningID(sdk.BigEndianToUint64(iterator.Value())),
		})
	}
	return mappings
}

// DeleteSigningIDMapping removes the mapping between tss signingID and bandtss signing id of the given key
func (k Keeper) DeleteSigningIDMapping(ctx sdk.Context, signingID tss.SigningID) {
	ctx.KVStore(k.storeKey).Delete(types.SigningIDMappingStoreKey(signingID))
}

// HandleCreateSigning creates a new signing process and returns the result.
func (k Keeper) HandleCreateSigning(
	ctx sdk.Context,
	content tsstypes.Content,
	sender sdk.AccAddress,
	feeLimit sdk.Coins,
) (types.SigningID, error) {
	currentGroupID := k.GetCurrentGroupID(ctx)
	if currentGroupID == 0 {
		return 0, types.ErrNoActiveGroup
	}

	replacement := k.GetReplacement(ctx)

	currentGroup, err := k.tssKeeper.GetGroup(ctx, currentGroupID)
	if err != nil {
		return 0, err
	}
	if currentGroup.Status != tsstypes.GROUP_STATUS_ACTIVE {
		return 0, types.ErrNoActiveGroup
	}

	// charged fee if necessary; If found any coins that exceed limit then return error
	feePerSigner := sdk.NewCoins()
	if sender.String() != k.authority {
		feePerSigner = k.GetParams(ctx).Fee
		totalFee := feePerSigner.MulInt(sdk.NewInt(int64(currentGroup.Threshold)))
		for _, fc := range totalFee {
			limitAmt := feeLimit.AmountOf(fc.Denom)
			if fc.Amount.GT(limitAmt) {
				return 0, types.ErrFeeExceedsLimit.Wrapf(
					"require: %s, limit: %s%s",
					fc.String(),
					limitAmt.String(),
					fc.Denom,
				)
			}
		}

		// transfer fee to module account.
		err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, sender, types.ModuleName, totalFee)
		if err != nil {
			return 0, err
		}
	}

	currentGroupSigning, err := k.tssKeeper.CreateSigning(ctx, currentGroup, content)
	if err != nil {
		return 0, err
	}

	replacingGroupSigningID := tss.SigningID(0)
	if replacement.Status == types.REPLACEMENT_STATUS_WAITING_REPLACE {
		replacingGroup, err := k.tssKeeper.GetGroup(ctx, replacement.NewGroupID)
		if err != nil {
			return 0, err
		}

		replacingGroupSigning, err := k.tssKeeper.CreateSigning(ctx, replacingGroup, content)
		if err != nil {
			return 0, err
		}

		replacingGroupSigningID = replacingGroupSigning.ID
	}

	// save signing info
	bandtssSigningID := k.AddSigning(ctx, types.Signing{
		Fee:                     feePerSigner,
		Requester:               sender.String(),
		CurrentGroupSigningID:   currentGroupSigning.ID,
		ReplacingGroupSigningID: replacingGroupSigningID,
	})

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSigningRequestCreated,
			sdk.NewAttribute(types.AttributeKeySigningID, fmt.Sprintf("%d", bandtssSigningID)),
			sdk.NewAttribute(types.AttributeKeyCurrentGroupID, fmt.Sprintf("%d", currentGroupID)),
			sdk.NewAttribute(types.AttributeKeyCurrentGroupSigningID, fmt.Sprintf("%d", currentGroupSigning.ID)),
			sdk.NewAttribute(types.AttributeKeyReplacingGroupID, fmt.Sprintf("%d", replacement.NewGroupID)),
			sdk.NewAttribute(types.AttributeKeyReplacingGroupSigningID, fmt.Sprintf("%d", replacingGroupSigningID)),
		),
	)

	return bandtssSigningID, nil
}

// RefundFee refunds the fee to the requester.
func (k Keeper) RefundFee(ctx sdk.Context, signing tsstypes.Signing, bandtssSigningID types.SigningID) error {
	bandtssSigning, err := k.GetSigning(ctx, bandtssSigningID)
	if err != nil {
		return err
	}

	// Check fee is not zero and this signing is the current signing ID.
	if bandtssSigning.Fee.IsZero() || signing.ID != bandtssSigning.CurrentGroupSigningID {
		return nil
	}

	// Refund fee to requester
	address := sdk.MustAccAddressFromBech32(bandtssSigning.Requester)
	feeCoins := bandtssSigning.Fee.MulInt(sdk.NewInt(int64(len(signing.AssignedMembers))))
	return k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, address, feeCoins)
}

func decodeSigningMappingKeyToSigningID(key []byte) tss.SigningID {
	kv.AssertKeyLength(key, 10)
	return tss.SigningID(sdk.BigEndianToUint64(key[2:]))
}
