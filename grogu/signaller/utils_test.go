package signaller

import (
	"testing"

	proto "github.com/bandprotocol/bothan/bothan-api/client/go-client/query"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"

	"github.com/bandprotocol/chain/v2/x/feeds/types"
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
		priceData      *proto.PriceData
		expectedResult types.SignalPrice
		expectingError bool
	}{
		{
			"Unsupported price status",
			&proto.PriceData{PriceStatus: proto.PriceStatus_PRICE_STATUS_UNSUPPORTED, SignalId: "signal1"},
			types.SignalPrice{PriceStatus: types.PriceStatusUnsupported, SignalID: "signal1", Price: 0},
			false,
		},
		{
			"Unavailable price status",
			&proto.PriceData{PriceStatus: proto.PriceStatus_PRICE_STATUS_UNAVAILABLE, SignalId: "signal2"},
			types.SignalPrice{PriceStatus: types.PriceStatusUnavailable, SignalID: "signal2", Price: 0},
			false,
		},
		{
			"Available price status",
			&proto.PriceData{
				PriceStatus: proto.PriceStatus_PRICE_STATUS_AVAILABLE,
				SignalId:    "signal3",
				Price:       "123.456",
			},
			types.SignalPrice{PriceStatus: types.PriceStatusAvailable, SignalID: "signal3", Price: 123456000000},
			false,
		},
		{
			"Invalid price value",
			&proto.PriceData{
				PriceStatus: proto.PriceStatus_PRICE_STATUS_AVAILABLE,
				SignalId:    "signal4",
				Price:       "invalid",
			},
			types.SignalPrice{},
			true,
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

func TestSafeConvert(t *testing.T) {
	tests := []struct {
		name           string
		price          string
		expectedResult uint64
		expectingError bool
	}{
		{"Empty string", "", 0, false},
		{"Valid price", "123.456", 123456000000, false},
		{"Negative price", "-123.456", 0, true},
		{"Above upper bound", "100000000001", 0, true},
		{"Invalid price format", "abc123", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := safeConvert(tt.price)
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
