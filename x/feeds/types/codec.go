package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// RegisterLegacyAminoCodec registers concrete types on the LegacyAmino codec
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	legacy.RegisterAminoMsg(cdc, &MsgVote{}, "feeds/MsgVote")
	legacy.RegisterAminoMsg(cdc, &MsgSubmitSignalPrices{}, "feeds/MsgSubmitSignalPrices")
	legacy.RegisterAminoMsg(cdc, &MsgUpdateReferenceSourceConfig{}, "feeds/MsgUpdateReferenceSourceConfig")
	legacy.RegisterAminoMsg(cdc, &MsgUpdateParams{}, "feeds/MsgUpdateParams")
}

// RegisterInterfaces register the feeds module interfaces to protobuf Any.
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgVote{},
		&MsgSubmitSignalPrices{},
		&MsgUpdateReferenceSourceConfig{},
		&MsgUpdateParams{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
