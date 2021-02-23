package common

import (
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/x/oracle/types"
)

func getData(clientCtx client.Context, bz []byte, ptr interface{}) error {
	var result types.QueryResult
	if err := json.Unmarshal(bz, &result); err != nil {
		return err
	}
	return clientCtx.LegacyAmino.UnmarshalJSON(result.Result, ptr)

}

func queryReporters(clientCtx client.Context, validator sdk.ValAddress) ([]sdk.AccAddress, int64, error) {
	bz, height, err := clientCtx.Query(fmt.Sprintf("custom/oracle/%s/%s", types.QueryReporters, validator))
	if err != nil {
		return nil, 0, err
	}
	var reporters []sdk.AccAddress
	err = getData(clientCtx, bz, &reporters)
	if err != nil {
		return nil, 0, err
	}
	return reporters, height, nil
}

func queryParams(clientCtx client.Context) (types.Params, int64, error) {
	bz, height, err := clientCtx.Query(fmt.Sprintf("custom/oracle/%s", types.QueryParams))
	if err != nil {
		return types.Params{}, 0, err
	}
	var params types.Params
	err = getData(clientCtx, bz, &params)
	if err != nil {
		return types.Params{}, 0, err
	}
	return params, height, nil
}

// TODO: Refactor this code with yoda
type VerificationMessage struct {
	ChainID    string           `json:"chain_id"`
	Validator  sdk.ValAddress   `json:"validator"`
	RequestID  types.RequestID  `json:"request_id"`
	ExternalID types.ExternalID `json:"external_id"`
}

func NewVerificationMessage(
	chainID string, validator sdk.ValAddress, requestID types.RequestID, externalID types.ExternalID,
) VerificationMessage {
	return VerificationMessage{
		ChainID:    chainID,
		Validator:  validator,
		RequestID:  requestID,
		ExternalID: externalID,
	}
}

func (msg VerificationMessage) GetSignBytes(legacyQuerierCdc *codec.LegacyAmino) []byte {
	return sdk.MustSortJSON(legacyQuerierCdc.MustMarshalJSON(msg))
}

type VerificationResult struct {
	ChainID      string             `json:"chain_id"`
	Validator    sdk.ValAddress     `json:"validator"`
	RequestID    types.RequestID    `json:"request_id,string"`
	ExternalID   types.ExternalID   `json:"external_id,string"`
	DataSourceID types.DataSourceID `json:"data_source_id,string"`
}

func VerifyRequest(
	clientCtx client.Context, chainID string, requestID types.RequestID,
	externalID types.ExternalID, validator sdk.ValAddress, reporterPubkey cryptotypes.PubKey, signature []byte,
) ([]byte, int64, error) {
	// Verify chain id
	if clientCtx.ChainID != chainID {
		return nil, 0, fmt.Errorf("Invalid Chain ID; expect %s, got %s", clientCtx.ChainID, chainID)
	}
	// Verify signature
	if !reporterPubkey.VerifySignature(
		NewVerificationMessage(
			chainID, validator, requestID, externalID,
		).GetSignBytes(clientCtx.LegacyAmino),
		signature,
	) {
		return nil, 0, fmt.Errorf("Signature verification failed")
	}

	// Check reporters
	reporters, _, err := queryReporters(clientCtx, validator)
	if err != nil {
		return nil, 0, err
	}

	reporter := sdk.AccAddress(reporterPubkey.Address().Bytes())
	isReporter := false
	for _, r := range reporters {
		if reporter.Equals(r) {
			isReporter = true
		}
	}
	if !isReporter {
		return nil, 0, fmt.Errorf("%s is not an authorized report of %s", reporter, validator)
	}

	request, height, err := queryRequest(clientCtx, requestID)
	if err != nil {
		return nil, 0, err
	}

	// Verify that this validator has been assigned to report this request
	assigned := false
	for _, requestedVal := range request.Request.RequestedValidators {
		val, _ := sdk.ValAddressFromBech32(requestedVal)
		if validator.Equals(val) {
			assigned = true
			break
		}
	}
	if !assigned {
		return nil, 0, fmt.Errorf("%s is not assigned for request ID %d", validator, requestID)
	}

	// Verify this request need this external id
	dataSourceID := types.DataSourceID(0)
	for _, rawRequest := range request.Request.RawRequests {
		if rawRequest.ExternalID == externalID {
			dataSourceID = rawRequest.DataSourceID
			break
		}
	}
	if dataSourceID == types.DataSourceID(0) {
		return nil, 0, fmt.Errorf("Invalid external ID %d for request ID %d", externalID, requestID)
	}

	// Verify validator hasn't reported on the request.
	reported := false
	for _, report := range request.Reports {
		repVal, _ := sdk.ValAddressFromBech32(report.Validator)
		if repVal.Equals(validator) {
			reported = true
			break
		}
	}

	if reported {
		return nil, 0, fmt.Errorf(
			"Validator %s already submitted data report for this request", validator,
		)
	}

	// Verify request has not been expired
	params, _, err := queryParams(clientCtx)
	if err != nil {
		return nil, 0, err
	}

	if request.Request.RequestHeight+int64(params.ExpirationBlockCount) < height {
		return nil, 0, fmt.Errorf("Request #%d is already expired", requestID)
	}
	bz, err := types.QueryOK(clientCtx.LegacyAmino, VerificationResult{
		ChainID:      chainID,
		Validator:    validator,
		RequestID:    requestID,
		ExternalID:   externalID,
		DataSourceID: dataSourceID,
	})
	return bz, height, err
}
