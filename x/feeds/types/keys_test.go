package types

import (
	"encoding/hex"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestDelegatorSignalStoreKey(t *testing.T) {
	acc, _ := sdk.AccAddressFromHexUnsafe("b80f2a5df7d5710b15622d1a9f1e3830ded5bda8")
	expect, _ := hex.DecodeString("04b80f2a5df7d5710b15622d1a9f1e3830ded5bda8")
	require.Equal(t, expect, DelegatorSignalStoreKey(acc))
}

func TestFeedStoreKey(t *testing.T) {
	expect, _ := hex.DecodeString("0142414e44")
	require.Equal(t, expect, FeedStoreKey("BAND"))
}

func TestValidatorPricesStoreKey(t *testing.T) {
	expect, _ := hex.DecodeString("0242414e44")
	require.Equal(t, expect, ValidatorPricesStoreKey("BAND"))
}

func TestValidatorPriceStoreKey(t *testing.T) {
	acc, _ := sdk.ValAddressFromHex("b80f2a5df7d5710b15622d1a9f1e3830ded5bda8")
	expect, _ := hex.DecodeString("0242414e44b80f2a5df7d5710b15622d1a9f1e3830ded5bda8")
	require.Equal(t, expect, ValidatorPriceStoreKey("BAND", acc))
}

func TestPriceStoreKey(t *testing.T) {
	expect, _ := hex.DecodeString("0342414e44")
	require.Equal(t, expect, PriceStoreKey("BAND"))
}

func TestFeedsByPowerIndexKey(t *testing.T) {
	expect, _ := hex.DecodeString("20000000000098968004bdbeb1bb")
	require.Equal(t, expect, FeedsByPowerIndexKey("BAND", 10e6))
}
