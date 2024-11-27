package types

import tsstypes "github.com/bandprotocol/chain/v3/x/tss/types"

// signature order types
const (
	SignatureOrderTypeOracleResult = "oracleResult"
)

// Implements Content Interface
var _ tsstypes.Content = &OracleResultSignatureOrder{}

// NewOracleResultSignatureOrder returns a new OracleResultSignatureOrder object
func NewOracleResultSignatureOrder(rid RequestID, encoder Encoder) *OracleResultSignatureOrder {
	return &OracleResultSignatureOrder{RequestID: rid, Encoder: encoder}
}

// OrderRoute returns the order router key
func (o *OracleResultSignatureOrder) OrderRoute() string { return RouterKey }

// OrderType returns type of signature order that should be "OracleResult"
func (o *OracleResultSignatureOrder) OrderType() string {
	return SignatureOrderTypeOracleResult
}

// IsInternal returns false for OracleResultSignatureOrder (allow user to submit this content type).
func (o *OracleResultSignatureOrder) IsInternal() bool { return false }

// ValidateBasic validates the request's title and description of the request signature
func (o *OracleResultSignatureOrder) ValidateBasic() error {
	if o.RequestID == 0 {
		return ErrInvalidRequestID
	}

	if _, ok := Encoder_name[int32(o.Encoder)]; !ok {
		return ErrInvalidOracleEncoder.Wrapf("invalid encoder: %d", o.Encoder)
	}

	if o.Encoder == ENCODER_UNSPECIFIED {
		return ErrInvalidOracleEncoder.Wrapf("encoder must be specified")
	}
	return nil
}
