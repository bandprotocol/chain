package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// RegisterInterfaces register the ibc transfer module interfaces to protobuf
// Any.
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil), &MsgRequestData{})
	registry.RegisterImplementations((*sdk.Msg)(nil), &MsgReportData{})
	registry.RegisterImplementations((*sdk.Msg)(nil), &MsgCreateDataSource{})
	registry.RegisterImplementations((*sdk.Msg)(nil), &MsgEditDataSource{})
	registry.RegisterImplementations((*sdk.Msg)(nil), &MsgCreateOracleScript{})
	registry.RegisterImplementations((*sdk.Msg)(nil), &MsgEditOracleScript{})
	registry.RegisterImplementations((*sdk.Msg)(nil), &MsgActivate{})
	registry.RegisterImplementations((*sdk.Msg)(nil), &MsgAddReporter{})
	registry.RegisterImplementations((*sdk.Msg)(nil), &MsgRemoveReporter{})
	// cdc.RegisterConcrete(OracleRequestPacketData{}, "oracle/OracleRequestPacketData", nil)
	// cdc.RegisterConcrete(OracleResponsePacketData{}, "oracle/OracleResponsePacketData", nil)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	// ModuleCdc references the global x/ibc-transfer module codec. Note, the codec
	// should ONLY be used in certain instances of tests and for JSON encoding.
	//
	// The actual codec used for serialization should be provided to x/ibc-transfer and
	// defined at the application level.
	ModuleCdc = codec.NewProtoCodec(codectypes.NewInterfaceRegistry())
)
