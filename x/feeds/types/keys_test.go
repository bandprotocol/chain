package types

import (
	"encoding/hex"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestDelegatorSignalStoreKey(t *testing.T) {
	acc, _ := sdk.AccAddressFromHexUnsafe("b80f2a5df7d5710b15622d1a9f1e3830ded5bda8")
	expect, _ := hex.DecodeString("03b80f2a5df7d5710b15622d1a9f1e3830ded5bda8")
	require.Equal(t, expect, DelegatorSignalStoreKey(acc))
}

func TestSignalTotalPowerStoreKey(t *testing.T) {
	expect, _ := hex.DecodeString("0442414e44")
	require.Equal(t, expect, SignalTotalPowerStoreKey("BAND"))
}

func TestValidatorPriceListStoreKey(t *testing.T) {
	expect, _ := hex.DecodeString("0131303030303030303031")
	require.Equal(t, expect, ValidatorPriceListStoreKey(sdk.ValAddress("1000000001")))
}

func TestPriceStoreKey(t *testing.T) {
	expect, _ := hex.DecodeString("0242414e44")
	require.Equal(t, expect, PriceStoreKey("BAND"))
}
