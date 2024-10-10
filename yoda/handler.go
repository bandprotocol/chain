package yoda

import (
	"encoding/hex"
	"strconv"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/crypto/tmhash"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"

	"github.com/bandprotocol/chain/v3/x/oracle/types"
)

type processingResult struct {
	rawReport types.RawReport
	version   string
	err       error
}

func MustAtoi(num string) int64 {
	result, err := strconv.ParseInt(num, 10, 64)
	if err != nil {
		panic(err)
	}
	return result
}

func handleTransaction(c *Context, l *Logger, tx abci.TxResult) {
	l.Debug(":eyes: Inspecting incoming transaction: %X", tmhash.Sum(tx.Tx))
	if tx.Result.Code != 0 {
		l.Debug(":alien: Skipping transaction with non-zero code: %d", tx.Result.Code)
		return
	}

	events := tx.Result.Events
	idStrs := GetEventValues(events, types.EventTypeRequest, types.AttributeKeyID)
	for _, idStr := range idStrs {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			l.Error(":cold_sweat: Failed to convert %s to integer with error: %s", c, idStr, err.Error())
			return
		}

		// If id is in pending requests list, then skip it.
		if c.pendingRequests[types.RequestID(id)] {
			l.Debug(":eyes: Request is in pending list, then skip")
			return
		}

		go handleRequest(c, l, types.RequestID(id))
	}
}

func handleRequest(c *Context, l *Logger, id types.RequestID) {
	l = l.With("rid", id)

	req, err := GetRequest(c, l, id)
	if err != nil {
		l.Error(":skull: Failed to get request with error: %s", c, err.Error())
		return
	}

	hasMe := false
	for _, val := range req.RequestedValidators {
		if val == c.validator.String() {
			hasMe = true
			break
		}
	}
	if !hasMe {
		l.Debug(":next_track_button: Skip request not related to this validator")
		return
	}

	l.Info(":delivery_truck: Processing request")

	keyIndex := c.nextKeyIndex()
	key := c.keys[keyIndex]

	var rawRequests []rawRequest

	// prepare raw requests
	for _, raw := range req.RawRequests {
		hash, err := GetDataSourceHash(c, l, raw.DataSourceID)
		if err != nil {
			l.Error(":skull: Failed to get data source hash with error: %s", c, err.Error())
			return
		}

		rawRequests = append(rawRequests, rawRequest{
			dataSourceID:   raw.DataSourceID,
			dataSourceHash: hash,
			externalID:     raw.ExternalID,
			calldata:       string(raw.Calldata),
		})
	}

	// process raw requests
	reports, execVersions := handleRawRequests(c, l, id, rawRequests, key)

	c.pendingMsgs <- ReportMsgWithKey{
		msg:         types.NewMsgReportData(id, reports, c.validator),
		execVersion: execVersions,
		keyIndex:    keyIndex,
		feeEstimationData: FeeEstimationData{
			askCount:    int64(len(req.RequestedValidators)),
			minCount:    int64(req.MinCount),
			callData:    req.Calldata,
			rawRequests: rawRequests,
			clientID:    req.ClientID,
		},
	}
}

func handleRawRequests(
	c *Context,
	l *Logger,
	id types.RequestID,
	reqs []rawRequest,
	key *keyring.Record,
) (reports []types.RawReport, execVersions []string) {
	resultsChan := make(chan processingResult, len(reqs))
	for _, req := range reqs {
		go handleRawRequest(
			c,
			l.With("did", req.dataSourceID, "eid", req.externalID),
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
	c *Context,
	l *Logger,
	req rawRequest,
	key *keyring.Record,
	id types.RequestID,
	processingResultCh chan processingResult,
) {
	c.updateHandlingGauge(1)
	defer c.updateHandlingGauge(-1)

	exec, err := GetExecutable(c, l, req.dataSourceHash)
	if err != nil {
		l.Error(":skull: Failed to load data source with error: %s", c, err.Error())
		processingResultCh <- processingResult{
			rawReport: types.NewRawReport(
				req.externalID, 255, []byte("FAIL_TO_LOAD_DATA_SOURCE"),
			),
			err: err,
		}
		return
	}

	vmsg := types.NewRequestVerification(cfg.ChainID, c.validator, id, req.externalID, req.dataSourceID)
	sig, pubkey, err := kb.Sign(key.Name, vmsg.GetSignBytes(), signing.SignMode_SIGN_MODE_DIRECT)
	if err != nil {
		l.Error(":skull: Failed to sign verify message: %s", c, err.Error())
		processingResultCh <- processingResult{
			rawReport: types.NewRawReport(req.externalID, 255, nil),
			err:       err,
		}
		return
	}

	result, err := c.executor.Exec(exec, req.calldata, map[string]interface{}{
		"BAND_CHAIN_ID":       vmsg.ChainID,
		"BAND_DATA_SOURCE_ID": strconv.Itoa(int(vmsg.DataSourceID)),
		"BAND_VALIDATOR":      vmsg.Validator,
		"BAND_REQUEST_ID":     strconv.Itoa(int(vmsg.RequestID)),
		"BAND_EXTERNAL_ID":    strconv.Itoa(int(vmsg.ExternalID)),
		"BAND_REPORTER":       hex.EncodeToString(pubkey.Bytes()),
		"BAND_SIGNATURE":      sig,
	})

	if err != nil {
		l.Error(":skull: Failed to execute data source script: %s", c, err.Error())
		processingResultCh <- processingResult{
			rawReport: types.NewRawReport(req.externalID, 255, nil),
			err:       err,
		}
		return
	} else {
		l.Debug(
			":sparkles: Query data done with calldata: %q, result: %q, exitCode: %d",
			req.calldata, result.Output, result.Code,
		)
		processingResultCh <- processingResult{
			rawReport: types.NewRawReport(req.externalID, result.Code, result.Output),
			version:   result.Version,
		}
	}
}
