package types

import (
	"sort"
)

// Constants representing multipliers
func getMultipliers() [5]uint64 {
	return [5]uint64{60, 40, 20, 11, 10}
}

// Constants representing sections
func getSections() [5]uint64 {
	return [5]uint64{1, 3, 7, 15, 32}
}

// PriceFeedInfo contains information about a price feed
type PriceFeedInfo struct {
	PriceStatus PriceStatus // PriceStatus represents the state of the price feed
	Power       uint64      // Power represents the power of the price feed
	Price       uint64      // Price represents the reported price
	Timestamp   int64       // Timestamp represents the time at which the price feed was reported
	Index       int64       // Index represents the index of the price feed
}

// NewPriceFeedInfo returns a new PriceFeedInfo
func NewPriceFeedInfo(
	priceStatus PriceStatus,
	power uint64,
	price uint64,
	timestamp int64,
	index int64,
) PriceFeedInfo {
	return PriceFeedInfo{
		PriceStatus: priceStatus,
		Power:       power,
		Price:       price,
		Timestamp:   timestamp,
		Index:       index,
	}
}

// FilterPriceFeedInfos filters price feed infos based on price status
func FilterPriceFeedInfos(pfInfos []PriceFeedInfo, opt PriceStatus) []PriceFeedInfo {
	filtered := []PriceFeedInfo{}
	for _, pfInfo := range pfInfos {
		if pfInfo.PriceStatus == opt {
			filtered = append(filtered, pfInfo)
		}
	}
	return filtered
}

// CalculatePricesPowers calculates total, available, unavailable, and unsupported powers
func CalculatePricesPowers(
	priceFeedInfos []PriceFeedInfo,
) (totalPower uint64, availablePower uint64, unavailablePower uint64, unsupportedPower uint64) {
	for _, pfInfo := range priceFeedInfos {
		totalPower += pfInfo.Power

		switch pfInfo.PriceStatus {
		case PriceStatusAvailable:
			availablePower += pfInfo.Power
		case PriceStatusUnavailable:
			unavailablePower += pfInfo.Power
		case PriceStatusUnsupported:
			unsupportedPower += pfInfo.Power
		}
	}
	return totalPower, availablePower, unavailablePower, unsupportedPower
}

// CalculateMedianPriceFeedInfo calculates the median price feed info by timestamp and power
func CalculateMedianPriceFeedInfo(priceFeedInfos []PriceFeedInfo) (uint64, error) {
	totalPower, _, _, _ := CalculatePricesPowers(priceFeedInfos)

	sort.Slice(priceFeedInfos, func(i, j int) bool {
		if priceFeedInfos[i].Timestamp == priceFeedInfos[j].Timestamp {
			if priceFeedInfos[i].Power == priceFeedInfos[j].Power {
				return priceFeedInfos[i].Index < priceFeedInfos[j].Index
			}
			return priceFeedInfos[i].Power > priceFeedInfos[j].Power
		}
		return priceFeedInfos[i].Timestamp > priceFeedInfos[j].Timestamp
	})

	multipliers := getMultipliers()
	sections := getSections()

	var wps []WeightedPrice
	currentSection := 0
	currentPower := uint64(0)
	for _, priceFeedInfo := range priceFeedInfos {
		leftPower := priceFeedInfo.Power * 32
		totalWeight := uint64(0)
		for ; currentSection < len(sections); currentSection++ {
			takePower := uint64(0)
			if currentPower+leftPower <= totalPower*sections[currentSection] {
				takePower = leftPower
			} else {
				takePower = totalPower*sections[currentSection] - currentPower
			}
			totalWeight += takePower * multipliers[currentSection]
			currentPower += takePower
			leftPower -= takePower
			if leftPower == 0 {
				break
			}
		}
		wps = append(
			wps,
			WeightedPrice{
				Price: priceFeedInfo.Price,
				Power: totalWeight,
			},
		)
	}

	return CalculateMedianWeightedPrice(wps)
}

// WeightedPrice represents a weighted price
type WeightedPrice struct {
	Power uint64 // Power represents the power for the price
	Price uint64 // Price represents the price
}

// CalculateMedianWeightedPrice calculates the median of weighted prices
func CalculateMedianWeightedPrice(wps []WeightedPrice) (uint64, error) {
	sort.Slice(wps, func(i, j int) bool {
		if wps[i].Price == wps[j].Price {
			return wps[i].Power < wps[j].Power
		}
		return wps[i].Price < wps[j].Price
	})

	totalPower := uint64(0)
	for _, wp := range wps {
		totalPower += wp.Power
	}

	currentPower := uint64(0)
	for _, wp := range wps {
		currentPower += wp.Power
		if currentPower*2 >= totalPower {
			return wp.Price, nil
		}
	}

	return 0, ErrInvalidWeightedPriceArray
}
