package types

import (
	tsstypes "github.com/bandprotocol/chain/v2/x/tss/types"
)

// request signature types
const (
	RequestSignatureTypeOracleResult string = "OracleResult"
)

func init() {
	tsstypes.RegisterRequestSignatureType(RequestSignatureTypeOracleResult)
	tsstypes.RegisterRequestSignatureTypeCodec(&OracleResultRequestSignature{}, "tss/OracleResultRequestSignature")
}

// Implements Content Interface
var _ tsstypes.Content = &OracleResultRequestSignature{}

func NewRequestSignature(rid RequestID) *OracleResultRequestSignature {
	return &OracleResultRequestSignature{RequestID: rid}
}

// RequestRoute returns the request router key
func (ors *OracleResultRequestSignature) RequestSignatureRoute() string { return RouterKey }

// RequestType is "OracleResult"
func (ors *OracleResultRequestSignature) RequestSignatureType() string {
	return RequestSignatureTypeOracleResult
}

// ValidateBasic validates the content's title and description of the request signature
func (ors *OracleResultRequestSignature) ValidateBasic() error { return nil }
