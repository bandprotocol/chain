package group_test

import (
	"encoding/hex"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"

	"github.com/bandprotocol/chain/v2/cylinder/workers/group"
	bandtsstypes "github.com/bandprotocol/chain/v2/x/bandtss/types"
	tsstypes "github.com/bandprotocol/chain/v2/x/tss/types"
)

func TestParseEvent(t *testing.T) {
	tests := []struct {
		name     string
		events   sdk.StringEvents
		evType   string
		expEvent *group.Event
		expError string
	}{
		{
			"success - create group event",
			sdk.StringifyEvents([]abci.Event{
				abci.Event(sdk.NewEvent(
					bandtsstypes.EventTypeCreateGroup,
					sdk.NewAttribute(bandtsstypes.AttributeKeyGroupID, "1"),
					sdk.NewAttribute(bandtsstypes.AttributeKeySize, "3"),
					sdk.NewAttribute(bandtsstypes.AttributeKeyThreshold, "2"),
					sdk.NewAttribute(bandtsstypes.AttributeKeyStatus, tsstypes.GROUP_STATUS_ROUND_1.String()),
					sdk.NewAttribute(bandtsstypes.AttributeKeyDKGContext, hex.EncodeToString([]byte("dkgContext"))),
					sdk.NewAttribute(bandtsstypes.AttributeKeyAddress, "member 1"),
					sdk.NewAttribute(bandtsstypes.AttributeKeyAddress, "member 2"),
					sdk.NewAttribute(bandtsstypes.AttributeKeyAddress, "member 3"),
				)),
			}),
			bandtsstypes.EventTypeCreateGroup,
			&group.Event{
				GroupID: 1,
			},
			"",
		},
		{
			"success - round1 success event",
			sdk.StringifyEvents([]abci.Event{
				abci.Event(sdk.NewEvent(
					tsstypes.EventTypeRound1Success,
					sdk.NewAttribute(tsstypes.AttributeKeyGroupID, "1"),
					sdk.NewAttribute(tsstypes.AttributeKeyStatus, tsstypes.GROUP_STATUS_ROUND_2.String()),
				)),
			}),
			tsstypes.EventTypeRound1Success,
			&group.Event{
				GroupID: 1,
			},
			"",
		},
		{
			"success - round2 success event",
			sdk.StringifyEvents([]abci.Event{
				abci.Event(sdk.NewEvent(
					tsstypes.EventTypeRound2Success,
					sdk.NewAttribute(tsstypes.AttributeKeyGroupID, "1"),
					sdk.NewAttribute(tsstypes.AttributeKeyStatus, tsstypes.GROUP_STATUS_ROUND_3.String()),
				)),
			}),
			tsstypes.EventTypeRound2Success,
			&group.Event{
				GroupID: 1,
			},
			"",
		},
		{
			"no event",
			sdk.StringifyEvents([]abci.Event{}),
			bandtsstypes.EventTypeCreateGroup,
			nil,
			"cannot find event with type",
		},
		{
			"no groupID",
			sdk.StringifyEvents([]abci.Event{
				abci.Event(sdk.NewEvent(
					bandtsstypes.EventTypeCreateGroup,
				)),
			}),
			bandtsstypes.EventTypeCreateGroup,
			nil,
			"cannot find event with type",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			event, err := group.ParseEvent(test.events, test.evType)
			assert.Equal(t, test.expEvent, event)
			if test.expError != "" {
				assert.ErrorContains(t, err, test.expError)
			}
		})
	}
}
