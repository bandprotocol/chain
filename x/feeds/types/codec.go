package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"

	tsstypes "github.com/bandprotocol/chain/v3/x/tss/types"
)

// RegisterLegacyAminoCodec registers concrete types on the LegacyAmino codec
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	legacy.RegisterAminoMsg(cdc, &MsgSubmitSignals{}, "feeds/MsgSubmitSignals")
	legacy.RegisterAminoMsg(cdc, &MsgSubmitSignalPrices{}, "feeds/MsgSubmitSignalPrices")
	legacy.RegisterAminoMsg(cdc, &MsgUpdateReferenceSourceConfig{}, "feeds/MsgUpdateReferenceSourceConfig")
	legacy.RegisterAminoMsg(cdc, &MsgUpdateParams{}, "feeds/MsgUpdateParams")

	cdc.RegisterConcrete(&FeedsSignatureOrder{}, "feeds/FeedsSignatureOrder", nil)
}

// RegisterInterfaces register the feeds module interfaces to protobuf Any.
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgSubmitSignals{},
		&MsgSubmitSignalPrices{},
		&MsgUpdateReferenceSourceConfig{},
		&MsgUpdateParams{},
	)

	registry.RegisterImplementations(
		(*tsstypes.Content)(nil),
		&FeedsSignatureOrder{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
