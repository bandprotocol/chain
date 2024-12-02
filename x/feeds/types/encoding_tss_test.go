package types_test

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bandprotocol/chain/v3/pkg/tss"
	"github.com/bandprotocol/chain/v3/x/feeds/types"
)

func TestEncoderPrefix(t *testing.T) {
	require.Equal(t, []byte(types.EncoderFixedPointABIPrefix), tss.Hash([]byte("FixedPointABI"))[:4])
	require.Equal(t, []byte(types.EncoderTickABIPrefix), tss.Hash([]byte("TickABI"))[:4])
}

func TestPriceEncoderEncodingABI(t *testing.T) {
	prices := []types.Price{
		{SignalID: "testSignal", Price: 100, Status: types.PRICE_STATUS_AVAILABLE},
	}

	result, err := types.EncodeTSS(prices, 123456789, types.ENCODER_FIXED_POINT_ABI)
	require.NoError(t, err)

	expected := "cba0ad5a000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000075bcd15000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000746573745369676e616c0000000000000000000000000000000000000000000000000000000000000064"
	require.Equal(t, expected, hex.EncodeToString(result))
}

func TestTickPriceEncoderEncodingABI(t *testing.T) {
	prices := []types.Price{
		{SignalID: "testSignal", Price: 1e10, Status: types.PRICE_STATUS_AVAILABLE},
	}

	result, err := types.EncodeTSS(prices, 123456789, types.ENCODER_TICK_ABI)
	require.NoError(t, err)

	expected := "db99b2b3000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000075bcd15000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000746573745369676e616c00000000000000000000000000000000000000000000000000000000000459f3"
	require.Equal(t, expected, hex.EncodeToString(result))
}

func TestToRelayPrices(t *testing.T) {
	signalIDAtom, err := types.StringToBytes32("CS:ATOM-USD")
	require.NoError(t, err)

	signalIDBand, err := types.StringToBytes32("CS:BAND-USD")
	require.NoError(t, err)

	// Define test cases
	testCases := []struct {
		name         string
		prices       []types.Price
		expectResult []types.RelayPrice
		expectError  error
	}{
		{
			name: "success case",
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
			expectResult: []types.RelayPrice{
				{SignalID: signalIDAtom, Price: 1e10},
				{SignalID: signalIDBand, Price: 1e8},
			},
			expectError: nil,
		},
		{
			name: "fail case - signalID is too long",
			prices: []types.Price{
				{
					SignalID:  "this-is-too-long-signal-id-that-cannot-be-converted",
					Price:     1e8,
					Timestamp: 123,
					Status:    types.PRICE_STATUS_AVAILABLE,
				},
			},
			expectError: types.ErrInvalidSignal,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			encoderPrices, err := types.ToRelayPrices(tc.prices)

			// Check the result
			if tc.expectError != nil {
				require.ErrorIs(t, err, tc.expectError)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectResult, encoderPrices)
			}
		})
	}
}

func TestToRelayTickPrices(t *testing.T) {
	signalIDAtom, err := types.StringToBytes32("CS:ATOM-USD")
	require.NoError(t, err)

	signalIDBand, err := types.StringToBytes32("CS:BAND-USD")
	require.NoError(t, err)

	// Define test cases
	testCases := []struct {
		name         string
		prices       []types.Price
		encoder      types.Encoder
		expectResult []types.RelayPrice
		expectError  error
	}{
		{
			name: "success case",
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
			expectResult: []types.RelayPrice{
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
			expectResult: []types.RelayPrice{
				{SignalID: signalIDAtom, Price: 0},
				{SignalID: signalIDBand, Price: 239116},
			},
		},
		{
			name: "fail case - signalID is too long",
			prices: []types.Price{
				{
					SignalID:  "this-is-too-long-signal-id-that-cannot-be-converted",
					Price:     1e8,
					Timestamp: 123,
					Status:    types.PRICE_STATUS_AVAILABLE,
				},
			},
			expectError: types.ErrInvalidSignal,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			encoderPrices, err := types.ToRelayTickPrices(tc.prices)

			// Check the result
			if tc.expectError != nil {
				require.ErrorIs(t, err, tc.expectError)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectResult, encoderPrices)
			}
		})
	}
}
