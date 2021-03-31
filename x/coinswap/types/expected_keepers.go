package types

import (
	"github.com/GeoDB-Limited/odincore/chain/x/oracle"
	sdk "github.com/cosmos/cosmos-sdk/types"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/supply/exported"
	supplyexported "github.com/cosmos/cosmos-sdk/x/supply/exported"
)

// SupplyKeeper defines the expected supply Keeper.
type SupplyKeeper interface {
	GetSupply(ctx sdk.Context) (supply exported.SupplyI)
	GetModuleAccount(ctx sdk.Context, name string) supplyexported.ModuleAccountI
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
	MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
}

type DistrKeeper interface {
	GetFeePool(ctx sdk.Context) (feePool distr.FeePool)
	SetFeePool(ctx sdk.Context, feePool distr.FeePool)
	FundCommunityPool(ctx sdk.Context, amount sdk.Coins, sender sdk.AccAddress) error
}

type OracleKeeper interface {
	GetOraclePool(ctx sdk.Context) (oraclePool oracle.OraclePool)
	SetOraclePool(ctx sdk.Context, oraclePool oracle.OraclePool)
	WithdrawOraclePool(ctx sdk.Context, amount sdk.Coins, recipient sdk.AccAddress) error
	FundOraclePool(ctx sdk.Context, amount sdk.Coins, sender sdk.AccAddress) error
}
