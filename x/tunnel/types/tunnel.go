package types

import (
	fmt "fmt"

	"github.com/cosmos/cosmos-sdk/codec/types"
	proto "github.com/cosmos/gogoproto/proto"
)

var _ types.UnpackInterfacesMessage = Tunnel{}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (t Tunnel) UnpackInterfaces(unpacker types.AnyUnpacker) error {
	var route RouteI
	return unpacker.UnpackAny(t.Route, &route)
}

// SetRoute sets the route of the tunnel.
func (t *Tunnel) SetRoute(route RouteI) error {
	msg, ok := route.(proto.Message)
	if !ok {
		return fmt.Errorf("can't proto marshal %T", msg)
	}
	any, err := types.NewAnyWithValue(msg)
	if err != nil {
		return err
	}
	t.Route = any

	return nil
}

// createPacket creates a new packet for the tunnel
func (t Tunnel) CreatePacket(createdAt int64) Packet {
	return NewPacket(
		t.ID,
		t.NonceCount,
		t.FeedType,
		t.SignalPriceInfos,
		createdAt,
	)
}

// IsTunnelInList checks if a tunnel with the given ID is in the list of tunnels.
func IsTunnelInList(id uint64, tunnels []Tunnel) bool {
	for _, tunnel := range tunnels {
		if tunnel.ID == id {
			return true
		}
	}
	return false
}
