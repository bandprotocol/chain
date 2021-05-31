package yoda

import (
	"encoding/hex"
	"math"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/x/oracle/types"
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
	baseRequestSize    = uint64(32)
	addressSize        = uint64(20)
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
	baseAuthAnteGas              = readParamGas*4 + readAccountGas*4 + writeAccountGas + signatureVerificationGasCost
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
	var requestedValidator []sdk.ValAddress
	for i := 0; i < int(f.askCount); i++ {
		addr := []byte{200, 199, 198, 197, 196, 195, 194, 193, 192, 191, 190, 189, 188, 187, 176, 165, 154, 143, 132, 181}
		requestedValidator = append(requestedValidator, addr)
	}
	var rawRequest []types.RawRequest
	for _, rreq := range f.rawRequests {
		calldata, _ := hex.DecodeString(rreq.calldata)
		rawRequest = append(rawRequest, types.RawRequest{
			ExternalID:   rreq.externalID,
			DataSourceID: rreq.dataSourceID,
			Calldata:     calldata,
		})
	}

	request := types.NewRequest(
		types.OracleScriptID(math.MaxInt64),
		f.callData,
		requestedValidator,
		uint64(f.minCount),
		math.MaxInt64,
		time.Now(),
		f.clientID,
		rawRequest,
		&types.IBCChannel{
			PortId:    "mockPortIDmockPortIDmockPortIDmockPortIDmockPortIDmockPortID",
			ChannelId: "mockChannelIDmockChannelIDmockChannelIDmockChannelIDmockChannelID",
		},
		math.MaxInt64,
	)

	return uint64(len(cdc.MustMarshalBinaryBare(&request)))
}

func getReportMsgByteLength(msg *types.MsgReportData) uint64 {
	report := types.NewReport(
		sdk.ValAddress(msg.Validator),
		true,
		msg.RawReports,
	)
	return uint64(len(cdc.MustMarshalBinaryBare(&report)))
}

func estimateReportHandlerGas(msg *types.MsgReportData, f FeeEstimationData) uint64 {
	reportByteLength := getReportMsgByteLength(msg)
	requestByteLength := getRequestMsgByteLength(f)

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

func estimateAuthAnteHandlerGas(c *Context, msgs []sdk.Msg, acc client.Account) uint64 {
	gas := uint64(baseAuthAnteGas)

	if acc.GetPubKey() == nil {
		gas += readAccountWithoutPublicKeyGas + writeAccountGas
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

	for i, msg := range msgs {
		msg, ok := msg.(*types.MsgReportData)
		if !ok {
			panic("Don't support non-report data message")
		}
		gas += estimateReportHandlerGas(msg, feeEstimations[i])
	}

	l.Debug(":fuel_pump: Estimated gas is %d", gas)

	return gas
}
