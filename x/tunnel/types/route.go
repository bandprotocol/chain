package types

import (
	"github.com/cosmos/gogoproto/proto"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RouteI defines a routing path to deliver data to the destination.
type RouteI interface {
	proto.Message

	ValidateBasic() error
	Fee() (sdk.Coins, error)
}
