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

func (h MultiTSSHooks) AfterCreatingGroupCompleted(ctx sdk.Context, group Group) {
	for i := range h {
		h[i].AfterCreatingGroupCompleted(ctx, group)
	}
}

func (h MultiTSSHooks) AfterCreatingGroupFailed(ctx sdk.Context, group Group) {
	for i := range h {
		h[i].AfterCreatingGroupFailed(ctx, group)
	}
}

func (h MultiTSSHooks) BeforeSetGroupExpired(ctx sdk.Context, group Group) {
	for i := range h {
		h[i].BeforeSetGroupExpired(ctx, group)
	}
}

func (h MultiTSSHooks) AfterReplacingGroupCompleted(ctx sdk.Context, replacement Replacement) {
	for i := range h {
		h[i].AfterReplacingGroupCompleted(ctx, replacement)
	}
}

func (h MultiTSSHooks) AfterReplacingGroupFailed(ctx sdk.Context, replacement Replacement) {
	for i := range h {
		h[i].AfterReplacingGroupFailed(ctx, replacement)
	}
}

func (h MultiTSSHooks) AfterSigningCreated(ctx sdk.Context, signing Signing) error {
	for i := range h {
		if err := h[i].AfterSigningCreated(ctx, signing); err != nil {
			return err
		}
	}

	return nil
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

func (h MultiTSSHooks) AfterHandleSetDEs(ctx sdk.Context, address sdk.AccAddress) error {
	for i := range h {
		if err := h[i].AfterHandleSetDEs(ctx, address); err != nil {
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

func (h MultiTSSHooks) BeforeSetSigningExpired(ctx sdk.Context, signing Signing) {
	for i := range h {
		h[i].BeforeSetSigningExpired(ctx, signing)
	}
}
