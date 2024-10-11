package types

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestValidatorPriceListStoreKey(t *testing.T) {
	// Prefix: 0x10
	expect, _ := hex.DecodeString("100a31303030303030303031")
	require.Equal(t, expect, ValidatorPriceListStoreKey(sdk.ValAddress("1000000001")))

	// Test with empty validator address
	emptyVal := sdk.ValAddress{}
	expectEmpty, _ := hex.DecodeString("10")
	require.Equal(t, expectEmpty, ValidatorPriceListStoreKey(emptyVal))
}

func TestPriceStoreKey(t *testing.T) {
	// Prefix: 0x11
	expect, _ := hex.DecodeString("1142414e44")
	require.Equal(t, expect, PriceStoreKey("BAND"))

	// Test with empty string
	expectEmpty, _ := hex.DecodeString("11")
	require.Equal(t, expectEmpty, PriceStoreKey(""))
}

func TestDelegatorSignalsStoreKey(t *testing.T) {
	// Prefix: 0x12
	acc, _ := sdk.AccAddressFromHexUnsafe("b80f2a5df7d5710b15622d1a9f1e3830ded5bda8")
	expect, _ := hex.DecodeString("1214b80f2a5df7d5710b15622d1a9f1e3830ded5bda8")
	require.Equal(t, expect, DelegatorSignalsStoreKey(acc))

	// Test with empty address
	emptyAcc := sdk.AccAddress{}
	expectEmpty, _ := hex.DecodeString("12")
	require.Equal(t, expectEmpty, DelegatorSignalsStoreKey(emptyAcc))
}

func TestSignalTotalPowerStoreKey(t *testing.T) {
	// Prefix: 0x13
	expect, _ := hex.DecodeString("1342414e44")
	require.Equal(t, expect, SignalTotalPowerStoreKey("BAND"))

	// Test with empty string
	expectEmpty, _ := hex.DecodeString("13")
	require.Equal(t, expectEmpty, SignalTotalPowerStoreKey(""))
}

func TestSignalTotalPowerByPowerIndexKey(t *testing.T) {
	// Prefix: 0x80
	// Test with signal ID "BAND" and power 100
	expect, _ := hex.DecodeString(
		"80000000000000006404bdbeb1bb",
	) // Signal ID "BAND" with power 100, Signal ID bytes negated
	require.Equal(t, expect, SignalTotalPowerByPowerIndexKey("BAND", 100))

	// Test with empty signal ID and zero power
	expectEmpty, _ := hex.DecodeString("80000000000000000000")
	require.Equal(t, expectEmpty, SignalTotalPowerByPowerIndexKey("", 0))

	// Test with empty signal ID and large power
	largePower := int64(9223372036854775807) // Max int64
	expectLargePower, _ := hex.DecodeString("807fffffffffffffff00")
	require.Equal(t, expectLargePower, SignalTotalPowerByPowerIndexKey("", largePower))
}
