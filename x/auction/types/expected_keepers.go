package types

import (
	coinswaptypes "github.com/GeoDB-Limited/odin-core/x/coinswap/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

type AccountKeeper interface {
	GetModuleAddress(name string) sdk.AccAddress
	SetModuleAccount(sdk.Context, types.ModuleAccountI)
	GetModuleAccount(ctx sdk.Context, moduleName string) types.ModuleAccountI
}

type CoinswapKeeper interface {
	SetParams(ctx sdk.Context, value coinswaptypes.Params)
	GetParams(ctx sdk.Context) (params coinswaptypes.Params)
	ExchangeDenom(ctx sdk.Context, fromDenom, toDenom string, amt sdk.Coin, requester sdk.AccAddress) error
}
