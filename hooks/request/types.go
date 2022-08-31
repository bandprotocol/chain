package request

import (
	"encoding/base64"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/oracle/types"
)

// RawReport is GORM model of RawReport proto message
type RawReport struct {
	ID uint `gorm:"primarykey"`
	// ReportID is a foreign key of Report model
	ReportID   uint `gorm:"index:idx_raw_report_report_id"`
	ExternalID int64
	// Data is proto's RawReport.data field encoded in base64
	Data     string `gorm:"size:1024"`
	ExitCode uint32
}

// Report is GORM model of Report proto message
type Report struct {
	ID uint `gorm:"primarykey"`
	// RequestID is a foreign key of Request model
	RequestID       uint `gorm:"index:idx_report_request_id"`
	Validator       string
	RawReports      []RawReport `gorm:"constraint:OnDelete:CASCADE"`
	InBeforeResolve bool
}

// RawRequest is GORM model of RawRequest proto message
type RawRequest struct {
	ID uint `gorm:"primarykey"`
	// RequestID is a foreign key of Request model
	RequestID    uint `gorm:"index:idx_raw_request_request_id"`
	ExternalID   int64
	DataSourceID int64
	// Calldata is proto's RawRequest.calldata field encoded in base64
	Calldata string `gorm:"size:1024"`
}

// RequestedValidator is GORM model of Request.requested_validators proto message
type RequestedValidator struct {
	ID uint `gorm:"primarykey"`
	// RequestID is a foreign key of Request model
	RequestID uint `gorm:"index:idx_requested_validator_request_id"`
	Address   string
}

// Request is GORM model of Request proto message combined with Result proto message
type Request struct {
	ID             uint  `gorm:"primarykey"`
	OracleScriptID int64 `gorm:"index:idx_min_count_ask_count_oracle_script_id_calldata"`
	// Calldata is proto's Request.calldata field encoded in base64
	Calldata            string               `gorm:"index:idx_min_count_ask_count_oracle_script_id_calldata;size:1024"`
	RequestedValidators []RequestedValidator `gorm:"constraint:OnDelete:CASCADE"`
	MinCount            uint64               `gorm:"index:idx_min_count_ask_count_oracle_script_id_calldata"`
	AskCount            uint64               `gorm:"index:idx_min_count_ask_count_oracle_script_id_calldata"`
	AnsCount            uint64
	RequestHeight       int64
	RequestTime         time.Time
	ClientID            string
	RawRequests         []RawRequest `gorm:"constraint:OnDelete:CASCADE"`
	IBCPortID           string
	IBCChannelID        string
	ExecuteGas          uint64
	Reports             []Report  `gorm:"constraint:OnDelete:CASCADE"`
	ResolveTime         time.Time `gorm:"index:idx_min_count_ask_count_oracle_script_id_calldata"`
	// ResolveStatus is proto's resolve_status but as enum string
	ResolveStatus string `gorm:"size:64"`
	// Result is proto's Result.result field encoded in OBI and base64
	Result string `gorm:"size:1024"`
}

// NewRawReport creates new instance of RawReport
func NewRawReport(requestID types.RequestID, externalID types.ExternalID, data []byte, exitCode uint32) RawReport {
	return RawReport{
		ReportID:   uint(requestID),
		ExternalID: int64(externalID),
		Data:       base64.StdEncoding.EncodeToString(data),
		ExitCode:   exitCode,
	}
}

// NewReport creates new instance of Report
func NewReport(requestID types.RequestID, valAddr string, rawReports []RawReport, inBeforeResolve bool) Report {
	return Report{
		RequestID:       uint(requestID),
		Validator:       valAddr,
		RawReports:      rawReports,
		InBeforeResolve: inBeforeResolve,
	}
}

// NewRawRequest creates new instance of RawRequest
func NewRawRequest(
	requestID types.RequestID,
	externalID types.ExternalID,
	dataSourceID types.DataSourceID,
	calldata []byte,
) RawRequest {
	return RawRequest{
		RequestID:    uint(requestID),
		ExternalID:   int64(externalID),
		DataSourceID: int64(dataSourceID),
		Calldata:     base64.StdEncoding.EncodeToString(calldata),
	}
}

// NewRequestedValidator creates new instance of RequestedValidator
func NewRequestedValidator(requestID types.RequestID, address string) RequestedValidator {
	return RequestedValidator{
		RequestID: uint(requestID),
		Address:   address,
	}
}

// NewRequest creates new instance of Request
func NewRequest(
	id types.RequestID,
	oracleScriptID types.OracleScriptID,
	calldata []byte,
	requestedValidators []RequestedValidator,
	minCount uint64,
	askCount uint64,
	clientID string,
	ansCount uint64,
	requestHeight int64,
	requestTimeUnix int64,
	rawRequests []RawRequest,
	reports []Report,
	ibcChannel string,
	ibcPort string,
	executeGas uint64,
	resolveTimeUnix int64,
	resolveStatus types.ResolveStatus,
	result []byte,
) Request {
	return Request{
		ID:                  uint(id),
		OracleScriptID:      int64(oracleScriptID),
		Calldata:            base64.StdEncoding.EncodeToString(calldata),
		RequestedValidators: requestedValidators,
		MinCount:            minCount,
		AskCount:            askCount,
		AnsCount:            ansCount,
		RequestHeight:       requestHeight,
		RequestTime:         time.Unix(int64(requestTimeUnix), 0),
		ClientID:            clientID,
		RawRequests:         rawRequests,
		Reports:             reports,
		IBCChannelID:        ibcChannel,
		IBCPortID:           ibcPort,
		ExecuteGas:          executeGas,
		ResolveTime:         time.Unix(resolveTimeUnix, 0),
		ResolveStatus:       resolveStatus.String(),
		Result:              base64.StdEncoding.EncodeToString(result),
	}
}

// QueryRequestResponse convert GORM's Request model to proto's QueryRequestResponse
func (r Request) QueryRequestResponse() types.QueryRequestResponse {
	// Request's calldata
	calldata, err := base64.StdEncoding.DecodeString(r.Calldata)
	if err != nil {
		panic(err)
	}

	// Requested validators
	var requestedValidators []sdk.ValAddress
	for _, rVal := range r.RequestedValidators {
		valAddr, err := sdk.ValAddressFromBech32(rVal.Address)
		if err != nil {
			panic(err)
		}
		requestedValidators = append(requestedValidators, valAddr)
	}

	// Raw requests
	var rawRequests []types.RawRequest
	for _, rr := range r.RawRequests {
		calldata, err := base64.StdEncoding.DecodeString(rr.Calldata)
		if err != nil {
			panic(err)
		}
		rawRequests = append(rawRequests, types.NewRawRequest(
			types.ExternalID(rr.ExternalID),
			types.DataSourceID(rr.DataSourceID),
			calldata,
		))
	}

	// Result data
	requestResult, err := base64.StdEncoding.DecodeString(r.Result)
	if err != nil {
		panic(err)
	}

	// Reports
	var reports []types.Report
	for _, dbReport := range r.Reports {
		// Raw reports
		var rawReports []types.RawReport
		for _, rr := range dbReport.RawReports {
			data, err := base64.StdEncoding.DecodeString(rr.Data)
			if err != nil {
				panic(err)
			}
			rawReports = append(rawReports, types.NewRawReport(
				types.ExternalID(rr.ExternalID),
				rr.ExitCode,
				data,
			))
		}
		valAddr, err := sdk.ValAddressFromBech32(dbReport.Validator)
		if err != nil {
			panic(err)
		}
		reports = append(reports, types.NewReport(
			valAddr,
			dbReport.InBeforeResolve,
			rawReports,
		))
	}

	// IBC Channel
	var ibcChannel *types.IBCChannel
	if len(r.IBCChannelID) > 0 {
		channelAndPort := types.NewIBCChannel(r.IBCPortID, r.IBCChannelID)
		ibcChannel = &channelAndPort
	}

	// Oracle request
	oracleRequest := types.NewRequest(
		types.OracleScriptID(r.OracleScriptID),
		calldata,
		requestedValidators,
		r.MinCount,
		r.RequestHeight,
		r.RequestTime,
		r.ClientID,
		rawRequests,
		ibcChannel,
		r.ExecuteGas,
	)

	// Oracle result for the above request
	oracleResult := types.NewResult(
		r.ClientID,
		types.OracleScriptID(r.OracleScriptID),
		calldata,
		r.AskCount,
		r.MinCount,
		types.RequestID(r.ID),
		r.AnsCount,
		r.RequestTime.Unix(),
		r.ResolveTime.Unix(),
		types.ResolveStatus(types.ResolveStatus_value[r.ResolveStatus]),
		requestResult,
	)

	// The whole response
	request := types.QueryRequestResponse{
		Request: &oracleRequest,
		Result:  &oracleResult,
		Reports: reports,
	}

	return request
}
