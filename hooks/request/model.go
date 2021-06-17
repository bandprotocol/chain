package request

import (
	"github.com/bandprotocol/chain/x/oracle/types"
)

// GenerateRequestModel converts proto's QueryRequestResponse to GORM's Request model
func GenerateRequestModel(data types.QueryRequestResponse) Request {
	oRequest := data.Request
	oReports := data.Reports
	oResult := data.Result

	requestID := oResult.RequestID

	// Requested validators
	var requestedValidators []RequestedValidator
	for _, reqVal := range oRequest.RequestedValidators {
		requestedValidators = append(requestedValidators, NewRequestedValidator(requestID, reqVal))
	}

	// IBC Channel & Port
	var ibcChannel string
	var ibcPort string
	if oRequest.IBCChannel != nil {
		ibcChannel = oRequest.IBCChannel.ChannelId
		ibcPort = oRequest.IBCChannel.PortId
	}

	// Raw requests
	var rawRequests []RawRequest
	for _, rawReq := range oRequest.RawRequests {
		rawRequests = append(rawRequests, NewRawRequest(
			requestID,
			rawReq.ExternalID,
			rawReq.DataSourceID,
			rawReq.Calldata,
		))
	}

	// Reports
	var reports []Report
	for _, report := range oReports {
		reports = append(reports, GenerateReportModel(requestID, report))
	}

	// Oracle request
	dbRequest := NewRequest(
		requestID,
		oRequest.OracleScriptID,
		oRequest.Calldata,
		requestedValidators,
		oResult.MinCount,
		oResult.AskCount,
		oResult.ClientID,
		oResult.AnsCount,
		oRequest.RequestHeight,
		oRequest.RequestTime,
		rawRequests,
		reports,
		ibcChannel,
		ibcPort,
		oRequest.ExecuteGas,
		oResult.ResolveTime,
		oResult.ResolveStatus,
		oResult.Result,
	)

	return dbRequest
}

// GenerateReportModel converts proto's Report to GORM's Report model
func GenerateReportModel(requestID types.RequestID, report types.Report) Report {
	var rawReports []RawReport
	for _, rawReport := range report.RawReports {
		rawReports = append(rawReports, NewRawReport(
			requestID,
			rawReport.ExternalID,
			rawReport.Data,
			rawReport.ExitCode,
		))
	}

	return NewReport(
		requestID,
		report.Validator,
		rawReports,
		report.InBeforeResolve,
	)
}
