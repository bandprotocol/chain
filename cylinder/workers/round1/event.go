package round1

import (
	"encoding/hex"
	"errors"
	"strconv"

	"github.com/bandprotocol/chain/v2/pkg/event"
	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/tss/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Event represents the parsed information from a create_group event.
type Event struct {
	GroupID    tss.GroupID
	Threshold  uint64
	DKGContext []byte
	Members    []string
}

// ParseEvent parses the ABCIMessageLog and extracts the relevant information to create a create_group event.
// It returns the parsed Event or an error if parsing fails.
func ParseEvent(log sdk.ABCIMessageLog) (*Event, error) {
	gidStr, err := event.GetEventValue(log, types.EventTypeCreateGroup, types.AttributeKeyGroupID)
	if err != nil {
		return nil, err
	}

	gid, err := strconv.ParseUint(gidStr, 10, 64)
	if err != nil {
		return nil, err
	}

	thresholdStr, err := event.GetEventValue(log, types.EventTypeCreateGroup, types.AttributeKeyThreshold)
	if err != nil {
		return nil, err
	}

	threshold, err := strconv.ParseUint(thresholdStr, 10, 64)
	if err != nil {
		return nil, err
	}

	dkgContextStr, err := event.GetEventValue(log, types.EventTypeCreateGroup, types.AttributeKeyDKGContext)
	if err != nil {
		return nil, err
	}

	dkgContext, err := hex.DecodeString(dkgContextStr)
	if err != nil {
		return nil, err
	}

	members := event.GetEventValues(log, types.EventTypeCreateGroup, types.AttributeKeyMember)

	return &Event{
		GroupID:    tss.GroupID(gid),
		Threshold:  threshold,
		DKGContext: dkgContext,
		Members:    members,
	}, nil
}

// getMemberID returns the member ID corresponding to the provided address.
// It searches through the event's members and returns the member ID if found.
// If no member with the given address is found, it returns an error.
func (e *Event) getMemberID(address string) (tss.MemberID, error) {
	for idx, member := range e.Members {
		if member == address {
			return tss.MemberID(idx + 1), nil
		}
	}

	return 0, errors.New("failed to find member in the event")
}
