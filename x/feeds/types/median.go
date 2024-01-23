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

func CalculateMedianPriceFeedInfo(pfInfos []PriceFeedInfo) (uint64, error) {
	n := len(pfInfos)
	if n == 0 {
		return 0, ErrNotEnoughPriceValidator
	}

	totalPower := uint64(0)
	for _, pfInfo := range pfInfos {
		totalPower += pfInfo.Power
	}

	sort.Slice(pfInfos, func(i, j int) bool {
		if pfInfos[i].Timestamp == pfInfos[j].Timestamp {
			return pfInfos[i].Power > pfInfos[j].Power
		}

		return pfInfos[i].Timestamp > pfInfos[j].Timestamp
	})

	multipliers := []uint64{60, 40, 20, 11, 10}
	sections := []uint64{1, 3, 7, 15, 32}

	var wps []WeightedPrice
	idxSection := 0
	currentPower := uint64(0)
	for _, pfInfo := range pfInfos {
		leftPower := pfInfo.Power * 32
		for ; idxSection < len(sections); idxSection++ {
			takePower := uint64(0)
			if currentPower+leftPower <= totalPower*sections[idxSection] {
				takePower = leftPower
			} else {
				takePower = totalPower*sections[idxSection] - currentPower
			}

			wps = append(
				wps,
				GetDeviationWeightedPrices(
					pfInfo.Price,
					pfInfo.Deviation,
					takePower*multipliers[idxSection],
				)...,
			)

			currentPower += takePower
			leftPower -= takePower

			if leftPower == 0 {
				break
			}
		}
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
	n := len(wps)
	if n == 0 {
		return 0, ErrNotEnoughPriceValidator
	}

	sort.Slice(wps, func(i, j int) bool {
		return wps[i].Price < wps[j].Price
	})

	totalPower := uint64(0)
	for _, wp := range wps {
		totalPower += wp.Power
	}

	price := wps[0].Price
	currentPower := wps[0].Power
	for i := 1; i < n; i++ {
		if currentPower >= totalPower/2 {
			break
		}
		currentPower += wps[i].Power
		price = wps[i].Price
	}

	return price, nil
}
