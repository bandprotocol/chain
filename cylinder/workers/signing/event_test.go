package signing_test

import (
	"encoding/hex"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"

	"github.com/bandprotocol/chain/v2/cylinder/workers/signing"
	"github.com/bandprotocol/chain/v2/pkg/tss"
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
					sdk.NewAttribute(types.AttributeKeyMessage, hex.EncodeToString([]byte("message"))),
					sdk.NewAttribute(types.AttributeKeyGroupPubNonce, hex.EncodeToString([]byte("groupPubNonce"))),
					sdk.NewAttribute(types.AttributeKeyMemberID, "1"),
					sdk.NewAttribute(types.AttributeKeyMember, "member 1"),
					sdk.NewAttribute(types.AttributeKeyBindingFactor, hex.EncodeToString([]byte("bindingFactor1"))),
					sdk.NewAttribute(types.AttributeKeyPubNonce, hex.EncodeToString([]byte("pubNonce1"))),
					sdk.NewAttribute(types.AttributeKeyPubD, hex.EncodeToString([]byte("pubD1"))),
					sdk.NewAttribute(types.AttributeKeyPubE, hex.EncodeToString([]byte("pubE1"))),
					sdk.NewAttribute(types.AttributeKeyMemberID, "2"),
					sdk.NewAttribute(types.AttributeKeyMember, "member 2"),
					sdk.NewAttribute(types.AttributeKeyBindingFactor, hex.EncodeToString([]byte("bindingFactor2"))),
					sdk.NewAttribute(types.AttributeKeyPubNonce, hex.EncodeToString([]byte("pubNonce2"))),
					sdk.NewAttribute(types.AttributeKeyPubD, hex.EncodeToString([]byte("pubD2"))),
					sdk.NewAttribute(types.AttributeKeyPubE, hex.EncodeToString([]byte("pubE2"))),
				),
			}),
			"member 2",
			&signing.Event{
				GroupID:       1,
				SigningID:     1,
				MemberIDs:     []tss.MemberID{1, 2},
				GroupPubNonce: []byte("groupPubNonce"),
				Data:          []byte("message"),
				BindingFactor: []byte("bindingFactor2"),
				PubDE: types.DE{
					PubD: []byte("pubD2"),
					PubE: []byte("pubE2"),
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
					sdk.NewAttribute(types.AttributeKeyMessage, hex.EncodeToString([]byte("message"))),
					sdk.NewAttribute(types.AttributeKeyGroupPubNonce, hex.EncodeToString([]byte("groupPubNonce"))),
					sdk.NewAttribute(types.AttributeKeyMemberID, "1"),
					sdk.NewAttribute(types.AttributeKeyMember, "member 1"),
					sdk.NewAttribute(types.AttributeKeyBindingFactor, hex.EncodeToString([]byte("bindingFactor1"))),
					sdk.NewAttribute(types.AttributeKeyPubNonce, hex.EncodeToString([]byte("pubNonce1"))),
					sdk.NewAttribute(types.AttributeKeyPubD, hex.EncodeToString([]byte("pubD1"))),
					sdk.NewAttribute(types.AttributeKeyPubE, hex.EncodeToString([]byte("pubE1"))),
					sdk.NewAttribute(types.AttributeKeyMemberID, "2"),
					sdk.NewAttribute(types.AttributeKeyMember, "member 2"),
					sdk.NewAttribute(types.AttributeKeyBindingFactor, hex.EncodeToString([]byte("bindingFactor2"))),
					sdk.NewAttribute(types.AttributeKeyPubNonce, hex.EncodeToString([]byte("pubNonce2"))),
					sdk.NewAttribute(types.AttributeKeyPubD, hex.EncodeToString([]byte("pubD2"))),
					sdk.NewAttribute(types.AttributeKeyPubE, hex.EncodeToString([]byte("pubE2"))),
				),
			}),
			"member 3",
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
