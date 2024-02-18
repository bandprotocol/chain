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

func (h MultiTSSHooks) AfterStatusUpdated(ctx sdk.Context, status Status) {
	for i := range h {
		h[i].AfterStatusUpdated(ctx, status)
	}
}
