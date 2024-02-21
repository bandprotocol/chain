package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	tsstypes "github.com/bandprotocol/chain/v2/x/tss/types"
)

type Router struct {
	handlers map[string]tsstypes.Handler
	sealed   bool
}

// NewRouter creates a new Router interface instance
func NewRouter() *Router {
	return &Router{
		handlers: make(map[string]tsstypes.Handler),
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
func (r *Router) AddRoute(path string, h tsstypes.Handler) *Router {
	if r.sealed {
		panic("router sealed; cannot add route handler")
	}

	if path == tsstypes.ReplaceGroupPath {
		panic(fmt.Sprintf("path (%s) is reserved for replacing group only", tsstypes.ReplaceGroupPath))
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
func (r *Router) GetRoute(path string) tsstypes.Handler {
	if !r.HasRoute(path) {
		panic(fmt.Sprintf("route \"%s\" does not exist", path))
	}

	return r.handlers[path]
}

func wrapHandler(path string, handler tsstypes.Handler) tsstypes.Handler {
	return func(ctx sdk.Context, req tsstypes.Content) ([]byte, error) {
		msg, err := handler(ctx, req)
		if err != nil {
			return nil, ErrHandleSignatureOrderFailed.Wrap(err.Error())
		}
		selector := tss.Hash([]byte(path))[:4]

		return append(selector, msg...), nil
	}
}
