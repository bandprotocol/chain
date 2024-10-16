package types_test

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bandprotocol/chain/v3/x/feeds/types"
)

func TestABIEncodeFeedsPriceData(t *testing.T) {
	sp := types.SignalPrice{
		SignalID: "testSignal",
		Price:    100,
	}
	f := types.FeedsPriceData{
		SignalPrices: []types.SignalPrice{sp},
		Timestamp:    123456789,
	}

	expected := "000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000075bcd15000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000746573745369676e616c0000000000000000000000000000000000000000000000000000000000000064"
	result, err := f.ABIEncode()
	require.NoError(t, err)
	require.Equal(t, expected, hex.EncodeToString(result))
}
