package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// RegisterLegacyAminoCodec registers the necessary x/oracle interfaces and concrete types
// on the provided LegacyAmino codec. These types are used for Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgRequestData{}, "oracle/Request", nil)
	cdc.RegisterConcrete(&MsgReportData{}, "oracle/Report", nil)
	cdc.RegisterConcrete(&MsgCreateDataSource{}, "oracle/CreateDataSource", nil)
	cdc.RegisterConcrete(&MsgEditDataSource{}, "oracle/EditDataSource", nil)
	cdc.RegisterConcrete(&MsgCreateOracleScript{}, "oracle/CreateOracleScript", nil)
	cdc.RegisterConcrete(&MsgEditOracleScript{}, "oracle/EditOracleScript", nil)
	cdc.RegisterConcrete(&MsgActivate{}, "oracle/Activate", nil)
	cdc.RegisterConcrete(&MsgAddReporter{}, "oracle/AddReporter", nil)
	cdc.RegisterConcrete(&MsgRemoveReporter{}, "oracle/RemoveReporter", nil)
	// cdc.RegisterConcrete(OracleRequestPacketData{}, "oracle/OracleRequestPacketData", nil)
	// cdc.RegisterConcrete(OracleResponsePacketData{}, "oracle/OracleResponsePacketData", nil)
}

// RegisterInterfaces register the oracle module interfaces to protobuf Any.
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgRequestData{},
		&MsgReportData{},
		&MsgCreateDataSource{},
		&MsgEditDataSource{},
		&MsgCreateOracleScript{},
		&MsgEditOracleScript{},
		&MsgActivate{},
		&MsgAddReporter{},
		&MsgRemoveReporter{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	amino = codec.NewLegacyAmino()

	// ModuleCdc references the global x/oracle module codec. Note, the codec
	// should ONLY be used in certain instances of tests and for JSON encoding.
	//
	// The actual codec used for serialization should be provided to x/oracle and
	// defined at the application level.
	ModuleCdc = codec.NewProtoCodec(codectypes.NewInterfaceRegistry())

	// AminoCdc is a amino codec created to support amino json compatible msgs.
	AminoCdc = codec.NewAminoCodec(amino)
)

func init() {
	RegisterLegacyAminoCodec(amino)
	cryptocodec.RegisterCrypto(amino)
	amino.Seal()
}
