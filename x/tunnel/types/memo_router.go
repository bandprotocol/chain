package types

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewRouterMemo creates a new RouterMemo object.
func NewRouterMemo(
	contract string,
	destinationChainID string,
	destinationContractAddress string,
	gasLimit uint64,
	payload string,
) RouterMemo {
	return RouterMemo{
		Wasm: &RouterWasm{
			Contract: contract,
			Msg: &RouterMsg{
				ReceiveBandData: &RouterPacket{
					DestChainID:         destinationChainID,
					DestContractAddress: destinationContractAddress,
					GasLimit:            gasLimit,
					Payload:             payload,
				},
			},
		},
	}
}

// JSONString returns the JSON string representation of the RouterMemo
func (r RouterMemo) JSONString() (string, error) {
	bz, err := json.Marshal(r)
	if err != nil {
		return "", err
	}

	return string(sdk.MustSortJSON(bz)), nil
}
