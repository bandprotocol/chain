package types

import (
	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Signature order types
const (
	SignatureOrderTypeText string = "Text"
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

// NewSignatureOrderHandler implements the Handler interface for tss module-based
// request signatures (ie. TextSignatureOrder)
func NewSignatureOrderHandler() Handler {
	return func(ctx sdk.Context, content Content) ([]byte, error) {
		switch c := content.(type) {
		case *TextSignatureOrder:
			return c.Message, nil

		default:
			return nil, errors.Wrapf(
				sdkerrors.ErrUnknownRequest,
				"unrecognized tss request signature message type: %s",
				c.OrderType(),
			)
		}
	}
}
