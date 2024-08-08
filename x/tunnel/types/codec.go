package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// RegisterLegacyAminoCodec registers the necessary x/tunnel interfaces and concrete types
// on the provided LegacyAmino codec. These types are used for Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	legacy.RegisterAminoMsg(cdc, &MsgUpdateParams{}, "tunnel/MsgUpdateParams")
	legacy.RegisterAminoMsg(cdc, &MsgCreateTunnel{}, "tunnel/MsgCreateTunnel")
	legacy.RegisterAminoMsg(cdc, &MsgActivateTunnel{}, "tunnel/MsgActivateTunnel")

	cdc.RegisterInterface((*Route)(nil), nil)
	cdc.RegisterConcrete(&TSSRoute{}, "tunnel/TSSRoute", nil)
	cdc.RegisterConcrete(&AxelarRoute{}, "tunnel/AxelarRoute", nil)
	cdc.RegisterConcrete(&IBCRoute{}, "tunnel/IBCRoute", nil)

	cdc.RegisterInterface((*Packet)(nil), nil)
}

// RegisterInterfaces registers the x/tunnel interfaces types with the interface registry
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgUpdateParams{},
		&MsgCreateTunnel{},
		&MsgActivateTunnel{},
	)

	registry.RegisterInterface(
		"tunnel.v1beta1.Route",
		(*Route)(nil),
		&TSSRoute{},
		&AxelarRoute{},
		&IBCRoute{},
	)

	registry.RegisterInterface(
		"tunnel.v1beta1.Packet",
		(*Packet)(nil),
		&TSSPacket{},
		&AxelarPacket{},
		&IBCPacket{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	amino = codec.NewLegacyAmino()

	// ModuleCdc references the global x/tunnel module codec. Note, the codec
	// should ONLY be used in certain instances of tests and for JSON encoding.
	//
	// The actual codec used for serialization should be provided to x/tunnel and
	// defined at the application level.
	ModuleCdc = codec.NewProtoCodec(codectypes.NewInterfaceRegistry())

	// AminoCdc is a amino codec created to support amino json compatible msgs.
	AminoCdc = codec.NewAminoCodec(amino)
)

func init() {
	RegisterLegacyAminoCodec(amino)
	cryptocodec.RegisterCrypto(amino)
	amino.Seal()
}
