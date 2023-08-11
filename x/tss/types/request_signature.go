package types

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// request signature types
const (
	RequestSignatureTypeDefault string = "default"
)

// Implements Content Interface
var _ Content = &DefaultRequestSignature{}

func NewDefaultRequestSignature(msg []byte) *DefaultRequestSignature {
	return &DefaultRequestSignature{Message: msg}
}

// RequestSignatureRoute returns the request router key
func (rs *DefaultRequestSignature) RequestSignatureRoute() string { return RouterKey }

// RequestSignatureType is "default"
func (rs *DefaultRequestSignature) RequestSignatureType() string { return RequestSignatureTypeDefault }

// ValidateBasic validates the content's title and description of the request signature
func (rs *DefaultRequestSignature) ValidateBasic() error { return nil }

var validRequestSignatureTypes = map[string]struct{}{
	RequestSignatureTypeDefault: {},
}

// RegisterRequestSignatureType registers a request signature type. It will panic if the type is
// already registered.
func RegisterRequestSignatureType(ty string) {
	if _, ok := validRequestSignatureTypes[ty]; ok {
		panic(fmt.Sprintf("already registered proposal type: %s", ty))
	}

	validRequestSignatureTypes[ty] = struct{}{}
}

// NewDefaultRequestSignatureHandler implements the Handler interface for tss module-based
// request signatures (ie. DefaultRequestSignature ). Since these are
// merely signaling mechanisms at the moment and do not affect state, it
// performs a no-op.
func NewDefaultRequestSignatureHandler() Handler {
	return func(ctx sdk.Context, content Content) ([]byte, error) {
		switch c := content.(type) {
		case *DefaultRequestSignature:
			return c.Message, nil

		default:
			return nil, sdkerrors.Wrapf(
				sdkerrors.ErrUnknownRequest,
				"unrecognized tss request signature message type: %s",
				c.RequestSignatureType(),
			)
		}
	}
}
