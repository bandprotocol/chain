package types_test

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bandprotocol/chain/v3/x/feeds/types"
)

func TestABIEncodeFeedsPriceData(t *testing.T) {
	p := types.Price{
		SignalID: "testSignal",
		Price:    100,
	}
	f := types.FeedsPriceData{
		Prices:    []types.Price{p},
		Timestamp: 123456789,
	}

	expected := "000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000075bcd15000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000746573745369676e616c0000000000000000000000000000000000000000000000000000000000000064"
	result, err := f.ABIEncode()
	require.NoError(t, err)
	require.Equal(t, expected, hex.EncodeToString(result))
}

func TestValidateEncoder(t *testing.T) {
	// validate encoder
	err := types.ValidateEncoder(1)
	require.NoError(t, err)

	// invalid encoder
	err = types.ValidateEncoder(999)
	require.Error(t, err)
}
