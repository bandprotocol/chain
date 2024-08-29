package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	tsstypes "github.com/bandprotocol/chain/v2/x/tss/types"
)

// AuthzKeeper defines the expected authz keeper. for query and testing only; don't use to
// create/remove grant on deliver tx
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

// AccountKeeper defines the expected account keeper (noalias)
type AccountKeeper interface {
	GetModuleAddress(name string) sdk.AccAddress
	GetModuleAccount(ctx sdk.Context, name string) authtypes.ModuleAccountI
	SetModuleAccount(sdk.Context, authtypes.ModuleAccountI)
}

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	GetAllBalances(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins

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

// DistrKeeper defines the expected distribution keeper.
type DistrKeeper interface {
	GetCommunityTax(ctx sdk.Context) (percent sdk.Dec)
	GetFeePool(ctx sdk.Context) (feePool distrtypes.FeePool)
	SetFeePool(ctx sdk.Context, feePool distrtypes.FeePool)
	AllocateTokensToValidator(ctx sdk.Context, val stakingtypes.ValidatorI, tokens sdk.DecCoins)
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
