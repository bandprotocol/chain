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
