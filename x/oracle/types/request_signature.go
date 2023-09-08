package types

import (
	tsstypes "github.com/bandprotocol/chain/v2/x/tss/types"
)

// request signature types
const (
	RequestSignatureTypeOracleResult string = "OracleResult"
)

func init() {
	tsstypes.RegisterRequestingSignatureType(RequestSignatureTypeOracleResult)
	tsstypes.RegisterRequestSignatureTypeCodec(&OracleResultRequestingSignature{}, "tss/OracleResultRequestingSignature")
}

// Implements Content Interface
var _ tsstypes.Content = &OracleResultRequestingSignature{}

func NewRequestingSignature(rid RequestID) *OracleResultRequestingSignature {
	return &OracleResultRequestingSignature{RequestID: rid}
}

// RequestingSignatureRoute returns the request router key
func (ors *OracleResultRequestingSignature) RequestingSignatureRoute() string { return RouterKey }

// RequestingSignatureType is "OracleResult"
func (ors *OracleResultRequestingSignature) RequestingSignatureType() string {
	return RequestSignatureTypeOracleResult
}

// ValidateBasic validates the content's title and description of the request signature
func (ors *OracleResultRequestingSignature) ValidateBasic() error { return nil }
