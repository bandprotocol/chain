package client_test

import (
	"fmt"
	"testing"

	"github.com/bandprotocol/chain/v2/cylinder/client"
	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/tss/types"
	"github.com/stretchr/testify/assert"
)

func TestGetRound1Data(t *testing.T) {
	tests := []struct {
		name               string
		queryGroupResponse *types.QueryGroupResponse
		memberID           tss.MemberID
		expectedData       types.Round1Data
		expectedError      error
	}{
		{
			name: "Existing Member ID",
			queryGroupResponse: &types.QueryGroupResponse{
				AllRound1Data: []*types.Round1Data{
					{},
				},
			},
			memberID:      1,
			expectedData:  types.Round1Data{},
			expectedError: nil,
		},
		{
			name: "Non-Existing Member ID",
			queryGroupResponse: &types.QueryGroupResponse{
				AllRound1Data: []*types.Round1Data{},
			},
			memberID:      2,
			expectedData:  types.Round1Data{},
			expectedError: fmt.Errorf("No MemberID(2) in the group"),
		},
		{
			name: "No data for Member yet",
			queryGroupResponse: &types.QueryGroupResponse{
				AllRound1Data: []*types.Round1Data{
					{},
					nil,
				},
			},
			memberID:      2,
			expectedData:  types.Round1Data{},
			expectedError: fmt.Errorf("No Round1Data from MemberID(2)"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			groupResponse := client.NewGroupResponse(test.queryGroupResponse)
			data, err := groupResponse.GetRound1Data(test.memberID)
			assert.Equal(t, test.expectedError, err)
			assert.Equal(t, test.expectedData, data)
		})
	}
}
