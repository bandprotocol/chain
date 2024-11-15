package signaller

import (
	"crypto/sha256"
	"fmt"
	"math"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	bothan "github.com/bandprotocol/bothan/bothan-api/client/go-client/proto/bothan/v1"

	"github.com/bandprotocol/chain/v3/x/feeds/types"
)

// isDeviated checks if the deviation between the old price and the new price
// exceeds a given threshold in basis points.

// Parameters:
// - deviationBasisPoint: the allowable deviation in basis points (1/1000th)
// - oldPrice: the original price
// - newPrice: the new price to compare against the original
//
// Returns:
//   - bool: true if the deviation is greater than or equal to the given threshold,
//     false otherwise
//
// The deviation is calculated as follows:
//  1. Calculate the absolute difference between the new price and the old price.
//  2. Compute the deviation in basis points by dividing the difference by the
//     original price and multiplying by 10000.
//  3. Check if the calculated deviation meets or exceeds the allowable deviation.
func isDeviated(deviationBasisPoint int64, oldPrice uint64, newPrice uint64) bool {
	// Calculate the deviation
	diff := math.Abs(float64(newPrice) - float64(oldPrice))
	dev := int64((diff * 10000) / float64(oldPrice))

	// Check if the new price deviation is meets or exceeds the bounds
	return deviationBasisPoint <= dev
}

func convertPriceData(price *bothan.Price) (types.SignalPrice, error) {
	switch price.Status {
	case bothan.Status_STATUS_UNSUPPORTED:
		return types.NewSignalPrice(
			types.SIGNAL_PRICE_STATUS_UNSUPPORTED,
			price.SignalId,
			0,
		), nil
	case bothan.Status_STATUS_UNAVAILABLE:
		return types.NewSignalPrice(
			types.SIGNAL_PRICE_STATUS_UNAVAILABLE,
			price.SignalId,
			0,
		), nil
	case bothan.Status_STATUS_AVAILABLE:
		return types.NewSignalPrice(
			types.SIGNAL_PRICE_STATUS_AVAILABLE,
			price.SignalId,
			price.Price,
		), nil
	default:
		// Handle unexpected price status
		return types.SignalPrice{}, fmt.Errorf("unexpected price status: %v", price.Status)
	}
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

	return time.Unix(timestamp+timeOffset, 0)
}

func sliceToMap[T any, K comparable](slice []T, keyFunc func(T) K) map[K]T {
	resultMap := make(map[K]T)
	for _, item := range slice {
		key := keyFunc(item)
		resultMap[key] = item
	}
	return resultMap
}
