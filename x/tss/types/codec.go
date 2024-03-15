package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	legacy.RegisterAminoMsg(cdc, &MsgSubmitDKGRound1{}, "tss/SubmitDKGRound1")
	legacy.RegisterAminoMsg(cdc, &MsgSubmitDKGRound2{}, "tss/SubmitDKGRound2")
	legacy.RegisterAminoMsg(cdc, &MsgComplain{}, "tss/Complaint")
	legacy.RegisterAminoMsg(cdc, &MsgConfirm{}, "tss/Confirm")
	legacy.RegisterAminoMsg(cdc, &MsgSubmitDEs{}, "tss/SubmitDEs")
	legacy.RegisterAminoMsg(cdc, &MsgSubmitSignature{}, "tss/SubmitSignature")
	legacy.RegisterAminoMsg(cdc, &MsgUpdateParams{}, "tss/UpdateParams")

	cdc.RegisterConcrete(&TextSignatureOrder{}, "tss/TextSignatureOrder", nil)
}

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
