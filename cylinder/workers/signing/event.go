package signing

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/pkg/event"
	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

// Event represents the data structure for request-sign events.
type Event struct {
	SigningID tss.SigningID
}

// ParseEvents parses the request-sign events from the given events.
// It extracts the signing information from the events and returns the parsed Events or an error if parsing fails.
func ParseEvents(events sdk.StringEvents) ([]Event, error) {
	sids, err := event.GetEventValuesUint64(events, types.EventTypeRequestSignature, types.AttributeKeySigningID)
	if err != nil {
		return nil, err
	}

	var eves []Event
	for _, sid := range sids {
		eves = append(eves, Event{
			SigningID: tss.SigningID(sid),
		})
	}

	return eves, nil
}
