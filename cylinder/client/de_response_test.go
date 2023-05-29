package client_test

// TODO-CYLINDER: Use the real query
// func TestGetRemaining(t *testing.T) {
// 	tests := []struct {
// 		name            string
// 		queryDEResponse *types.QueryDEResponse
// 		exp             uint64
// 	}{
// 		{
// 			name: "No DE left",
// 			queryDEResponse: &types.QueryDEResponse{
// 				Head: 10,
// 				Tail: 10,
// 			},
// 			exp: 0,
// 		},
// 		{
// 			name: "Has some DE left",
// 			queryDEResponse: &types.QueryDEResponse{
// 				Head: 10,
// 				Tail: 20,
// 			},
// 			exp: 10,
// 		},
// 	}

// 	for _, test := range tests {
// 		t.Run(test.name, func(t *testing.T) {
// 			deResponse := client.NewDEResponse(test.queryDEResponse)
// 			remaining := deResponse.GetRemaining()
// 			assert.Equal(t, test.exp, remaining)
// 		})
// 	}
// }
