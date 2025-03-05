package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewIBCHookMemo creates a new IBCHookMemo instance.
func NewIBCHookMemo(
	contract string,
	packet TunnelPricesPacketData,
) *IBCHookMemo {
	return &IBCHookMemo{
		Wasm: &IBCHookWasm{
			Contract: contract,
			Msg: &IBCHookMsg{
				ReceivePacket: &IBCHookPacket{
					Packet: &packet,
				},
			},
		},
	}
}

// JSONString returns the JSON string representation of the IBCHookMemo
func (r IBCHookMemo) JSONString() string {
	return string(sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&r)))
}
