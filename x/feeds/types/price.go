package types

import (
	"fmt"
	"math"
)

const (
	TICK_SIZE float64 = 1.0001       // Equivalent to 10^(-4)
	MAX_PRICE float64 = 2.421902e11  // Equivalent to 1.0001 ** ((2**19) - 1 - (2**18))
	MIN_PRICE float64 = 4.128986e-12 // Equivalent to 1.0001 ** (1 - (2**18))
	OFFSET    float64 = 262144       // Equivalent to 2**18
	BILLION   uint64  = 1e9          // Equivalent to 10^9
)

// ToTick converts the price to tick
func (p *Price) ToTick() error {
	price, err := PriceToTick(ConvertToRealPrice(p.Price))
	if err != nil {
		return err
	}

	p.Price = price
	return nil
}

// ConvertToRealPrice converts the price multiplied by 1e9 to real price
func ConvertToRealPrice(price uint64) float64 {
	realPrice := float64(price) / float64(BILLION)
	return realPrice
}

// PriceToTick converts the price to tick
func PriceToTick(price float64) (uint64, error) {
	// Check if price is less than or equal to zero to prevent NaN results
	if price <= 0 {
		return 0, fmt.Errorf("price must be greater than 0")
	}

	// For safely convert from i64 to u64 since the price is already checked
	// to ensure `tick` is always a positive value
	if price < MIN_PRICE || price > MAX_PRICE {
		return 0, fmt.Errorf("price out of range")
	}

	return uint64(math.Round(math.Log(price)/math.Log(TICK_SIZE)) + OFFSET), nil
}
