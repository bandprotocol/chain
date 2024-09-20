package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	bandtsstypes "github.com/bandprotocol/chain/v2/x/bandtss/types"
	feedstypes "github.com/bandprotocol/chain/v2/x/feeds/types"
	tsstypes "github.com/bandprotocol/chain/v2/x/tss/types"
)

// AccountKeeper defines the expected account keeper (noalias)
type AccountKeeper interface {
	GetModuleAddress(name string) sdk.AccAddress
	GetModuleAccount(ctx sdk.Context, name string) authtypes.ModuleAccountI
	SetModuleAccount(sdk.Context, authtypes.ModuleAccountI)

	GetAccount(ctx sdk.Context, addr sdk.AccAddress) authtypes.AccountI
	NewAccount(ctx sdk.Context, account authtypes.AccountI) authtypes.AccountI
	SetAccount(ctx sdk.Context, account authtypes.AccountI)
}

type BankKeeper interface {
	GetAllBalances(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins

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
}

type FeedsKeeper interface {
	GetCurrentPrices(ctx sdk.Context) (prices []feedstypes.Price)
}

type BandtssKeeper interface {
	CreateDirectSigningRequest(
		ctx sdk.Context,
		content tsstypes.Content,
		memo string,
		sender sdk.AccAddress,
		feeLimit sdk.Coins,
	) (bandtsstypes.SigningID, error)
}
