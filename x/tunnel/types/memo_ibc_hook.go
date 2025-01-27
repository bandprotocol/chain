package types

import (
	"encoding/json"

	feedstypes "github.com/bandprotocol/chain/v3/x/feeds/types"
)

// ReceiveBandData represents the data structure of the IBC hook message.
type ReceiveBandData struct {
	TunnelID  uint64             `json:"tunnel_id"`
	Sequence  uint64             `json:"sequence"`
	Prices    []feedstypes.Price `json:"prices"`
	CreatedAt int64              `json:"created_at"`
}

// IBCHookMsg represents the message structure of the IBC hook message.
type IBCHookMsg struct {
	ReceiveBandData ReceiveBandData `json:"receive_band_data"`
}

// IBCHookWasm represents the WASM contract and its associated message.
type IBCHookWasm struct {
	Contract string     `json:"contract"`
	Msg      IBCHookMsg `json:"msg"`
}

// IBCHookMemo represents the IBC hook memo structure.
type IBCHookMemo struct {
	Wasm IBCHookWasm `json:"wasm"`
}

// NewIBCHookMemo creates a new IBCHookMemo instance.
func NewIBCHookMemo(
	contract string,
	tunnelID uint64,
	sequence uint64,
	prices []feedstypes.Price,
	createdAt int64,
) IBCHookMemo {
	return IBCHookMemo{
		Wasm: IBCHookWasm{
			Contract: contract,
			Msg: IBCHookMsg{
				ReceiveBandData: ReceiveBandData{
					TunnelID:  tunnelID,
					Sequence:  sequence,
					Prices:    prices,
					CreatedAt: createdAt,
				},
			},
		},
	}
}

// String marshals the IBCHookMemo into a JSON string.
func (r IBCHookMemo) String() (string, error) {
	j, err := json.Marshal(r)
	if err != nil {
		return "", err
	}
	return string(j), nil
}
