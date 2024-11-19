package types

import "github.com/cosmos/gogoproto/proto"

// PacketContentI defines the interface for packet content.
type PacketContentI interface {
	proto.Message
}
