package keeper

import (
	"github.com/bandprotocol/chain/x/oracle/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// FeeCollector define fee collector
type FeeCollector interface {
	Collect(sdk.Context, sdk.Coins, sdk.AccAddress) error
}

type feeCollector struct {
	bankKeeper types.BankKeeper
	payer      sdk.AccAddress
	collected  map[string]sdk.Int
	limit      map[string]sdk.Int
}

func (coll *feeCollector) Collect(ctx sdk.Context, coins sdk.Coins, treasury sdk.AccAddress) error {
	for _, c := range coins {
		if _, found := coll.collected[c.Denom]; !found {
			coll.collected[c.Denom] = sdk.NewInt(0)
		}

		coll.collected[c.Denom] = coll.collected[c.Denom].Add(c.Amount)
		collected := coll.collected[c.Denom]

		limit := sdk.NewInt(0)
		if cLimit, found := coll.limit[c.Denom]; found {
			limit = cLimit
		}

		if collected.GT(limit) {
			return sdkerrors.Wrapf(types.ErrNotEnoughFee, "require: %d, max: %d", collected.Int64(), limit.Int64())
		}
	}

	return coll.bankKeeper.SendCoins(ctx, coll.payer, treasury, coins)
}

// NewFeeCollector create new fee collector
func NewFeeCollector(bankKeeper types.BankKeeper, feeLimit sdk.Coins, payer sdk.AccAddress) FeeCollector {
	limit := map[string]sdk.Int{}

	for _, c := range feeLimit {
		if _, found := limit[c.Denom]; !found {
			limit[c.Denom] = sdk.NewInt(0)
		}

		limit[c.Denom] = limit[c.Denom].Add(c.Amount)
	}

	return &feeCollector{
		bankKeeper: bankKeeper,
		payer:      payer,
		collected:  map[string]sdk.Int{},
		limit:      limit,
	}
}
