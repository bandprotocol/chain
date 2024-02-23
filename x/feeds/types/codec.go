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
	legacy.RegisterAminoMsg(cdc, &MsgSignalSymbols{}, "feeds/MsgSignalSymbols")
	legacy.RegisterAminoMsg(cdc, &MsgSubmitPrices{}, "feeds/MsgSubmitPrices")
	legacy.RegisterAminoMsg(cdc, &MsgUpdatePriceService{}, "feeds/MsgUpdatePriceService")
	legacy.RegisterAminoMsg(cdc, &MsgUpdateParams{}, "feeds/MsgUpdateParams")
}

func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgSignalSymbols{},
		&MsgSubmitPrices{},
		&MsgUpdatePriceService{},
		&MsgUpdateParams{},
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
