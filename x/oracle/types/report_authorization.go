package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
)

var (
	_ authz.Authorization = &ReportAuthorization{}
)

// NewSendAuthorization creates a new SendAuthorization object.
func NewReportAuthorization() *ReportAuthorization {
	return &ReportAuthorization{}
}

// MsgTypeURL implements Authorization.MsgTypeURL.
func (a ReportAuthorization) MsgTypeURL() string {
	return sdk.MsgTypeURL(&MsgReportData{})
}

// Accept implements Authorization.Accept.
func (a ReportAuthorization) Accept(ctx sdk.Context, _ sdk.Msg) (authz.AcceptResponse, error) {
	return authz.AcceptResponse{Accept: true, Delete: false}, nil
}

// ValidateBasic implements Authorization.ValidateBasic.
func (a ReportAuthorization) ValidateBasic() error {
	return nil
}
