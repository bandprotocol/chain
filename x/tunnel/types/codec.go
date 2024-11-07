package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// ModuleCdc references the global x/tunnel module codec. Note, the codec
// should ONLY be used in certain instances of tests and for JSON encoding.
//
// The actual codec used for serialization should be provided to x/tunnel and
// defined at the application level.
var ModuleCdc = codec.NewProtoCodec(codectypes.NewInterfaceRegistry())

// RegisterLegacyAminoCodec registers the necessary x/tunnel interfaces and concrete types
// on the provided LegacyAmino codec. These types are used for Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	legacy.RegisterAminoMsg(cdc, &MsgCreateTunnel{}, "tunnel/MsgCreateTunnel")
	legacy.RegisterAminoMsg(cdc, &MsgEditTunnel{}, "tunnel/MsgEditTunnel")
	legacy.RegisterAminoMsg(cdc, &MsgActivate{}, "tunnel/MsgActivate")
	legacy.RegisterAminoMsg(cdc, &MsgDeactivate{}, "tunnel/MsgDeactivate")
	legacy.RegisterAminoMsg(cdc, &MsgTriggerTunnel{}, "tunnel/MsgTriggerTunnel")
	legacy.RegisterAminoMsg(cdc, &MsgDepositTunnel{}, "tunnel/MsgDepositTunnel")
	legacy.RegisterAminoMsg(cdc, &MsgWithdrawTunnel{}, "tunnel/MsgWithdrawTunnel")
	legacy.RegisterAminoMsg(cdc, &MsgUpdateParams{}, "tunnel/MsgUpdateParams")

	cdc.RegisterInterface((*RouteI)(nil), nil)
	cdc.RegisterConcrete(&TSSRoute{}, "tunnel/TSSRoute", nil)
	cdc.RegisterInterface((*PacketContentI)(nil), nil)
	cdc.RegisterConcrete(&TSSPacketContent{}, "tunnel/TSSPacketContent", nil)
}

// RegisterInterfaces registers the x/tunnel interfaces types with the interface registry
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgCreateTunnel{},
		&MsgEditTunnel{},
		&MsgActivate{},
		&MsgDeactivate{},
		&MsgTriggerTunnel{},
		&MsgDepositTunnel{},
		&MsgWithdrawTunnel{},
		&MsgUpdateParams{},
	)

	registry.RegisterInterface(
		"tunnel.v1beta1.RouteI",
		(*RouteI)(nil),
		&TSSRoute{},
	)

	registry.RegisterInterface(
		"tunnel.v1beta1.PacketContentI",
		(*PacketContentI)(nil),
		&TSSPacketContent{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
