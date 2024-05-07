package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
)

// AuthzKeeper defines the expected authz keeper. for query and testing only don't use to create/remove grant on deliver tx
type AuthzKeeper interface {
	GetAuthorization(
		ctx sdk.Context,
		grantee sdk.AccAddress,
		granter sdk.AccAddress,
		msgType string,
	) (authz.Authorization, *time.Time)
	SaveGrant(
		ctx sdk.Context,
		grantee, granter sdk.AccAddress,
		authorization authz.Authorization,
		expiration *time.Time,
	) error
}

// RollingseedKeeper defines the expected rollingseed keeper
type RollingseedKeeper interface {
	GetRollingSeed(ctx sdk.Context) []byte
}

// AccountKeeper defines the expected account keeper (noalias)
type AccountKeeper interface {
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) types.AccountI

	GetModuleAddress(name string) sdk.AccAddress
	GetModuleAccount(ctx sdk.Context, name string) types.ModuleAccountI
}

// TSSHooks event hooks for tss validator object (noalias)
type TSSHooks interface {
	// Must be called when a group is created successfully.
	AfterCreatingGroupCompleted(ctx sdk.Context, group Group) error

	// Must be called when a group creation.
	AfterCreatingGroupFailed(ctx sdk.Context, group Group) error

	// Must be called before setting group status to expired.
	BeforeSetGroupExpired(ctx sdk.Context, group Group) error

	// Must be called when a signing request is unsuccessfully signed.
	AfterSigningFailed(ctx sdk.Context, signing Signing) error

	// Must be called when a signing request is successfully signed by selected members.
	AfterSigningCompleted(ctx sdk.Context, signing Signing) error

	// Must be called before setting signing status to expired.
	BeforeSetSigningExpired(ctx sdk.Context, signing Signing) error
}
