package replacer_test

import (
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"

	"github.com/bandprotocol/chain/v2/cylinder/workers/replacer"
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

func TestParseEvent(t *testing.T) {
	tests := []struct {
		name     string
		events   sdk.StringEvents
		expEvent *replacer.Event
		expError string
	}{
		{
			"success",
			sdk.StringifyEvents([]abci.Event{
				abci.Event(sdk.NewEvent(
					types.EventTypeReplaceSuccess,
					sdk.NewAttribute(types.AttributeKeyFromGroupID, "2"),
					sdk.NewAttribute(types.AttributeKeyToGroupID, "1"),
				)),
			}),
			&replacer.Event{
				FromGroupID: 2,
				ToGroupID:   1,
			},
			"",
		},
		{
			"no event",
			sdk.StringifyEvents([]abci.Event{}),
			nil,
			"Cannot find event with type",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			event, err := replacer.ParseEvent(test.events)
			assert.Equal(t, test.expEvent, event)
			if test.expError != "" {
				assert.ErrorContains(t, err, test.expError)
			}
		})
	}
}
