package de

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/pkg/event"
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

// Event represents the data structure for submit_sign events.
type Event struct {
	PubDE types.DE
}

// ParseSubmitSignEvent parses the submit_sign event from the given events.
// It extracts the public D and E from the log and returns the parsed Event or an error if parsing fails.
func ParseSubmitSignEvent(events sdk.StringEvents) (*Event, error) {
	pubD, err := event.GetEventValueBytes(events, types.EventTypeSubmitSign, types.AttributeKeyPubD)
	if err != nil {
		return nil, err
	}

	pubE, err := event.GetEventValueBytes(events, types.EventTypeSubmitSign, types.AttributeKeyPubE)
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
