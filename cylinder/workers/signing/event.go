package signing

import (
	"github.com/bandprotocol/chain/v2/pkg/event"
	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/tss/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Event represents the data structure for round3 events.
// TODO-CYLINDER: use real type
type Event struct {
	GroupID       tss.GroupID
	SigningID     tss.SigningID
	MemberIDs     []tss.MemberID
	GroupPubNonce tss.PublicKey
	Data          []byte
	Bytes         []byte
	PubDE         types.DE
}

// ParseEvent parses the round3 event from the given message log.
// It extracts the group ID from the log and returns the parsed Event or an error if parsing fails.
// TODO-CYLINDER: use real type, parse more
func ParseEvent(log sdk.ABCIMessageLog, address string) (*Event, error) {
	gid, err := event.GetEventValueUint64(log, types.EventTypeRequestSign, types.AttributeKeyGroupID)
	if err != nil {
		return nil, err
	}

	sid, err := event.GetEventValueUint64(log, types.EventTypeRequestSign, types.AttributeKeySigningID)
	if err != nil {
		return nil, err
	}

	groupPubNonce, err := event.GetEventValueBytes(log, types.EventTypeRequestSign, types.AttributeKeyGroupPubNonce)
	if err != nil {
		return nil, err
	}

	bytes, err := event.GetEventValueBytes(log, types.EventTypeRequestSign, types.AttributeBytes)
	if err != nil {
		return nil, err
	}

	data, err := event.GetEventValueBytes(log, types.EventTypeRequestSign, types.AttributeMessage)
	if err != nil {
		return nil, err
	}

	midInts, err := event.GetEventValuesUint64(log, types.EventTypeRequestSign, types.AttributeKeyMemberID)
	if err != nil {
		return nil, err
	}

	pubDs, err := event.GetEventValuesBytes(log, types.EventTypeRequestSign, types.AttributeKeyPublicD)
	if err != nil {
		return nil, err
	}

	pubEs, err := event.GetEventValuesBytes(log, types.EventTypeRequestSign, types.AttributeKeyPublicE)
	if err != nil {
		return nil, err
	}

	var pubD, pubE tss.PublicKey
	var mids []tss.MemberID

	members := event.GetEventValues(log, types.EventTypeRequestSign, types.AttributeKeyMember)
	for i, member := range members {
		mids = append(mids, tss.MemberID(midInts[i]))
		if member == address {
			pubD = pubDs[i]
			pubE = pubEs[i]
		}
	}

	return &Event{
		GroupID:       tss.GroupID(gid),
		SigningID:     tss.SigningID(sid),
		MemberIDs:     mids,
		GroupPubNonce: groupPubNonce,
		Data:          data,
		Bytes:         bytes,
		PubDE: types.DE{
			PubD: pubD,
			PubE: pubE,
		},
	}, nil
}
