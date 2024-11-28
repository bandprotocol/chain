package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// RegisterLegacyAminoCodec registers the necessary x/tss interfaces and concrete types
// on the provided LegacyAmino codec. These types are used for Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	legacy.RegisterAminoMsg(cdc, &MsgSubmitDKGRound1{}, "tss/MsgSubmitDKGRound1")
	legacy.RegisterAminoMsg(cdc, &MsgSubmitDKGRound2{}, "tss/MsgSubmitDKGRound2")
	legacy.RegisterAminoMsg(cdc, &MsgComplain{}, "tss/MsgComplaint")
	legacy.RegisterAminoMsg(cdc, &MsgConfirm{}, "tss/MsgConfirm")
	legacy.RegisterAminoMsg(cdc, &MsgSubmitDEs{}, "tss/MsgSubmitDEs")
	legacy.RegisterAminoMsg(cdc, &MsgSubmitSignature{}, "tss/MsgSubmitSignature")
	legacy.RegisterAminoMsg(cdc, &MsgUpdateParams{}, "tss/MsgUpdateParams")

	cdc.RegisterInterface((*Originator)(nil), nil)
	cdc.RegisterInterface((*Content)(nil), nil)

	cdc.RegisterConcrete(&TextSignatureOrder{}, "tss/TextSignatureOrder", nil)
	cdc.RegisterConcrete(&DirectOriginator{}, "tss/DirectOriginator", nil)
	cdc.RegisterConcrete(&TunnelOriginator{}, "tss/TunnelOriginator", nil)
	cdc.RegisterConcrete(Params{}, "tss/Params", nil)
}

// RegisterInterfaces registers the x/tss interfaces types with the interface registry
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgSubmitDKGRound1{},
		&MsgSubmitDKGRound2{},
		&MsgComplain{},
		&MsgConfirm{},
		&MsgSubmitDEs{},
		&MsgSubmitSignature{},
		&MsgUpdateParams{},
	)

	registry.RegisterInterface(
		"tss.v1beta1.Content",
		(*Content)(nil),
		&TextSignatureOrder{},
	)

	registry.RegisterInterface(
		"tss.v1beta1.Originator",
		(*Originator)(nil),
		&DirectOriginator{}, &TunnelOriginator{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
