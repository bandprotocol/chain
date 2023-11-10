package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// Content defines an interface that a request signature type must implement.
// Content can have additional fields, which will handled by a requestSignature's Handler.
type Content interface {
	RequestingSignatureRoute() string
	RequestingSignatureType() string
	ValidateBasic() error
	String() string
}

// Handler defines a function that handles a signature request.
type Handler func(ctx sdk.Context, content Content) ([]byte, error)

// Router define a struct that hold prefix and handler of the route
type Route struct {
	Prefix  byte
	Handler Handler
}

func NewRoute(prefix byte, handler Handler) Route {
	return Route{
		Prefix:  prefix,
		Handler: handler,
	}
}
