package client_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bandprotocol/chain/v3/cylinder/client"
	"github.com/bandprotocol/chain/v3/x/tss/types"
)

func TestIsActive(t *testing.T) {
	tests := []struct {
		name           string
		members        []types.Member
		address        string
		expectedActive bool
		expectedError  error
	}{
		{
			name: "Active member",
			members: []types.Member{
				{
					Address:  "band address 1",
					IsActive: true,
				},
				{
					Address:  "band address 2",
					IsActive: false,
				},
			},
			address:        "band address 1",
			expectedActive: true,
			expectedError:  nil,
		},
		{
			name: "Inactive member",
			members: []types.Member{
				{
					Address:  "band address 1",
					IsActive: true,
				},
				{
					Address:  "band address 2",
					IsActive: false,
				},
			},
			address:        "band address 2",
			expectedActive: false,
			expectedError:  nil,
		},
		{
			name: "Member not found",
			members: []types.Member{
				{
					Address:  "band address 1",
					IsActive: true,
				},
				{
					Address:  "band address 2",
					IsActive: false,
				},
			},
			address:        "band address 3",
			expectedActive: false,
			expectedError:  fmt.Errorf("member not found"),
		},
		{
			name:           "Empty members list",
			members:        []types.Member{},
			address:        "band address 1",
			expectedActive: false,
			expectedError:  fmt.Errorf("member not found"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			membersResponse := client.MembersResponse{
				Members: test.members,
			}
			isActive, err := membersResponse.IsActive(test.address)
			assert.Equal(t, test.expectedError, err)
			assert.Equal(t, test.expectedActive, isActive)
		})
	}
}
