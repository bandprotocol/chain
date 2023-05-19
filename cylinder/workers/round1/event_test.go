package round1_test

import (
	"errors"
	"testing"

	"github.com/bandprotocol/chain/v2/cylinder/workers/round1"
	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/stretchr/testify/assert"
)

func TestGetMemberID(t *testing.T) {
	tests := []struct {
		name        string
		event       round1.Event
		expMemberID tss.MemberID
		expError    error
	}{
		{
			"not in the member",
			round1.Event{
				Members: []string{
					"b",
					"c",
				},
			},
			0,
			errors.New("failed to find member in the event"),
		},
		{
			"first member",
			round1.Event{
				Members: []string{
					"a",
					"b",
					"c",
				},
			},
			1,
			nil,
		},
		{
			"last member",
			round1.Event{
				Members: []string{
					"b",
					"c",
					"a",
				},
			},
			3,
			nil,
		},
		{
			"no member in the group",
			round1.Event{
				Members: []string{},
			},
			0,
			errors.New("failed to find member in the event"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			memberID, err := test.event.GetMemberID("a")
			assert.Equal(t, test.expError, err)
			assert.Equal(t, test.expMemberID, memberID)
		})
	}
}
