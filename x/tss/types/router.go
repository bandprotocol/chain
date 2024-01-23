package types

import (
	"fmt"

	"cosmossdk.io/errors"
	tsslib "github.com/bandprotocol/chain/v2/pkg/tss"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Router struct {
	handlers map[string]Handler
	sealed   bool
}

// NewRouter creates a new Router interface instance
func NewRouter() *Router {
	return &Router{
		handlers: make(map[string]Handler),
	}
}

// Seal seals the router which prohibits any subsequent route handlers to be
// added. Seal will panic if called more than once.
func (r *Router) Seal() {
	if r.sealed {
		panic("router already sealed")
	}
	r.sealed = true
}

// AddRoute adds request signature handler for a given path and prefix. It returns the Router
// so AddRoute calls can be linked. It will panic if the router is sealed.
func (r *Router) AddRoute(path string, h Handler) *Router {
	if r.sealed {
		panic("router sealed; cannot add route handler")
	}

	if path == ReplaceGroupPath {
		panic(fmt.Sprintf("prefix (%x) is reserved for replacing group only", ReplaceGroupMsgPrefix))
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

// HasRoute returns true if the router has a path registered or false otherwise.
func (r *Router) HasRoute(path string) bool {
	_, ok := r.handlers[path]
	return ok
}

// GetRoute returns a Handler for a given path.
func (r *Router) GetRoute(path string) Handler {
	if !r.HasRoute(path) {
		panic(fmt.Sprintf("route \"%s\" does not exist", path))
	}

	return r.handlers[path]
}

func wrapHandler(path string, handler Handler) Handler {
	return func(ctx sdk.Context, req Content) ([]byte, error) {
		msg, err := handler(ctx, req)
		if err != nil {
			return nil, errors.Wrap(ErrHandleSignatureOrderFailed, err.Error())
		}
		selector := tsslib.Hash([]byte(path))[:4]

		return append(selector, msg...), nil
	}
}
