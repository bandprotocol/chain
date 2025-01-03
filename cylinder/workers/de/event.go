package de

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/pkg/event"
	"github.com/bandprotocol/chain/v3/x/tss/types"
)

// PubDE represents the data structure for public D,E being retrieved from events.
type PubDE struct {
	PubDE types.DE
}

// ParsePubDEFromEvents parses the events into PubDE struct from the given events and event type.
// It extracts the public D and E from the log and returns the parsed Events or an error if parsing fails.
func ParsePubDEFromEvents(events sdk.StringEvents, eventType string) ([]PubDE, error) {
	pubDs, err := event.GetEventValuesBytes(events, eventType, types.AttributeKeyPubD)
	if err != nil {
		return nil, err
	}

	pubEs, err := event.GetEventValuesBytes(events, eventType, types.AttributeKeyPubE)
	if err != nil {
		return nil, err
	}

	if len(pubDs) != len(pubEs) {
		return nil, errors.New("length of public D and e are not equal")
	}

	var pubDEs []PubDE
	for i, pubD := range pubDs {
		pubDEs = append(pubDEs, PubDE{
			PubDE: types.DE{
				PubD: pubD,
				PubE: pubEs[i],
			},
		})
	}

	return pubDEs, nil
}
