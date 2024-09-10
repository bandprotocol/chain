package tickmath

import (
	"errors"
	"fmt"
	"math/big"
)

const (
	MaxTick int64 = 262143   // Equivalent to 2**18 - 1
	MinTick int64 = -MaxTick // Equivalent to -2**18 + 1
	Offset  int64 = 262144   // Equivalent to 2**18
)

var (
	priceX96AtBinaryTicks = getPricesX96AtBinaryTicks()

	maxUint192, _ = new(big.Int).SetString("ffffffffffffffffffffffffffffffffffffffffffffffff", 16)
	maxUint96, _  = new(big.Int).SetString("ffffffffffffffffffffffff", 16)
	maxUint64, _  = new(big.Int).SetString("ffffffffffffffff", 16)
	q96, _        = new(big.Int).SetString("1000000000000000000000000", 16)
	zero          = new(big.Int).SetUint64(0)
	one           = new(big.Int).SetUint64(1)
	billion       = new(big.Int).SetUint64(1000000000)
)

// TickToPrice converts the tick to price with 10^9 precision. It will return an error
// if the tick is out of range or the tick is so large that cannot be converted to uint64.
// NOTE: the result is rounded up to the nearest integer, this is aligned with the UniswapV3 calculation.
func TickToPrice(tick int64) (uint64, error) {
	priceX96, err := tickToPriceX96(tick)
	if err != nil {
		return 0, err
	}

	// round up the price and convert to uint64
	// we round up in the division so PriceX1E9ToTick of the output price is always consistent
	// var price *big.Int
	price := new(big.Int).Div(priceX96, q96)
	if price.Cmp(zero) <= 0 {
		return 0, fmt.Errorf("price out of range")
	}

	if new(big.Int).Rem(priceX96, q96).Cmp(zero) > 0 {
		priceNextTickX96 := new(big.Int).Div(new(big.Int).Mul(priceX96, big.NewInt(10001)), big.NewInt(10000))
		priceNextTick := new(big.Int).Div(priceNextTickX96, q96)

		if priceNextTick.Cmp(price) > 0 {
			price = new(big.Int).Add(price, one)
		}
	}

	if price.Cmp(maxUint64) > 0 {
		return 0, fmt.Errorf("price out of range")
	}
	return price.Uint64(), nil
}

// PriceToTick converts the price to tick, it will return the nearest tick that yields
// less than or equal to the given price.
// ref: https://en.wikipedia.org/wiki/Binary_logarithm#Iterative_approximation
func PriceToTick(price uint64) (uint64, error) {
	if price == 0 {
		return 0, errors.New("price must be greater than 0")
	}

	// find the most significant bit (msb) of the log2(price);
	msb := uint64(0)
	p := price
	bits := []uint64{4294967295, 65535, 255, 15, 3, 1}
	for i, bit := range bits {
		if p > bit {
			n := uint64(1 << (len(bits) - i - 1))
			msb += n
			p >>= n
		}
	}

	// find the remaining r = price / 2^msb and shift significant bits to 2^31;
	r := uint64(0)
	if msb >= 32 {
		r = price >> (msb - 31)
	} else {
		r = price << (31 - msb)
	}

	// approximate log2(r) using iterations of base-2 logarithm with 16-bit precision;
	log2 := int64(msb) << 16
	for i := 0; i < 16; i++ {
		r = (r * r) >> 31
		f := r >> 32
		log2 |= int64(f) << (15 - i)
		r >>= f
	}

	// convert to tick value;
	// tick = (log2 - log2(10^9) *2^16) *  (1/log2(1.0001))/(2^16/2^32)
	log1p0001 := (log2 - 1959352) * 454283648
	tick := log1p0001 >> 32
	if tick > MaxTick || tick < MinTick {
		return 0, fmt.Errorf("tick out of range")
	}

	// the result will differ by 1 tick if the price is not exactly at the tick value;
	// it will return the largest tick whose price are less than or equal to the given price.
	// NOTE: cannot use the previous result divided by 1.0001 as the fraction has been reduced.
	expectPriceX96 := new(big.Int).Mul(new(big.Int).SetUint64(price), q96)
	for i := int64(1); i >= 0; i-- {
		t := tick + i
		pX96, err := tickToPriceX96(t)
		if err == nil && pX96.Cmp(expectPriceX96) <= 0 {
			return uint64(t + Offset), nil
		}
	}

	return uint64(tick - 1 + Offset), nil
}

// mulShift multiplies two big.Int and shifts the result to the right by 96 bits.
// It returns a new big.Int object.
func mulShift(val *big.Int, mulBy *big.Int) *big.Int {
	return new(big.Int).Rsh(new(big.Int).Mul(val, mulBy), 96)
}

// tickToPriceX96 converts the tick to price in x96 (2^96) * 10^9 format.
func tickToPriceX96(tick int64) (*big.Int, error) {
	if tick > MaxTick || tick < MinTick {
		return nil, fmt.Errorf("tick out of range")
	}

	absTick := tick
	if tick < 0 {
		absTick = -tick
	}

	// multiply the price ratio at each binary tick
	priceX96 := new(big.Int).Set(q96)
	for i, pX96 := range priceX96AtBinaryTicks {
		if absTick&(1<<uint(i)) != 0 {
			priceX96 = mulShift(priceX96, pX96)
		}
	}

	// inverse the price if tick is positive.
	if tick > 0 {
		priceX96 = new(big.Int).Div(maxUint192, priceX96)
	}

	priceX96 = new(big.Int).Mul(priceX96, billion)

	return priceX96, nil
}

// getPricesX96AtBinaryTicks returns the prices at each binary tick in x96 format.
// the prices are in the term of 1.0001^-(2^i) * 2^96.
func getPricesX96AtBinaryTicks() []*big.Int {
	x96Hexes := []string{
		"fff97272373d413259a46990",
		"fff2e50f5f656932ef12357c",
		"ffe5caca7e10e4e61c3624ea",
		"ffcb9843d60f6159c9db5883",
		"ff973b41fa98c081472e6896",
		"ff2ea16466c96a3843ec78b3",
		"fe5dee046a99a2a811c461f1",
		"fcbe86c7900a88aedcffc83b",
		"f987a7253ac413176f2b074c",
		"f3392b0822b70005940c7a39",
		"e7159475a2c29b7443b29c7f",
		"d097f3bdfd2022b8845ad8f7",
		"a9f746462d870fdf8a65dc1f",
		"70d869a156d2a1b890bb3df6",
		"31be135f97d08fd981231505",
		"9aa508b5b7a84e1c677de54",
		"5d6af8dedb81196699c329",
		"2216e584f5fa1ea92604",
	}

	prices := make([]*big.Int, 0, len(x96Hexes))
	for _, x96Hex := range x96Hexes {
		p, ok := new(big.Int).SetString(x96Hex, 16)
		if !ok {
			panic("failed to parse hex string")
		}
		prices = append(prices, p)
	}

	return prices
}
