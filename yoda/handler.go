package yoda

import (
	"encoding/hex"
	"fmt"
	"github.com/GeoDB-Limited/odin-core/yoda/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"strconv"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/tmhash"

	"github.com/GeoDB-Limited/odin-core/hooks/common"
	"github.com/GeoDB-Limited/odin-core/x/oracle/types"
)

type processingResult struct {
	rawReport types.RawReport
	version   string
	err       error
}

func handleTransaction(ctx *Context, logger *Logger, tx abci.TxResult) {
	logger.Debug(":eyes: Inspecting incoming transaction: %X", tmhash.Sum(tx.Tx))
	if tx.Result.Code != 0 {
		logger.Debug(":alien: Skipping transaction with non-zero code: %d", tx.Result.Code)
		return
	}

	logs, err := sdk.ParseABCILogs(tx.Result.Log)
	if err != nil {
		logger.Error(":cold_sweat: Failed to parse transaction logs with error: %s", ctx, err.Error())
		return
	}

	for _, log := range logs {
		messageType, err := GetEventValue(log, sdk.EventTypeMessage, sdk.AttributeKeyAction)
		if err != nil {
			logger.Error(":cold_sweat: Failed to get message action type with error: %s", ctx, err.Error())
			continue
		}

		if messageType == (types.MsgRequestData{}).Type() {
			go handleRequestLog(ctx, logger, log)
		} else if messageType == (channeltypes.MsgRecvPacket{}).Type() {
			// Try to get request id from packet. If not then return error.
			_, err := GetEventValue(log, types.EventTypeRequest, types.AttributeKeyID)
			if err != nil {
				logger.Debug(":ghost: Skipping non-request packet")
				return
			}
			go handleRequestLog(ctx, logger, log)
		} else {
			logger.Debug(":ghost: Skipping non-{request/packet} type: %s", messageType)
		}
	}
}

func handleRequestLog(ctx *Context, logger *Logger, msgLog sdk.ABCIMessageLog) {
	idStr, err := GetEventValue(msgLog, types.EventTypeRequest, types.AttributeKeyID)
	if err != nil {
		logger.Error(":cold_sweat: Failed to parse request id with error: %s", ctx, err.Error())
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		logger.Error(":cold_sweat: Failed to convert %s to integer with error: %s", ctx, idStr, err.Error())
		return
	}

	logger = logger.With("rid", id)

	// If id is in pending requests list, then skip it.
	if ctx.pendingRequests[types.RequestID(id)] {
		logger.Debug(":eyes: Request is in pending list, then skip")
		return
	}

	// Skip if not related to this validator
	validators := GetEventValues(msgLog, types.EventTypeRequest, types.AttributeKeyValidator)
	hasMe := false
	for _, validator := range validators {
		if validator == ctx.validator.String() {
			hasMe = true
			break
		}
	}

	if !hasMe {
		logger.Debug(":next_track_button: Skip request not related to this validator")
		return
	}

	logger.Info(":delivery_truck: Processing incoming request event")

	reqs, err := GetRawRequests(ctx, logger, msgLog)
	if err != nil {
		logger.Error(":skull: Failed to parse raw requests with error: %s", ctx, err.Error())
	}

	keyIndex := ctx.nextKeyIndex()
	key := ctx.keys[keyIndex]

	reports, execVersions := handleRawRequests(ctx, logger, types.RequestID(id), reqs, key)

	rawAskCount := GetEventValues(msgLog, types.EventTypeRequest, types.AttributeKeyAskCount)
	if len(rawAskCount) != 1 {
		panic(sdkerrors.Wrap(errors.ErrEventValueDoesNotExist, "failed to get ask count"))
	}
	askCount := common.Atoi(rawAskCount[0])

	rawMinCount := GetEventValues(msgLog, types.EventTypeRequest, types.AttributeKeyMinCount)
	if len(rawMinCount) != 1 {
		panic(sdkerrors.Wrap(errors.ErrEventValueDoesNotExist, "failed to get min count"))
	}
	minCount := common.Atoi(rawMinCount[0])

	rawCallData := GetEventValues(msgLog, types.EventTypeRequest, types.AttributeKeyCalldata)
	if len(rawCallData) != 1 {
		panic(sdkerrors.Wrap(errors.ErrEventValueDoesNotExist, "failed to get call data"))
	}
	callData, err := hex.DecodeString(rawCallData[0])
	if err != nil {
		logger.Error(":skull: Fail to parse call data: %s", ctx, err.Error())
	}

	var clientID string
	rawClientID := GetEventValues(msgLog, types.EventTypeRequest, types.AttributeKeyClientID)
	if len(rawClientID) > 0 {
		clientID = rawClientID[0]
	}

	ctx.pendingMsgs <- ReportMsgWithKey{
		msg:         types.NewMsgReportData(types.RequestID(id), reports, ctx.validator, key.GetAddress()),
		execVersion: execVersions,
		keyIndex:    keyIndex,
		feeEstimationData: FeeEstimationData{
			askCount:    askCount,
			minCount:    minCount,
			callData:    callData,
			rawRequests: reqs,
			clientID:    clientID,
			reports:     reports,
		},
	}
}

func handlePendingRequest(ctx *Context, logger *Logger, id types.RequestID) {
	req, err := GetRequest(ctx, logger, id)
	if err != nil {
		logger.Error(":skull: Failed to get request with error: %s", ctx, err.Error())
		return
	}

	logger.Info(":delivery_truck: Processing pending request")

	keyIndex := ctx.nextKeyIndex()
	key := ctx.keys[keyIndex]

	var rawRequests []rawRequest

	// prepare raw requests
	for _, raw := range req.RawRequests {
		ds, err := GetDataSource(ctx, logger, raw.DataSourceID)
		if err != nil {
			logger.Error(":skull: Failed to get data source hash with error: %s", ctx, err.Error())
			return
		}

		hash, ok := ctx.dataSourceCache.Load(raw.DataSourceID)
		if !ok {
			logger.Error(":skull: couldn't load data source id from cache", ctx)
			panic(sdkerrors.Wrap(errors.ErrInvalidCacheLoading, "couldn't load data source id from cache"))
		}

		rawRequests = append(rawRequests, rawRequest{
			dataSourceID:   raw.DataSourceID,
			dataSourceHash: hash.(string),
			externalID:     raw.ExternalID,
			calldata:       string(raw.Calldata),
			dataSource:     ds,
		})
	}

	// process raw requests
	reports, execVersions := handleRawRequests(ctx, logger, id, rawRequests, key)

	ctx.pendingMsgs <- ReportMsgWithKey{
		msg:         types.NewMsgReportData(id, reports, ctx.validator, key.GetAddress()),
		execVersion: execVersions,
		keyIndex:    keyIndex,
		feeEstimationData: FeeEstimationData{
			askCount:    int64(len(req.RequestedValidators)),
			minCount:    int64(req.MinCount),
			callData:    req.Calldata,
			rawRequests: rawRequests,
			clientID:    req.ClientID,
			reports:     reports,
		},
	}
}

func handleRawRequests(
	ctx *Context,
	logger *Logger,
	id types.RequestID,
	reqs []rawRequest,
	key keyring.Info,
) (reports []types.RawReport, execVersions []string) {
	resultsChan := make(chan processingResult, len(reqs))
	for _, req := range reqs {
		go handleRawRequest(
			ctx, logger.With("did", req.dataSourceID, "eid", req.externalID),
			req,
			key,
			id,
			resultsChan,
		)
	}

	versions := map[string]bool{}
	for range reqs {
		result := <-resultsChan
		reports = append(reports, result.rawReport)

		if result.err == nil {
			versions[result.version] = true
		}
	}

	for version := range versions {
		execVersions = append(execVersions, version)
	}

	return
}

func handleRawRequest(
	ctx *Context,
	logger *Logger,
	req rawRequest,
	key keyring.Info,
	id types.RequestID,
	processingResultCh chan processingResult,
) {
	ctx.updateHandlingGauge(1)
	defer ctx.updateHandlingGauge(-1)

	exec, err := GetExecutable(ctx, logger, req.dataSourceHash)
	if err != nil {
		logger.Error(":skull: Failed to load data source with error: %s", ctx, err.Error())
		processingResultCh <- processingResult{
			rawReport: types.NewRawReport(
				req.externalID, 255, []byte("FAIL_TO_LOAD_DATA_SOURCE"),
			),
			err: err,
		}
		return
	}

	vmsg := NewVerificationMessage(yoda.config.ChainID, ctx.validator, id, req.externalID)
	sig, pubkey, err := yoda.keybase.Sign(key.GetName(), vmsg.GetSignBytes())
	if err != nil {
		logger.Error(":skull: Failed to sign verify message: %s", ctx, err.Error())
		processingResultCh <- processingResult{
			rawReport: types.NewRawReport(req.externalID, 255, nil),
			err:       err,
		}
		return
	}

	result, err := ctx.executor.Exec(
		exec,
		fmt.Sprintf("\"%s\" %s", req.dataSource.Owner, req.calldata),
		map[string]interface{}{
			"BAND_CHAIN_ID":    vmsg.ChainID,
			"BAND_VALIDATOR":   vmsg.Validator.String(),
			"BAND_REQUEST_ID":  strconv.Itoa(int(vmsg.RequestID)),
			"BAND_EXTERNAL_ID": strconv.Itoa(int(vmsg.ExternalID)),
			"BAND_REPORTER":    sdk.MustBech32ifyPubKey(sdk.Bech32PubKeyTypeAccPub, pubkey),
			"BAND_SIGNATURE":   sig,
		},
	)

	if err != nil {
		logger.Error(":skull: Failed to execute data source script: %s", ctx, err.Error())
		processingResultCh <- processingResult{
			rawReport: types.NewRawReport(req.externalID, 255, nil),
			err:       err,
		}
		return
	} else {
		logger.Debug(
			":sparkles: Query data done with calldata: %q, result: %q, exitCode: %d",
			req.calldata, result.Output, result.Code,
		)
		processingResultCh <- processingResult{
			rawReport: types.NewRawReport(req.externalID, result.Code, result.Output),
			version:   result.Version,
		}
	}
}
