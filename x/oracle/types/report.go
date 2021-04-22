package types

import sdk "github.com/cosmos/cosmos-sdk/types"

func NewReport(
	Validator sdk.ValAddress,
	InBeforeResolve bool,
	RawReports []RawReport,
) Report {
	return Report{
		Validator:       Validator.String(),
		InBeforeResolve: InBeforeResolve,
		RawReports:      RawReports,
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
