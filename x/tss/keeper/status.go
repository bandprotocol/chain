package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/tss/types"
)

// SetActive sets the member status to active
func (k Keeper) SetActive(ctx sdk.Context, address sdk.AccAddress) error {
	status := k.GetStatus(ctx, address)

	if status.Status == types.MEMBER_STATUS_ACTIVE {
		return nil
	} else if status.Status == types.MEMBER_STATUS_INACTIVE {
		penaltyDuration := k.InactivePenaltyDuration(ctx)
		if status.Since.Add(penaltyDuration).After(ctx.BlockTime()) {
			return types.ErrTooSoonToActivate
		}
	} else if status.Status == types.MEMBER_STATUS_JAIL {
		penaltyDuration := k.JailPenaltyDuration(ctx)
		if status.Since.Add(penaltyDuration).After(ctx.BlockTime()) {
			return types.ErrTooSoonToActivate
		}
	}

	status.Status = types.MEMBER_STATUS_ACTIVE
	status.Address = address.String()
	status.Since = ctx.BlockTime()
	status.LastActive = status.Since
	k.SetStatus(ctx, status)

	return nil
}

// SetLastActive sets last active of the member
func (k Keeper) SetLastActive(ctx sdk.Context, address sdk.AccAddress) error {
	status := k.GetStatus(ctx, address)

	if status.Status != types.MEMBER_STATUS_ACTIVE {
		return types.ErrInvalidStatus
	}

	status.LastActive = ctx.BlockTime()
	k.SetStatus(ctx, status)

	return nil
}

// SetInactive sets the member status to inactive
func (k Keeper) SetInactive(ctx sdk.Context, address sdk.AccAddress) {
	status := k.GetStatus(ctx, address)

	if status.Status == types.MEMBER_STATUS_INACTIVE {
		return
	}

	status.Status = types.MEMBER_STATUS_INACTIVE
	status.Address = address.String()
	status.Since = ctx.BlockTime()
	k.SetStatus(ctx, status)

	return
}

// SetJail sets the member status to jail
func (k Keeper) SetJail(ctx sdk.Context, address sdk.AccAddress) {
	status := k.GetStatus(ctx, address)

	if status.Status == types.MEMBER_STATUS_JAIL {
		return
	}

	status.Status = types.MEMBER_STATUS_JAIL
	status.Address = address.String()
	status.Since = ctx.BlockTime()
	k.SetStatus(ctx, status)

	return
}
