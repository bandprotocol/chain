package replacer

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/pkg/event"
	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

// Event represents the data structure for replace_success events.
type Event struct {
	FromGroupID tss.GroupID
	ToGroupID   tss.GroupID
}

// ParseEvent parses the replace_success event from the given events.
// It extracts the replacement information from the events and returns the parsed Event or an error if parsing fails.
func ParseEvent(events sdk.StringEvents) (*Event, error) {
	fromGroupID, err := event.GetEventValueUint64(events, types.EventTypeReplaceSuccess, types.AttributeKeyFromGroupID)
	if err != nil {
		return nil, err
	}

	toGroupID, err := event.GetEventValueUint64(events, types.EventTypeReplaceSuccess, types.AttributeKeyToGroupID)
	if err != nil {
		return nil, err
	}

	return &Event{
		FromGroupID: tss.GroupID(fromGroupID),
		ToGroupID:   tss.GroupID(toGroupID),
	}, nil
}
