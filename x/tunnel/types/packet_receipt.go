package types

import "github.com/cosmos/gogoproto/proto"

// PacketReceiptI defines an interface for confirming the delivery of a packet to its destination via the specified route.
type PacketReceiptI interface {
	proto.Message
}
