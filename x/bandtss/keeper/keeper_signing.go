package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/pkg/ctxcache"
	"github.com/bandprotocol/chain/v3/pkg/tss"
	"github.com/bandprotocol/chain/v3/x/bandtss/types"
	tsstypes "github.com/bandprotocol/chain/v3/x/tss/types"
)

// CreateDirectSigningRequest creates a new signing process and returns the result.
func (k Keeper) CreateDirectSigningRequest(
	ctx sdk.Context,
	content tsstypes.Content,
	memo string,
	sender sdk.AccAddress,
	feeLimit sdk.Coins,
) (types.SigningID, error) {
	originator := tsstypes.NewDirectOriginator(ctx.ChainID(), sender.String(), memo)
	return k.createSigningRequest(ctx, &originator, content, sender, feeLimit)
}

func (k Keeper) CreateTunnelSigningRequest(
	ctx sdk.Context,
	tunnelID uint64,
	destinationChainID string,
	destinationContractAddr string,
	content tsstypes.Content,
	sender sdk.AccAddress,
	feeLimit sdk.Coins,
) (types.SigningID, error) {
	originator := tsstypes.NewTunnelOriginator(
		ctx.ChainID(),
		tunnelID,
		destinationChainID,
		destinationContractAddr,
	)
	return k.createSigningRequest(ctx, &originator, content, sender, feeLimit)
}

// createSigningRequest creates a new signing process and returns the result.
func (k Keeper) createSigningRequest(
	ctx sdk.Context,
	originator tsstypes.Originator,
	content tsstypes.Content,
	sender sdk.AccAddress,
	feeLimit sdk.Coins,
) (types.SigningID, error) {
	currentGroupID := k.GetCurrentGroup(ctx).GroupID
	incomingGroupID := k.GetIncomingGroupID(ctx)
	if currentGroupID == 0 && incomingGroupID == 0 {
		return 0, types.ErrNoActiveGroup
	}

	// charged fee if necessary; If found any coins that exceed limit then return error
	feePerSigner := sdk.NewCoins()
	totalFee := sdk.NewCoins()
	if sender.String() != k.authority && currentGroupID != 0 {
		currentGroup, err := k.tssKeeper.GetGroup(ctx, currentGroupID)
		if err != nil {
			return 0, err
		}

		feePerSigner = k.GetParams(ctx).FeePerSigner
		totalFee = feePerSigner.MulInt(math.NewInt(int64(currentGroup.Threshold)))
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

	currentGroupSigningID := tss.SigningID(0)
	incomingGroupSigningID := tss.SigningID(0)
	if currentGroupID != 0 {
		signingID, err := k.tssKeeper.RequestSigning(ctx, currentGroupID, originator, content)
		if err != nil {
			return 0, err
		}
		currentGroupSigningID = signingID
	}

	// create signing request for incoming group if any. In case of error, emit event and continue
	// the process, as the signing request for incoming group is optional.
	if incomingGroupID != 0 {
		createSigningFunc := func(ctx sdk.Context) (err error) {
			incomingGroupSigningID, err = k.tssKeeper.RequestSigning(ctx, incomingGroupID, originator, content)
			return err
		}

		if err := ctxcache.ApplyFuncIfNoError(ctx, createSigningFunc); err != nil {
			codespace, code, _ := errorsmod.ABCIInfo(err, false)
			ctx.EventManager().EmitEvent(sdk.NewEvent(
				types.EventTypeCreateSigningFailed,
				sdk.NewAttribute(types.AttributeKeyGroupID, fmt.Sprintf("%d", incomingGroupID)),
				sdk.NewAttribute(types.AttributeKeySigningErrReason, err.Error()),
				sdk.NewAttribute(types.AttributeKeySigningErrCodespace, codespace),
				sdk.NewAttribute(types.AttributeKeySigningErrCode, fmt.Sprintf("%d", code)),
			))
		}
	}

	// save signing info
	bandtssSigningID := k.AddSigning(ctx, feePerSigner, sender, currentGroupSigningID, incomingGroupSigningID)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSigningRequestCreated,
			sdk.NewAttribute(types.AttributeKeySigningID, fmt.Sprintf("%d", bandtssSigningID)),
			sdk.NewAttribute(types.AttributeKeyCurrentGroupID, fmt.Sprintf("%d", currentGroupID)),
			sdk.NewAttribute(types.AttributeKeyCurrentGroupSigningID, fmt.Sprintf("%d", currentGroupSigningID)),
			sdk.NewAttribute(types.AttributeKeyIncomingGroupID, fmt.Sprintf("%d", incomingGroupID)),
			sdk.NewAttribute(types.AttributeKeyIncomingGroupSigningID, fmt.Sprintf("%d", incomingGroupSigningID)),
			sdk.NewAttribute(types.AttributeKeyTotalFee, totalFee.String()),
		),
	)

	return bandtssSigningID, nil
}

// =====================================
// Signing store
// =====================================

// SetSigningCount sets the number of bandtss signing count to the given value.
func (k Keeper) SetSigningCount(ctx sdk.Context, count uint64) {
	ctx.KVStore(k.storeKey).Set(types.SigningCountStoreKey, sdk.Uint64ToBigEndian(count))
}

// GetSigningCount returns the current number of all bandtss signing ever existed.
func (k Keeper) GetSigningCount(ctx sdk.Context) uint64 {
	return sdk.BigEndianToUint64(ctx.KVStore(k.storeKey).Get(types.SigningCountStoreKey))
}

// SetSigning sets a signing of the bandtss module.
func (k Keeper) SetSigning(ctx sdk.Context, signing types.Signing) {
	ctx.KVStore(k.storeKey).Set(types.SigningInfoStoreKey(signing.ID), k.cdc.MustMarshal(&signing))
}

// AddSigning adds the signing data to the store and returns the new Signing ID.
func (k Keeper) AddSigning(
	ctx sdk.Context,
	feePerSigner sdk.Coins,
	sender sdk.AccAddress,
	currentGroupSigningID tss.SigningID,
	incomingGroupSigningID tss.SigningID,
) types.SigningID {
	id := types.SigningID(k.GetSigningCount(ctx) + 1)
	signing := types.NewSigning(id, feePerSigner, sender, currentGroupSigningID, incomingGroupSigningID)
	k.SetSigning(ctx, signing)

	if currentGroupSigningID != 0 {
		k.SetSigningIDMapping(ctx, currentGroupSigningID, id)
	}
	if incomingGroupSigningID != 0 {
		k.SetSigningIDMapping(ctx, incomingGroupSigningID, id)
	}

	k.SetSigningCount(ctx, uint64(id))
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

// DeleteSigningIDMapping removes the mapping between tss signingID and bandtss signing id of the given key
func (k Keeper) DeleteSigningIDMapping(ctx sdk.Context, signingID tss.SigningID) {
	ctx.KVStore(k.storeKey).Delete(types.SigningIDMappingStoreKey(signingID))
}
