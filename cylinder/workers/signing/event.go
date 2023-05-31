package signing

import (
	"strconv"

	"github.com/bandprotocol/chain/v2/pkg/event"
	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/tss/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Event represents the data structure for round3 events.
// TODO-CYLINDER: use real type
type Event struct {
	GroupID       tss.GroupID
	SigningID     tss.GroupID // tss.SigningID
	MemberIDs     []tss.MemberID
	GroupPubNonce tss.PublicKey
	Data          []byte
	Bytes         []byte
	PubDE         types.DE
}

// ParseEvent parses the round3 event from the given message log.
// It extracts the group ID from the log and returns the parsed Event or an error if parsing fails.
// TODO-CYLINDER: use real type, parse more
func ParseEvent(log sdk.ABCIMessageLog) (*Event, error) {
	gidStr, err := event.GetEventValue(log, types.EventTypeRound2Success, types.AttributeKeyGroupID)
	if err != nil {
		return nil, err
	}

	gid, err := strconv.ParseUint(gidStr, 10, 64)
	if err != nil {
		return nil, err
	}

	sidStr, err := event.GetEventValue(log, types.EventTypeRound2Success, types.AttributeKeyGroupID)
	if err != nil {
		return nil, err
	}

	sid, err := strconv.ParseUint(sidStr, 10, 64)
	if err != nil {
		return nil, err
	}

	return &Event{
		GroupID:       tss.GroupID(gid),
		SigningID:     tss.GroupID(sid),
		MemberIDs:     []tss.MemberID{},
		GroupPubNonce: tss.PublicKey{},
		Data:          []byte{},
		Bytes:         []byte{},
		PubDE:         types.DE{},
	}, nil
}
