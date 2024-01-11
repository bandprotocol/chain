package client_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bandprotocol/chain/v2/cylinder/client"
	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

func TestGetRound1Info(t *testing.T) {
	tests := []struct {
		name               string
		queryGroupResponse *types.QueryGroupResponse
		memberID           tss.MemberID
		expectedData       types.Round1Info
		expectedError      error
	}{
		{
			name: "Existing MemberID",
			queryGroupResponse: &types.QueryGroupResponse{
				Round1Infos: []types.Round1Info{
					{
						MemberID: 1,
					},
				},
			},
			memberID:      1,
			expectedData:  types.Round1Info{MemberID: 1},
			expectedError: nil,
		},
		{
			name: "No data from MemberID",
			queryGroupResponse: &types.QueryGroupResponse{
				Round1Infos: []types.Round1Info{
					{
						MemberID: 1,
					},
				},
			},
			memberID:      2,
			expectedData:  types.Round1Info{},
			expectedError: fmt.Errorf("no Round1Info from MemberID(2)"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			groupResponse := client.NewGroupResponse(test.queryGroupResponse)
			data, err := groupResponse.GetRound1Info(test.memberID)
			assert.Equal(t, test.expectedError, err)
			assert.Equal(t, test.expectedData, data)
		})
	}
}

func TestGetRound2Infos(t *testing.T) {
	tests := []struct {
		name               string
		queryGroupResponse *types.QueryGroupResponse
		memberID           tss.MemberID
		expectedData       types.Round2Info
		expectedError      error
	}{
		{
			name: "Existing MemberID",
			queryGroupResponse: &types.QueryGroupResponse{
				Round2Infos: []types.Round2Info{
					{
						MemberID: 1,
					},
				},
			},
			memberID:      1,
			expectedData:  types.Round2Info{MemberID: 1},
			expectedError: nil,
		},
		{
			name: "No data from MemberID",
			queryGroupResponse: &types.QueryGroupResponse{
				Round2Infos: []types.Round2Info{
					{
						MemberID: 1,
					},
				},
			},
			memberID:      2,
			expectedData:  types.Round2Info{},
			expectedError: fmt.Errorf("no Round2Info from MemberID(2)"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			groupResponse := client.NewGroupResponse(test.queryGroupResponse)
			data, err := groupResponse.GetRound2Info(test.memberID)
			assert.Equal(t, test.expectedError, err)
			assert.Equal(t, test.expectedData, data)
		})
	}
}

func TestGetEncryptedSecretShare(t *testing.T) {
	tests := []struct {
		name               string
		queryGroupResponse *types.QueryGroupResponse
		senderID           tss.MemberID
		receiverID         tss.MemberID
		expectedShare      tss.EncSecretShare
		expectedError      error
	}{
		{
			name: "Existing share",
			queryGroupResponse: &types.QueryGroupResponse{
				Round2Infos: []types.Round2Info{
					{
						MemberID:              1,
						EncryptedSecretShares: tss.EncSecretShares{[]byte("share1"), []byte("share2")},
					},
				},
			},
			senderID:      1,
			receiverID:    1,
			expectedShare: []byte("share1"),
			expectedError: nil,
		},
		{
			name: "Invalid ReceiverID",
			queryGroupResponse: &types.QueryGroupResponse{
				Round2Infos: []types.Round2Info{
					{
						MemberID:              1,
						EncryptedSecretShares: tss.EncSecretShares{[]byte("share1"), []byte("share2")},
					},
				},
			},
			senderID:      2,
			receiverID:    1,
			expectedShare: nil,
			expectedError: fmt.Errorf("no Round2Info from MemberID(2)"),
		},
		{
			name: "Invalid SenderID",
			queryGroupResponse: &types.QueryGroupResponse{
				Round2Infos: []types.Round2Info{
					{
						MemberID:              1,
						EncryptedSecretShares: tss.EncSecretShares{[]byte("share1"), []byte("share2")},
					},
				},
			},
			senderID:      1,
			receiverID:    4,
			expectedShare: nil,
			expectedError: fmt.Errorf("no encrypted secret share from MemberID(1) to MemberID(4)"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			groupResponse := client.NewGroupResponse(test.queryGroupResponse)
			share, err := groupResponse.GetEncryptedSecretShare(test.senderID, test.receiverID)
			assert.Equal(t, test.expectedError, err)
			assert.Equal(t, test.expectedShare, share)
		})
	}
}
