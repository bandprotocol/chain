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
			name: "Existing MemberID",
			queryGroupResponse: &types.QueryGroupResponse{
				AllRound1Data: []types.Round1Data{
					{
						MemberID: 1,
					},
				},
			},
			memberID:      1,
			expectedData:  types.Round1Data{MemberID: 1},
			expectedError: nil,
		},
		{
			name: "No data from MemberID",
			queryGroupResponse: &types.QueryGroupResponse{
				AllRound1Data: []types.Round1Data{
					{
						MemberID: 1,
					},
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
