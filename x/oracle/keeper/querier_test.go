package oraclekeeper_test

// TODO: fix test
//func TestQueryPendingRequests(t *testing.T) {
//	app, ctx, k := testapp.CreateTestInput(true)
//
//	// Add 3 requests
//	k.SetRequestLastExpired(ctx, 40)
//	k.SetRequest(ctx, 41, defaultRequest())
//	k.SetRequest(ctx, 42, defaultRequest())
//	k.SetRequest(ctx, 43, defaultRequest())
//	k.SetRequestCount(ctx, 43)
//
//	// Fulfill some requests
//	k.SetReport(ctx, 41, oracletypes.NewReport(testapp.Validators[0].ValAddress, true, nil))
//	k.SetReport(ctx, 42, oracletypes.NewReport(testapp.Validators[1].ValAddress, true, nil))
//
//	q := oraclkeeper.NewQuerier(k, app.LegacyAmino())
//
//	tests := []struct {
//		name     string
//		args     []string
//		expected oracletypes.PendingResolveList
//	}{
//
//		{
//			name:     "Get all pending requests",
//			args:     []string{},
//			expected: oracletypes.PendingResolveList{RequestIds: []int64{41, 42, 43}},
//		},
//		{
//			name:     "Get pending requests for Validators[0]",
//			args:     []string{testapp.Validators[0].ValAddress.String()},
//			expected: oracletypes.PendingResolveList{RequestIds: []int64{42, 43}},
//		},
//		{
//			name:     "Get pending requests for Validators[1]",
//			args:     []string{testapp.Validators[1].ValAddress.String()},
//			expected: oracletypes.PendingResolveList{RequestIds: []int64{41, 43}},
//		},
//		{
//			name:     "Get pending requests for Validators[2]",
//			args:     []string{testapp.Validators[2].ValAddress.String()},
//			expected: oracletypes.PendingResolveList{RequestIds: nil},
//		},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			raw, err := q(ctx, append([]string{oracletypes.QueryPendingRequests}, tt.args...), abci.RequestQuery{})
//			require.NoError(t, err)
//
//			var queryRequest commontypes.QueryResult
//			require.NoError(t, json.Unmarshal(raw, &queryRequest))
//
//			var pending oracletypes.PendingResolveList
//			oracletypes.ModuleCdc.MustUnmarshalJSON(queryRequest.Result, &pending)
//
//			require.Equal(t, tt.expected, pending)
//		})
//	}
//}
