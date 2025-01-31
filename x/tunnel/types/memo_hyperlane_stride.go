package types

import "encoding/json"

// Dispatch represents the dispatch details in the Hyperlane stride message.
type Dispatch struct {
	DestDomain    uint64 `json:"dest_domain"`
	RecipientAddr string `json:"recipient_addr"`
	MsgBody       string `json:"msg_body"`
}

// HyperlaneStrideMsg represents the message structure containing dispatch details.
type HyperlaneStrideMsg struct {
	Dispatch Dispatch `json:"dispatch"`
}

// HyperlaneStrideWasm represents the WASM contract and its associated message.
type HyperlaneStrideWasm struct {
	Contract string             `json:"contract"`
	Msg      HyperlaneStrideMsg `json:"msg"`
}

// HyperlaneStrideMemo represents the HyperlaneStride memo structure.
type HyperlaneStrideMemo struct {
	Wasm HyperlaneStrideWasm `json:"wasm"`
}

// NewHyperlaneStrideMemo creates a new HyperlaneStrideMemo with the provided contract, destination domain, recipient address, and message body.
func NewHyperlaneStrideMemo(
	contract string,
	destDomain uint64,
	recipientAddr string,
	msgBody string,
) HyperlaneStrideMemo {
	return HyperlaneStrideMemo{
		Wasm: HyperlaneStrideWasm{
			Contract: contract,
			Msg: HyperlaneStrideMsg{
				Dispatch: Dispatch{
					DestDomain:    destDomain,
					RecipientAddr: recipientAddr,
					MsgBody:       msgBody,
				},
			},
		},
	}
}

// String marshals the HyperlaneStrideMemo into a JSON string.
func (r HyperlaneStrideMemo) String() (string, error) {
	j, err := json.Marshal(r)
	if err != nil {
		return "", err
	}
	return string(j), nil
}
