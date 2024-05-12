package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Content defines an interface that a signature order must implement. It contains information
// such as the type and routing information for the appropriate handler to process the order.
// Content can have additional fields, which is handled by an order's Handler.
type Content interface {
	OrderRoute() string
	OrderType() string

	ValidateBasic() error
	String() string
}

// Handler defines a function that receive signature order and return message that should to be signed.
type Handler func(ctx sdk.Context, content Content) ([]byte, error)
