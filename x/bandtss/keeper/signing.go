package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/bandtss/types"
	tsstypes "github.com/bandprotocol/chain/v2/x/tss/types"
)

// HandleCreateSigning creates a new signing process and returns the result.
func (k Keeper) HandleCreateSigning(
	ctx sdk.Context,
	groupID tss.GroupID,
	content tsstypes.Content,
	sender sdk.AccAddress,
	feeLimit sdk.Coins,
) (*tsstypes.Signing, error) {
	// Execute the handler to process the request.
	msg, err := k.tssKeeper.HandleSigningContent(ctx, content)
	if err != nil {
		return nil, err
	}

	group, err := k.tssKeeper.GetActiveGroup(ctx, groupID)
	if err != nil {
		return nil, err
	}

	// charged fee if necessary
	fee := sdk.NewCoins()
	if sender.String() != k.authority {
		fee = group.Fee

		// If found any coins that exceed limit then return error
		feeCoins := group.Fee.MulInt(sdk.NewInt(int64(group.Threshold)))
		for _, fc := range feeCoins {
			limitAmt := feeLimit.AmountOf(fc.Denom)
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

	signing, err := k.tssKeeper.CreateSigning(ctx, group, msg, fee, sender)
	if err != nil {
		return nil, err
	}

	return signing, nil
}
