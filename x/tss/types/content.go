package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// Content defines an interface that a request signature type must implement.
// Content can have additional fields, which will handled by a requestSignature's Handler.
type Content interface {
	RequestSignatureRoute() string
	RequestSignatureType() string
	ValidateBasic() error
	String() string
}

// Handler defines a function that handles a signature request.
type Handler func(ctx sdk.Context, content Content) ([]byte, error)
