package round1

import (
	"encoding/hex"
	"strconv"

	"github.com/bandprotocol/chain/v2/pkg/event"
	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/tss/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Event struct {
	GroupID    tss.GroupID
	Threshold  uint64
	DKGContext []byte
	Members    []string
}

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
