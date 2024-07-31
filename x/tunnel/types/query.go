package types

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	proto "github.com/cosmos/gogoproto/proto"
)

var (
	_ codectypes.UnpackInterfacesMessage = &QueryPacketsResponse{}
	_ codectypes.UnpackInterfacesMessage = &QueryPacketResponse{}
)

// Packet defines a type that implements the Packet interface
type Packet interface {
	proto.Message
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (m *QueryPacketsResponse) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	for _, p := range m.Packets {
		var packet Packet
		if err := unpacker.UnpackAny(p, &packet); err != nil {
			return err
		}
	}
	return nil
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (m *QueryPacketResponse) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	var packet Packet
	return unpacker.UnpackAny(m.Packet, &packet)
}
