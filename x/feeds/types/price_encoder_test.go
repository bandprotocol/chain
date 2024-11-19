package types_test

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bandprotocol/chain/v3/x/feeds/types"
)

func TestPriceEncoderEncodingABI(t *testing.T) {
	prices := []types.Price{
		{SignalID: "testSignal", Price: 100, Status: types.PRICE_STATUS_AVAILABLE},
	}

	priceEncoders, err := types.ToPriceEncoders(prices, types.ENCODER_FIXED_POINT_ABI)
	require.NoError(t, err)

	expected := "000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000075bcd15000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000746573745369676e616c0000000000000000000000000000000000000000000000000000000000000064"
	result, err := priceEncoders.EncodeABI(123456789)
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

func TestToPriceEncoders(t *testing.T) {
	signalIDAtom, err := types.StringToBytes32("CS:ATOM-USD")
	require.NoError(t, err)

	signalIDBand, err := types.StringToBytes32("CS:BAND-USD")
	require.NoError(t, err)

	// Define test cases
	testCases := []struct {
		name         string
		prices       []types.Price
		encoder      types.Encoder
		expectResult types.PriceEncoders
		expectError  error
	}{
		{
			name: "success case - fixed-point abi encoder",
			prices: []types.Price{
				{
					SignalID:  "CS:ATOM-USD",
					Price:     1e10,
					Timestamp: 123,
					Status:    types.PRICE_STATUS_AVAILABLE,
				},
				{
					SignalID:  "CS:BAND-USD",
					Price:     1e8,
					Timestamp: 123,
					Status:    types.PRICE_STATUS_AVAILABLE,
				},
			},
			encoder: types.ENCODER_FIXED_POINT_ABI,
			expectResult: types.PriceEncoders{
				{SignalID: signalIDAtom, Price: 1e10},
				{SignalID: signalIDBand, Price: 1e8},
			},
			expectError: nil,
		},
		{
			name: "success case - tick abi encoder",
			prices: []types.Price{
				{
					SignalID:  "CS:ATOM-USD",
					Price:     1e10,
					Timestamp: 123,
					Status:    types.PRICE_STATUS_AVAILABLE,
				},
				{
					SignalID:  "CS:BAND-USD",
					Price:     1e8,
					Timestamp: 123,
					Status:    types.PRICE_STATUS_AVAILABLE,
				},
			},
			encoder: types.ENCODER_TICK_ABI,
			expectResult: types.PriceEncoders{
				{SignalID: signalIDAtom, Price: 285171},
				{SignalID: signalIDBand, Price: 239116},
			},
			expectError: nil,
		},
		{
			name: "success case - price status not in current feeds",
			prices: []types.Price{
				{
					SignalID:  "CS:ATOM-USD",
					Price:     0,
					Timestamp: 0,
					Status:    types.PRICE_STATUS_NOT_IN_CURRENT_FEEDS,
				},
				{
					SignalID:  "CS:BAND-USD",
					Price:     1e8,
					Timestamp: 123,
					Status:    types.PRICE_STATUS_AVAILABLE,
				},
			},
			encoder: types.ENCODER_FIXED_POINT_ABI,
			expectResult: types.PriceEncoders{
				{SignalID: signalIDAtom, Price: 0},
				{SignalID: signalIDBand, Price: 1e8},
			},
		},
		{
			name: "fail case - encoder type unspecified",
			prices: []types.Price{
				{
					SignalID:  "CS:BAND-USD",
					Price:     1e8,
					Timestamp: 123,
					Status:    types.PRICE_STATUS_AVAILABLE,
				},
			},
			encoder:      types.ENCODER_UNSPECIFIED,
			expectResult: types.PriceEncoders{},
			expectError:  types.ErrInvalidEncoder,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			encoderPrices, err := types.ToPriceEncoders(tc.prices, tc.encoder)

			// Check the result
			if tc.expectError != nil {
				require.ErrorContains(t, err, tc.expectError.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectResult, encoderPrices)
			}
		})
	}
}
