package parser

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/pkg/event"
	"github.com/bandprotocol/chain/v3/pkg/tss"
	"github.com/bandprotocol/chain/v3/x/tss/types"
)

// RequestSignatureEvent represents the data structure for request-sign events.
type RequestSignatureEvent struct {
	SigningID tss.SigningID
}

// ParseRequestSignatureEvents parses the request-sign events from the given events.
// It extracts the signing information from the events and returns the parsed Events or an error if parsing fails.
func ParseRequestSignatureEvents(events sdk.StringEvents) ([]RequestSignatureEvent, error) {
	sids, err := event.GetEventValuesUint64(events, types.EventTypeRequestSignature, types.AttributeKeySigningID)
	if err != nil {
		return nil, err
	}

	var eves []RequestSignatureEvent
	for _, sid := range sids {
		eves = append(eves, RequestSignatureEvent{
			SigningID: tss.SigningID(sid),
		})
	}

	return eves, nil
}
