package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
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

	cdc.RegisterConcrete(&TextSignatureOrder{}, "tss/TextSignatureOrder", nil)
}

// RegisterInterfaces register the tss module interfaces to protobuf Any.
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
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
}

// RegisterRequestSignatureTypeCodec registers an external signature request type defined
// in another module for the internal ModuleCdc. This allows the MsgRequestSignature
// to be correctly Amino encoded and decoded.
//
// NOTE: This should only be used for applications that are still using a concrete
// Amino codec for serialization.
func RegisterSignatureOrderTypeCodec(o interface{}, name string) {
	amino.RegisterConcrete(o, name, nil)
}

var (
	amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewAminoCodec(amino)
)

func init() {
	RegisterLegacyAminoCodec(amino)
	sdk.RegisterLegacyAminoCodec(amino)
}
