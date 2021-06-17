package yoda

import (
	"github.com/GeoDB-Limited/odin-core/yoda/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/GeoDB-Limited/odin-core/x/oracle/types"
)

// Constant used to estimate gas price of reports transaction.
const (
	// cosmos
	baseFixedGas        = uint64(37764)
	baseTransactionSize = uint64(200)
	txCostPerByte       = uint64(5) // Using DefaultTxSizeCostPerByte of Odin

	readingBaseCost = uint64(1000)
	writingBaseCost = uint64(2000)

	readingCostPerByte = uint64(3)
	writingCostPerByte = uint64(30)

	payingFeeCost = uint64(16500)

	// odin
	baseReportCost    = uint64(4024)
	addingPendingCost = uint64(4500)

	baseDecCoinSize = uint64(64) // approximate value

	baseRequestSize = uint64(32)
	addressSize     = uint64(20)

	baseRawRequestSize = uint64(16)
)

func estimateTxSize(msgs []sdk.Msg) uint64 {
	// base tx + reports
	size := baseTransactionSize

	for _, msg := range msgs {
		msg, ok := msg.(*types.MsgReportData)
		if !ok {
			panic(sdkerrors.Wrap(errors.ErrNotReportedDataMsg, "failed to estimate tx size"))
		}

		ser := cdc.MustMarshalBinaryBare(msg)
		size += uint64(len(ser))
	}

	return size
}

func estimateStoringReportCost(msg sdk.Msg) uint64 {
	cost := writingBaseCost
	cost += uint64(len(cdc.MustMarshalBinaryBare(msg))) * writingCostPerByte

	return cost
}

func estimateReadingRequestCost(f FeeEstimationData) uint64 {
	cost := readingBaseCost

	size := baseRequestSize
	size += uint64(len(f.callData))
	size += uint64(f.askCount) * addressSize
	size += uint64(len(f.clientID))

	for _, r := range f.rawRequests {
		size += baseRawRequestSize + uint64(len(r.calldata))
	}

	cost += size * readingCostPerByte

	return cost
}

func estimateReadingDataSourceCost(f FeeEstimationData) uint64 {
	cost := readingBaseCost

	dataSourceMap := make(map[types.ExternalID]types.DataSource)
	for _, rawRequest := range f.rawRequests {
		dataSourceMap[rawRequest.externalID] = rawRequest.dataSource
	}

	for _, rawRep := range f.reports {
		cost += uint64(len(cdc.MustMarshalBinaryBare(dataSourceMap[rawRep.ExternalID]))) * readingCostPerByte
	}

	return cost
}

func estimateReportHandleCost(msg sdk.Msg, f FeeEstimationData) uint64 {
	cost := baseReportCost

	// read request twice
	cost += 2 * estimateReadingRequestCost(f)

	// read oracle params
	cost += readingBaseCost + readingCostPerByte*baseDecCoinSize

	// write reward
	writeRewardCost := writingBaseCost + addressSize*writingCostPerByte + baseDecCoinSize*writingCostPerByte

	// read reward
	readRewardCost := readingBaseCost + addressSize*readingCostPerByte + baseDecCoinSize*readingCostPerByte

	// store reward for every provider
	storeRewardCost := writeRewardCost + 2*readRewardCost

	cost += storeRewardCost * uint64(len(f.reports))

	// we need to read data sources
	cost += estimateReadingDataSourceCost(f)

	// write report once
	cost += estimateStoringReportCost(msg)

	// count report
	countingPerReportCost := 30 + readingCostPerByte*uint64(len(cdc.MustMarshalBinaryBare(msg)))

	// reach min count and have to update pending list
	costWhenReachMinCount := countingPerReportCost*uint64(f.minCount+1) + addingPendingCost

	// reach ask count but don't have to update pending list
	costWhenReachAskCount := countingPerReportCost * uint64(f.askCount+1)

	if costWhenReachMinCount > costWhenReachAskCount {
		cost += costWhenReachMinCount
	} else {
		cost += costWhenReachAskCount
	}

	return cost
}

func estimateGas(c *Context, msgs []sdk.Msg, feeEstimations []FeeEstimationData) uint64 {
	gas := baseFixedGas

	txSize := estimateTxSize(msgs)
	gas += txCostPerByte * txSize

	// process paying fee
	if len(c.gasPrices) > 0 {
		gas += payingFeeCost
	}

	for i := range msgs {
		gas += estimateReportHandleCost(msgs[i], feeEstimations[i])
	}

	return gas
}
