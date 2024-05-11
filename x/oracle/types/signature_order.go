package types

import (
	tsstypes "github.com/bandprotocol/chain/v2/x/tss/types"
)

// signature order types
const (
	SignatureOrderTypeOracleResult = "OracleResult"
)

func init() {
	tsstypes.RegisterSignatureOrderTypeCodec(
		&OracleResultSignatureOrder{},
		"oracle/OracleResultSignatureOrder",
	)
}

// Implements Content Interface
var _ tsstypes.Content = &OracleResultSignatureOrder{}

func NewOracleResultSignatureOrder(rid RequestID, encodeType EncodeType) *OracleResultSignatureOrder {
	return &OracleResultSignatureOrder{RequestID: rid, EncodeType: encodeType}
}

// OrderRoute returns the order router key
func (o *OracleResultSignatureOrder) OrderRoute() string { return RouterKey }

// OrderType returns type of signature order that should be "OracleResult"
func (o *OracleResultSignatureOrder) OrderType() string {
	return SignatureOrderTypeOracleResult
}

// ValidateBasic validates the request's title and description of the request signature
func (o *OracleResultSignatureOrder) ValidateBasic() error {
	if o.RequestID == 0 {
		return ErrInvalidRequestID
	}

	if o.EncodeType == ENCODE_TYPE_UNSPECIFIED {
		return ErrInvalidTSSEncodeType
	}
	return nil
}
