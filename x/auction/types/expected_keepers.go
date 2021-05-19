package types

import (
	coinswaptypes "github.com/GeoDB-Limited/odin-core/x/coinswap/types"
	oracletypes "github.com/GeoDB-Limited/odin-core/x/oracle/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

type AccountKeeper interface {
	GetModuleAddress(name string) sdk.AccAddress
	SetModuleAccount(sdk.Context, types.ModuleAccountI)
	GetModuleAccount(ctx sdk.Context, moduleName string) types.ModuleAccountI
}

type OracleKeeper interface {
	GetOraclePool(ctx sdk.Context) (oraclePool oracletypes.OraclePool)
}

type CoinswapKeeper interface {
	Exchange(ctx sdk.Context, amt sdk.Coin, pair coinswaptypes.Exchange, requester sdk.AccAddress) error
}
