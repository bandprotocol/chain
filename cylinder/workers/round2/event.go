package round2

import (
	"strconv"

	"github.com/bandprotocol/chain/v2/pkg/event"
	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/tss/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Event struct {
	GroupID tss.GroupID
}

func ParseEvent(log sdk.ABCIMessageLog) (*Event, error) {
	gidStr, err := event.GetEventValue(log, types.EventTypeRound1Success, types.AttributeKeyGroupID)
	if err != nil {
		return nil, err
	}

	gid, err := strconv.ParseUint(gidStr, 10, 64)
	if err != nil {
		return nil, err
	}

	return &Event{
		GroupID: tss.GroupID(gid),
	}, nil
}
