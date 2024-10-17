package testutil

import gomock "go.uber.org/mock/gomock"

// ChainGoMockCalls is a helper function to chain multiple gomock calls together.
func ChainGoMockCalls(calls ...*gomock.Call) *gomock.Call {
	for i := 1; i < len(calls); i++ {
		calls[i].After(calls[i-1])
	}

	// return the last call
	return calls[len(calls)-1]
}
