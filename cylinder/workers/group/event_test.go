package group_test

import (
	"encoding/hex"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"

	"github.com/bandprotocol/chain/v2/cylinder/workers/group"
	"github.com/bandprotocol/chain/v2/x/tss/types"
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
					types.EventTypeCreateGroup,
					sdk.NewAttribute(types.AttributeKeyGroupID, "1"),
					sdk.NewAttribute(types.AttributeKeySize, "3"),
					sdk.NewAttribute(types.AttributeKeyThreshold, "2"),
					sdk.NewAttribute(types.AttributeKeyStatus, types.GROUP_STATUS_ROUND_1.String()),
					sdk.NewAttribute(types.AttributeKeyDKGContext, hex.EncodeToString([]byte("dkgContext"))),
					sdk.NewAttribute(types.AttributeKeyAddress, "member 1"),
					sdk.NewAttribute(types.AttributeKeyAddress, "member 2"),
					sdk.NewAttribute(types.AttributeKeyAddress, "member 3"),
				)),
			}),
			types.EventTypeCreateGroup,
			&group.Event{
				GroupID: 1,
			},
			"",
		},
		{
			"success - round1 success event",
			sdk.StringifyEvents([]abci.Event{
				abci.Event(sdk.NewEvent(
					types.EventTypeRound1Success,
					sdk.NewAttribute(types.AttributeKeyGroupID, "1"),
					sdk.NewAttribute(types.AttributeKeyStatus, types.GROUP_STATUS_ROUND_2.String()),
				)),
			}),
			types.EventTypeRound1Success,
			&group.Event{
				GroupID: 1,
			},
			"",
		},
		{
			"success - round2 success event",
			sdk.StringifyEvents([]abci.Event{
				abci.Event(sdk.NewEvent(
					types.EventTypeRound2Success,
					sdk.NewAttribute(types.AttributeKeyGroupID, "1"),
					sdk.NewAttribute(types.AttributeKeyStatus, types.GROUP_STATUS_ROUND_3.String()),
				)),
			}),
			types.EventTypeRound2Success,
			&group.Event{
				GroupID: 1,
			},
			"",
		},
		{
			"no event",
			sdk.StringifyEvents([]abci.Event{}),
			types.EventTypeCreateGroup,
			nil,
			"cannot find event with type",
		},
		{
			"no groupID",
			sdk.StringifyEvents([]abci.Event{
				abci.Event(sdk.NewEvent(
					types.EventTypeCreateGroup,
				)),
			}),
			types.EventTypeCreateGroup,
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
