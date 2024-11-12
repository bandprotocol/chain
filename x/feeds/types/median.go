package types

import (
	"cmp"
	"slices"
)

const PowerScalingFactor = 32 // Scaling factor to avoid floating-point calculations

// getMultipliers returns predefined multiplier constants.
func getMultipliers() [5]uint64 {
	return [5]uint64{60, 40, 20, 11, 10}
}

// getSections returns predefined section constants.
func getSections() [5]uint64 {
	return [5]uint64{1, 3, 7, 15, 32}
}

// ValidatorPriceInfo represents a single entry of price information from a validator.
// It includes the reported price, associated power (weight), and timestamp.
type ValidatorPriceInfo struct {
	SignalPriceStatus SignalPriceStatus // indicates the validity or state of the price entry
	Power             uint64            // power or weight of this entry in calculations
	Price             uint64            // reported price value
	Timestamp         int64             // Unix timestamp for when this entry was recorded
}

// NewValidatorPriceInfo creates a new instance of ValidatorPriceInfo.
func NewValidatorPriceInfo(
	signalPriceStatus SignalPriceStatus,
	power uint64,
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
) (totalPower uint64, availablePower uint64, unavailablePower uint64, unsupportedPower uint64) {
	for _, priceInfo := range validatorPriceInfos {
		totalPower += priceInfo.Power

		switch priceInfo.SignalPriceStatus {
		case SignalPriceStatusAvailable:
			availablePower += priceInfo.Power
		case SignalPriceStatusUnavailable:
			unavailablePower += priceInfo.Power
		case SignalPriceStatusUnsupported:
			unsupportedPower += priceInfo.Power
		}
	}
	return
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
	totalPower := uint64(0)
	for _, priceInfo := range validatorPriceInfos {
		if priceInfo.SignalPriceStatus == SignalPriceStatusAvailable {
			validPrices = append(validPrices, priceInfo)
			totalPower += priceInfo.Power
		}
	}

	// Step 2: Sort valid entries by timestamp (descending) and by power (descending).
	slices.SortStableFunc(validPrices, func(priceA, priceB ValidatorPriceInfo) int {
		// primary comparison: timestamp (descending)
		if cmpResult := cmp.Compare(priceB.Timestamp, priceA.Timestamp); cmpResult != 0 {
			return cmpResult
		}
		// secondary comparison: power (descending)
		return cmp.Compare(priceB.Power, priceA.Power)
	})

	// Step 3: Define multipliers and sections for sectional weighting.
	multipliers := getMultipliers()
	sections := getSections()

	var weightedPrices []WeightedPrice
	currentPower := uint64(0)
	sectionIndex := 0

	// Step 4: Distribute each entry’s power across sections.
	for _, priceInfo := range validPrices {
		leftPower := priceInfo.Power * PowerScalingFactor // scale up power to avoid floating-point calculations
		totalWeight := uint64(0)                          // accumulated weight for this entry

		// distribute the entry's power across remaning sections, starting from the current `sectionIndex`
		for ; sectionIndex < len(sections); sectionIndex++ {
			// calculate the power limit for the current section
			sectionLimit := totalPower * sections[sectionIndex]

			// determine how much power to take from this section, based on available capacity
			takePower := uint64(0)
			if currentPower+leftPower <= sectionLimit {
				takePower = leftPower
			} else {
				takePower = sectionLimit - currentPower
			}

			// accumulate the weighted power for this entry based on the section's multiplier
			totalWeight += takePower * multipliers[sectionIndex]

			// update current power and remaining power for this entry
			currentPower += takePower
			leftPower -= takePower

			// if all power has been distributed for this entry, exit the loop
			if leftPower == 0 {
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
	Weight uint64 // weight of the price
	Price  uint64 // actual price value
}

// NewWeightedPrice creates and returns a new WeightedPrice instance.
func NewWeightedPrice(weight uint64, price uint64) WeightedPrice {
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
		return cmp.Compare(a.Weight, b.Weight)
	})

	// calculate total weight
	totalWeight := uint64(0)
	for _, wp := range weightedPrices {
		totalWeight += wp.Weight
	}

	// find median by accumulating weights until reaching the midpoint
	cumulativeWeight := uint64(0)
	for _, wp := range weightedPrices {
		cumulativeWeight += wp.Weight
		if cumulativeWeight*2 >= totalWeight {
			return wp.Price, nil
		}
	}

	// return an error if median cannot be determined
	return 0, ErrInvalidWeightedPrices
}
