package client_test

import (
	"fmt"
	"testing"

	"github.com/bandprotocol/chain/v2/cylinder/client"
	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/tss/types"
	"github.com/stretchr/testify/assert"
)

func TestGetRound1Commitment(t *testing.T) {
	tests := []struct {
		name               string
		queryGroupResponse *types.QueryGroupResponse
		memberID           tss.MemberID
		expectedCommitment *types.Round1Commitments
		expectedError      error
	}{
		{
			name: "Existing Member ID",
			queryGroupResponse: &types.QueryGroupResponse{
				AllRound1Commitments: map[uint64]types.Round1Commitments{
					1: {},
				},
			},
			memberID:           1,
			expectedCommitment: &types.Round1Commitments{},
			expectedError:      nil,
		},
		{
			name: "Non-Existing Member ID",
			queryGroupResponse: &types.QueryGroupResponse{
				AllRound1Commitments: map[uint64]types.Round1Commitments{},
			},
			memberID:           2,
			expectedCommitment: nil,
			expectedError:      fmt.Errorf("No Round1Commitment from MemberID(2)"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			groupResponse := client.NewGroupResponse(test.queryGroupResponse)

			commitment, err := groupResponse.GetRound1Commitment(test.memberID)

			assert.Equal(t, test.expectedError, err)
			assert.Equal(t, test.expectedCommitment, commitment)
		})
	}
}
