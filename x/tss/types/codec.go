package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgCreateGroup{}, "tss/CreateGroup", nil)
	cdc.RegisterConcrete(&MsgSubmitDKGRound1{}, "tss/SubmitDKGRound1", nil)
	cdc.RegisterConcrete(&MsgSubmitDKGRound2{}, "tss/SubmitDKGRound2", nil)
	cdc.RegisterConcrete(&MsgComplain{}, "tss/Complaint", nil)
	cdc.RegisterConcrete(&MsgConfirm{}, "tss/Confirm", nil)
	cdc.RegisterConcrete(&MsgSubmitDEs{}, "tss/SubmitDEs", nil)
	cdc.RegisterConcrete(&MsgRequestSignature{}, "tss/RequestSign", nil)
	cdc.RegisterConcrete(&MsgSubmitSignature{}, "tss/SubmitSignature", nil)
	cdc.RegisterConcrete(&MsgActivate{}, "tss/Activate", nil)
	cdc.RegisterConcrete(&DefaultRequestSignature{}, "tss/DefaultRequestSignature", nil)
	cdc.RegisterConcrete(&MsgActive{}, "tss/Active", nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgCreateGroup{},
		&MsgSubmitDKGRound1{},
		&MsgSubmitDKGRound2{},
		&MsgComplain{},
		&MsgConfirm{},
		&MsgSubmitDEs{},
		&MsgRequestSignature{},
		&MsgSubmitSignature{},
		&MsgActivate{},
		&MsgActive{},
	)
	registry.RegisterInterface(
		"tss.v1beta1.Content",
		(*Content)(nil),
		&DefaultRequestSignature{},
	)
}

// RegisterRequestSignatureTypeCodec registers an external request signature content type defined
// in another module for the internal ModuleCdc. This allows the MsgRequestSignature
// to be correctly Amino encoded and decoded.
//
// NOTE: This should only be used for applications that are still using a concrete
// Amino codec for serialization.
func RegisterRequestSignatureTypeCodec(o interface{}, name string) {
	amino.RegisterConcrete(o, name, nil)
}

var (
	amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)

func init() {
	RegisterLegacyAminoCodec(amino)
	sdk.RegisterLegacyAminoCodec(amino)
}
