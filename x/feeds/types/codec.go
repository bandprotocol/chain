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
	legacy.RegisterAminoMsg(cdc, &MsgSubmitSignals{}, "feeds/MsgSubmitSignals")
	legacy.RegisterAminoMsg(cdc, &MsgSubmitPrices{}, "feeds/MsgSubmitPrices")
	legacy.RegisterAminoMsg(cdc, &MsgUpdatePriceService{}, "feeds/MsgUpdatePriceService")
	legacy.RegisterAminoMsg(cdc, &MsgUpdateParams{}, "feeds/MsgUpdateParams")
}

// RegisterInterfaces register the feeds module interfaces to protobuf Any.
func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgSubmitSignals{},
		&MsgSubmitPrices{},
		&MsgUpdatePriceService{},
		&MsgUpdateParams{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	amino = codec.NewLegacyAmino()

	// ModuleCdc references the global x/feeds module codec. Note, the codec
	// should ONLY be used in certain instances of tests and for JSON encoding.
	//
	// The actual codec used for serialization should be provided to x/feeds and
	// defined at the application level
	ModuleCdc = codec.NewAminoCodec(amino)
)

func init() {
	RegisterLegacyAminoCodec(amino)
	sdk.RegisterLegacyAminoCodec(amino)
}