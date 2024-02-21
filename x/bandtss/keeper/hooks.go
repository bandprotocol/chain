package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/bandtss/types"
	tsstypes "github.com/bandprotocol/chain/v2/x/tss/types"
)

// Wrapper struct
type Hooks struct {
	k Keeper
}

var _ tsstypes.TSSHooks = Hooks{}

// Create new distribution hooks
func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}

func (h Hooks) AfterGroupActivated(ctx sdk.Context, group tsstypes.Group) {}

func (h Hooks) AfterGroupFailedToActivate(ctx sdk.Context, group tsstypes.Group) {}

func (h Hooks) AfterGroupReplaced(ctx sdk.Context, replacement tsstypes.Replacement) {}

func (h Hooks) AfterGroupFailedToReplace(ctx sdk.Context, replacement tsstypes.Replacement) {}

func (h Hooks) AfterStatusUpdated(ctx sdk.Context, status types.Status) {}

func (h Hooks) AfterSigningFailed(ctx sdk.Context, signing tsstypes.Signing) {
	if signing.Fee.IsZero() {
		return
	}

	address := sdk.MustAccAddressFromBech32(signing.Requester)
	feeCoins := signing.Fee.MulInt(sdk.NewInt(int64(len(signing.AssignedMembers))))

	// Refund fee to requester
	err := h.k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, address, feeCoins)
	if err != nil {
		panic(err) // Error is not possible
	}
}

func (h Hooks) AfterSigningCompleted(ctx sdk.Context, signing tsstypes.Signing) {
	// Send fee to assigned members.
	for _, am := range signing.AssignedMembers {
		address := sdk.MustAccAddressFromBech32(am.Address)
		if err := h.k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, address, signing.Fee); err != nil {
			panic(err) // Error is not possible
		}
	}
}

func (h Hooks) AfterSigningInitiated(ctx sdk.Context, signing tsstypes.Signing) error {
	feeCoins := signing.Fee.MulInt(sdk.NewInt(int64(len(signing.AssignedMembers))))
	if feeCoins.IsZero() {
		return nil
	}

	address, err := sdk.AccAddressFromBech32(signing.Requester)
	if err != nil {
		return err
	}

	err = h.k.bankKeeper.SendCoinsFromAccountToModule(ctx, address, types.ModuleName, feeCoins)
	if err != nil {
		return err
	}

	return nil
}

func (h Hooks) AfterHandleSetDEs(ctx sdk.Context, address sdk.AccAddress) error {
	// only update status if the member was paused
	status := h.k.GetStatus(ctx, address)
	if status.Status != types.MEMBER_STATUS_PAUSED {
		return nil
	}

	// if DE is still empty, keep its status as is.
	left := h.k.tssKeeper.GetDECount(ctx, address)
	if left == 0 {
		return nil
	}

	// Set status to active and update the status in tssKeeper
	err := h.k.SetActiveStatus(ctx, address)
	if err != nil {
		return err
	}

	return nil
}

func (h Hooks) BeforeSetGroupExpired(ctx sdk.Context, group tsstypes.Group) error {
	penalizedMembers, err := h.k.tssKeeper.GetPenalizedMembersExpiredGroup(ctx, group)
	if err != nil {
		return err
	}

	for _, m := range penalizedMembers {
		h.k.SetJailStatus(ctx, m)
	}

	return nil
}

func (h Hooks) AfterPollDE(ctx sdk.Context, member sdk.AccAddress) error {
	left := h.k.tssKeeper.GetDECount(ctx, member)
	if left == 0 {
		h.k.SetPausedStatus(ctx, member)
		h.k.tssKeeper.SetMemberStatus(ctx, member, false)
	}

	return nil
}

func (h Hooks) BeforeSetSigningExpired(ctx sdk.Context, signing tsstypes.Signing) error {
	penalizedMembers, err := h.k.tssKeeper.GetPenalizedMembersExpiredSigning(ctx, signing)
	if err != nil {
		return err
	}

	for _, m := range penalizedMembers {
		h.k.SetInactiveStatus(ctx, m)
	}

	return nil
}
