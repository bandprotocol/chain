package de_test

import (
	"encoding/hex"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"

	"github.com/bandprotocol/chain/v2/cylinder/workers/de"
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

func TestParseSubmitSignEvents(t *testing.T) {
	tests := []struct {
		name     string
		events   sdk.StringEvents
		expEvent []de.Event
		expError string
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
			[]de.Event{
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
			[]de.Event{
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
			[]de.Event{
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
			nil,
			"encoding/hex: invalid byte",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			events, err := de.ParseSubmitSignEvents(test.events)
			assert.Equal(t, test.expEvent, events)
			if test.expError != "" {
				assert.ErrorContains(t, err, test.expError)
			}
		})
	}
}
