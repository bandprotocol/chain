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
	msg, err := k.tssKeeper.HandleSigningContent(ctx, input.Content)
	if err != nil {
		return nil, err
	}

	tssInput := tsstypes.CreateSigningInput{
		GroupID:      input.GroupID,
		Message:      msg,
		IsFeeCharged: input.Sender.String() != k.authority,
		FeePayer:     input.Sender,
		FeeLimit:     input.FeeLimit,
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
