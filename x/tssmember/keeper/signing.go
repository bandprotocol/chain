package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	tsstypes "github.com/bandprotocol/chain/v2/x/tss/types"
	"github.com/bandprotocol/chain/v2/x/tssmember/types"
)

// handleRequestSign initiates the signing process by requesting signatures from assigned members.
// It assigns assigned members randomly, computes necessary values, and emits appropriate events.
func (k Keeper) HandleRequestSign(
	ctx sdk.Context,
	groupID tss.GroupID,
	msg []byte,
	feePayer sdk.AccAddress,
	feeLimit sdk.Coins,
) (*tsstypes.Signing, error) {
	// Verify if the group status is active.
	group, err := k.tssKeeper.GetActiveGroup(ctx, groupID)
	if err != nil {
		return nil, err
	}

	fee := sdk.NewCoins()
	// Charge fee if requester is not authority address
	if feePayer.String() != k.authority {
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

	input := tsstypes.CreateSigningInput{
		Group:    group,
		Message:  msg,
		Fee:      fee,
		FeePayer: feePayer,
	}
	result, err := k.tssKeeper.CreateSigning(ctx, input)
	if err != nil {
		return nil, err
	}

	return &result.Signing, nil
}
