package yoda

import (
	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/x/oracle/types"
)

// Constant used to estimate gas price of reports transaction.
const (

	// Request components
	baseRequestSize    = uint64(32)
	addressSize        = uint64(20)
	baseRawRequestSize = uint64(16)

	// auth ante handlers procedures
	baseAuthAnteGas     = uint64(34656)
	payingFeeGasCost    = uint64(19834)
	baseTransactionSize = uint64(253)
	txCostPerByte       = uint64(5) // Using DefaultTxSizeCostPerByte of BandChain

	readParamGas                   = uint64(5066)
	readAccountGas                 = uint64(1528)
	readAccountWithoutPublicKeyGas = uint64(1309)
	setAccountGas                  = uint64(7280)
)

func getTxByteLength(msgs []sdk.Msg) uint64 {
	// base tx + reports
	size := baseTransactionSize

	for _, msg := range msgs {
		msg, ok := msg.(*types.MsgReportData)
		if !ok {
			panic("Don't support non-report data message")
		}

		ser := cdc.MustMarshalBinaryBare(msg)
		size += uint64(len(ser))
	}

	return size
}

func getRequestMsgByteLength(f FeeEstimationData) uint64 {
	size := baseRequestSize
	size += uint64(len(f.callData))
	size += uint64(f.askCount) * addressSize
	size += uint64(len(f.clientID))

	for _, r := range f.rawRequests {
		size += baseRawRequestSize + uint64(len(r.calldata))
	}

	return size
}

func getReportMsgByteLength(msg sdk.Msg) uint64 {
	return uint64(len(cdc.MustMarshalBinaryBare(msg)))
}

func estimateReportHandlerGas(msg sdk.Msg, f FeeEstimationData) uint64 {
	reportByteLength := getReportMsgByteLength(msg)
	requestByteLength := getRequestMsgByteLength(f)

	cost := 6*requestByteLength + 33*reportByteLength + 8041

	costWhenReachAskCountFirst := 3*reportByteLength*uint64(f.askCount) + 30*uint64(f.askCount)
	costWhenReachMinCountFirst := 3*reportByteLength*uint64(f.minCount) + 30*uint64(f.minCount) + 7791

	if costWhenReachMinCountFirst > costWhenReachAskCountFirst {
		cost += costWhenReachMinCountFirst
	} else {
		cost += costWhenReachAskCountFirst
	}

	return cost
}

func estimateAuthAnteHandlerGas(c *Context, msgs []sdk.Msg, acc client.Account) uint64 {
	gas := uint64(baseAuthAnteGas)

	if acc == nil || acc.GetPubKey() == nil {
		gas += readAccountWithoutPublicKeyGas + setAccountGas
	} else {
		gas += readAccountGas
	}

	txByteLength := getTxByteLength(msgs)
	gas += txCostPerByte * txByteLength

	if len(c.gasPrices) > 0 {
		gas += payingFeeGasCost
	}

	return gas
}

func estimateGas(c *Context, msgs []sdk.Msg, feeEstimations []FeeEstimationData, acc client.Account, l *Logger) uint64 {
	gas := estimateAuthAnteHandlerGas(c, msgs, acc)

	for i := range msgs {
		gas += estimateReportHandlerGas(msgs[i], feeEstimations[i])
	}

	l.Info(":fuel_pump: Estimated gas is %d", gas)

	return gas
}
