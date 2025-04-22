package client_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/bandprotocol/chain/v3/cylinder/client"
	bandtsstypes "github.com/bandprotocol/chain/v3/x/bandtss/types"
)

func TestFindMembersByAddress(t *testing.T) {
	tests := []struct {
		name            string
		members         []*bandtsstypes.Member
		address         string
		expectedMembers []bandtsstypes.Member
		expectedError   error
	}{
		{
			name: "Member found",
			members: []*bandtsstypes.Member{
				{
					Address:  "band address 1",
					GroupID:  1,
					IsActive: true,
					Since:    time.Unix(38044418, 0),
				},
				{
					Address:  "band address 2",
					GroupID:  1,
					IsActive: true,
					Since:    time.Unix(38044418, 0),
				},
			},
			address: "band address 1",
			expectedMembers: []bandtsstypes.Member{
				{
					Address:  "band address 1",
					GroupID:  1,
					IsActive: true,
					Since:    time.Unix(38044418, 0),
				},
			},
			expectedError: nil,
		},
		{
			name: "Member not found",
			members: []*bandtsstypes.Member{
				{
					Address:  "band address 1",
					GroupID:  1,
					IsActive: true,
					Since:    time.Unix(38044418, 0),
				},
				{
					Address:  "band address 2",
					GroupID:  1,
					IsActive: true,
					Since:    time.Unix(38044418, 0),
				},
			},
			address:         "band address 3",
			expectedMembers: []bandtsstypes.Member{},
			expectedError:   fmt.Errorf("member not found"),
		},
		{
			name:            "Empty members list",
			members:         []*bandtsstypes.Member{},
			address:         "band address 1",
			expectedMembers: []bandtsstypes.Member{},
			expectedError:   fmt.Errorf("member not found"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			membersResponse := client.MembersResponse{
				Members: test.members,
			}
			members, err := membersResponse.FindMembersByAddress(test.address)
			assert.Equal(t, test.expectedError, err)
			assert.Equal(t, test.expectedMembers, members)
		})
	}
}
