package de

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/pkg/event"
	"github.com/bandprotocol/chain/v3/x/tss/types"
)

// Event represents the data structure for submit_sign events.
type Event struct {
	PubDE types.DE
}

// ParseSubmitSignEvents parses the submit_sign events from the given events.
// It extracts the public D and E from the log and returns the parsed Events or an error if parsing fails.
func ParseSubmitSignEvents(events sdk.StringEvents) ([]Event, error) {
	pubDs, err := event.GetEventValuesBytes(events, types.EventTypeSubmitSignature, types.AttributeKeyPubD)
	if err != nil {
		return nil, err
	}

	pubEs, err := event.GetEventValuesBytes(events, types.EventTypeSubmitSignature, types.AttributeKeyPubE)
	if err != nil {
		return nil, err
	}

	if len(pubDs) != len(pubEs) {
		return nil, errors.New("length of public D and e are not equal")
	}

	var eves []Event
	for i, pubD := range pubDs {
		eves = append(eves, Event{
			PubDE: types.DE{
				PubD: pubD,
				PubE: pubEs[i],
			},
		})
	}

	return eves, nil
}
