package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// ModuleCdc references the global x/tunnel module codec. Note, the codec
// should ONLY be used in certain instances of tests and for JSON encoding.
//
// The actual codec used for serialization should be provided to x/tunnel and
// defined at the application level.
var ModuleCdc = codec.NewProtoCodec(codectypes.NewInterfaceRegistry())

// RegisterLegacyAminoCodec registers the necessary x/tunnel interfaces and concrete types
// on the provided LegacyAmino codec. These types are used for Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	legacy.RegisterAminoMsg(cdc, &MsgCreateTunnel{}, "tunnel/MsgCreateTunnel")
	legacy.RegisterAminoMsg(cdc, &MsgUpdateRoute{}, "tunnel/MsgUpdateRoute")
	legacy.RegisterAminoMsg(cdc, &MsgUpdateSignalsAndInterval{}, "tunnel/MsgUpdateSignalsAndInterval")
	legacy.RegisterAminoMsg(cdc, &MsgWithdrawFeePayerFunds{}, "tunnel/MsgWithdrawFeePayerFunds")
	legacy.RegisterAminoMsg(cdc, &MsgActivateTunnel{}, "tunnel/MsgActivateTunnel")
	legacy.RegisterAminoMsg(cdc, &MsgDeactivateTunnel{}, "tunnel/MsgDeactivateTunnel")
	legacy.RegisterAminoMsg(cdc, &MsgTriggerTunnel{}, "tunnel/MsgTriggerTunnel")
	legacy.RegisterAminoMsg(cdc, &MsgDepositToTunnel{}, "tunnel/MsgDepositToTunnel")
	legacy.RegisterAminoMsg(cdc, &MsgWithdrawFromTunnel{}, "tunnel/MsgWithdrawFromTunnel")
	legacy.RegisterAminoMsg(cdc, &MsgUpdateParams{}, "tunnel/MsgUpdateParams")

	cdc.RegisterInterface((*RouteI)(nil), nil)
	cdc.RegisterConcrete(&TSSRoute{}, "tunnel/TSSRoute", nil)
	cdc.RegisterConcrete(&IBCRoute{}, "tunnel/IBCRoute", nil)
	cdc.RegisterConcrete(&IBCHookRoute{}, "tunnel/IBCHookRoute", nil)
	cdc.RegisterConcrete(&RouterRoute{}, "tunnel/RouterRoute", nil)
	cdc.RegisterConcrete(&AxelarRoute{}, "tunnel/AxelarRoute", nil)

	cdc.RegisterInterface((*PacketReceiptI)(nil), nil)
	cdc.RegisterConcrete(&TSSPacketReceipt{}, "tunnel/TSSPacketReceipt", nil)
	cdc.RegisterConcrete(&IBCPacketReceipt{}, "tunnel/IBCPacketReceipt", nil)
	cdc.RegisterConcrete(&IBCHookPacketReceipt{}, "tunnel/IBCHookPacketReceipt", nil)
	cdc.RegisterConcrete(&RouterPacketReceipt{}, "tunnel/RouterPacketReceipt", nil)
	cdc.RegisterConcrete(&AxelarPacketReceipt{}, "tunnel/AxelarPacketReceipt", nil)

	cdc.RegisterConcrete(Params{}, "tunnel/Params", nil)
}

// RegisterInterfaces registers the x/tunnel interfaces types with the interface registry
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgCreateTunnel{},
		&MsgUpdateRoute{},
		&MsgUpdateSignalsAndInterval{},
		&MsgWithdrawFeePayerFunds{},
		&MsgActivateTunnel{},
		&MsgDeactivateTunnel{},
		&MsgTriggerTunnel{},
		&MsgDepositToTunnel{},
		&MsgWithdrawFromTunnel{},
		&MsgUpdateParams{},
	)

	registry.RegisterInterface(
		"tunnel.v1beta1.RouteI",
		(*RouteI)(nil),
		&TSSRoute{},
		&IBCRoute{},
		&IBCHookRoute{},
		&RouterRoute{},
		&AxelarRoute{},
	)

	registry.RegisterInterface(
		"tunnel.v1beta1.PacketReceiptI",
		(*PacketReceiptI)(nil),
		&TSSPacketReceipt{},
		&IBCPacketReceipt{},
		&IBCHookPacketReceipt{},
		&RouterPacketReceipt{},
		&AxelarPacketReceipt{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
