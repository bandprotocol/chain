package types

import "github.com/cosmos/gogoproto/proto"

// Route defines a routing path to deliver data to the destination.
type Route interface {
	proto.Message

	ValidateBasic() error
	String() string
}
