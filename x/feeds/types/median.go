package types

import (
	"sort"
)

func getMultipliers() [5]uint64 {
	return [5]uint64{60, 40, 20, 11, 10}
}

func getSections() [5]uint64 {
	return [5]uint64{1, 3, 7, 15, 32}
}

type PriceFeedInfo struct {
	PriceOption PriceOption
	Power       uint64
	Price       uint64
	Deviation   uint64
	Timestamp   int64
	Index       int64
}

func FilterPriceFeedInfos(pfInfos []PriceFeedInfo, opt PriceOption) []PriceFeedInfo {
	filtered := []PriceFeedInfo{}
	for _, pfInfo := range pfInfos {
		if pfInfo.PriceOption == opt {
			filtered = append(filtered, pfInfo)
		}
	}
	return filtered
}

func CalculatePricesPowers(
	priceFeedInfos []PriceFeedInfo,
) (totalPower uint64, availablePower uint64, unavailablePower uint64, unsupportedPower uint64) {
	for _, pfInfo := range priceFeedInfos {
		totalPower += pfInfo.Power

		switch pfInfo.PriceOption {
		case PriceOptionAvailable:
			availablePower += pfInfo.Power
		case PriceOptionUnavailable:
			unavailablePower += pfInfo.Power
		case PriceOptionUnsupported:
			unsupportedPower += pfInfo.Power
		}
	}
	return totalPower, availablePower, unavailablePower, unsupportedPower
}

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
			GetDeviationWeightedPrices(
				priceFeedInfo.Price,
				priceFeedInfo.Deviation,
				totalWeight,
			)...,
		)
	}

	return CalculateMedianWeightedPrice(wps)
}

func GetDeviationWeightedPrices(price uint64, deviation uint64, power uint64) []WeightedPrice {
	return []WeightedPrice{{
		Price: price,
		Power: power,
	}, {
		Price: price - deviation,
		Power: power,
	}, {
		Price: price + deviation,
		Power: power,
	}}
}

type WeightedPrice struct {
	Power uint64
	Price uint64
}

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
