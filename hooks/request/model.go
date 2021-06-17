package request

import (
	"github.com/bandprotocol/chain/x/oracle/types"
)

// GenerateRequestModel converts proto's QueryRequestResponse to GORM's Request model
func GenerateRequestModel(data types.QueryRequestResponse) Request {
	pbRequest := data.Request
	pbReports := data.Reports
	pbResult := data.Result

	requestID := pbResult.RequestID

	// Requested validators
	var requestedValidators []RequestedValidator
	for _, reqVal := range pbRequest.RequestedValidators {
		requestedValidators = append(requestedValidators, NewRequestedValidator(requestID, reqVal))
	}

	// IBC Channel & Port
	var ibcChannel string
	var ibcPort string
	if pbRequest.IBCChannel != nil {
		ibcChannel = pbRequest.IBCChannel.ChannelId
		ibcPort = pbRequest.IBCChannel.PortId
	}

	// Raw requests
	var rawRequests []RawRequest
	for _, rawReq := range pbRequest.RawRequests {
		rawRequests = append(rawRequests, NewRawRequest(
			requestID,
			rawReq.ExternalID,
			rawReq.DataSourceID,
			rawReq.Calldata,
		))
	}

	// Reports
	var reports []Report
	for _, report := range pbReports {
		reports = append(reports, GenerateReportModel(requestID, report))
	}

	// Oracle request
	dbRequest := NewRequest(
		requestID,
		pbRequest.OracleScriptID,
		pbRequest.Calldata,
		requestedValidators,
		pbResult.MinCount,
		pbResult.AskCount,
		pbResult.ClientID,
		pbResult.AnsCount,
		pbRequest.RequestHeight,
		pbRequest.RequestTime,
		rawRequests,
		reports,
		ibcChannel,
		ibcPort,
		pbRequest.ExecuteGas,
		pbResult.ResolveTime,
		pbResult.ResolveStatus,
		pbResult.Result,
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
