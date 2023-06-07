package client_test

import (
	"testing"

	"github.com/bandprotocol/chain/v2/cylinder/client"
	"github.com/bandprotocol/chain/v2/x/tss/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/stretchr/testify/assert"
)

func TestGetRemaining(t *testing.T) {
	tests := []struct {
		name            string
		queryDEResponse *types.QueryDEResponse
		exp             uint64
	}{
		{
			name: "No DE left",
			queryDEResponse: &types.QueryDEResponse{
				Pagination: &query.PageResponse{
					Total: 0,
				},
			},
			exp: 0,
		},
		{
			name: "Has some DE left",
			queryDEResponse: &types.QueryDEResponse{
				Pagination: &query.PageResponse{
					Total: 10,
				},
			},
			exp: 10,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			deResponse := client.NewDEResponse(test.queryDEResponse)
			remaining := deResponse.GetRemaining()
			assert.Equal(t, test.exp, remaining)
		})
	}
}
