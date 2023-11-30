package types

import (
	"bytes"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ Router = (*router)(nil)

// Router implements a TSS Handler router.
type Router interface {
	AddRoute(r string, h Route) (rtr Router)
	HasRoute(r string) bool
	GetRoute(path string) (h Route)
	Seal()
}

type router struct {
	routes   map[string]Route
	prefixes map[string]Route
	sealed   bool
}

// NewRouter creates a new Router interface instance
func NewRouter() Router {
	return &router{
		routes:   make(map[string]Route),
		prefixes: make(map[string]Route),
	}
}

// Seal seals the router which prohibits any subsequent route handlers to be
// added. Seal will panic if called more than once.
func (rtr *router) Seal() {
	if rtr.sealed {
		panic("router already sealed")
	}
	rtr.sealed = true
}

// AddRoute adds request signature handler for a given path. It returns the Router
// so AddRoute calls can be linked. It will panic if the router is sealed.
func (rtr *router) AddRoute(path string, h Route) Router {
	if rtr.sealed {
		panic("router sealed; cannot add route handler")
	}

	if bytes.Equal(h.Prefix, ReplaceGroupMsgPrefix) {
		panic(fmt.Sprintf("prefix (%x) is reserved for replacing group only", ReplaceGroupMsgPrefix))
	}

	if !sdk.IsAlphaNumeric(path) {
		panic("route expressions can only contain alphanumeric characters")
	}

	if rtr.HasRoute(path) {
		panic(fmt.Sprintf("route %s has already been initialized", path))
	}

	if rtr.HasPrefix(h.Prefix) {
		panic(fmt.Sprintf("prefix %s has already been initialized", path))
	}

	rtr.routes[path] = h
	rtr.prefixes[string(h.Prefix)] = h
	return rtr
}

// HasRoute returns true if the router has a path registered or false otherwise.
func (rtr *router) HasRoute(path string) bool {
	_, ok := rtr.routes[path]
	return ok
}

// HasPrefix returns true if the router has a prefix registered or false otherwise.
func (rtr *router) HasPrefix(prefix []byte) bool {
	_, ok := rtr.prefixes[string(prefix)]
	return ok
}

// GetRoute returns a Handler for a given path.
func (rtr *router) GetRoute(path string) Route {
	if !rtr.HasRoute(path) {
		panic(fmt.Sprintf("route \"%s\" does not exist", path))
	}

	return rtr.routes[path]
}
