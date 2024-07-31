package types

import (
	"encoding/hex"
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestKeyStoreKey(t *testing.T) {
	keyName := "keyName"
	expect, err := hex.DecodeString("01" + hex.EncodeToString([]byte(keyName)))
	require.NoError(t, err)
	require.Equal(t, expect, KeyStoreKey(keyName))
}

func TestLocksStoreKey(t *testing.T) {
	hexAddress := "b80f2a5df7d5710b15622d1a9f1e3830ded5bda8"
	acc, err := sdk.AccAddressFromHexUnsafe(hexAddress)
	require.NoError(t, err)

	expect, err := hex.DecodeString("02" + "14" + hexAddress)
	require.NoError(t, err)
	require.Equal(t, expect, LocksStoreKey(acc))
}

func TestLockStoreKey(t *testing.T) {
	keyName := "keyName"

	hexAddress := "b80f2a5df7d5710b15622d1a9f1e3830ded5bda8"
	acc, err := sdk.AccAddressFromHexUnsafe(hexAddress)
	require.NoError(t, err)

	expect, err := hex.DecodeString("02" + "14" + hexAddress + hex.EncodeToString([]byte(keyName)))
	require.NoError(t, err)
	require.Equal(t, expect, LockStoreKey(acc, keyName))
}

func TestLocksByAmountIndexKey(t *testing.T) {
	hexAddress := "b80f2a5df7d5710b15622d1a9f1e3830ded5bda8"
	acc, err := sdk.AccAddressFromHexUnsafe(hexAddress)
	require.NoError(t, err)

	expect, err := hex.DecodeString("10" + "14" + hexAddress)
	require.NoError(t, err)
	require.Equal(t, expect, LocksByAmountIndexKey(acc))
}

func TestLockByAmountIndexKey(t *testing.T) {
	keyName := "keyName"

	hexAddress := "b80f2a5df7d5710b15622d1a9f1e3830ded5bda8"
	acc, err := sdk.AccAddressFromHexUnsafe(hexAddress)
	require.NoError(t, err)

	lock := Lock{
		LockerAddress:  acc.String(),
		Key:            keyName,
		Amount:         sdkmath.NewInt(100),
		PosRewardDebts: sdk.NewDecCoins(),
		NegRewardDebts: sdk.NewDecCoins(),
	}

	expect, err := hex.DecodeString(
		"10" + "14" + hexAddress + "0000000000000064" + hex.EncodeToString([]byte(keyName)),
	)
	require.NoError(t, err)
	require.Equal(t, expect, LockByAmountIndexKey(lock))
}
