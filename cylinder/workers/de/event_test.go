package de_test

import (
	"encoding/hex"
	"testing"

	"github.com/bandprotocol/chain/v2/cylinder/workers/de"
	"github.com/bandprotocol/chain/v2/x/tss/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
)

func TestParseSubmitSignEvent(t *testing.T) {
	tests := []struct {
		name     string
		log      sdk.ABCIMessageLog
		expEvent *de.Event
		expError string
	}{
		{
			"success",
			sdk.NewABCIMessageLog(0, "", sdk.Events{
				sdk.NewEvent(
					types.EventTypeSubmitSign,
					sdk.NewAttribute(types.AttributeKeySigningID, "1"),
					sdk.NewAttribute(types.AttributeKeyGroupID, "1"),
					sdk.NewAttribute(types.AttributeKeyMemberID, "2"),
					sdk.NewAttribute(types.AttributeKeyMember, "member 1"),
					sdk.NewAttribute(types.AttributeKeyPubD, hex.EncodeToString([]byte("pubD"))),
					sdk.NewAttribute(types.AttributeKeyPubE, hex.EncodeToString([]byte("pubE"))),
					sdk.NewAttribute(types.AttributeKeySignature, hex.EncodeToString([]byte("signature"))),
				),
			}),
			&de.Event{
				PubDE: types.DE{
					PubD: []byte("pubD"),
					PubE: []byte("pubE"),
				},
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
					types.EventTypeSubmitSign,
					sdk.NewAttribute(types.AttributeKeySigningID, "1"),
					sdk.NewAttribute(types.AttributeKeyGroupID, "1"),
					sdk.NewAttribute(types.AttributeKeyMemberID, "2"),
					sdk.NewAttribute(types.AttributeKeyMember, "member 1"),
					sdk.NewAttribute(types.AttributeKeyPubD, "invalid hex"),
					sdk.NewAttribute(types.AttributeKeyPubE, hex.EncodeToString([]byte("pubE"))),
					sdk.NewAttribute(types.AttributeKeySignature, hex.EncodeToString([]byte("signature"))),
				),
			}),
			nil,
			"encoding/hex: invalid byte",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			event, err := de.ParseSubmitSignEvent(test.log)
			assert.Equal(t, test.expEvent, event)
			if test.expError != "" {
				assert.ErrorContains(t, err, test.expError)
			}
		})
	}
}
