package types

import (
	"context"
	"time"

	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	bandtsstypes "github.com/bandprotocol/chain/v3/x/bandtss/types"
	tsstypes "github.com/bandprotocol/chain/v3/x/tss/types"
)

// AccountKeeper defines the expected account keeper.
type AccountKeeper interface {
	GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
	GetModuleAccount(ctx context.Context, moduleName string) sdk.ModuleAccountI
}

// BankKeeper defines the expected bank keeper.
type BankKeeper interface {
	GetAllBalances(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	SendCoinsFromModuleToModule(ctx context.Context, senderModule, recipientModule string, amt sdk.Coins) error
	SendCoins(ctx context.Context, from, to sdk.AccAddress, amt sdk.Coins) error
	SpendableCoins(ctx context.Context, addr sdk.AccAddress) sdk.Coins
}

// StakingKeeper defines the expected staking keeper.
type StakingKeeper interface {
	ValidatorByConsAddr(context.Context, sdk.ConsAddress) (stakingtypes.ValidatorI, error)
	IterateBondedValidatorsByPower(context.Context,
		func(index int64, validator stakingtypes.ValidatorI) (stop bool)) error
	Validator(context.Context, sdk.ValAddress) (stakingtypes.ValidatorI, error)
}

// DistrKeeper defines the expected distribution keeper.
type DistrKeeper interface {
	GetCommunityTax(ctx context.Context) (math.LegacyDec, error)
	AllocateTokensToValidator(ctx context.Context, val stakingtypes.ValidatorI, tokens sdk.DecCoins) error
	FundCommunityPool(ctx context.Context, amount sdk.Coins, sender sdk.AccAddress) error
}

// PortKeeper defines the expected IBC port keeper
type PortKeeper interface {
	BindPort(ctx sdk.Context, portID string) *capabilitytypes.Capability
}

// AuthzKeeper defines the expected authz keeper. for query and testing only don't use to create/remove grant on deliver tx
type AuthzKeeper interface {
	DispatchActions(ctx context.Context, grantee sdk.AccAddress, msgs []sdk.Msg) ([][]byte, error)
	GetAuthorization(
		ctx context.Context,
		grantee, granter sdk.AccAddress,
		msgType string,
	) (authz.Authorization, *time.Time)
	GetAuthorizations(ctx context.Context, grantee, granter sdk.AccAddress) ([]authz.Authorization, error)
	SaveGrant(
		ctx context.Context,
		grantee, granter sdk.AccAddress,
		authorization authz.Authorization,
		expiration *time.Time,
	) error
	DeleteGrant(ctx context.Context, grantee, granter sdk.AccAddress, msgType string) error
	GranterGrants(ctx context.Context, req *authz.QueryGranterGrantsRequest) (*authz.QueryGranterGrantsResponse, error)
}

// RollingseedKeeper defines the expected rollingseed keeper
type RollingseedKeeper interface {
	GetRollingSeed(ctx sdk.Context) []byte
}

// BandtssKeeper defines the expected bandtss keeper.
type BandtssKeeper interface {
	CreateDirectSigningRequest(
		ctx sdk.Context,
		content tsstypes.Content,
		memo string,
		sender sdk.AccAddress,
		feeLimit sdk.Coins,
	) (bandtsstypes.SigningID, error)
}
