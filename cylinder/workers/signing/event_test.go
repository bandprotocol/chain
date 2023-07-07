package signing_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"

	"github.com/bandprotocol/chain/v2/cylinder/workers/signing"
	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/pkg/tss/testutil"
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

func TestParseEvent(t *testing.T) {
	tests := []struct {
		name     string
		log      sdk.ABCIMessageLog
		address  string
		expEvent *signing.Event
		expError string
	}{
		{
			"success",
			sdk.NewABCIMessageLog(0, "", sdk.Events{
				sdk.NewEvent(
					types.EventTypeRequestSign,
					sdk.NewAttribute(types.AttributeKeyGroupID, "1"),
					sdk.NewAttribute(types.AttributeKeySigningID, "1"),
					sdk.NewAttribute(types.AttributeKeyMessage, "aaaaa0"),
					sdk.NewAttribute(types.AttributeKeyGroupPubNonce, "aaaaa1"),
					sdk.NewAttribute(types.AttributeKeyMemberID, "1"),
					sdk.NewAttribute(types.AttributeKeyMember, "banda"),
					sdk.NewAttribute(types.AttributeKeyBindingFactor, "aaaaa2"),
					sdk.NewAttribute(types.AttributeKeyPubNonces, "aaaaa2"),
					sdk.NewAttribute(types.AttributeKeyPubD, "aaaaa3"),
					sdk.NewAttribute(types.AttributeKeyPubE, "aaaaa4"),
					sdk.NewAttribute(types.AttributeKeyMemberID, "2"),
					sdk.NewAttribute(types.AttributeKeyMember, "bandb"),
					sdk.NewAttribute(types.AttributeKeyBindingFactor, "aaaaa5"),
					sdk.NewAttribute(types.AttributeKeyPubNonces, "aaaaa6"),
					sdk.NewAttribute(types.AttributeKeyPubD, "aaaaa7"),
					sdk.NewAttribute(types.AttributeKeyPubE, "aaaaa8"),
				),
			}),
			"bandb",
			&signing.Event{
				GroupID:       1,
				SigningID:     1,
				MemberIDs:     []tss.MemberID{1, 2},
				GroupPubNonce: testutil.HexDecode("aaaaa1"),
				Data:          testutil.HexDecode("aaaaa0"),
				BindingFactor: testutil.HexDecode("aaaaa5"),
				PubDE: types.DE{
					PubD: testutil.HexDecode("aaaaa7"),
					PubE: testutil.HexDecode("aaaaa8"),
				},
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
					types.EventTypeRequestSign,
					sdk.NewAttribute(types.AttributeKeyGroupID, "1"),
					sdk.NewAttribute(types.AttributeKeySigningID, "1"),
					sdk.NewAttribute(types.AttributeKeyMessage, "aaaaa0"),
					sdk.NewAttribute(types.AttributeKeyGroupPubNonce, "aaaaa1"),
					sdk.NewAttribute(types.AttributeKeyMemberID, "1"),
					sdk.NewAttribute(types.AttributeKeyMember, "banda"),
					sdk.NewAttribute(types.AttributeKeyBindingFactor, "aaaaa2"),
					sdk.NewAttribute(types.AttributeKeyPubNonces, "aaaaa2"),
					sdk.NewAttribute(types.AttributeKeyPubD, "aaaaa3"),
					sdk.NewAttribute(types.AttributeKeyPubE, "aaaaa4"),
					sdk.NewAttribute(types.AttributeKeyMemberID, "2"),
					sdk.NewAttribute(types.AttributeKeyMember, "bandb"),
					sdk.NewAttribute(types.AttributeKeyBindingFactor, "aaaaa5"),
					sdk.NewAttribute(types.AttributeKeyPubNonces, "aaaaa6"),
					sdk.NewAttribute(types.AttributeKeyPubD, "aaaaa7"),
					sdk.NewAttribute(types.AttributeKeyPubE, "aaaaa8"),
				),
			}),
			"bandc",
			nil,
			"failed to find member in the event",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			event, err := signing.ParseEvent(test.log, test.address)
			assert.Equal(t, test.expEvent, event)
			if test.expError != "" {
				assert.ErrorContains(t, err, test.expError)
			}
		})
	}
}
