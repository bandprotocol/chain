package types

import (
	oracletypes "github.com/GeoDB-Limited/odin-core/x/oracle/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank/exported"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
)

// AccountKeeper defines the expected account keeper.
type AccountKeeper interface {
	GetModuleAccount(ctx sdk.Context, name string) authtypes.ModuleAccountI
}

// BankKeeper defines the expected supply Keeper.
type BankKeeper interface {
	GetSupply(ctx sdk.Context) (supply exported.SupplyI)
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
	MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
}

type DistrKeeper interface {
	GetFeePool(ctx sdk.Context) (feePool distrtypes.FeePool)
	SetFeePool(ctx sdk.Context, feePool distrtypes.FeePool)
	FundCommunityPool(ctx sdk.Context, amount sdk.Coins, sender sdk.AccAddress) error
}

type OracleKeeper interface {
	GetOraclePool(ctx sdk.Context) (oraclePool oracletypes.OraclePool)
	SetOraclePool(ctx sdk.Context, oraclePool oracletypes.OraclePool)
	WithdrawOraclePool(ctx sdk.Context, amount sdk.Coins, recipient sdk.AccAddress) error
	FundOraclePool(ctx sdk.Context, amount sdk.Coins, sender sdk.AccAddress) error
}
