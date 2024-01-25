package types

import (
	"sort"
)

type PriceFeedInfo struct {
	Power     uint64
	Price     uint64
	Deviation uint64
	Timestamp int64
}

func CalculateMedianPriceFeedInfo(pfInfos []PriceFeedInfo) uint64 {
	totalPower := uint64(0)
	for _, pfInfo := range pfInfos {
		totalPower += pfInfo.Power
	}

	// TODO: recheck
	sort.Slice(pfInfos, func(i, j int) bool {
		if pfInfos[i].Timestamp == pfInfos[j].Timestamp {
			if pfInfos[i].Power == pfInfos[j].Power {
				return i < j
			}

			return pfInfos[i].Power > pfInfos[j].Power
		}

		return pfInfos[i].Timestamp > pfInfos[j].Timestamp
	})

	multipliers := []uint64{60, 40, 20, 11, 10}
	sections := []uint64{1, 3, 7, 15, 32}

	var wps []WeightedPrice
	currentSection := 0
	currentPower := uint64(0)
	for _, pfInfo := range pfInfos {
		leftPower := pfInfo.Power * 32
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
				pfInfo.Price,
				pfInfo.Deviation,
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

func CalculateMedianWeightedPrice(wps []WeightedPrice) uint64 {
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
			return wp.Price
		}
	}

	// TODO: check if should panic or not
	return 0
}
