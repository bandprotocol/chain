package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewReport(
	validator sdk.ValAddress,
	inBeforeResolve bool,
	rawReports []RawReport,
) Report {
	return Report{
		Validator:       validator.String(),
		InBeforeResolve: inBeforeResolve,
		RawReports:      rawReports,
	}
}

func NewRawReport(
	externalID ExternalID,
	exitCode uint32,
	data []byte,
) RawReport {
	return RawReport{
		ExternalID: externalID,
		ExitCode:   exitCode,
		Data:       data,
	}
}
