package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

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

// SetSigning sets a signing of the Bandtss module.
func (k Keeper) SetSigning(ctx sdk.Context, signing types.Signing) {
	ctx.KVStore(k.storeKey).Set(types.SigningStoreKey(signing.ID), k.cdc.MustMarshal(&signing))
}

// AddSigning adds the Signing data to the store and returns the new Signing ID.
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
	bz := ctx.KVStore(k.storeKey).Get(types.SigningStoreKey(id))
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
	return sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.SigningStoreKeyPrefix)
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

// DeleteSigning removes the bandtss signing of the given id
func (k Keeper) DeleteSigning(ctx sdk.Context, id types.SigningID) {
	ctx.KVStore(k.storeKey).Delete(types.SigningStoreKey(id))
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
	return sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.SigningStoreKeyPrefix)
}

// GetSigningIDMappings retrieves all signingID mapping items of the store.
func (k Keeper) GetSigningIDMappings(ctx sdk.Context) []types.SigningIDMappingGenesis {
	var mappings []types.SigningIDMappingGenesis
	iterator := k.GetSigningRequestIDMappingIterator(ctx)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		mappings = append(mappings, types.SigningIDMappingGenesis{
			SigningID:        tss.SigningID(sdk.BigEndianToUint64(iterator.Key())),
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
	// Execute the handler to process the request.
	msg, err := k.tssKeeper.HandleSigningContent(ctx, content)
	if err != nil {
		return 0, err
	}

	currentGroupID := k.GetCurrentGroupID(ctx)
	replacingGroupID := k.GetReplacingGroupID(ctx)

	currentGroup, err := k.tssKeeper.GetGroup(ctx, currentGroupID)
	if err != nil {
		return 0, err
	}

	// charged fee if necessary; If found any coins that exceed limit then return error
	feePerSigner := sdk.NewCoins()
	if sender.String() != k.authority {
		feePerSigner = k.GetParams(ctx).Fee
		totalFee := feePerSigner.MulInt(sdk.NewInt(int64(currentGroup.Threshold)))
		for _, fc := range totalFee {
			limitAmt := feeLimit.AmountOf(fc.Denom)
			if fc.Amount.GT(limitAmt) {
				return 0, types.ErrNotEnoughFee.Wrapf(
					"require: %s, limit: %s%s",
					fc.String(),
					limitAmt.String(),
					fc.Denom,
				)
			}
		}

		// transfer fee to module account.
		if !totalFee.IsZero() {
			err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, sender, types.ModuleName, totalFee)
			if err != nil {
				return 0, err
			}
		}
	}

	currentGroupSigning, err := k.tssKeeper.CreateSigning(ctx, currentGroup, msg)
	if err != nil {
		return 0, err
	}

	replacingGroupSigningID := tss.SigningID(0)
	if replacingGroupID != 0 {
		replacingGroup, err := k.tssKeeper.GetGroup(ctx, replacingGroupID)
		if err != nil {
			return 0, err
		}

		replacingGroupSigning, err := k.tssKeeper.CreateSigning(ctx, replacingGroup, msg)
		if err != nil {
			return 0, err
		}

		replacingGroupSigningID = replacingGroupSigning.ID
	}

	// save signingInfo
	bandtssSigningID := k.AddSigning(ctx, types.Signing{
		Fee:                     feePerSigner,
		Requester:               sender.String(),
		CurrentGroupID:          k.GetCurrentGroupID(ctx),
		CurrentGroupSigningID:   currentGroupSigning.ID,
		ReplacingGroupID:        replacingGroupID,
		ReplacingGroupSigningID: replacingGroupSigningID,
	})

	return bandtssSigningID, nil
}

// CheckRefundFee refunds the fee to the requester.
func (k Keeper) CheckRefundFee(ctx sdk.Context, signing tsstypes.Signing) error {
	bandtssSigningID := k.GetSigningIDMapping(ctx, signing.ID)
	if bandtssSigningID == 0 {
		return types.ErrSigningNotFound
	}

	bandtssSigning, err := k.GetSigning(ctx, bandtssSigningID)
	if err != nil {
		return err
	}

	if bandtssSigning.Fee.IsZero() || signing.GroupID != bandtssSigning.CurrentGroupID {
		return nil
	}

	// Refund fee to requester
	address := sdk.MustAccAddressFromBech32(bandtssSigning.Requester)
	feeCoins := bandtssSigning.Fee.MulInt(sdk.NewInt(int64(len(signing.AssignedMembers))))
	return k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, address, feeCoins)
}
