package types

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	bandtsstypes "github.com/bandprotocol/chain/v3/x/bandtss/types"
	feedstypes "github.com/bandprotocol/chain/v3/x/feeds/types"
	tsstypes "github.com/bandprotocol/chain/v3/x/tss/types"
)

// AccountKeeper defines the expected account keeper (noalias)
type AccountKeeper interface {
	GetModuleAddress(name string) sdk.AccAddress
	GetModuleAccount(ctx context.Context, name string) sdk.ModuleAccountI
	SetModuleAccount(ctx context.Context, moduleAccount sdk.ModuleAccountI)

	GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
	NewAccount(ctx context.Context, account sdk.AccountI) sdk.AccountI
	SetAccount(ctx context.Context, account sdk.AccountI)
}

type BankKeeper interface {
	GetAllBalances(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	SpendableCoins(ctx context.Context, addr sdk.AccAddress) sdk.Coins

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
}

type FeedsKeeper interface {
	GetAllPrices(ctx sdk.Context) (prices []feedstypes.Price)
	GetPrices(ctx sdk.Context, signalIDs []string) (prices []feedstypes.Price)
}

type BandtssKeeper interface {
	CreateTunnelSigningRequest(
		ctx sdk.Context,
		tunnelID uint64,
		destinationChainID string,
		destinationContractAddr string,
		content tsstypes.Content,
		sender sdk.AccAddress,
		feeLimit sdk.Coins,
	) (bandtsstypes.SigningID, error)
	GetSigningFee(ctx sdk.Context) (sdk.Coins, error)
}
