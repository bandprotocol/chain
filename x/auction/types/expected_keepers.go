package types

import (
	coinswaptypes "github.com/GeoDB-Limited/odin-core/x/coinswap/types"
	oracletypes "github.com/GeoDB-Limited/odin-core/x/oracle/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

// AccountKeeper defines the expected account Keeper.
type AccountKeeper interface {
	GetModuleAddress(name string) sdk.AccAddress
	SetModuleAccount(sdk.Context, types.ModuleAccountI)
	GetModuleAccount(ctx sdk.Context, moduleName string) types.ModuleAccountI
}

// OracleKeeper defines the expected oracle Keeper.
type OracleKeeper interface {
	GetOraclePool(ctx sdk.Context) (oraclePool oracletypes.OraclePool)
}

// CoinswapKeeper defines the expected coinswap Keeper.
type CoinswapKeeper interface {
	Convert(ctx sdk.Context, amt sdk.Coin, pair coinswaptypes.Exchange) (sdk.Coin, error)
	Exchange(ctx sdk.Context, initialAmt sdk.Coin, convertedAmt sdk.Coin, requester sdk.AccAddress) error
}

// BankKeeper defines the expected bank Keeper.
type BankKeeper interface {
	BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
	MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
}
