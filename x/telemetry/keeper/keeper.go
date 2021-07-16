package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"sort"
)

type Keeper struct {
	cdc        codec.BinaryMarshaler
	bankKeeper bankkeeper.ViewKeeper
}

func NewKeeper(cdc codec.BinaryMarshaler, bk bankkeeper.ViewKeeper) Keeper {
	return Keeper{
		cdc:        cdc,
		bankKeeper: bk,
	}
}

func (k Keeper) GetPaginatedBalances(ctx sdk.Context, denom string, desc bool, pagination *query.PageRequest) ([]banktypes.Balance, uint64) {
	balances := k.bankKeeper.GetAccountsBalances(ctx)

	sort.Slice(balances, func(i, j int) bool {
		if desc {
			return balances[j].GetCoins().AmountOf(denom).LT(balances[i].GetCoins().AmountOf(denom))
		}
		return balances[i].GetCoins().AmountOf(denom).LT(balances[j].GetCoins().AmountOf(denom))
	})

	if pagination.GetOffset() >= uint64(len(balances)) {
		return []banktypes.Balance{}, 0
	}

	maxLimit := pagination.GetLimit()
	if pagination.GetOffset()+pagination.GetLimit() >= uint64(len(balances)) {
		maxLimit = uint64(len(balances)) - pagination.GetOffset()
	}

	return balances[pagination.GetOffset() : pagination.GetOffset()+maxLimit], uint64(len(balances))
}
