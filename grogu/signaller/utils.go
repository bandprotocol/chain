package signaller

import (
	"crypto/sha256"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	proto "github.com/bandprotocol/bothan/bothan-api/client/go-client/query"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/feeds/types"
)

// isDeviated checks if the deviation between the old price and the new price
// exceeds a given threshold in thousandths.

// Parameters:
// - deviationInThousandth: the allowable deviation in thousandths (1/1000th)
// - oldPrice: the original price
// - newPrice: the new price to compare against the original
//
// Returns:
//   - bool: true if the deviation is greater than or equal to the given threshold,
//     false otherwise
//
// The deviation is calculated as follows:
//  1. Calculate the absolute difference between the new price and the old price.
//  2. Compute the deviation in thousandths by dividing the difference by the
//     original price and multiplying by 1000.
//  3. Check if the calculated deviation meets or exceeds the allowable deviation.
func isDeviated(deviationInThousandth int64, oldPrice uint64, newPrice uint64) bool {
	// Calculate the deviation
	diff := math.Abs(float64(newPrice) - float64(oldPrice))
	dev := int64((diff / float64(oldPrice)) * 1000)

	// Check if the new price deviation is meets or exceeds the bounds
	return deviationInThousandth <= dev
}

func convertPriceData(priceData *proto.PriceData) (types.SubmitPrice, error) {
	switch priceData.PriceStatus {
	case proto.PriceStatus_PRICE_STATUS_UNSPECIFIED:
		// This should never happen
		panic("unspecified price status")
	case proto.PriceStatus_PRICE_STATUS_UNSUPPORTED:
		return types.SubmitPrice{
			PriceStatus: types.PriceStatusUnsupported,
			SignalID:    priceData.SignalId,
			Price:       0,
		}, nil
	case proto.PriceStatus_PRICE_STATUS_UNAVAILABLE:
		return types.SubmitPrice{
			PriceStatus: types.PriceStatusUnavailable,
			SignalID:    priceData.SignalId,
			Price:       0,
		}, nil
	case proto.PriceStatus_PRICE_STATUS_AVAILABLE:
		price, err := safeConvert(priceData.Price)
		if err != nil {
			return types.SubmitPrice{}, err
		}
		return types.SubmitPrice{
			PriceStatus: types.PriceStatusAvailable,
			SignalID:    priceData.SignalId,
			Price:       price,
		}, nil
	default:
		// Handle unexpected price status
		return types.SubmitPrice{}, fmt.Errorf("unexpected price status: %v", priceData.PriceStatus)
	}
}

func safeConvert(price string) (uint64, error) {
	if price == "" {
		return 0, nil
	}

	parsedPrice, err := strconv.ParseFloat(strings.TrimSpace(price), 64)
	if err != nil {
		return 0, err
	}

	if parsedPrice < 0 {
		return 0, fmt.Errorf("price is negative")
	}

	if parsedPrice > UpperBound {
		return 0, fmt.Errorf("price is above allowable limit")
	}

	return uint64(parsedPrice * Multiplier), nil
}

// calculateAssignedTime calculates the assigned time for a validator to send prices
//
// The assigned time is calculated as follows:
//  1. Hash the validator address and timestamp using SHA256.
//  2. Calculate the offset by taking the modulo of the hashed value with distribution offset percentage and adding distribution percentage start.
//  3. Calculate the time offset by multiplying the interval with the offset and dividing by 100.
//  4. Add the time offset to the received timestamp to get the assigned time.
func calculateAssignedTime(
	valAddr sdk.ValAddress,
	interval int64,
	timestamp int64,
	dpOffset uint64,
	dpStart uint64,
) time.Time {
	hashed := sha256.Sum256(append(valAddr.Bytes(), sdk.Uint64ToBigEndian(uint64(timestamp))...))
	offset := sdk.BigEndianToUint64(
		hashed[:],
	)%dpOffset + dpStart
	timeOffset := interval * int64(offset) / 100
	// add time buffer to ensure the assigned time is not too early
	return time.Unix(timestamp, 0).Add(time.Duration(timeOffset) * time.Second)
}

func sliceToMap[T any, K comparable](slice []T, keyFunc func(T) K) map[K]T {
	resultMap := make(map[K]T)
	for _, item := range slice {
		key := keyFunc(item)
		resultMap[key] = item
	}
	return resultMap
}
