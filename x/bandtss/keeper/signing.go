package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/bandtss/types"
	tsstypes "github.com/bandprotocol/chain/v2/x/tss/types"
)

// HandleCreateSigning creates a new signing process and returns the result.
func (k Keeper) HandleCreateSigning(
	ctx sdk.Context,
	input types.HandleCreateSigningInput,
) (*types.HandleCreateSigningResult, error) {
	// Execute the handler to process the request.
	msg, err := k.HandleSigningContent(ctx, input.Content)
	if err != nil {
		return nil, err
	}

	group, err := k.tssKeeper.GetActiveGroup(ctx, input.GroupID)
	if err != nil {
		return nil, err
	}

	// charged fee if necessary
	fee := sdk.NewCoins()
	if input.Sender.String() != k.authority {
		fee = group.Fee

		// If found any coins that exceed limit then return error
		feeCoins := group.Fee.MulInt(sdk.NewInt(int64(group.Threshold)))
		for _, fc := range feeCoins {
			limitAmt := input.FeeLimit.AmountOf(fc.Denom)
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

	tssInput := tsstypes.CreateSigningInput{
		Group:    group,
		Message:  msg,
		Fee:      fee,
		FeePayer: input.Sender,
	}

	result, err := k.tssKeeper.CreateSigning(ctx, tssInput)
	if err != nil {
		return nil, err
	}

	return &types.HandleCreateSigningResult{
		Message: msg,
		Signing: result.Signing,
	}, nil
}

func (k Keeper) HandleSigningContent(
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
