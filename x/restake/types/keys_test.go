package types

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestVaultStoreKey(t *testing.T) {
	key := "key"
	expect, err := hex.DecodeString("10" + hex.EncodeToString([]byte(key)))
	require.NoError(t, err)
	require.Equal(t, expect, VaultStoreKey(key))
}

func TestStakeStoreKey(t *testing.T) {
	hexAddress := "b80f2a5df7d5710b15622d1a9f1e3830ded5bda8"
	acc, err := sdk.AccAddressFromHexUnsafe(hexAddress)
	require.NoError(t, err)

	expect, err := hex.DecodeString("12" + "14" + hexAddress)
	require.NoError(t, err)
	require.Equal(t, expect, StakeStoreKey(acc))
}

func TestLocksByAddressStoreKey(t *testing.T) {
	hexAddress := "b80f2a5df7d5710b15622d1a9f1e3830ded5bda8"
	acc, err := sdk.AccAddressFromHexUnsafe(hexAddress)
	require.NoError(t, err)

	expect, err := hex.DecodeString("11" + "14" + hexAddress)
	require.NoError(t, err)
	require.Equal(t, expect, LocksByAddressStoreKey(acc))
}

func TestLockStoreKey(t *testing.T) {
	key := "key"

	hexAddress := "b80f2a5df7d5710b15622d1a9f1e3830ded5bda8"
	acc, err := sdk.AccAddressFromHexUnsafe(hexAddress)
	require.NoError(t, err)

	expect, err := hex.DecodeString("11" + "14" + hexAddress + hex.EncodeToString([]byte(key)))
	require.NoError(t, err)
	require.Equal(t, expect, LockStoreKey(acc, key))
}

func TestLocksByPowerIndexKey(t *testing.T) {
	hexAddress := "b80f2a5df7d5710b15622d1a9f1e3830ded5bda8"
	acc, err := sdk.AccAddressFromHexUnsafe(hexAddress)
	require.NoError(t, err)

	expect, err := hex.DecodeString("80" + "14" + hexAddress)
	require.NoError(t, err)
	require.Equal(t, expect, LocksByPowerIndexKey(acc))
}

func TestLockByPowerIndexKey(t *testing.T) {
	key := "key"

	hexAddress := "b80f2a5df7d5710b15622d1a9f1e3830ded5bda8"
	acc, err := sdk.AccAddressFromHexUnsafe(hexAddress)
	require.NoError(t, err)

	lock := Lock{
		StakerAddress: acc.String(),
		Key:           key,
		Power:         sdkmath.NewInt(100),
	}

	expect, err := hex.DecodeString(
		"80" + "14" + hexAddress + "0000000000000064" + hex.EncodeToString([]byte(key)),
	)
	require.NoError(t, err)
	require.Equal(t, expect, LockByPowerIndexKey(lock))
}

func TestSplitLockByPowerIndexKey(t *testing.T) {
	key := "key"

	hexAddress := "b80f2a5df7d5710b15622d1a9f1e3830ded5bda8"
	expAddr, err := sdk.AccAddressFromHexUnsafe(hexAddress)
	expPower := sdkmath.NewInt(100)
	require.NoError(t, err)

	indexKey, err := hex.DecodeString(
		"80" + "14" + hexAddress + "0000000000000064" + hex.EncodeToString([]byte(key)),
	)
	require.NoError(t, err)

	addr, power := SplitLockByPowerIndexKey(indexKey)
	require.Equal(t, expAddr, addr)
	require.Equal(t, expPower, power)
}
