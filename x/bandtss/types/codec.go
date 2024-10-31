package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"

	tsstypes "github.com/bandprotocol/chain/v3/x/tss/types"
)

// RegisterLegacyAminoCodec registers the necessary x/bandtss interfaces and concrete types
// on the provided LegacyAmino codec. These types are used for Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	legacy.RegisterAminoMsg(cdc, &MsgTransitionGroup{}, "bandtss/MsgTransitionGroup")
	legacy.RegisterAminoMsg(cdc, &MsgForceTransitionGroup{}, "bandtss/MsgForceTransitionGroup")
	legacy.RegisterAminoMsg(cdc, &MsgRequestSignature{}, "bandtss/MsgRequestSignature")
	legacy.RegisterAminoMsg(cdc, &MsgActivate{}, "bandtss/MsgActivate")
	legacy.RegisterAminoMsg(cdc, &MsgHeartbeat{}, "bandtss/MsgHeartbeat")
	legacy.RegisterAminoMsg(cdc, &MsgUpdateParams{}, "bandtss/MsgUpdateParams")
}

// RegisterInterfaces registers the x/tss interfaces types with the interface registry
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgTransitionGroup{},
		&MsgForceTransitionGroup{},
		&MsgRequestSignature{},
		&MsgActivate{},
		&MsgHeartbeat{},
		&MsgUpdateParams{},
	)

	registry.RegisterImplementations(
		(*tsstypes.Content)(nil),
		&GroupTransitionSignatureOrder{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
