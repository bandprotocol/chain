package oraclekeeper

import (
	oracletypes "github.com/GeoDB-Limited/odin-core/x/oracle/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type FeeCollector interface {
	Collect(sdk.Context, sdk.Coins) error
	Collected() sdk.Coins
}

type RewardCollector interface {
	Collect(sdk.Context, sdk.Coins, sdk.AccAddress) error
	CalculateReward([]byte, sdk.Coins) sdk.Coins
	Collected() sdk.Coins
}

// CollectFee subtract fee from fee payer and send them to treasury
func (k Keeper) CollectFee(
	ctx sdk.Context, payer sdk.AccAddress, feeLimit sdk.Coins, askCount uint64, rawRequests []oracletypes.RawRequest,
) (sdk.Coins, error) {

	collector := newFeeCollector(k, feeLimit, payer)

	for _, r := range rawRequests {

		ds := k.MustGetDataSource(ctx, r.DataSourceID)
		if ds.Fee.Empty() {
			continue
		}

		fee := sdk.NewCoins()
		for _, c := range ds.Fee {
			c.Amount = c.Amount.Mul(sdk.NewInt(int64(askCount)))
			fee = fee.Add(c)
		}

		if err := collector.Collect(ctx, fee); err != nil {
			return nil, err
		}

		accumulatedPaymentsForData := k.GetAccumulatedPaymentsForData(ctx)
		accumulatedAmount := accumulatedPaymentsForData.AccumulatedAmount.Add(fee...)
		auctionThreshold := k.auctionKeeper.GetThreshold(ctx)
		for accumulatedAmount.IsAllGTE(auctionThreshold) {
			if err := k.auctionKeeper.BuyCoins(ctx); err != nil {
				return nil, sdkerrors.Wrapf(err, "failed to process auction")
			}
			accumulatedAmount = accumulatedAmount.Sub(auctionThreshold)
		}
		accumulatedPaymentsForData.AccumulatedAmount = accumulatedAmount
		k.SetAccumulatedPaymentsForData(ctx, accumulatedPaymentsForData)
	}

	return collector.Collected(), nil
}

// CollectReward subtract reward from fee pool and sends it to the data providers for reporting data
func (k Keeper) CollectReward(
	ctx sdk.Context, rawReports []oracletypes.RawReport, rawRequests []oracletypes.RawRequest,
) (sdk.Coins, error) {
	collector := newRewardCollector(k, k.bankKeeper)
	oracleParams := k.GetParams(ctx)

	rawReportsMap := make(map[oracletypes.ExternalID]oracletypes.RawReport)
	for _, rawRep := range rawReports {
		rawReportsMap[rawRep.ExternalID] = rawRep
	}

	accumulatedDataProvidersRewards := k.GetAccumulatedDataProvidersRewards(ctx)
	accumulatedAmount := accumulatedDataProvidersRewards.AccumulatedAmount
	currentRewardPerByte := accumulatedDataProvidersRewards.CurrentRewardPerByte

	for _, rawReq := range rawRequests {
		rawRep, ok := rawReportsMap[rawReq.GetExternalID()]
		if !ok {
			// this request had no report
			continue
		}

		ds := k.MustGetDataSource(ctx, rawReq.GetDataSourceID())
		dsOwnerAddr, err := sdk.AccAddressFromBech32(ds.Owner)
		if err != nil {
			return nil, sdkerrors.Wrapf(err, "parsing data source owner address: %s", dsOwnerAddr)
		}

		var reward sdk.Coins
		for {
			reward = collector.CalculateReward(rawRep.Data, currentRewardPerByte)
			if reward.Add(accumulatedAmount...).IsAllLT(oracleParams.DataProviderRewardThreshold.Amount) {
				break
			}

			currentRewardPerByte, _ = sdk.NewDecCoinsFromCoins(currentRewardPerByte...).MulDec(
				sdk.NewDec(1).Sub(oracleParams.RewardDecreasingFraction),
			).TruncateDecimal()
		}

		accumulatedAmount = accumulatedAmount.Add(reward...)
		err = collector.Collect(ctx, reward, dsOwnerAddr)
		if err != nil {
			return nil, err
		}
	}

	k.SetAccumulatedDataProvidersRewards(
		ctx,
		oracletypes.NewDataProvidersAccumulatedRewards(currentRewardPerByte, accumulatedAmount),
	)

	return collector.Collected(), nil
}
