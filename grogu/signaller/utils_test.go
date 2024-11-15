package signaller

import (
	"testing"

	"github.com/stretchr/testify/assert"

	sdk "github.com/cosmos/cosmos-sdk/types"

	bothan "github.com/bandprotocol/bothan/bothan-api/client/go-client/proto/bothan/v1"

	"github.com/bandprotocol/chain/v3/x/feeds/types"
)

func TestIsDeviated(t *testing.T) {
	tests := []struct {
		name                string
		deviationBasisPoint int64
		oldPrice            uint64
		newPrice            uint64
		expectedDeviated    bool
	}{
		{"No deviation", 100, 1000, 1000, false},
		{"Below threshold", 100, 1000, 1001, false},
		{"Exact threshold", 100, 1000, 1010, true},
		{"Above threshold", 100, 1000, 1100, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isDeviated(tt.deviationBasisPoint, tt.oldPrice, tt.newPrice)
			assert.Equal(t, tt.expectedDeviated, result)
		})
	}
}

func TestConvertPriceData(t *testing.T) {
	tests := []struct {
		name           string
		priceData      *bothan.Price
		expectedResult types.SignalPrice
		expectingError bool
	}{
		{
			"Unsupported price status",
			&bothan.Price{Status: bothan.Status_STATUS_UNSUPPORTED, SignalId: "signal1"},
			types.SignalPrice{Status: types.SIGNAL_PRICE_STATUS_UNSUPPORTED, SignalID: "signal1", Price: 0},
			false,
		},
		{
			"Unavailable price status",
			&bothan.Price{Status: bothan.Status_STATUS_UNAVAILABLE, SignalId: "signal2"},
			types.SignalPrice{Status: types.SIGNAL_PRICE_STATUS_UNAVAILABLE, SignalID: "signal2", Price: 0},
			false,
		},
		{
			"Available price status",
			&bothan.Price{
				Status:   bothan.Status_STATUS_AVAILABLE,
				SignalId: "signal3",
				Price:    123456000000,
			},
			types.SignalPrice{Status: types.SIGNAL_PRICE_STATUS_AVAILABLE, SignalID: "signal3", Price: 123456000000},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertPriceData(tt.priceData)
			if tt.expectingError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}
		})
	}
}

func TestCalculateAssignedTime(t *testing.T) {
	valAddr := sdk.ValAddress([]byte("validator1"))
	interval := int64(3600) // 1 hour
	timestamp := int64(100000000)

	dpOffset := uint64(30)
	dpStart := uint64(50)

	result := calculateAssignedTime(valAddr, interval, timestamp, dpOffset, dpStart)
	assert.Equal(t, int64(100002016), result.Unix())
}

func TestSliceToMap(t *testing.T) {
	type testStruct struct {
		ID    string
		Value int
	}
	slice := []testStruct{
		{ID: "a", Value: 1},
		{ID: "b", Value: 2},
		{ID: "c", Value: 3},
	}

	keyFunc := func(ts testStruct) string {
		return ts.ID
	}

	expectedMap := map[string]testStruct{
		"a": {ID: "a", Value: 1},
		"b": {ID: "b", Value: 2},
		"c": {ID: "c", Value: 3},
	}

	result := sliceToMap(slice, keyFunc)
	assert.Equal(t, expectedMap, result)
}
