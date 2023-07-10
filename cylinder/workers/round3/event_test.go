package round3_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"

	"github.com/bandprotocol/chain/v2/cylinder/workers/round3"
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

func TestParseEvent(t *testing.T) {
	tests := []struct {
		name     string
		log      sdk.ABCIMessageLog
		expEvent *round3.Event
		expError string
	}{
		{
			"success",
			sdk.NewABCIMessageLog(0, "", sdk.Events{
				sdk.NewEvent(
					types.EventTypeRound2Success,
					sdk.NewAttribute(types.AttributeKeyGroupID, "1"),
					sdk.NewAttribute(types.AttributeKeyStatus, types.GROUP_STATUS_ROUND_3.String()),
				),
			}),
			&round3.Event{
				GroupID: 1,
			},
			"",
		},
		{
			"no event",
			sdk.NewABCIMessageLog(0, "", sdk.Events{}),
			nil,
			"Cannot find event with type",
		},
		{
			"invalid value",
			sdk.NewABCIMessageLog(0, "", sdk.Events{
				sdk.NewEvent(
					types.EventTypeRound2Success,
					sdk.NewAttribute(types.AttributeKeyGroupID, "invalid uint64"),
					sdk.NewAttribute(types.AttributeKeyStatus, types.GROUP_STATUS_ROUND_2.String()),
				),
			}),
			nil,
			"strconv.ParseUint",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			event, err := round3.ParseEvent(test.log)
			assert.Equal(t, test.expEvent, event)
			if test.expError != "" {
				assert.ErrorContains(t, err, test.expError)
			}
		})
	}
}
