package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// ModuleCdc references the global x/oracle module codec. Note, the codec
// should ONLY be used in certain instances of tests and for JSON encoding.
//
// The actual codec used for serialization should be provided to x/oracle and
// defined at the application level.
var ModuleCdc = codec.NewProtoCodec(codectypes.NewInterfaceRegistry())

// RegisterLegacyAminoCodec registers the necessary x/oracle interfaces and concrete types
// on the provided LegacyAmino codec. These types are used for Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	legacy.RegisterAminoMsg(cdc, &MsgRequestData{}, "oracle/Request")
	legacy.RegisterAminoMsg(cdc, &MsgReportData{}, "oracle/Report")
	legacy.RegisterAminoMsg(cdc, &MsgCreateDataSource{}, "oracle/CreateDataSource")
	legacy.RegisterAminoMsg(cdc, &MsgEditDataSource{}, "oracle/EditDataSource")
	legacy.RegisterAminoMsg(cdc, &MsgCreateOracleScript{}, "oracle/CreateOracleScript")
	legacy.RegisterAminoMsg(cdc, &MsgEditOracleScript{}, "oracle/EditOracleScript")
	legacy.RegisterAminoMsg(cdc, &MsgActivate{}, "oracle/Activate")
	legacy.RegisterAminoMsg(cdc, &MsgUpdateParams{}, "oracle/UpdateParams")
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
		&MsgUpdateParams{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
