package types

import (
	"context"
	"time"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/bandprotocol/chain/v3/pkg/tss"
	tsstypes "github.com/bandprotocol/chain/v3/x/tss/types"
)

// AuthzKeeper defines the expected authz keeper. for query and testing only; don't use to
// create/remove grant on deliver tx
type AuthzKeeper interface {
	GetAuthorization(
		ctx context.Context,
		grantee sdk.AccAddress,
		granter sdk.AccAddress,
		msgType string,
	) (authz.Authorization, *time.Time)
	SaveGrant(
		ctx context.Context,
		grantee, granter sdk.AccAddress,
		authorization authz.Authorization,
		expiration *time.Time,
	) error
}

// AccountKeeper defines the expected account keeper (noalias)
type AccountKeeper interface {
	GetModuleAddress(name string) sdk.AccAddress
	GetModuleAccount(ctx context.Context, name string) sdk.ModuleAccountI
	SetModuleAccount(context.Context, sdk.ModuleAccountI)
}

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	GetAllBalances(ctx context.Context, addr sdk.AccAddress) sdk.Coins

	SendCoinsFromModuleToAccount(
		ctx context.Context,
		senderModule string,
		recipientAddr sdk.AccAddress,
		amt sdk.Coins,
	) error
	SendCoinsFromAccountToModule(
		ctx context.Context,
		senderAddr sdk.AccAddress,
		recipientModule string,
		amt sdk.Coins,
	) error
	SendCoinsFromModuleToModule(ctx context.Context, senderModule, recipientModule string, amt sdk.Coins) error
}

// DistrKeeper defines the expected distribution keeper.
type DistrKeeper interface {
	GetCommunityTax(ctx context.Context) (math.LegacyDec, error)
	AllocateTokensToValidator(ctx context.Context, val stakingtypes.ValidatorI, tokens sdk.DecCoins) error
	FundCommunityPool(ctx context.Context, amount sdk.Coins, sender sdk.AccAddress) error
}

type FeePoolManager interface {
	GetFeePool(ctx context.Context) (collections.Item[distrtypes.FeePool], error)
	SetFeePool(ctx context.Context, feePool distrtypes.FeePool)
}

// StakingKeeper defines the expected staking keeper.
type StakingKeeper interface {
	MaxValidators(ctx context.Context) (res uint32, err error)
	ValidatorByConsAddr(context.Context, sdk.ConsAddress) (stakingtypes.ValidatorI, error)
	IterateBondedValidatorsByPower(
		ctx context.Context,
		fn func(index int64, validator stakingtypes.ValidatorI) (stop bool),
	) error
}

// TSSKeeper defines the expected tss keeper (noalias)
type TSSKeeper interface {
	CreateGroup(
		ctx sdk.Context,
		members []sdk.AccAddress,
		threshold uint64,
		moduleOwner string,
	) (tss.GroupID, error)

	RequestSigning(
		ctx sdk.Context,
		groupID tss.GroupID,
		originator tsstypes.Originator,
		content tsstypes.Content,
	) (tss.SigningID, error)

	MustGetMembers(ctx sdk.Context, groupID tss.GroupID) []tsstypes.Member
	GetMemberByAddress(ctx sdk.Context, groupID tss.GroupID, address string) (tsstypes.Member, error)
	ActivateMember(ctx sdk.Context, groupID tss.GroupID, address sdk.AccAddress) error
	DeactivateMember(ctx sdk.Context, groupID tss.GroupID, address sdk.AccAddress) error

	GetDEQueue(ctx sdk.Context, address sdk.AccAddress) tsstypes.DEQueue
	GetGroup(ctx sdk.Context, groupID tss.GroupID) (tsstypes.Group, error)
	MustGetGroup(ctx sdk.Context, groupID tss.GroupID) tsstypes.Group

	GetSigning(ctx sdk.Context, signingID tss.SigningID) (tsstypes.Signing, error)
	MustGetSigning(ctx sdk.Context, signingID tss.SigningID) tsstypes.Signing
	GetSigningResult(ctx sdk.Context, signingID tss.SigningID) (*tsstypes.SigningResult, error)
}
