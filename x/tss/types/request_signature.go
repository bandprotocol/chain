package types

import (
	fmt "fmt"

	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// request signature types
const (
	RequestSignatureTypeText string = "text"
)

// Implements Content Interface
var _ Content = &TextRequestingSignature{}

func NewTextRequestingSignature(msg []byte) *TextRequestingSignature {
	return &TextRequestingSignature{Message: msg}
}

// RequestingSignatureRoute returns the request router key
func (rs *TextRequestingSignature) RequestingSignatureRoute() string { return RouterKey }

// RequestSignatureType is "default"
func (rs *TextRequestingSignature) RequestingSignatureType() string { return RequestSignatureTypeText }

// ValidateBasic validates the content's title and description of the request signature
func (rs *TextRequestingSignature) ValidateBasic() error { return nil }

var validRequestingSignatureTypes = map[string]struct{}{
	RequestSignatureTypeText: {},
}

// RegisterRequestingSignatureType registers a request signature type. It will panic if the type is
// already registered.
func RegisterRequestingSignatureType(ty string) {
	if _, ok := validRequestingSignatureTypes[ty]; ok {
		panic(fmt.Sprintf("already registered proposal type: %s", ty))
	}

	validRequestingSignatureTypes[ty] = struct{}{}
}

// NewRequestingSignatureHandler implements the Handler interface for tss module-based
// request signatures (ie. TextRequestingSignature ). Since these are
// merely signaling mechanisms at the moment and do not affect state, it
// performs a no-op.
func NewRequestingSignatureHandler() Handler {
	return func(ctx sdk.Context, content Content) ([]byte, error) {
		switch c := content.(type) {
		case *TextRequestingSignature:
			return c.Message, nil

		default:
			return nil, errors.Wrapf(
				sdkerrors.ErrUnknownRequest,
				"unrecognized tss request signature message type: %s",
				c.RequestingSignatureType(),
			)
		}
	}
}
