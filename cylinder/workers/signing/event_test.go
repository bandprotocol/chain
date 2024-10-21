package signing_test

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"

	abci "github.com/cometbft/cometbft/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/cylinder/workers/signing"
	"github.com/bandprotocol/chain/v3/x/tss/types"
)

func TestParseEvent(t *testing.T) {
	tests := []struct {
		name     string
		events   sdk.StringEvents
		expEvent []signing.Event
		expError string
	}{
		{
			"success",
			sdk.StringifyEvents([]abci.Event{
				abci.Event(sdk.NewEvent(
					types.EventTypeRequestSignature,
					sdk.NewAttribute(types.AttributeKeyGroupID, "1"),
					sdk.NewAttribute(types.AttributeKeySigningID, "1"),
					sdk.NewAttribute(types.AttributeKeyMessage, hex.EncodeToString([]byte("message"))),
					sdk.NewAttribute(types.AttributeKeyGroupPubNonce, hex.EncodeToString([]byte("groupPubNonce"))),
					sdk.NewAttribute(types.AttributeKeyMemberID, "1"),
					sdk.NewAttribute(types.AttributeKeyAddress, "member 1"),
					sdk.NewAttribute(types.AttributeKeyBindingFactor, hex.EncodeToString([]byte("bindingFactor1"))),
					sdk.NewAttribute(types.AttributeKeyPubNonce, hex.EncodeToString([]byte("pubNonce1"))),
					sdk.NewAttribute(types.AttributeKeyPubD, hex.EncodeToString([]byte("pubD1"))),
					sdk.NewAttribute(types.AttributeKeyPubE, hex.EncodeToString([]byte("pubE1"))),
					sdk.NewAttribute(types.AttributeKeyMemberID, "2"),
					sdk.NewAttribute(types.AttributeKeyAddress, "member 2"),
					sdk.NewAttribute(types.AttributeKeyBindingFactor, hex.EncodeToString([]byte("bindingFactor2"))),
					sdk.NewAttribute(types.AttributeKeyPubNonce, hex.EncodeToString([]byte("pubNonce2"))),
					sdk.NewAttribute(types.AttributeKeyPubD, hex.EncodeToString([]byte("pubD2"))),
					sdk.NewAttribute(types.AttributeKeyPubE, hex.EncodeToString([]byte("pubE2"))),
				)),
			}),
			[]signing.Event{
				{
					SigningID: 1,
				},
			},
			"",
		},
		{
			"success - two events",
			sdk.StringifyEvents([]abci.Event{
				abci.Event(sdk.NewEvent(
					types.EventTypeRequestSignature,
					sdk.NewAttribute(types.AttributeKeyGroupID, "1"),
					sdk.NewAttribute(types.AttributeKeySigningID, "1"),
					sdk.NewAttribute(types.AttributeKeyMessage, hex.EncodeToString([]byte("message"))),
					sdk.NewAttribute(types.AttributeKeyGroupPubNonce, hex.EncodeToString([]byte("groupPubNonce"))),
					sdk.NewAttribute(types.AttributeKeyMemberID, "1"),
					sdk.NewAttribute(types.AttributeKeyAddress, "member 1"),
					sdk.NewAttribute(types.AttributeKeyBindingFactor, hex.EncodeToString([]byte("bindingFactor1"))),
					sdk.NewAttribute(types.AttributeKeyPubNonce, hex.EncodeToString([]byte("pubNonce1"))),
					sdk.NewAttribute(types.AttributeKeyPubD, hex.EncodeToString([]byte("pubD1"))),
					sdk.NewAttribute(types.AttributeKeyPubE, hex.EncodeToString([]byte("pubE1"))),
					sdk.NewAttribute(types.AttributeKeyMemberID, "2"),
					sdk.NewAttribute(types.AttributeKeyAddress, "member 2"),
					sdk.NewAttribute(types.AttributeKeyBindingFactor, hex.EncodeToString([]byte("bindingFactor2"))),
					sdk.NewAttribute(types.AttributeKeyPubNonce, hex.EncodeToString([]byte("pubNonce2"))),
					sdk.NewAttribute(types.AttributeKeyPubD, hex.EncodeToString([]byte("pubD2"))),
					sdk.NewAttribute(types.AttributeKeyPubE, hex.EncodeToString([]byte("pubE2"))),
				)),
				abci.Event(sdk.NewEvent(
					types.EventTypeRequestSignature,
					sdk.NewAttribute(types.AttributeKeyGroupID, "1"),
					sdk.NewAttribute(types.AttributeKeySigningID, "2"),
					sdk.NewAttribute(types.AttributeKeyMessage, hex.EncodeToString([]byte("message"))),
					sdk.NewAttribute(types.AttributeKeyGroupPubNonce, hex.EncodeToString([]byte("groupPubNonce"))),
					sdk.NewAttribute(types.AttributeKeyMemberID, "1"),
					sdk.NewAttribute(types.AttributeKeyAddress, "member 1"),
					sdk.NewAttribute(types.AttributeKeyBindingFactor, hex.EncodeToString([]byte("bindingFactor1"))),
					sdk.NewAttribute(types.AttributeKeyPubNonce, hex.EncodeToString([]byte("pubNonce1"))),
					sdk.NewAttribute(types.AttributeKeyPubD, hex.EncodeToString([]byte("pubD1"))),
					sdk.NewAttribute(types.AttributeKeyPubE, hex.EncodeToString([]byte("pubE1"))),
					sdk.NewAttribute(types.AttributeKeyMemberID, "2"),
					sdk.NewAttribute(types.AttributeKeyAddress, "member 2"),
					sdk.NewAttribute(types.AttributeKeyBindingFactor, hex.EncodeToString([]byte("bindingFactor2"))),
					sdk.NewAttribute(types.AttributeKeyPubNonce, hex.EncodeToString([]byte("pubNonce2"))),
					sdk.NewAttribute(types.AttributeKeyPubD, hex.EncodeToString([]byte("pubD2"))),
					sdk.NewAttribute(types.AttributeKeyPubE, hex.EncodeToString([]byte("pubE2"))),
				)),
			}),
			[]signing.Event{
				{
					SigningID: 1,
				},
				{
					SigningID: 2,
				},
			},
			"",
		},
		{
			"success - two events (merge type)",
			sdk.StringifyEvents([]abci.Event{
				abci.Event(sdk.NewEvent(
					types.EventTypeRequestSignature,
					sdk.NewAttribute(types.AttributeKeyGroupID, "1"),
					sdk.NewAttribute(types.AttributeKeySigningID, "1"),
					sdk.NewAttribute(types.AttributeKeyMessage, hex.EncodeToString([]byte("message"))),
					sdk.NewAttribute(types.AttributeKeyGroupPubNonce, hex.EncodeToString([]byte("groupPubNonce"))),
					sdk.NewAttribute(types.AttributeKeyMemberID, "1"),
					sdk.NewAttribute(types.AttributeKeyAddress, "member 1"),
					sdk.NewAttribute(types.AttributeKeyBindingFactor, hex.EncodeToString([]byte("bindingFactor1"))),
					sdk.NewAttribute(types.AttributeKeyPubNonce, hex.EncodeToString([]byte("pubNonce1"))),
					sdk.NewAttribute(types.AttributeKeyPubD, hex.EncodeToString([]byte("pubD1"))),
					sdk.NewAttribute(types.AttributeKeyPubE, hex.EncodeToString([]byte("pubE1"))),
					sdk.NewAttribute(types.AttributeKeyMemberID, "2"),
					sdk.NewAttribute(types.AttributeKeyAddress, "member 2"),
					sdk.NewAttribute(types.AttributeKeyBindingFactor, hex.EncodeToString([]byte("bindingFactor2"))),
					sdk.NewAttribute(types.AttributeKeyPubNonce, hex.EncodeToString([]byte("pubNonce2"))),
					sdk.NewAttribute(types.AttributeKeyPubD, hex.EncodeToString([]byte("pubD2"))),
					sdk.NewAttribute(types.AttributeKeyPubE, hex.EncodeToString([]byte("pubE2"))),
					sdk.NewAttribute(types.AttributeKeyGroupID, "1"),
					sdk.NewAttribute(types.AttributeKeySigningID, "2"),
					sdk.NewAttribute(types.AttributeKeyMessage, hex.EncodeToString([]byte("message"))),
					sdk.NewAttribute(types.AttributeKeyGroupPubNonce, hex.EncodeToString([]byte("groupPubNonce"))),
					sdk.NewAttribute(types.AttributeKeyMemberID, "1"),
					sdk.NewAttribute(types.AttributeKeyAddress, "member 1"),
					sdk.NewAttribute(types.AttributeKeyBindingFactor, hex.EncodeToString([]byte("bindingFactor1"))),
					sdk.NewAttribute(types.AttributeKeyPubNonce, hex.EncodeToString([]byte("pubNonce1"))),
					sdk.NewAttribute(types.AttributeKeyPubD, hex.EncodeToString([]byte("pubD1"))),
					sdk.NewAttribute(types.AttributeKeyPubE, hex.EncodeToString([]byte("pubE1"))),
					sdk.NewAttribute(types.AttributeKeyMemberID, "2"),
					sdk.NewAttribute(types.AttributeKeyAddress, "member 2"),
					sdk.NewAttribute(types.AttributeKeyBindingFactor, hex.EncodeToString([]byte("bindingFactor2"))),
					sdk.NewAttribute(types.AttributeKeyPubNonce, hex.EncodeToString([]byte("pubNonce2"))),
					sdk.NewAttribute(types.AttributeKeyPubD, hex.EncodeToString([]byte("pubD2"))),
					sdk.NewAttribute(types.AttributeKeyPubE, hex.EncodeToString([]byte("pubE2"))),
				)),
			}),
			[]signing.Event{
				{
					SigningID: 1,
				},
				{
					SigningID: 2,
				},
			},
			"",
		},
		{
			"no event",
			sdk.StringifyEvents([]abci.Event{}),
			nil,
			"",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			events, err := signing.ParseEvents(test.events)
			assert.Equal(t, test.expEvent, events)
			if test.expError != "" {
				assert.ErrorContains(t, err, test.expError)
			}
		})
	}
}
