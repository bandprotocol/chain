package types

import (
	"context"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/pkg/tss"
	tsstypes "github.com/bandprotocol/chain/v3/x/tss/types"
)

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
	FundCommunityPool(ctx context.Context, amount sdk.Coins, sender sdk.AccAddress) error
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
