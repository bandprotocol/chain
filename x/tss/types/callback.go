package types

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/pkg/tss"
)

// CallbackRouter is a struct that holds a map of TSSCallback objects for each module.
type CallbackRouter struct {
	routes map[string]TSSCallback
	sealed bool
}

// NewCallbackRouter creates a new CallbackRouter instance.
func NewCallbackRouter() *CallbackRouter {
	return &CallbackRouter{
		routes: make(map[string]TSSCallback),
	}
}

// Seal seals the CallbackRouter which prohibits any subsequent TSSCallback to be added.
// Seal will panic if called more than once.
func (cbr *CallbackRouter) Seal() {
	if cbr.sealed {
		panic(errors.New("callback router is already sealed"))
	}
	cbr.sealed = true
}

// Sealed returns whether the CallbackRouter can be changed or not.
func (cbr CallbackRouter) Sealed() bool {
	return cbr.sealed
}

// AddRoute adds TSSCallback for a given module name. It returns the CallbackRouter
// so that the function can be chained. It will panic if the CallbackRouter is sealed.
func (cbr *CallbackRouter) AddRoute(module string, cbs TSSCallback) *CallbackRouter {
	if cbr.sealed {
		panic(fmt.Errorf("callback router sealed; cannot register %s route callbacks", module))
	}
	if !sdk.IsAlphaNumeric(module) {
		panic(errors.New("callback route expressions can only contain alphanumeric characters"))
	}
	if cbr.HasRoute(module) {
		panic(fmt.Errorf("route %s has already been registered", module))
	}

	cbr.routes[module] = cbs
	return cbr
}

// HasRoute returns whether the given module is registered.
func (cbr *CallbackRouter) HasRoute(module string) bool {
	_, ok := cbr.routes[module]
	return ok
}

// GetRoute returns a TSSCallback for a given module.
func (cbr *CallbackRouter) GetRoute(module string) (TSSCallback, bool) {
	if !cbr.HasRoute(module) {
		return nil, false
	}
	return cbr.routes[module], true
}

// TSSCallback defines the expected interface for a callback object that registered
// in the callbackRouter.
type TSSCallback interface {
	// Must be called when a group is created successfully.
	OnGroupCreationCompleted(ctx sdk.Context, groupID tss.GroupID)

	// Must be called after members fails to create a group.
	OnGroupCreationFailed(ctx sdk.Context, groupID tss.GroupID)

	// Must be called before setting group status to expired.
	OnGroupCreationExpired(ctx sdk.Context, groupID tss.GroupID)

	// Must be called after a signing request is unsuccessfully signed.
	OnSigningFailed(ctx sdk.Context, signingID tss.SigningID)

	// Must be called after a signing request is successfully signed by selected members.
	OnSigningCompleted(ctx sdk.Context, signingID tss.SigningID, assignedMembers []sdk.AccAddress)

	// Must be called before a retry that occurs due to expiration of the signing process.
	OnSigningTimeout(ctx sdk.Context, signingID tss.SigningID, idleMembers []sdk.AccAddress)
}
