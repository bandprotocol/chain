package types

// Signature order types
const (
	SignatureOrderTypeText string = "text"
)

// Implements SignatureRequest Interface
var _ Content = &TextSignatureOrder{}

func NewTextSignatureOrder(msg []byte) *TextSignatureOrder {
	return &TextSignatureOrder{Message: msg}
}

// OrderRoute returns the order router key
func (rs *TextSignatureOrder) OrderRoute() string { return RouterKey }

// OrderType of TextSignatureOrder is "Text"
func (rs *TextSignatureOrder) OrderType() string {
	return SignatureOrderTypeText
}

// ValidateBasic performs no-op for this type
func (rs *TextSignatureOrder) ValidateBasic() error { return nil }
