package de

import (
	"github.com/bandprotocol/chain/v2/pkg/event"
	"github.com/bandprotocol/chain/v2/x/tss/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Event represents the data structure for submit_sign events.
type Event struct {
	PubDE types.DE
}

// ParseSubmitSignEvent parses the submit_sign event from the given message log.
// It extracts the public D and E from the log and returns the parsed Event or an error if parsing fails.
func ParseSubmitSignEvent(log sdk.ABCIMessageLog) (*Event, error) {
	pubD, err := event.GetEventValueBytes(log, types.EventTypeSubmitSign, types.AttributeKeyPubD)
	if err != nil {
		return nil, err
	}

	pubE, err := event.GetEventValueBytes(log, types.EventTypeSubmitSign, types.AttributeKeyPubE)
	if err != nil {
		return nil, err
	}

	return &Event{
		PubDE: types.DE{
			PubD: pubD,
			PubE: pubE,
		},
	}, nil
}
