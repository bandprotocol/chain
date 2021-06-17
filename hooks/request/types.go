package request

import (
	"encoding/base64"
	"time"

	"github.com/bandprotocol/chain/x/oracle/types"
	"gorm.io/gorm"
)

// RawReport is GORM model of RawReport proto message
type RawReport struct {
	gorm.Model
	// ReportID is a foreign key of Report model
	ReportID   uint
	ExternalID int64
	// Data is proto's RawReport.data field encoded in base64
	Data     string `gorm:"size:1024"`
	ExitCode uint32
}

// Report is GORM model of Report proto message
type Report struct {
	gorm.Model
	// RequestID is a foreign key of Request model
	RequestID       uint
	Validator       string
	RawReports      []RawReport `gorm:"constraint:OnDelete:CASCADE"`
	InBeforeResolve bool
}

// RawRequest is GORM model of RawRequest proto message
type RawRequest struct {
	gorm.Model
	// RequestID is a foreign key of Request model
	RequestID    uint
	ExternalID   int64
	DataSourceID int64
	// Calldata is proto's RawRequest.calldata field encoded in base64
	Calldata string `gorm:"size:1024"`
}

// RequestedValidator is GORM model of Request.requested_validators proto message
type RequestedValidator struct {
	gorm.Model
	// RequestID is a foreign key of Request model
	RequestID uint
	Address   string
}

// Request is GORM model of Request proto message combined with Result proto message
type Request struct {
	gorm.Model
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

// QueryRequestResponse convert GORM's Request model to proto's QueryRequestResponse
func (r Request) QueryRequestResponse() types.QueryRequestResponse {
	// Request's calldata
	calldata, err := base64.StdEncoding.DecodeString(r.Calldata)
	if err != nil {
		panic(err)
	}

	// Requested validators
	var requestedValidators []string
	for _, rVal := range r.RequestedValidators {
		requestedValidators = append(requestedValidators, rVal.Address)
	}

	// Raw requests
	var rawRequests []types.RawRequest
	for _, rr := range r.RawRequests {
		calldata, err := base64.StdEncoding.DecodeString(rr.Calldata)
		if err != nil {
			panic(err)
		}
		rawRequests = append(rawRequests, types.RawRequest{
			ExternalID:   types.ExternalID(rr.ExternalID),
			DataSourceID: types.DataSourceID(rr.DataSourceID),
			Calldata:     calldata,
		})
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
			rawReports = append(rawReports, types.RawReport{
				ExternalID: types.ExternalID(rr.ExternalID),
				Data:       data,
				ExitCode:   rr.ExitCode,
			})
		}
		report := types.Report{
			Validator:       dbReport.Validator,
			InBeforeResolve: dbReport.InBeforeResolve,
			RawReports:      rawReports,
		}
		reports = append(reports, report)
	}

	// IBC Channel
	var ibcChannel *types.IBCChannel
	if len(r.IBCChannelID) > 0 {
		ibcChannel = &types.IBCChannel{
			PortId:    r.IBCPortID,
			ChannelId: r.IBCChannelID,
		}
	}

	// The whole response
	request := types.QueryRequestResponse{
		Request: &types.Request{
			OracleScriptID:      types.OracleScriptID(r.OracleScriptID),
			Calldata:            calldata,
			MinCount:            r.MinCount,
			RequestHeight:       r.RequestHeight,
			RequestTime:         uint64(r.RequestTime.Unix()),
			ClientID:            r.ClientID,
			IBCChannel:          ibcChannel,
			ExecuteGas:          r.ExecuteGas,
			RequestedValidators: requestedValidators,
			RawRequests:         rawRequests,
		},
		Result: &types.Result{
			ClientID:       r.ClientID,
			OracleScriptID: types.OracleScriptID(r.OracleScriptID),
			Calldata:       calldata,
			AskCount:       r.AskCount,
			MinCount:       r.MinCount,
			AnsCount:       r.AnsCount,
			RequestID:      types.RequestID(r.ID),
			RequestTime:    r.RequestTime.Unix(),
			ResolveTime:    r.ResolveTime.Unix(),
			ResolveStatus:  types.ResolveStatus(types.ResolveStatus_value[r.ResolveStatus]),
			Result:         requestResult,
		},
		Reports: reports,
	}

	return request
}

// GenerateRequestModel converts proto's QueryRequestResponse to GORM's Request model
func GenerateRequestModel(data types.QueryRequestResponse) Request {
	request := data.Request
	reports := data.Reports
	result := data.Result

	// Oracle requests
	dbRequest := Request{
		Model: gorm.Model{
			ID: uint(result.RequestID),
		},
		OracleScriptID: int64(request.OracleScriptID),
		Calldata:       base64.StdEncoding.EncodeToString(request.Calldata),
		MinCount:       result.MinCount,
		AskCount:       result.AskCount,
		AnsCount:       result.AnsCount,
		RequestHeight:  request.RequestHeight,
		RequestTime:    time.Unix(int64(request.RequestTime), 0),
		ClientID:       result.ClientID,
		ResolveTime:    time.Unix(result.ResolveTime, 0),
		ResolveStatus:  result.ResolveStatus.String(),
		ExecuteGas:     request.ExecuteGas,
		Result:         base64.StdEncoding.EncodeToString(result.Result),
	}

	// IBC channel
	if request.IBCChannel != nil {
		dbRequest.IBCChannelID = request.IBCChannel.ChannelId
		dbRequest.IBCPortID = request.IBCChannel.PortId
	}

	// Requested validators
	for _, reqVal := range request.RequestedValidators {
		dbRequest.RequestedValidators = append(dbRequest.RequestedValidators, RequestedValidator{
			Address: reqVal,
		})
	}

	// Raw requests
	for _, rawReq := range request.RawRequests {
		dbRequest.RawRequests = append(dbRequest.RawRequests, RawRequest{
			ExternalID:   int64(rawReq.ExternalID),
			DataSourceID: int64(rawReq.DataSourceID),
			Calldata:     base64.StdEncoding.EncodeToString(rawReq.Calldata),
		})
	}

	// Reports
	for _, report := range reports {
		dbRequest.Reports = append(dbRequest.Reports, GenerateReportModel(result.RequestID, report))
	}

	return dbRequest
}

// GenerateReportModel converts proto's Report to GORM's Report model
func GenerateReportModel(requestID types.RequestID, report types.Report) Report {
	result := Report{
		RequestID:       uint(requestID),
		Validator:       report.Validator,
		InBeforeResolve: report.InBeforeResolve,
	}

	for _, rawReport := range report.RawReports {
		result.RawReports = append(result.RawReports, RawReport{
			ExternalID: int64(rawReport.ExternalID),
			Data:       base64.StdEncoding.EncodeToString(rawReport.Data),
			ExitCode:   rawReport.ExitCode,
		})
	}

	return result
}
