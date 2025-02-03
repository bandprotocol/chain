package types

import "encoding/json"

type AxelarMessageType int

const (
	// AxelarMessageTypeUnrecognized means coin type is unrecognized by axelar
	AxelarMessageTypeUnrecognized AxelarMessageType = iota
	// AxelarMessageTypeGeneralMessage is a pure axelar message
	AxelarMessageTypeGeneralMessage
	// AxelarMessageTypeGeneralMessageWithToken is a general axelar message with token
	AxelarMessageTypeGeneralMessageWithToken
	// AxelarMessageTypeSendToken is a direct token transfer
	AxelarMessageTypeSendToken
)

// AxelarFee is used to pay relayer for executing cross chain message
type AxelarFee struct {
	Amount          string  `json:"amount"`
	Recipient       string  `json:"recipient"`
	RefundRecipient *string `json:"refund_recipient"`
}

// NewAxelarFee creates a new AxelarFee instance.
func NewAxelarFee(
	amount string,
	recipient string,
	refundRecipient *string,
) AxelarFee {
	return AxelarFee{
		Amount:          amount,
		Recipient:       recipient,
		RefundRecipient: refundRecipient,
	}
}

// AxelarMemo is attached in ICS20 packet memo field for axelar cross chain message
type AxelarMemo struct {
	DestinationChain   string            `json:"destination_chain"`
	DestinationAddress string            `json:"destination_address"`
	Payload            []byte            `json:"payload"`
	Type               AxelarMessageType `json:"type"`
	Fee                *AxelarFee        `json:"fee"` // Optional
}

// NewAxelarMemo creates a new AxelarMemo instance.
func NewAxelarMemo(
	destinationChain string,
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
	return string(j), nil
}
