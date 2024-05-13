package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/feeds/types"
)

// ValidateSubmitPricesRequest validates price submission.
func (k Keeper) ValidateSubmitPricesRequest(
	ctx sdk.Context,
	blockTime int64,
	req *types.MsgSubmitPrices,
	val sdk.ValAddress,
) error {
	isValid := k.IsBondedValidator(ctx, val)
	if !isValid {
		return types.ErrNotBondedValidator
	}

	status := k.oracleKeeper.GetValidatorStatus(ctx, val)
	if !status.IsActive {
		return types.ErrOracleStatusNotActive.Wrapf("val: %s", val.String())
	}

	if types.AbsInt64(req.Timestamp-blockTime) > k.GetParams(ctx).AllowableBlockTimeDiscrepancy {
		return types.ErrInvalidTimestamp.Wrapf(
			"block_time: %d, timestamp: %d",
			blockTime,
			req.Timestamp,
		)
	}
	return nil
}

// NewValidatorPrice creates new ValidatorPrice.
func (k Keeper) NewValidatorPrice(
	ctx sdk.Context,
	blockTime int64,
	price types.SubmitPrice,
	val sdk.ValAddress,
	cooldownTime int64,
) (types.ValidatorPrice, error) {
	f, err := k.GetFeed(ctx, price.SignalID)
	if err != nil {
		return types.ValidatorPrice{}, err
	}

	// check if price is not too fast
	priceVal, err := k.GetValidatorPrice(ctx, price.SignalID, val)
	if err == nil && blockTime < priceVal.Timestamp+cooldownTime {
		return types.ValidatorPrice{}, types.ErrPriceTooFast.Wrapf(
			"signal_id: %s, old: %d, new: %d, interval: %d",
			price.SignalID,
			priceVal.Timestamp,
			blockTime,
			f.Interval,
		)
	}

	return types.ValidatorPrice{
		PriceStatus: price.PriceStatus,
		Validator:   val.String(),
		SignalID:    price.SignalID,
		Price:       price.Price,
		Timestamp:   blockTime,
	}, nil
}
