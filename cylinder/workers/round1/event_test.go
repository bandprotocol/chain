package round1_test

import (
	"errors"
	"testing"

	"github.com/bandprotocol/chain/v2/cylinder/workers/round1"
	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/pkg/tss/testutil"
	"github.com/bandprotocol/chain/v2/x/tss/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
)

func TestParseEvent(t *testing.T) {
	tests := []struct {
		name     string
		log      sdk.ABCIMessageLog
		address  string
		expEvent *round1.Event
		expError string
	}{
		{
			"success",
			sdk.NewABCIMessageLog(0, "", sdk.Events{
				sdk.NewEvent(
					types.EventTypeCreateGroup,
					sdk.NewAttribute(types.AttributeKeyGroupID, "1"),
					sdk.NewAttribute(types.AttributeKeySize, "3"),
					sdk.NewAttribute(types.AttributeKeyThreshold, "2"),
					sdk.NewAttribute(types.AttributeKeyStatus, types.GROUP_STATUS_ROUND_1.String()),
					sdk.NewAttribute(types.AttributeKeyDKGContext, "aaaaaa"),
					sdk.NewAttribute(types.AttributeKeyMember, "banda"),
					sdk.NewAttribute(types.AttributeKeyMember, "bandb"),
					sdk.NewAttribute(types.AttributeKeyMember, "bandc"),
				),
			}),
			"bandb",
			&round1.Event{
				GroupID:    1,
				MemberID:   2,
				Threshold:  2,
				DKGContext: testutil.HexDecode("aaaaaa"),
			},
			"",
		},
		{
			"no event",
			sdk.NewABCIMessageLog(0, "", sdk.Events{}),
			"bandb",
			nil,
			"Cannot find event with type",
		},
		{
			"invalid member",
			sdk.NewABCIMessageLog(0, "", sdk.Events{
				sdk.NewEvent(
					types.EventTypeCreateGroup,
					sdk.NewAttribute(types.AttributeKeyGroupID, "1"),
					sdk.NewAttribute(types.AttributeKeySize, "3"),
					sdk.NewAttribute(types.AttributeKeyThreshold, "2"),
					sdk.NewAttribute(types.AttributeKeyStatus, types.GROUP_STATUS_ROUND_1.String()),
					sdk.NewAttribute(types.AttributeKeyDKGContext, "aaaaaa"),
					sdk.NewAttribute(types.AttributeKeyMember, "banda"),
					sdk.NewAttribute(types.AttributeKeyMember, "bandb"),
					sdk.NewAttribute(types.AttributeKeyMember, "bandc"),
				),
			}),
			"bandd",
			nil,
			"failed to find member in the event",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			event, err := round1.ParseEvent(test.log, test.address)
			assert.Equal(t, test.expEvent, event)
			if test.expError != "" {
				assert.ErrorContains(t, err, test.expError)
			}
		})
	}
}

func TestGetMemberID(t *testing.T) {
	tests := []struct {
		name        string
		members     []string
		expMemberID tss.MemberID
		expError    error
	}{
		{
			"not in the member",
			[]string{
				"b",
				"c",
			},
			0,
			errors.New("failed to find member in the event"),
		},
		{
			"first member",
			[]string{
				"a",
				"b",
				"c",
			},
			1,
			nil,
		},
		{
			"last member",
			[]string{
				"b",
				"c",
				"a",
			},
			3,
			nil,
		},
		{
			"no member in the group",
			[]string{},
			0,
			errors.New("failed to find member in the event"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			memberID, err := round1.GetMemberID(test.members, "a")
			assert.Equal(t, test.expError, err)
			assert.Equal(t, test.expMemberID, memberID)
		})
	}
}
