package types

import "github.com/cosmos/gogoproto/proto"

// PacketI defines a type that implements the Packet interface
type PacketContentI interface {
	proto.Message
}
