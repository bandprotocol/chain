package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	tsstypes "github.com/bandprotocol/chain/v2/x/tss/types"
)

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

// DistrKeeper defines the expected distribution keeper.
type DistrKeeper interface {
	GetCommunityTax(ctx sdk.Context) (percent sdk.Dec)
	GetFeePool(ctx sdk.Context) (feePool distrtypes.FeePool)
	SetFeePool(ctx sdk.Context, feePool distrtypes.FeePool)
	AllocateTokensToValidator(ctx sdk.Context, val stakingtypes.ValidatorI, tokens sdk.DecCoins)
}

// RollingseedKeeper defines the expected rollingseed keeper
type RollingseedKeeper interface {
	GetRollingSeed(ctx sdk.Context) []byte
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

	CreateSigning(
		ctx sdk.Context,
		group tsstypes.Group,
		message []byte,
	) (*tsstypes.Signing, error)

	MustGetMembers(ctx sdk.Context, groupID tss.GroupID) []tsstypes.Member
	GetMemberByAddress(ctx sdk.Context, groupID tss.GroupID, address string) (tsstypes.Member, error)
	ActivateMember(ctx sdk.Context, groupID tss.GroupID, address sdk.AccAddress) error
	DeactivateMember(ctx sdk.Context, groupID tss.GroupID, address sdk.AccAddress) error

	GetDECount(ctx sdk.Context, address sdk.AccAddress) uint64
	GetGroup(ctx sdk.Context, groupID tss.GroupID) (tsstypes.Group, error)

	GetPenalizedMembersExpiredGroup(ctx sdk.Context, group tsstypes.Group) ([]sdk.AccAddress, error)
	GetPenalizedMembersExpiredSigning(ctx sdk.Context, signing tsstypes.Signing) ([]sdk.AccAddress, error)

	GetSigning(ctx sdk.Context, signingID tss.SigningID) (tsstypes.Signing, error)
	HandleSigningContent(ctx sdk.Context, content tsstypes.Content) ([]byte, error)
	GetSigningResult(ctx sdk.Context, signingID tss.SigningID) (*tsstypes.SigningResult, error)
}
