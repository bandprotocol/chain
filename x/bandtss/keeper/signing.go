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

	group, err := k.tssKeeper.GetGroup(ctx, groupID)
	if err != nil {
		return nil, err
	}
	if group.Status != tsstypes.GROUP_STATUS_ACTIVE {
		return nil, tsstypes.ErrGroupIsNotActive.Wrap("group status is not active")
	}

	// charged fee if necessary; If found any coins that exceed limit then return error
	totalFee := sdk.NewCoins()
	if sender.String() != k.authority {
		totalFee = k.GetParams(ctx).Fee.MulInt(sdk.NewInt(int64(group.Threshold)))
		for _, fc := range totalFee {
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

	signing, err := k.tssKeeper.CreateSigning(ctx, group, msg)
	if err != nil {
		return nil, err
	}

	// transfer fee to module account.
	if !totalFee.IsZero() {
		err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, sender, types.ModuleName, totalFee)
		if err != nil {
			return nil, err
		}
	}

	// save signingFee
	k.SetSigningFee(ctx, types.SigningFee{
		SigningID: signing.ID,
		Fee:       totalFee,
		Requester: sender.String(),
	})

	return signing, nil
}

func (k Keeper) RefundFee(ctx sdk.Context, signing tsstypes.Signing) error {
	signingFee, err := k.GetSigningFee(ctx, signing.ID)
	if err != nil {
		return err
	}

	if signingFee.Fee.IsZero() {
		return nil
	}

	// Refund fee to requester
	address := sdk.MustAccAddressFromBech32(signingFee.Requester)
	feeCoins := signingFee.Fee.MulInt(sdk.NewInt(int64(len(signing.AssignedMembers))))
	return k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, address, feeCoins)
}
