package types

import "encoding/json"

// RouterReceiveBandData represents the payload of the Router message.
type RouterReceiveBandData struct {
	DestChainID         string `json:"dest_chain_id"`
	DestContractAddress string `json:"dest_contract_address"`
	GasLimit            uint64 `json:"gas_limit"`
	GasPrice            uint64 `json:"gas_price"`
	Payload             string `json:"payload"`
}

// RouterMsg represents the wasm contract call message.
type RouterMsg struct {
	ReceiveBandData RouterReceiveBandData `json:"receive_band_data"`
}

// RouterWasm represents the WASM contract and its associated message.
type RouterWasm struct {
	Contract string    `json:"contract"`
	Msg      RouterMsg `json:"msg"`
}

// RouterMemo represents the Router memo structure.
type RouterMemo struct {
	Wasm RouterWasm `json:"wasm"`
}

// NewRouterMemo creates a new RouterMemo object.
func NewRouterMemo(
	contract string,
	destinationChainID string,
	destinationContractAddress string,
	gasLimit uint64,
	gasPrice uint64,
	payload string,
) RouterMemo {
	return RouterMemo{
		Wasm: RouterWasm{
			Contract: contract,
			Msg: RouterMsg{
				ReceiveBandData: RouterReceiveBandData{
					DestChainID:         destinationChainID,
					DestContractAddress: destinationContractAddress,
					GasLimit:            gasLimit,
					GasPrice:            gasPrice,
					Payload:             payload,
				},
			},
		},
	}
}

// String marshals the RouterMemo into a JSON string.
func (r RouterMemo) String() (string, error) {
	j, err := json.Marshal(r)
	if err != nil {
		return "", err
	}
	return string(j), nil
}
