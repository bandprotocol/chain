package types

import (
	"time"

	github_com_cosmos_cosmos_sdk_types "github.com/cosmos/cosmos-sdk/types"
)

func NewDataSource(
	Owner github_com_cosmos_cosmos_sdk_types.AccAddress,
	Name string,
	Description string,
	Filename string,
) DataSource {
	return DataSource{
		Owner:       Owner.String(),
		Name:        Name,
		Description: Description,
		Filename:    Filename,
	}
}

func NewOracleScript(
	Owner github_com_cosmos_cosmos_sdk_types.AccAddress,
	Name string,
	Description string,
	Filename string,
	Schema string,
	SourceCodeURL string,
) OracleScript {
	return OracleScript{
		Owner:         Owner.String(),
		Name:          Name,
		Description:   Description,
		Filename:      Filename,
		Schema:        Schema,
		SourceCodeURL: SourceCodeURL,
	}
}

func NewRawRequest(
	ExternalID ExternalID,
	DataSourceID DataSourceID,
	Calldata []byte,
) RawRequest {
	return RawRequest{
		ExternalID:   ExternalID,
		DataSourceID: DataSourceID,
		Calldata:     Calldata,
	}
}

func NewRawReport(
	ExternalID ExternalID,
	ExitCode uint32,
	Data []byte,
) RawReport {
	return RawReport{
		ExternalID: ExternalID,
		ExitCode:   ExitCode,
		Data:       Data,
	}
}

func NewRequest(
	OracleScriptID OracleScriptID,
	Calldata []byte,
	RequestedValidators []github_com_cosmos_cosmos_sdk_types.ValAddress,
	MinCount uint64,
	RequestHeight int64,
	RequestTime time.Time,
	ClientID string,
	RawRequests []RawRequest,
) Request {
	requestedVals := make([]string, len(RequestedValidators))
	for idx, reqVal := range RequestedValidators {
		requestedVals[idx] = reqVal.String()
	}
	return Request{
		OracleScriptID:      OracleScriptID,
		Calldata:            Calldata,
		RequestedValidators: requestedVals,
		MinCount:            MinCount,
		RequestHeight:       RequestHeight,
		RequestTime:         uint64(RequestTime.UnixNano()),
		ClientID:            ClientID,
		RawRequests:         RawRequests,
	}
}

func NewReport(
	Validator github_com_cosmos_cosmos_sdk_types.ValAddress,
	InBeforeResolve bool,
	RawReports []RawReport,
) Report {
	return Report{
		Validator:       Validator.String(),
		InBeforeResolve: InBeforeResolve,
		RawReports:      RawReports,
	}
}

func NewValidatorStatus(
	IsActive bool,
	Since time.Time,
) ValidatorStatus {
	return ValidatorStatus{
		IsActive: IsActive,
		Since:    Since,
	}
}
