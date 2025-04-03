package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewIBCHookMemo creates a new IBCHookMemo instance.
func NewIBCHookMemo(
	contract string,
	packet TunnelPricesPacketData,
) IBCHookMemo {
	return IBCHookMemo{
		Wasm: IBCHookMemo_Payload{
			Contract: contract,
			Msg: IBCHookMemo_Payload_Msg{
				ReceivePacket: IBCHookMemo_Payload_Msg_ReceivePacket{
					Packet: packet,
				},
			},
		},
	}
}

// JSONString returns the JSON string representation of the IBCHookMemo
func (r IBCHookMemo) JSONString() string {
	return string(sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&r)))
}
