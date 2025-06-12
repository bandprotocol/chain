package de_test

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"

	abci "github.com/cometbft/cometbft/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/cylinder/workers/de"
	"github.com/bandprotocol/chain/v3/x/tss/types"
)

func TestParseSubmitSignEvents(t *testing.T) {
	tests := []struct {
		name      string
		events    sdk.StringEvents
		eventType string
		expEvent  []de.PubDE
		expError  string
	}{
		{
			"success",
			sdk.StringifyEvents([]abci.Event{
				abci.Event(sdk.NewEvent(
					types.EventTypeSubmitSignature,
					sdk.NewAttribute(types.AttributeKeySigningID, "1"),
					sdk.NewAttribute(types.AttributeKeyGroupID, "1"),
					sdk.NewAttribute(types.AttributeKeyMemberID, "2"),
					sdk.NewAttribute(types.AttributeKeyAddress, "member 1"),
					sdk.NewAttribute(types.AttributeKeyPubD, hex.EncodeToString([]byte("pubD"))),
					sdk.NewAttribute(types.AttributeKeyPubE, hex.EncodeToString([]byte("pubE"))),
					sdk.NewAttribute(types.AttributeKeySignature, hex.EncodeToString([]byte("signature"))),
				)),
			}),
			types.EventTypeSubmitSignature,
			[]de.PubDE{
				{
					PubDE: types.DE{
						PubD: []byte("pubD"),
						PubE: []byte("pubE"),
					},
				},
			},
			"",
		},
		{
			"success - two events",
			sdk.StringifyEvents([]abci.Event{
				abci.Event(sdk.NewEvent(
					types.EventTypeSubmitSignature,
					sdk.NewAttribute(types.AttributeKeySigningID, "1"),
					sdk.NewAttribute(types.AttributeKeyGroupID, "1"),
					sdk.NewAttribute(types.AttributeKeyMemberID, "2"),
					sdk.NewAttribute(types.AttributeKeyAddress, "member 1"),
					sdk.NewAttribute(types.AttributeKeyPubD, hex.EncodeToString([]byte("pubD 1"))),
					sdk.NewAttribute(types.AttributeKeyPubE, hex.EncodeToString([]byte("pubE 1"))),
					sdk.NewAttribute(types.AttributeKeySignature, hex.EncodeToString([]byte("signature"))),
				)),
				abci.Event(sdk.NewEvent(
					types.EventTypeSubmitSignature,
					sdk.NewAttribute(types.AttributeKeySigningID, "2"),
					sdk.NewAttribute(types.AttributeKeyGroupID, "1"),
					sdk.NewAttribute(types.AttributeKeyMemberID, "2"),
					sdk.NewAttribute(types.AttributeKeyAddress, "member 1"),
					sdk.NewAttribute(types.AttributeKeyPubD, hex.EncodeToString([]byte("pubD 2"))),
					sdk.NewAttribute(types.AttributeKeyPubE, hex.EncodeToString([]byte("pubE 2"))),
					sdk.NewAttribute(types.AttributeKeySignature, hex.EncodeToString([]byte("signature"))),
				)),
			}),
			types.EventTypeSubmitSignature,
			[]de.PubDE{
				{
					PubDE: types.DE{
						PubD: []byte("pubD 1"),
						PubE: []byte("pubE 1"),
					},
				},
				{
					PubDE: types.DE{
						PubD: []byte("pubD 2"),
						PubE: []byte("pubE 2"),
					},
				},
			},
			"",
		},
		{
			"success - two events (merge type)",
			sdk.StringifyEvents([]abci.Event{
				abci.Event(sdk.NewEvent(
					types.EventTypeSubmitSignature,
					sdk.NewAttribute(types.AttributeKeySigningID, "1"),
					sdk.NewAttribute(types.AttributeKeyGroupID, "1"),
					sdk.NewAttribute(types.AttributeKeyMemberID, "2"),
					sdk.NewAttribute(types.AttributeKeyAddress, "member 1"),
					sdk.NewAttribute(types.AttributeKeyPubD, hex.EncodeToString([]byte("pubD 1"))),
					sdk.NewAttribute(types.AttributeKeyPubE, hex.EncodeToString([]byte("pubE 1"))),
					sdk.NewAttribute(types.AttributeKeySignature, hex.EncodeToString([]byte("signature"))),
					sdk.NewAttribute(types.AttributeKeySigningID, "2"),
					sdk.NewAttribute(types.AttributeKeyGroupID, "1"),
					sdk.NewAttribute(types.AttributeKeyMemberID, "2"),
					sdk.NewAttribute(types.AttributeKeyAddress, "member 1"),
					sdk.NewAttribute(types.AttributeKeyPubD, hex.EncodeToString([]byte("pubD 2"))),
					sdk.NewAttribute(types.AttributeKeyPubE, hex.EncodeToString([]byte("pubE 2"))),
					sdk.NewAttribute(types.AttributeKeySignature, hex.EncodeToString([]byte("signature"))),
				)),
			}),
			types.EventTypeSubmitSignature,
			[]de.PubDE{
				{
					PubDE: types.DE{
						PubD: []byte("pubD 1"),
						PubE: []byte("pubE 1"),
					},
				},
				{
					PubDE: types.DE{
						PubD: []byte("pubD 2"),
						PubE: []byte("pubE 2"),
					},
				},
			},
			"",
		},
		{
			"no event",
			sdk.StringifyEvents([]abci.Event{}),
			types.EventTypeSubmitSignature,
			nil,
			"",
		},
		{
			"invalid value",
			sdk.StringifyEvents([]abci.Event{
				abci.Event(sdk.NewEvent(
					types.EventTypeSubmitSignature,
					sdk.NewAttribute(types.AttributeKeySigningID, "1"),
					sdk.NewAttribute(types.AttributeKeyGroupID, "1"),
					sdk.NewAttribute(types.AttributeKeyMemberID, "2"),
					sdk.NewAttribute(types.AttributeKeyAddress, "member 1"),
					sdk.NewAttribute(types.AttributeKeyPubD, "invalid hex"),
					sdk.NewAttribute(types.AttributeKeyPubE, hex.EncodeToString([]byte("pubE"))),
					sdk.NewAttribute(types.AttributeKeySignature, hex.EncodeToString([]byte("signature"))),
				)),
			}),
			types.EventTypeSubmitSignature,
			nil,
			"encoding/hex: invalid byte",
		},
		{
			"success - de_deleted",
			sdk.StringifyEvents([]abci.Event{
				abci.Event(sdk.NewEvent(
					types.EventTypeDEDeleted,
					sdk.NewAttribute(types.AttributeKeyAddress, "member 1"),
					sdk.NewAttribute(types.AttributeKeyPubD, hex.EncodeToString([]byte("pubD 1"))),
					sdk.NewAttribute(types.AttributeKeyPubE, hex.EncodeToString([]byte("pubE 1"))),
				)),
				abci.Event(sdk.NewEvent(
					types.EventTypeDEDeleted,
					sdk.NewAttribute(types.AttributeKeyAddress, "member 1"),
					sdk.NewAttribute(types.AttributeKeyPubD, hex.EncodeToString([]byte("pubD 2"))),
					sdk.NewAttribute(types.AttributeKeyPubE, hex.EncodeToString([]byte("pubE 2"))),
				)),
			}),
			types.EventTypeDEDeleted,
			[]de.PubDE{
				{
					PubDE: types.DE{
						PubD: []byte("pubD 1"),
						PubE: []byte("pubE 1"),
					},
				},
				{
					PubDE: types.DE{
						PubD: []byte("pubD 2"),
						PubE: []byte("pubE 2"),
					},
				},
			},
			"",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			events, err := de.ParsePubDEFromEvents(test.events, test.eventType)
			assert.Equal(t, test.expEvent, events)
			if test.expError != "" {
				assert.ErrorContains(t, err, test.expError)
			}
		})
	}
}

func TestCountAssignedSignings(t *testing.T) {
	tests := []struct {
		name    string
		events  sdk.StringEvents
		address string
		expOut  int64
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
			"member 1",
			1,
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
			"member 1",
			2,
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
					sdk.NewAttribute(types.AttributeKeyAddress, "member 3"),
					sdk.NewAttribute(types.AttributeKeyBindingFactor, hex.EncodeToString([]byte("bindingFactor2"))),
					sdk.NewAttribute(types.AttributeKeyPubNonce, hex.EncodeToString([]byte("pubNonce2"))),
					sdk.NewAttribute(types.AttributeKeyPubD, hex.EncodeToString([]byte("pubD2"))),
					sdk.NewAttribute(types.AttributeKeyPubE, hex.EncodeToString([]byte("pubE2"))),
				)),
			}),
			"member 1",
			2,
		},
		{
			"no event",
			sdk.StringifyEvents([]abci.Event{}),
			"member 3",
			0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := de.CountAssignedSignings(test.events, test.address)
			assert.Equal(t, test.expOut, actual)
		})
	}
}
