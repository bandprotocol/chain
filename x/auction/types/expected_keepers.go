package types

import (
	coinswaptypes "github.com/GeoDB-Limited/odin-core/x/coinswap/types"
	oracletypes "github.com/GeoDB-Limited/odin-core/x/oracle/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// OracleKeeper defines the expected oracle Keeper.
type OracleKeeper interface {
	GetAccumulatedPaymentsForData(ctx sdk.Context) (payments oracletypes.AccumulatedPaymentsForData)
}

// CoinswapKeeper defines the expected coinswap Keeper.
type CoinswapKeeper interface {
	AddExchangeRate(ctx sdk.Context, exchange coinswaptypes.Exchange) error
	RemoveExchangeRate(ctx sdk.Context, exchange coinswaptypes.Exchange) error
}
