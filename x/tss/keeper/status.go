package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

// SetActive sets the member status to active
func (k Keeper) SetActive(ctx sdk.Context, address sdk.AccAddress, groupID tss.GroupID) error {
	status, err := k.GetStatus(ctx, address, groupID)
	if err != nil {
		return err
	}

	if status.IsActive {
		return nil
	}

	penaltyDuration := k.InactivePenaltyDuration(ctx)
	if status.Since.Add(penaltyDuration).After(ctx.BlockTime()) {
		return types.ErrTooSoonToActivate
	}

	status.IsActive = true
	status.Since = ctx.BlockTime()
	k.SetStatus(ctx, address, status)

	return nil
}

// SetInActive sets the member status to inactive
func (k Keeper) SetInActive(ctx sdk.Context, address sdk.AccAddress, groupID tss.GroupID) error {
	status, err := k.GetStatus(ctx, address, groupID)
	if err != nil {
		return err
	}

	if !status.IsActive {
		return nil
	}

	status.IsActive = false
	status.Since = ctx.BlockTime()
	k.SetStatus(ctx, address, status)

	return nil
}
