package types

import (
	"cmp"
	"slices"

	sdkmath "cosmossdk.io/math"
)

// getPowerScalingFactor returns the scaling factor used to avoid floating-point calculations.
func getPowerScalingFactor() sdkmath.Int {
	return sdkmath.NewInt(32)
}

// getMultipliers returns predefined multiplier constants.
func getMultipliers() [5]sdkmath.Int {
	return [5]sdkmath.Int{
		sdkmath.NewInt(60),
		sdkmath.NewInt(40),
		sdkmath.NewInt(20),
		sdkmath.NewInt(11),
		sdkmath.NewInt(10),
	}
}

// getSections returns predefined section constants.
func getSections() [5]sdkmath.Int {
	return [5]sdkmath.Int{
		sdkmath.NewInt(1),
		sdkmath.NewInt(3),
		sdkmath.NewInt(7),
		sdkmath.NewInt(15),
		sdkmath.NewInt(32),
	}
}

// ValidatorPriceInfo represents a single entry of price information from a validator.
// It includes the reported price, associated power (weight), and timestamp.
type ValidatorPriceInfo struct {
	SignalPriceStatus SignalPriceStatus // indicates the validity or state of the price entry
	Power             sdkmath.Int       // power or weight of this entry in calculations
	Price             uint64            // reported price value
	Timestamp         int64             // Unix timestamp for when this entry was recorded
}

// NewValidatorPriceInfo creates a new instance of ValidatorPriceInfo.
func NewValidatorPriceInfo(
	signalPriceStatus SignalPriceStatus,
	power sdkmath.Int,
	price uint64,
	timestamp int64,
) ValidatorPriceInfo {
	return ValidatorPriceInfo{
		SignalPriceStatus: signalPriceStatus,
		Power:             power,
		Price:             price,
		Timestamp:         timestamp,
	}
}

// CalculatePricesPowers calculates total, available, unavailable, and unsupported powers
func CalculatePricesPowers(
	validatorPriceInfos []ValidatorPriceInfo,
) (sdkmath.Int, sdkmath.Int, sdkmath.Int, sdkmath.Int) {
	totalPower := sdkmath.NewInt(0)
	availablePower := sdkmath.NewInt(0)
	unavailablePower := sdkmath.NewInt(0)
	unsupportedPower := sdkmath.NewInt(0)

	for _, priceInfo := range validatorPriceInfos {
		totalPower = totalPower.Add(priceInfo.Power)

		switch priceInfo.SignalPriceStatus {
		case SIGNAL_PRICE_STATUS_AVAILABLE:
			availablePower = availablePower.Add(priceInfo.Power)
		case SIGNAL_PRICE_STATUS_UNAVAILABLE:
			unavailablePower = unavailablePower.Add(priceInfo.Power)
		case SIGNAL_PRICE_STATUS_UNSUPPORTED:
			unsupportedPower = unsupportedPower.Add(priceInfo.Power)
		}
	}
	return totalPower, availablePower, unavailablePower, unsupportedPower
}

// MedianValidatorPriceInfos calculates a time-weighted and power-weighted median price
// from ValidatorPriceInfo entries, prioritizing recent timestamps and higher power values.
//
// Algorithm Overview:
//
//  1. **Filter and Sum Power**: Filter entries with available prices and sum their power
//     to set a baseline for section capacities.
//
//  2. **Sort Entries**: Sort entries by timestamp (newest first) and, within equal timestamps, by power
//     (highest first). This ensures recent, high-power entries are prioritized.
//
//  3. **Set Multipliers and Sections**: Define multipliers and sections for weighting. Each section
//     has a capacity based on a fraction of the total power, and earlier sections have higher multipliers
//     to favor recent entries.
//
// 4. **Distribute Power Across Sections**: For each entry, distribute power across sections:
//   - Start from the section where the previous entry left off, progressing until all power is allocated.
//   - For each section, calculate the maximum power that can be taken without exceeding its capacity.
//     Apply the section’s multiplier to this power to accumulate a weighted total for the entry.
//
// 5. **Store Weighted Prices**: Calculate the weighted price for each entry and store it in a list.
//
//  6. **Compute Median**: Calculate the median of weighted prices, yielding a time- and power-weighted
//     median price that reflects the most recent and influential entries.
func MedianValidatorPriceInfos(validatorPriceInfos []ValidatorPriceInfo) (uint64, error) {
	// Step 1: Filter entries with available prices and calculate total power for valid entries.
	var validPrices []ValidatorPriceInfo
	totalPower := sdkmath.NewInt(0)
	for _, priceInfo := range validatorPriceInfos {
		if priceInfo.SignalPriceStatus == SIGNAL_PRICE_STATUS_AVAILABLE {
			validPrices = append(validPrices, priceInfo)
			totalPower = totalPower.Add(priceInfo.Power)
		}
	}

	// Step 2: Sort valid entries by timestamp (descending) and by power (descending).
	slices.SortStableFunc(validPrices, func(priceA, priceB ValidatorPriceInfo) int {
		// primary comparison: timestamp (descending)
		if cmpResult := cmp.Compare(priceB.Timestamp, priceA.Timestamp); cmpResult != 0 {
			return cmpResult
		}
		// secondary comparison: power (descending)
		return priceB.Power.BigInt().Cmp(priceA.Power.BigInt())
	})

	// Step 3: Define multipliers and sections for sectional weighting.
	multipliers := getMultipliers()
	sections := getSections()

	var weightedPrices []WeightedPrice
	currentPower := sdkmath.NewInt(0)
	sectionIndex := 0

	// Step 4: Distribute each entry’s power across sections.
	for _, priceInfo := range validPrices {
		// scale up power to avoid floating-point calculations
		leftPower := getPowerScalingFactor().Mul(priceInfo.Power)
		// accumulated weight for this entry
		totalWeight := sdkmath.NewInt(0)

		// distribute the entry's power across remaining sections, starting from the current `sectionIndex`
		for ; sectionIndex < len(sections); sectionIndex++ {
			// calculate the power limit for the current section
			sectionLimit := totalPower.Mul(sections[sectionIndex])

			// determine how much power to take from this section, based on available capacity
			var takePower sdkmath.Int
			if currentPower.Add(leftPower).LTE(sectionLimit) {
				takePower = leftPower
			} else {
				takePower = sectionLimit.Sub(currentPower)
			}

			// accumulate the weighted power for this entry based on the section's multiplier
			totalWeight = totalWeight.Add(takePower.Mul(multipliers[sectionIndex]))

			// update current power and remaining power for this entry
			currentPower = currentPower.Add(takePower)
			leftPower = leftPower.Sub(takePower)

			// if all power has been distributed for this entry, exit the loop
			if leftPower.IsZero() {
				break
			}
		}

		// Step 5: Store the calculated weighted price for this entry
		weightedPrices = append(weightedPrices, NewWeightedPrice(totalWeight, priceInfo.Price))
	}

	// Step 6: Calculate and return the median of weighted prices.
	return MedianWeightedPrice(weightedPrices)
}

// WeightedPrice represents a price with an associated weight.
type WeightedPrice struct {
	Weight sdkmath.Int // weight of the price
	Price  uint64      // actual price value
}

// NewWeightedPrice creates and returns a new WeightedPrice instance.
func NewWeightedPrice(weight sdkmath.Int, price uint64) WeightedPrice {
	return WeightedPrice{
		Weight: weight,
		Price:  price,
	}
}

// MedianWeightedPrice finds the median price from a list of weighted prices.
func MedianWeightedPrice(weightedPrices []WeightedPrice) (uint64, error) {
	// sort by Price (ascending), breaking ties by Weight (ascending)
	slices.SortStableFunc(weightedPrices, func(a, b WeightedPrice) int {
		if cmpResult := cmp.Compare(a.Price, b.Price); cmpResult != 0 {
			return cmpResult
		}
		return a.Weight.BigInt().Cmp(b.Weight.BigInt())
	})

	// calculate total weight
	totalWeight := sdkmath.NewInt(0)
	for _, wp := range weightedPrices {
		totalWeight = totalWeight.Add(wp.Weight)
	}

	// find median by accumulating weights until reaching the midpoint
	cumulativeWeight := sdkmath.NewInt(0)
	for _, wp := range weightedPrices {
		cumulativeWeight = cumulativeWeight.Add(wp.Weight)
		if cumulativeWeight.MulRaw(2).GTE(totalWeight) {
			return wp.Price, nil
		}
	}

	// return an error if median cannot be determined
	return 0, ErrInvalidWeightedPrices
}
