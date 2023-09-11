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

// ParseEvent parses the request-sign event from the given events.
// It extracts the signing information from the events and returns the parsed Event or an error if parsing fails.
func ParseEvent(events sdk.StringEvents) (*Event, error) {
	sid, err := event.GetEventValueUint64(events, types.EventTypeRequestSignature, types.AttributeKeySigningID)
	if err != nil {
		return nil, err
	}

	return &Event{
		SigningID: tss.SigningID(sid),
	}, nil
}
