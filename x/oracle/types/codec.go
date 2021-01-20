package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// RegisterCodec registers the module's concrete types on the codec.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(MsgRequestData{}, "oracle/Request", nil)
	cdc.RegisterConcrete(MsgReportData{}, "oracle/Report", nil)
	cdc.RegisterConcrete(MsgCreateDataSource{}, "oracle/CreateDataSource", nil)
	cdc.RegisterConcrete(MsgEditDataSource{}, "oracle/EditDataSource", nil)
	cdc.RegisterConcrete(MsgCreateOracleScript{}, "oracle/CreateOracleScript", nil)
	cdc.RegisterConcrete(MsgEditOracleScript{}, "oracle/EditOracleScript", nil)
	cdc.RegisterConcrete(MsgActivate{}, "oracle/Activate", nil)
	cdc.RegisterConcrete(MsgAddReporter{}, "oracle/AddReporter", nil)
	cdc.RegisterConcrete(MsgRemoveReporter{}, "oracle/RemoveReporter", nil)
	cdc.RegisterConcrete(OracleRequestPacketData{}, "oracle/OracleRequestPacketData", nil)
	cdc.RegisterConcrete(OracleResponsePacketData{}, "oracle/OracleResponsePacketData", nil)
}

func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgRequestData{},
		&MsgReportData{},
		&MsgCreateDataSource{},
		&MsgEditDataSource{},
		&MsgCreateOracleScript{},
		&MsgEditOracleScript{},
		&MsgActivate{},
		&MsgAddReporter{},
		&MsgAddReporter{},
		&MsgRemoveReporter{},
		&OracleRequestPacketData{},
		&OracleResponsePacketData{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var amino = codec.NewLegacyAmino()
var ModuleCdc = codec.NewAminoCodec(amino)

func init() {
	RegisterLegacyAminoCodec(amino)
	amino.Seal()
}
