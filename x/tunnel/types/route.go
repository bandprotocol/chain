package types

import "github.com/cosmos/gogoproto/proto"

// Route represents the interface of various Route types implemented
// by other modules.
type Route interface {
	proto.Message

	ValidateBasic() error
	String() string
}
