package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/bandprotocol/chain/v2/x/oracle/types"
)

// FeeCollector define fee collector
type FeeCollector interface {
	Collect(sdk.Context, sdk.Coins, sdk.AccAddress) error
	Collected() sdk.Coins
}

type feeCollector struct {
	bankKeeper types.BankKeeper
	payer      sdk.AccAddress
	collected  sdk.Coins
	limit      sdk.Coins
}

func newFeeCollector(bankKeeper types.BankKeeper, feeLimit sdk.Coins, payer sdk.AccAddress) FeeCollector {

	return &feeCollector{
		bankKeeper: bankKeeper,
		payer:      payer,
		collected:  sdk.NewCoins(),
		limit:      feeLimit,
	}
}

func (coll *feeCollector) Collect(ctx sdk.Context, coins sdk.Coins, treasury sdk.AccAddress) error {
	coll.collected = coll.collected.Add(coins...)

	// If found any collected coin that exceed limit then return error
	for _, c := range coll.collected {
		limitAmt := coll.limit.AmountOf(c.Denom)
		if c.Amount.GT(limitAmt) {
			return sdkerrors.Wrapf(types.ErrNotEnoughFee, "require: %s, max: %s%s", c.String(), limitAmt.String(), c.Denom)
		}
	}

	// Actual send coins
	return coll.bankKeeper.SendCoins(ctx, coll.payer, treasury, coins)
}

func (coll *feeCollector) Collected() sdk.Coins {
	return coll.collected
}
