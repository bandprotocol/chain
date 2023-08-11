package types

import (
	tsstypes "github.com/bandprotocol/chain/v2/x/tss/types"
)

// request signature types
const (
	RequestSignatureByRequestIDType string = "RequestSignatureByRequestID"
)

func init() {
	tsstypes.RegisterRequestSignatureType(RequestSignatureByRequestIDType)
	tsstypes.RegisterRequestSignatureTypeCodec(&RequestSignatureByRequestID{}, "tss/RequestSignatureByRequestID")
}

// Implements Content Interface
var _ tsstypes.Content = &RequestSignatureByRequestID{}

func NewRequestSignatureByRequestID(rid RequestID) *RequestSignatureByRequestID {
	return &RequestSignatureByRequestID{RequestID: rid}
}

// RequestRoute returns the request router key
func (rs *RequestSignatureByRequestID) RequestSignatureRoute() string { return RouterKey }

// RequestType is "default"
func (rs *RequestSignatureByRequestID) RequestSignatureType() string {
	return RequestSignatureByRequestIDType
}

// ValidateBasic validates the content's title and description of the request signature
func (rs *RequestSignatureByRequestID) ValidateBasic() error { return nil }
