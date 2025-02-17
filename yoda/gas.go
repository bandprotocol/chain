package yoda

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/x/oracle/types"
)

// Constant used to estimate gas price of reports transaction.
const (
	// Cosmos default gas
	readFlatGas     = 1000
	readGasPerByte  = 3
	writeFlatGas    = 2000
	writeGasPerByte = 30
	iterateFlatGas  = 30
	hasFlatGas      = 1000

	// Request components
	baseRequestSize    = uint64(170)
	addressSize        = uint64(52)
	baseRawRequestSize = uint64(16)

	// Auth's ante handlers keepers operations
	authParamsByteLength           = 22
	accountByteLength              = 176
	accountWithoutPubKeyByteLength = 103

	readParamGas                   = readFlatGas*5 + authParamsByteLength*readGasPerByte
	readAccountGas                 = readFlatGas + accountByteLength*readGasPerByte
	readAccountWithoutPublicKeyGas = readFlatGas + accountWithoutPubKeyByteLength*readGasPerByte
	writeAccountGas                = writeFlatGas + accountByteLength*writeGasPerByte

	// Auth's ante handlers procedures
	baseAuthAnteGas              = readParamGas*4 + readAccountGas*4 + writeAccountGas + signatureVerificationGasCost + readAccountWithoutPublicKeyGas + writeAccountGas
	payingFeeGasCost             = uint64(19834)
	baseTransactionSize          = uint64(253)
	txCostPerByte                = uint64(5)    // Using DefaultTxSizeCostPerByte of BandChain
	signatureVerificationGasCost = uint64(1000) // for secp256k1 signature, which more than ed21559

	// Report Data byte lengths
	pendingRequestIDByteLength   = 9
	requestIDByteLength          = 11
	pendingResolveListByteLength = 137 // The list have 15 request IDs

	// Report Data handlers
	baseReportDataHandlerGas = hasFlatGas*3 + readFlatGas*3 + requestIDByteLength*readGasPerByte + writeFlatGas
	readPendingListGas       = pendingResolveListByteLength*readGasPerByte + readFlatGas
	writePendingListGas      = (pendingResolveListByteLength+pendingRequestIDByteLength)*writeGasPerByte + writeFlatGas
)

func getTxByteLength(cdc codec.Codec, msgs []sdk.Msg) uint64 {
	// base tx + reports
	size := baseTransactionSize

	for _, msg := range msgs {
		msg, ok := msg.(*types.MsgReportData)
		if !ok {
			panic("Don't support non-report data message")
		}

		ser := cdc.MustMarshal(msg)
		size += uint64(len(ser))
	}

	return size
}

func getRequestByteLength(f FeeEstimationData) uint64 {
	size := baseRequestSize
	size += uint64(len(f.callData))
	size += uint64(f.askCount) * addressSize
	size += uint64(len(f.clientID))

	for _, r := range f.rawRequests {
		size += baseRawRequestSize + uint64(len(r.calldata))
	}

	return size
}

func getReportByteLength(cdc codec.Codec, msg *types.MsgReportData) uint64 {
	report := types.NewReport(
		sdk.ValAddress(msg.Validator),
		true,
		msg.RawReports,
	)
	return uint64(len(cdc.MustMarshal(&report)))
}

func estimateReportHandlerGas(cdc codec.Codec, msg *types.MsgReportData, f FeeEstimationData) uint64 {
	reportByteLength := getReportByteLength(cdc, msg)
	requestByteLength := getRequestByteLength(f)

	cost := 2*readGasPerByte*requestByteLength + writeGasPerByte*reportByteLength + baseReportDataHandlerGas

	costWhenReachAskCountFirst := (reportByteLength*readGasPerByte + iterateFlatGas) * (uint64(f.askCount) + 1)
	costWhenReachMinCountFirst := (reportByteLength*readGasPerByte+iterateFlatGas)*(uint64(f.minCount)+1) + readPendingListGas + writePendingListGas

	if costWhenReachMinCountFirst > costWhenReachAskCountFirst {
		cost += costWhenReachMinCountFirst
	} else {
		cost += costWhenReachAskCountFirst
	}

	return cost
}

func estimateAuthAnteHandlerGas(c *Context, msgs []sdk.Msg) uint64 {
	gas := baseAuthAnteGas

	txByteLength := getTxByteLength(c.encodingConfig.Codec, msgs)
	gas += txCostPerByte * txByteLength

	if len(c.gasPrices) > 0 {
		gas += payingFeeGasCost
	}

	return gas
}

func estimateGas(c *Context, l *Logger, msgs []sdk.Msg, feeEstimations []FeeEstimationData) uint64 {
	gas := estimateAuthAnteHandlerGas(c, msgs)

	for i, msg := range msgs {
		msg, ok := msg.(*types.MsgReportData)
		if !ok {
			panic("Don't support non-report data message")
		}
		gas += estimateReportHandlerGas(c.encodingConfig.Codec, msg, feeEstimations[i])
	}

	l.Debug(":fuel_pump: Estimated gas is %d", gas)

	return gas
}
