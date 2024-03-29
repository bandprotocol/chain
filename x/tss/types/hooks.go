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

func (h MultiTSSHooks) AfterCreatingGroupCompleted(ctx sdk.Context, group Group) error {
	for i := range h {
		if err := h[i].AfterCreatingGroupCompleted(ctx, group); err != nil {
			return err
		}
	}

	return nil
}

func (h MultiTSSHooks) AfterCreatingGroupFailed(ctx sdk.Context, group Group) error {
	for i := range h {
		if err := h[i].AfterCreatingGroupFailed(ctx, group); err != nil {
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

func (h MultiTSSHooks) AfterReplacingGroupCompleted(ctx sdk.Context, replacement Replacement) error {
	for i := range h {
		if err := h[i].AfterReplacingGroupCompleted(ctx, replacement); err != nil {
			return err
		}
	}

	return nil
}

func (h MultiTSSHooks) AfterReplacingGroupFailed(ctx sdk.Context, replacement Replacement) error {
	for i := range h {
		if err := h[i].AfterReplacingGroupFailed(ctx, replacement); err != nil {
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

func (h MultiTSSHooks) BeforeSetSigningExpired(ctx sdk.Context, signing Signing) error {
	for i := range h {
		if err := h[i].BeforeSetSigningExpired(ctx, signing); err != nil {
			return err
		}
	}

	return nil
}
