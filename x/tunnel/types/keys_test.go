package types_test

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/x/tunnel/types"
)

func TestTunnelStoreKey(t *testing.T) {
	expect, _ := hex.DecodeString("110000000000000001")
	require.Equal(t, expect, types.TunnelStoreKey(1))
}

func TestActiveTunnelIDStoreKey(t *testing.T) {
	expect, _ := hex.DecodeString("100000000000000001")
	require.Equal(t, expect, types.ActiveTunnelIDStoreKey(1))
}

func TestTunnelPacketsStoreKey(t *testing.T) {
	expect, _ := hex.DecodeString("120000000000000001")
	require.Equal(t, expect, types.TunnelPacketsStoreKey(1))
}

func TestTunnelPacketStoreKey(t *testing.T) {
	expect, _ := hex.DecodeString("1200000000000000010000000000000002")
	require.Equal(t, expect, types.TunnelPacketStoreKey(1, 2))
}

func TestLatestSignalPricesStoreKey(t *testing.T) {
	expect, _ := hex.DecodeString("130000000000000001")
	require.Equal(t, expect, types.LatestSignalPricesStoreKey(1))
}

func TestDepositsStoreKey(t *testing.T) {
	expect, _ := hex.DecodeString("140000000000000001")
	require.Equal(t, expect, types.DepositsStoreKey(1))
}

func TestDepositStoreKey(t *testing.T) {
	depositor := sdk.AccAddress([]byte("addr1"))
	expect, _ := hex.DecodeString("140000000000000001056164647231")
	require.Equal(t, expect, types.DepositStoreKey(1, depositor))
}

func TestParamsKey(t *testing.T) {
	expect, _ := hex.DecodeString("90")
	require.Equal(t, expect, types.ParamsKey)
}
