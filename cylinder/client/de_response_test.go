package client_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/bandprotocol/chain/v3/cylinder/client"
	"github.com/bandprotocol/chain/v3/x/tss/types"
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
			deResponse := client.NewDEResponse(test.queryDEResponse, 10)
			remaining := deResponse.GetRemaining()
			assert.Equal(t, test.exp, remaining)
		})
	}
}
