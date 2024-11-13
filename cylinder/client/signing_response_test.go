package client_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bandprotocol/chain/v3/cylinder/client"
	"github.com/bandprotocol/chain/v3/pkg/tss"
	"github.com/bandprotocol/chain/v3/x/tss/types"
)

func TestGetMemberIDs(t *testing.T) {
	tests := []struct {
		name                 string
		querySigningResponse *types.QuerySigningResponse
		expect               []tss.MemberID
	}{
		{
			name: "One member",
			querySigningResponse: &types.QuerySigningResponse{
				SigningResult: types.SigningResult{
					Signing: types.Signing{},
					CurrentSigningAttempt: &types.SigningAttempt{
						AssignedMembers: []types.AssignedMember{{MemberID: 1}},
					},
				},
			},
			expect: []tss.MemberID{1},
		},
		{
			name: "No data from MemberID",
			querySigningResponse: &types.QuerySigningResponse{
				SigningResult: types.SigningResult{
					Signing: types.Signing{},
					CurrentSigningAttempt: &types.SigningAttempt{
						AssignedMembers: []types.AssignedMember{{MemberID: 2}, {MemberID: 3}},
					},
				},
			},
			expect: []tss.MemberID{2, 3},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			signingResponse := client.NewSigningResponse(test.querySigningResponse)
			mids := signingResponse.GetMemberIDs()
			assert.Equal(t, test.expect, mids)
		})
	}
}

func TestGetAssignedMember(t *testing.T) {
	tests := []struct {
		name                 string
		querySigningResponse *types.QuerySigningResponse
		address              string
		expectedValue        types.AssignedMember
		expectedError        error
	}{
		{
			name: "Existing MemberID",
			querySigningResponse: &types.QuerySigningResponse{
				SigningResult: types.SigningResult{
					Signing: types.Signing{},
					CurrentSigningAttempt: &types.SigningAttempt{
						AssignedMembers: []types.AssignedMember{
							{MemberID: 1, Address: "band address 1"},
							{MemberID: 2, Address: "band address 2"},
						},
					},
				},
			},
			address: "band address 2",
			expectedValue: types.AssignedMember{
				MemberID: 2,
				Address:  "band address 2",
			},
			expectedError: nil,
		},
		{
			name: "No member",
			querySigningResponse: &types.QuerySigningResponse{
				SigningResult: types.SigningResult{
					Signing: types.Signing{},
					CurrentSigningAttempt: &types.SigningAttempt{
						AssignedMembers: []types.AssignedMember{
							{MemberID: 1, Address: "band address 1"},
							{MemberID: 2, Address: "band address 2"},
						},
					},
				},
			},
			address:       "band address 3",
			expectedValue: types.AssignedMember{},
			expectedError: fmt.Errorf("band address 3 is not the assigned member"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			signingResponse := client.NewSigningResponse(test.querySigningResponse)
			assignedMember, err := signingResponse.GetAssignedMember(test.address)
			assert.Equal(t, test.expectedError, err)
			assert.Equal(t, test.expectedValue, assignedMember)
		})
	}
}
