package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/bandprotocol/chain/v2/x/tss/types"
)

// HandleInactiveValidators handle inactive validators by inactive validator that has not been activated for a while.
func (k Keeper) HandleInactiveValidators(ctx sdk.Context) {
	// Only process every x (max number of validators) blocks
	maxValidators := k.stakingKeeper.MaxValidators(ctx)
	if ctx.BlockHeight()%int64(maxValidators) != 0 {
		return
	}

	// Set inactive for validator that last active exceeds active duration.
	k.stakingKeeper.IterateBondedValidatorsByPower(
		ctx,
		func(_ int64, validator stakingtypes.ValidatorI) (stop bool) {
			address := sdk.AccAddress(validator.GetOperator())
			status := k.GetStatus(ctx, address)

			if status.Status == types.MEMBER_STATUS_ACTIVE &&
				ctx.BlockTime().After(status.LastActive.Add(k.GetParams(ctx).ActiveDuration)) {
				k.SetInactive(ctx, address)

				ctx.EventManager().EmitEvent(sdk.NewEvent(
					types.EventTypeActivate,
					sdk.NewAttribute(types.AttributeKeyAddress, address.String()),
				))
			}

			return false
		},
	)
}

// SetActive sets the member status to active
func (k Keeper) SetActive(ctx sdk.Context, address sdk.AccAddress) error {
	status := k.GetStatus(ctx, address)
	if status.Status == types.MEMBER_STATUS_ACTIVE {
		return nil
	}

	params := k.GetParams(ctx)
	var penaltyDuration time.Duration
	if status.Status == types.MEMBER_STATUS_INACTIVE {
		penaltyDuration = params.InactivePenaltyDuration
	} else if status.Status == types.MEMBER_STATUS_JAIL {
		penaltyDuration = params.JailPenaltyDuration
	}

	if status.Since.Add(penaltyDuration).After(ctx.BlockTime()) {
		return types.ErrTooSoonToActivate
	}

	status.Status = types.MEMBER_STATUS_ACTIVE
	status.Address = address.String()
	status.Since = ctx.BlockTime()
	status.LastActive = status.Since
	k.SetMemberStatus(ctx, status)

	return nil
}

// SetLastActive sets last active of the member
func (k Keeper) SetLastActive(ctx sdk.Context, address sdk.AccAddress) error {
	status := k.GetStatus(ctx, address)

	if status.Status != types.MEMBER_STATUS_ACTIVE {
		return types.ErrInvalidStatus
	}

	status.LastActive = ctx.BlockTime()
	k.SetMemberStatus(ctx, status)

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
	k.SetMemberStatus(ctx, status)

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
	k.SetMemberStatus(ctx, status)

	return
}
