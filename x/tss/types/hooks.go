package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// combine multiple tss hooks, all hook functions are run in array sequence
var _ TSSHooks = &MultiTSSHooks{}

type MultiTSSHooks []TSSHooks

func NewMultiTSSHooks(hooks ...TSSHooks) MultiTSSHooks {
	return hooks
}

func (h MultiTSSHooks) AfterGroupActivated(ctx sdk.Context, group Group) error {
	for i := range h {
		if err := h[i].AfterGroupActivated(ctx, group); err != nil {
			return err
		}
	}

	return nil
}

func (h MultiTSSHooks) AfterGroupFailedToActivate(ctx sdk.Context, group Group) error {
	for i := range h {
		if err := h[i].AfterGroupFailedToActivate(ctx, group); err != nil {
			return err
		}
	}

	return nil
}

func (h MultiTSSHooks) AfterGroupReplaced(ctx sdk.Context, replacement Replacement) error {
	for i := range h {
		if err := h[i].AfterGroupReplaced(ctx, replacement); err != nil {
			return err
		}
	}

	return nil
}

func (h MultiTSSHooks) AfterGroupFailedToReplace(ctx sdk.Context, replacement Replacement) error {
	for i := range h {
		if err := h[i].AfterGroupFailedToReplace(ctx, replacement); err != nil {
			return err
		}
	}

	return nil
}

func (h MultiTSSHooks) AfterSigningFailed(ctx sdk.Context, signing Signing) error {
	for i := range h {
		if err := h[i].AfterSigningFailed(ctx, signing); err != nil {
			return err
		}
	}

	return nil
}

func (h MultiTSSHooks) AfterSigningCompleted(ctx sdk.Context, signing Signing) error {
	for i := range h {
		if err := h[i].AfterSigningCompleted(ctx, signing); err != nil {
			return err
		}
	}

	return nil
}

func (h MultiTSSHooks) AfterSigningCreated(ctx sdk.Context, signing Signing) error {
	for i := range h {
		if err := h[i].AfterSigningCreated(ctx, signing); err != nil {
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
