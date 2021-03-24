package keeper

import (
	"github.com/bandprotocol/chain/x/oracle/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// FeeCollector define fee collector
type FeeCollector interface {
	Collect(sdk.Context, sdk.Coins, sdk.AccAddress) error
	Collected() sdk.Coins
}

type feeCollector struct {
	bankKeeper types.BankKeeper
	payer      sdk.AccAddress
	collected  map[string]sdk.Coin
	limit      map[string]sdk.Coin
}

func (coll *feeCollector) Collect(ctx sdk.Context, coins sdk.Coins, treasury sdk.AccAddress) error {
	for _, c := range coins {
		if _, found := coll.collected[c.Denom]; !found {
			coll.collected[c.Denom] = sdk.NewCoin(c.Denom, sdk.ZeroInt())
		}

		coll.collected[c.Denom] = coll.collected[c.Denom].Add(c)
		collected := coll.collected[c.Denom]

		limit := sdk.NewCoin(c.Denom, sdk.ZeroInt())
		if cLimit, found := coll.limit[c.Denom]; found {
			limit = cLimit
		}

		if collected.IsGTE(limit) && !collected.Equal(limit) { // Need GT but have no
			return sdkerrors.Wrapf(types.ErrNotEnoughFee, "require: %s, max: %s", collected.String(), limit.String())
		}
	}

	// Actual send coins
	return coll.bankKeeper.SendCoins(ctx, coll.payer, treasury, coins)
}

func (coll *feeCollector) Collected() sdk.Coins {
	coins := sdk.NewCoins()

	for _, c := range coll.collected {
		coins = append(coins, c)
	}

	return coins.Sort()
}

func newFeeCollector(bankKeeper types.BankKeeper, feeLimit sdk.Coins, payer sdk.AccAddress) FeeCollector {
	limit := map[string]sdk.Coin{}

	// Coins is sorted and there are no duplicated denom
	for _, c := range feeLimit {
		limit[c.Denom] = c
	}

	return &feeCollector{
		bankKeeper: bankKeeper,
		payer:      payer,
		collected:  map[string]sdk.Coin{},
		limit:      limit,
	}
}
