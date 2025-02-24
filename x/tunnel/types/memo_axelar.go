package types

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// AxelarMemo is attached in ICS20 packet memo field for axelar cross chain message
type AxelarMemo struct {
	DestinationChain   ChainName         `json:"destination_chain"`
	DestinationAddress string            `json:"destination_address"`
	Payload            WasmBytes         `json:"payload"`
	Type               AxelarMessageType `json:"type"`
	Fee                *AxelarFee        `json:"fee"` // Optional
}

// NewAxelarMemo creates a new AxelarMemo instance.
func NewAxelarMemo(
	destinationChain ChainName,
	destinationAddress string,
	payload []byte,
	messageType AxelarMessageType,
	fee *AxelarFee,
) AxelarMemo {
	return AxelarMemo{
		DestinationChain:   destinationChain,
		DestinationAddress: destinationAddress,
		Payload:            payload,
		Type:               messageType,
		Fee:                fee,
	}
}

// String marshals the AxelarMemo into a JSON string.
func (r AxelarMemo) String() (string, error) {
	j, err := json.Marshal(r)
	if err != nil {
		return "", err
	}

	sj, err := sdk.SortJSON(j)
	if err != nil {
		return "", err
	}

	return string(sj), nil
}
