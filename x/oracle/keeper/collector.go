package oraclekeeper

import (
	"github.com/GeoDB-Limited/odin-core/x/oracle/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type FeeCollector interface {
	Collect(sdk.Context, sdk.Coins, sdk.AccAddress) error
	Collected() sdk.Coins
}

type RewardCollector interface {
	Collect(sdk.Context, sdk.DecCoins, sdk.AccAddress) error
	CalculateReward([]byte, sdk.DecCoins) sdk.DecCoins
	Collected() sdk.DecCoins
}

// CollectFee subtract fee from fee payer and send them to treasury
func (k Keeper) CollectFee(ctx sdk.Context, payer sdk.AccAddress, feeLimit sdk.Coins, askCount uint64, rawRequests []types.RawRequest) (sdk.Coins, error) {

	collector := newFeeCollector(k.bankKeeper, feeLimit, payer)

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

		treasury, err := sdk.AccAddressFromBech32(ds.Treasury)
		if err != nil {
			return nil, err
		}

		if err := collector.Collect(ctx, fee, treasury); err != nil {
			return nil, err
		}
	}

	return collector.Collected(), nil
}

// CollectReward subtract reward from fee pool and sends it to the data providers for reporting data
func (k Keeper) CollectReward(ctx sdk.Context, rawReports []types.RawReport, rawRequests []types.RawRequest) (sdk.DecCoins, error) {

	collector := newRewardCollector(k, k.bankKeeper)

	rawReportsMap := make(map[types.ExternalID]types.RawReport)
	for _, rawRep := range rawReports {
		rawReportsMap[rawRep.ExternalID] = rawRep
	}

	dataProviderRewardPerByte := k.GetDataProviderRewardPerByteParam(ctx)
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

		err = collector.Collect(ctx, collector.CalculateReward(rawRep.Data, sdk.NewDecCoins(dataProviderRewardPerByte)), dsOwnerAddr)
		if err != nil {
			return nil, err
		}
	}

	return collector.Collected(), nil
}
