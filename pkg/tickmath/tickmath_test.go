package tickmath_test

import (
	"fmt"
	"math"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/bandprotocol/chain/v3/pkg/tickmath"
)

// PriceToTickUsingLog converts the price to tick
func PriceToTickUsingLog(priceX1E9 uint64) (uint64, error) {
	price := float64(priceX1E9) / float64(1000000000)

	// Check if price is less than or equal to zero to prevent NaN results
	if price <= 0 {
		return 0, fmt.Errorf("price must be greater than 0")
	}

	tick := int64(math.Floor(math.Log(price) / math.Log(float64(1.0001))))
	if tick > tickmath.MaxTick || tick < tickmath.MinTick {
		return 0, fmt.Errorf("tick out of range")
	}

	return uint64(tick + tickmath.Offset), nil
}

func TestTickToPrice(t *testing.T) {
	testcases := []struct {
		name   string
		tick   int64
		result uint64
		err    error
	}{
		{
			name: "error case t=300000",
			tick: 300000,
			err:  fmt.Errorf("tick out of range"),
		},
		{
			name: "error case result is outside uint64",
			tick: 240000,
			err:  fmt.Errorf("price out of range"),
		},
		{
			name: "error case result is too small",
			tick: -240000,
			err:  fmt.Errorf("price out of range"),
		},
		{
			name: "normal case t=-207244",
			tick: -207244,
			err:  fmt.Errorf("price out of range"),
		},
		{
			name:   "normal case t=-207243",
			tick:   -207243,
			result: 1,
		},
		{
			name:   "normal case t=-1",
			tick:   -1,
			result: 999900010,
		},
		{
			name:   "normal case t=-25429",
			tick:   -25429,
			result: 78648017,
		},
		{
			name:   "normal case t=-76394",
			tick:   -76394,
			result: 481301,
		},
		{
			name:   "normal case t=1",
			tick:   1,
			result: 1000100000,
		},
		{
			name:   "normal case t=123456",
			tick:   123456,
			result: 229804163267350,
		},
		{
			name:   "normal case p=17128715048840092360",
			tick:   235651,
			result: 17126996933373237467,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := tickmath.TickToPrice(tc.tick)
			require.Equal(t, tc.err, err)
			require.Equal(t, tc.result, result)
		})
	}
}

func TestSmallPriceToTick(t *testing.T) {
	for i := 0; i <= 1000; i++ {
		price := uint64(i)
		tick, err := tickmath.PriceToTick(price)
		tickLog, errLog := PriceToTickUsingLog(price)
		if errLog != nil {
			require.Equal(t, err, errLog, fmt.Sprintf("price: %d", price))
		} else {
			require.Equal(t, tick, tickLog, fmt.Sprintf("price: %d", price))
		}
	}
}

func TestPriceToTick(t *testing.T) {
	testcases := []struct {
		name   string
		price  uint64
		result uint64
		err    error
	}{
		{
			name:   "normal case p=1.0",
			price:  1000000000,
			result: 262144,
		},
		{
			name:   "normal case p=1e-9",
			price:  1,
			result: 54900,
		},
		{
			name:   "normal case p=1/1.0001 + minor",
			price:  999900010,
			result: 262143,
		},
		{
			name:   "normal case p=1/1.0001 - minor",
			price:  999900009,
			result: 262142,
		},
		{
			name:   "normal case p=1.0001 + minor",
			price:  1000100001,
			result: 262145,
		},
		{
			name:   "normal case p=1.0001",
			price:  1000100000,
			result: 262145,
		},
		{
			name:   "normal case p=1.0001 - minor",
			price:  1000099999,
			result: 262144,
		},
		{
			name:   "normal case p=2^64-1",
			price:  uint64(18446744073709551615),
			result: 498537,
		},
		{
			name:   "normal case p=1.0001^2042 + minor",
			price:  uint64(1226570000),
			result: 264186,
		},
		{
			name:   "normal case p=1.0001^123456",
			price:  uint64(229804163267350),
			result: 385600,
		},
		{
			name:   "normal case p=481300",
			price:  481300,
			result: 185749,
		},
		{
			name:   "normal case p=481301",
			price:  481301,
			result: 185750,
		},
		{
			name:   "normal case p=17128715048840092360",
			price:  17128715048840092360,
			result: 497796,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := tickmath.PriceToTick(tc.price)
			require.Equal(t, tc.err, err)
			require.Equal(t, tc.result, result)

			resultLog, errLog := PriceToTickUsingLog(tc.price)
			require.Equal(t, tc.result, resultLog)
			require.Equal(t, tc.err, errLog)
		})
	}
}

func TestPriceToTickRandomly(t *testing.T) {
	// Use a unique random seed each test instance and log it if the tests fail.
	seed := time.Now().Unix()
	rng := rand.New(rand.NewSource(seed))
	defer func(t *testing.T, seed int64) {
		if t.Failed() {
			t.Logf("random seed: %d", seed)
		}
	}(t, seed)

	// random normal prices
	for i := 0; i < 1000; i++ {
		price := rng.Uint64()

		tick, err := tickmath.PriceToTick(price)
		tickLog, errLog := PriceToTickUsingLog(price)
		if errLog != nil {
			require.Equal(t, errLog, err, fmt.Sprintf("price: %d", price))
		} else {
			require.Equal(t, tickLog, tick, fmt.Sprintf("price: %d", price))
		}
	}

	// random small prices
	for i := 0; i < 1000; i++ {
		price := rng.Uint64() % 1000000

		tick, err := tickmath.PriceToTick(price)
		tickLog, errLog := PriceToTickUsingLog(price)
		if errLog != nil {
			require.Equal(t, errLog, err, fmt.Sprintf("price: %d", price))
		} else {
			require.Equal(t, tickLog, tick, fmt.Sprintf("price: %d", price))
		}
	}
}
