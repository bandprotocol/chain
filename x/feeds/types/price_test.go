package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bandprotocol/chain/v2/x/feeds/types"
)

func TestConvertToRealPrice(t *testing.T) {
	// Test normal case
	price := uint64(1000000000) // 1 when converted to real price
	realPrice := types.ConvertToRealPrice(price)
	require.Equal(t, 1.0, realPrice)

	// Test price equals zero
	price = uint64(0)
	realPrice = types.ConvertToRealPrice(price)
	require.Equal(t, 0.0, realPrice)
}

func TestPriceToTick(t *testing.T) {
	// Test normal case 1
	price := 1.0
	tick, err := types.PriceToTick(price)
	require.NoError(t, err)
	require.Equal(t, uint64(262144), tick)

	// Test normal case 2 (at MAX_PRICE)
	price = types.MAX_PRICE
	tick, err = types.PriceToTick(price)
	require.NoError(t, err)
	require.Equal(t, uint64(524287), tick)

	// Test normal case 3 (at MIN_PRICE)
	price = types.MIN_PRICE
	tick, err = types.PriceToTick(price)
	require.NoError(t, err)
	require.Equal(t, uint64(1), tick)

	// Test price equals zero
	price = 0.0
	_, err = types.PriceToTick(price)
	require.Error(t, err)

	// Test negative price
	price = -1.0
	_, err = types.PriceToTick(price)
	require.Error(t, err)
}
