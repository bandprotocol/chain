package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	tsstypes "github.com/bandprotocol/chain/v2/x/tss/types"
)

// RegisterLegacyAminoCodec registers the necessary x/bandtss interfaces and concrete types
// on the provided LegacyAmino codec. These types are used for Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	legacy.RegisterAminoMsg(cdc, &MsgTransitionGroup{}, "bandtss/MsgTransitionGroup")
	legacy.RegisterAminoMsg(cdc, &MsgForceReplaceGroup{}, "bandtss/MsgForceReplaceGroup")
	legacy.RegisterAminoMsg(cdc, &MsgRequestSignature{}, "bandtss/MsgRequestSignature")
	legacy.RegisterAminoMsg(cdc, &MsgActivate{}, "bandtss/MsgActivate")
	legacy.RegisterAminoMsg(cdc, &MsgHeartbeat{}, "bandtss/MsgHeartbeat")
	legacy.RegisterAminoMsg(cdc, &MsgUpdateParams{}, "bandtss/MsgUpdateParams")
}

// RegisterInterfaces register the bandtss module interfaces to protobuf Any.
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgTransitionGroup{},
		&MsgForceReplaceGroup{},
		&MsgRequestSignature{},
		&MsgActivate{},
		&MsgHeartbeat{},
		&MsgUpdateParams{},
	)

	registry.RegisterImplementations(
		(*tsstypes.Content)(nil),
		&GroupTransitionSignatureOrder{},
	)
}

var (
	amino = codec.NewLegacyAmino()

	// ModuleCdc references the global x/bandtss module codec. Note, the codec
	// should ONLY be used in certain instances of tests and for JSON encoding.
	//
	// The actual codec used for serialization should be provided to x/bandtss and
	// defined at the application level
	ModuleCdc = codec.NewAminoCodec(amino)
)

func init() {
	RegisterLegacyAminoCodec(amino)
	sdk.RegisterLegacyAminoCodec(amino)
}
