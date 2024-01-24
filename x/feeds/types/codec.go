package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// RegisterLegacyAminoCodec registers concrete types on the LegacyAmino codec
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	legacy.RegisterAminoMsg(cdc, &MsgUpdateSymbols{}, "feeds/MsgUpdateSymbols")
	legacy.RegisterAminoMsg(cdc, &MsgRemoveSymbols{}, "feeds/MsgRemoveSymbols")
	legacy.RegisterAminoMsg(cdc, &MsgSubmitPrices{}, "feeds/MsgSubmitPrices")
	legacy.RegisterAminoMsg(cdc, &MsgUpdatePriceService{}, "feeds/MsgUpdatePriceService")
}

func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgUpdateSymbols{},
		&MsgRemoveSymbols{},
		&MsgSubmitPrices{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewAminoCodec(amino)
)

func init() {
	RegisterLegacyAminoCodec(amino)
	sdk.RegisterLegacyAminoCodec(amino)
}
