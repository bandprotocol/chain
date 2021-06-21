package keeper_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/bandprotocol/chain/v2/testing/testapp"
	"github.com/bandprotocol/chain/v2/x/oracle/keeper"
	"github.com/bandprotocol/chain/v2/x/oracle/types"
)

func TestQueryPendingRequests(t *testing.T) {
	app, ctx, k := testapp.CreateTestInput(true)

	// Add 3 requests
	k.SetRequestLastExpired(ctx, 40)
	k.SetRequest(ctx, 41, defaultRequest())
	k.SetRequest(ctx, 42, defaultRequest())
	k.SetRequest(ctx, 43, defaultRequest())
	k.SetRequestCount(ctx, 43)

	// Fulfill some requests
	k.SetReport(ctx, 41, types.NewReport(testapp.Validators[0].ValAddress, true, nil))
	k.SetReport(ctx, 42, types.NewReport(testapp.Validators[1].ValAddress, true, nil))

	q := keeper.NewQuerier(k, app.LegacyAmino())

	tests := []struct {
		name     string
		args     []string
		expected types.PendingResolveList
	}{
		{
			name:     "Get all pending requests",
			args:     []string{},
			expected: types.PendingResolveList{RequestIds: []int64{41, 42, 43}},
		},
		{
			name:     "Get pending requests for Validators[0]",
			args:     []string{testapp.Validators[0].ValAddress.String()},
			expected: types.PendingResolveList{RequestIds: []int64{42, 43}},
		},
		{
			name:     "Get pending requests for Validators[1]",
			args:     []string{testapp.Validators[1].ValAddress.String()},
			expected: types.PendingResolveList{RequestIds: []int64{41, 43}},
		},
		{
			name:     "Get pending requests for Validators[2]",
			args:     []string{testapp.Validators[2].ValAddress.String()},
			expected: types.PendingResolveList{RequestIds: nil},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raw, err := q(ctx, append([]string{types.QueryPendingRequests}, tt.args...), abci.RequestQuery{})
			require.NoError(t, err)

			var queryRequest types.QueryResult
			require.NoError(t, json.Unmarshal(raw, &queryRequest))

			var pending types.PendingResolveList
			types.ModuleCdc.MustUnmarshalJSON(queryRequest.Result, &pending)

			require.Equal(t, tt.expected, pending)
		})
	}
}
