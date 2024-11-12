package types

import tsstypes "github.com/bandprotocol/chain/v3/x/tss/types"

// signature order types
const (
	SignatureOrderTypeOracleResult = "OracleResult"
)

// Implements Content Interface
var _ tsstypes.Content = &OracleResultSignatureOrder{}

// NewOracleResultSignatureOrder returns a new OracleResultSignatureOrder object
func NewOracleResultSignatureOrder(rid RequestID, encodeType EncodeType) *OracleResultSignatureOrder {
	return &OracleResultSignatureOrder{RequestID: rid, EncodeType: encodeType}
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

	if _, ok := EncodeType_name[int32(o.EncodeType)]; !ok {
		return ErrInvalidOracleEncodeType.Wrapf("invalid encode: %s", o.EncodeType)
	}

	if o.EncodeType == ENCODE_TYPE_UNSPECIFIED {
		return ErrInvalidOracleEncodeType.Wrapf("encode type must be specified")
	}
	return nil
}
