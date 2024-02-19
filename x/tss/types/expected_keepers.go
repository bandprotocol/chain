package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
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

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	GetAllBalances(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin

	SendCoinsFromModuleToAccount(
		ctx sdk.Context,
		senderModule string,
		recipientAddr sdk.AccAddress,
		amt sdk.Coins,
	) error
	SendCoinsFromAccountToModule(
		ctx sdk.Context,
		senderAddr sdk.AccAddress,
		recipientModule string,
		amt sdk.Coins,
	) error
	SendCoinsFromModuleToModule(ctx sdk.Context, senderModule, recipientModule string, amt sdk.Coins) error
}

// StakingKeeper defines the expected staking keeper.
type StakingKeeper interface {
	MaxValidators(ctx sdk.Context) (res uint32)
	ValidatorByConsAddr(sdk.Context, sdk.ConsAddress) stakingtypes.ValidatorI
	IterateBondedValidatorsByPower(
		ctx sdk.Context,
		fn func(index int64, validator stakingtypes.ValidatorI) (stop bool),
	)
}

// DistrKeeper defines the expected distribution keeper.
type DistrKeeper interface {
	GetCommunityTax(ctx sdk.Context) (percent sdk.Dec)
	GetFeePool(ctx sdk.Context) (feePool distrtypes.FeePool)
	SetFeePool(ctx sdk.Context, feePool distrtypes.FeePool)
	AllocateTokensToValidator(ctx sdk.Context, val stakingtypes.ValidatorI, tokens sdk.DecCoins)
}

// TSSHooks event hooks for staking validator object (noalias)
type TSSHooks interface {
	// Must be called when a group is ready; no error is returned due to the endblock process.
	AfterGroupActivated(ctx sdk.Context, group Group)

	// Must be called when a group cannot be created successfully or is expired; no error is returned
	// due to the endblock process.
	AfterGroupFailedToActivate(ctx sdk.Context, group Group)

	// Must be called when a group is replaced; no error is returned due to the endblock process.
	AfterGroupReplaced(ctx sdk.Context, replacement Replacement)

	// Must be called when a group cannot be replaced; no error is returned due to the endblock process.
	AfterGroupFailedToReplace(ctx sdk.Context, replacement Replacement)

	// Must be called when a member status is updated; no error is returned due to the endblock process.
	AfterStatusUpdated(ctx sdk.Context, status Status)

	// Must be called when a signing request is unsuccessfully signed.
	AfterSigningFailed(ctx sdk.Context, signing Signing)

	// Must be called when a signing request is successfully signed by selected members.
	AfterSigningCompleted(ctx sdk.Context, signing Signing)

	// Must be called when a signing request is initiated.
	AfterSigningInitiated(ctx sdk.Context, signing Signing) error
}
