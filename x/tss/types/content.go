package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	tsslib "github.com/bandprotocol/chain/v3/pkg/tss"
)

type ContentRouter struct {
	handlers map[string]Handler
	sealed   bool
}

// NewContentRouter creates a new Router interface instance
func NewContentRouter() *ContentRouter {
	return &ContentRouter{
		handlers: make(map[string]Handler),
	}
}

// Seal seals the content router which prohibits any subsequent route handlers to be
// added. Seal will panic if called more than once.
func (r *ContentRouter) Seal() {
	if r.sealed {
		panic("router already sealed")
	}
	r.sealed = true
}

// AddRoute adds request signature handler for a given path and prefix. It returns the ContentRouter
// so AddRoute calls can be linked. It will panic if the router is sealed.
func (r *ContentRouter) AddRoute(path string, h Handler) *ContentRouter {
	if r.sealed {
		panic("router sealed; cannot add route handler")
	}

	if !sdk.IsAlphaNumeric(path) {
		panic("route expressions can only contain alphanumeric characters")
	}

	if r.HasRoute(path) {
		panic(fmt.Sprintf("route %s has already been initialized", path))
	}

	r.handlers[path] = wrapHandler(path, h)
	return r
}

// HasRoute returns true if the content router has a path registered or false otherwise.
func (r *ContentRouter) HasRoute(path string) bool {
	_, ok := r.handlers[path]
	return ok
}

// GetRoute returns a Handler for a given path.
func (r *ContentRouter) GetRoute(path string) Handler {
	if !r.HasRoute(path) {
		panic(fmt.Sprintf("route \"%s\" does not exist", path))
	}

	return r.handlers[path]
}

// wrapHandler returns a function that converts content into message bytes.
// It prefixes the message with a selector, which consists of first 4 bytes of the hashed path.
func wrapHandler(path string, handler Handler) Handler {
	return func(ctx sdk.Context, req Content) ([]byte, error) {
		msg, err := handler(ctx, req)
		if err != nil {
			return nil, ErrHandleSignatureOrderFailed.Wrap(err.Error())
		}
		selector := tsslib.Hash([]byte(path))[:4]

		return append(selector, msg...), nil
	}
}

// Content defines an interface that a signature order must implement. It contains information
// such as the type and routing information for the appropriate handler to process the order.
// Content can have additional fields, which is handled by an order's Handler.
type Content interface {
	OrderRoute() string
	OrderType() string
	IsInternal() bool

	ValidateBasic() error
	String() string
}

// Handler defines a function that receive signature order and return message that should to be signed.
type Handler func(ctx sdk.Context, content Content) ([]byte, error)
