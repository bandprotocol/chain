package group

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/pkg/event"
	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

// Event represents the parsed information from a create_group event.
type Event struct {
	GroupID tss.GroupID
}

// ParseEvent parses the event from the given events.
// It extracts the event information from the events and returns the parsed Event or an error if parsing fails.
func ParseEvent(events sdk.StringEvents, evType string) (*Event, error) {
	gid, err := event.GetEventValueUint64(events, evType, types.AttributeKeyGroupID)
	if err != nil {
		return nil, err
	}

	return &Event{
		GroupID: tss.GroupID(gid),
	}, nil
}
