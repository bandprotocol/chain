package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// combine multiple tss hooks, all hook functions are run in array sequence
var _ TSSHooks = &MultiTSSHooks{}

type MultiTSSHooks []TSSHooks

func NewMultiStakingHooks(hooks ...TSSHooks) MultiTSSHooks {
	return hooks
}

func (h MultiTSSHooks) AfterGroupActivated(ctx sdk.Context, group Group) {
	for i := range h {
		h[i].AfterGroupActivated(ctx, group)
	}
}

func (h MultiTSSHooks) AfterGroupFailedToActivate(ctx sdk.Context, group Group) {
	for i := range h {
		h[i].AfterGroupFailedToActivate(ctx, group)
	}
}

func (h MultiTSSHooks) AfterGroupReplaced(ctx sdk.Context, replacement Replacement) {
	for i := range h {
		h[i].AfterGroupReplaced(ctx, replacement)
	}
}

func (h MultiTSSHooks) AfterGroupFailedToReplace(ctx sdk.Context, replacement Replacement) {
	for i := range h {
		h[i].AfterGroupFailedToReplace(ctx, replacement)
	}
}

func (h MultiTSSHooks) AfterSigningFailed(ctx sdk.Context, signing Signing) {
	for i := range h {
		h[i].AfterSigningFailed(ctx, signing)
	}
}

func (h MultiTSSHooks) AfterSigningCompleted(ctx sdk.Context, signing Signing) {
	for i := range h {
		h[i].AfterSigningCompleted(ctx, signing)
	}
}

func (h MultiTSSHooks) AfterSigningInitiated(ctx sdk.Context, signing Signing) error {
	for i := range h {
		if err := h[i].AfterSigningInitiated(ctx, signing); err != nil {
			return err
		}
	}

	return nil
}

func (h MultiTSSHooks) AfterHandleSetDEs(ctx sdk.Context, address sdk.AccAddress) error {
	for i := range h {
		if err := h[i].AfterHandleSetDEs(ctx, address); err != nil {
			return err
		}
	}

	return nil
}

func (h MultiTSSHooks) BeforeSetGroupExpired(ctx sdk.Context, group Group) error {
	for i := range h {
		if err := h[i].BeforeSetGroupExpired(ctx, group); err != nil {
			return err
		}
	}

	return nil
}

func (h MultiTSSHooks) AfterPollDE(ctx sdk.Context, member sdk.AccAddress) error {
	for i := range h {
		if err := h[i].AfterPollDE(ctx, member); err != nil {
			return err
		}
	}

	return nil
}

func (h MultiTSSHooks) BeforeSetSigningExpired(ctx sdk.Context, signing Signing) error {
	for i := range h {
		if err := h[i].BeforeSetSigningExpired(ctx, signing); err != nil {
			return err
		}
	}

	return nil
}
