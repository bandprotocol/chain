package types_test

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bandprotocol/chain/v3/pkg/tss"
	feedstypes "github.com/bandprotocol/chain/v3/x/feeds/types"
	"github.com/bandprotocol/chain/v3/x/tunnel/types"
)

func TestEncoderPrefix(t *testing.T) {
	require.Equal(t, []byte(types.EncoderFixedPointABIPrefix), tss.Hash([]byte("FixedPointABI"))[:4])
	require.Equal(t, []byte(types.EncoderTickABIPrefix), tss.Hash([]byte("TickABI"))[:4])
}

func TestEncodeTSSFixedPrice(t *testing.T) {
	expectedMsg := ("cba0ad5a" +
		"0000000000000000000000000000000000000000000000000000000000000020" +
		"0000000000000000000000000000000000000000000000000000000000000003" +
		"0000000000000000000000000000000000000000000000000000000000000060" +
		"000000000000000000000000000000000000000000000000000000000000007b" +
		"0000000000000000000000000000000000000000000000000000000000000001" +
		"00000000000000000000000000000000000000000000007369676e616c5f3031" +
		"0000000000000000000000000000000000000000000000000000000000000002")

	msg, err := types.EncodeTSS(
		3,
		[]feedstypes.Price{
			{SignalID: "signal_01", Price: 2, Status: feedstypes.PRICE_STATUS_AVAILABLE},
		},
		123,
		feedstypes.ENCODER_FIXED_POINT_ABI,
	)
	require.NoError(t, err)

	require.Equal(t, expectedMsg, hex.EncodeToString(msg))
}

func TestEncodeTSSTick(t *testing.T) {
	expectedMsg := ("db99b2b3" +
		"0000000000000000000000000000000000000000000000000000000000000020" +
		"0000000000000000000000000000000000000000000000000000000000000003" +
		"0000000000000000000000000000000000000000000000000000000000000060" +
		"000000000000000000000000000000000000000000000000000000000000007b" +
		"0000000000000000000000000000000000000000000000000000000000000001" +
		"00000000000000000000000000000000000000000000007369676e616c5f3031" +
		"000000000000000000000000000000000000000000000000000000000000f188")

	msg, err := types.EncodeTSS(
		3,
		[]feedstypes.Price{
			{SignalID: "signal_01", Price: 2, Status: feedstypes.PRICE_STATUS_AVAILABLE},
		},
		123,
		feedstypes.ENCODER_TICK_ABI,
	)
	require.NoError(t, err)

	require.Equal(t, expectedMsg, hex.EncodeToString(msg))
}
