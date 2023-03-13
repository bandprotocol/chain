package price_test

import (
	"encoding/hex"
	"testing"

	"github.com/bandprotocol/chain/v2/hooks/price"
	"github.com/stretchr/testify/require"
)

func hexToBytes(hexstr string) []byte {
	b, err := hex.DecodeString(hexstr)
	if err != nil {
		panic(err)
	}
	return b
}

func TestSuccessMustDecodeResultLegacy(t *testing.T) {
	expected := price.CommonOutput{
		Symbols:    []string{"BTC", "DAI", "DOT", "ETH", "LINK", "SUSHI", "UNI", "USDC", "USDT"},
		Rates:      []uint64{21968554120000, 990000000, 5764054360, 1578222740000, 6468030290, 1119460000, 5862500000, 990847600, 1004000000},
		Multiplier: 1000000000,
	}

	// This instance was obtained from https://www.cosmoscan.io/request/17304888
	calldata := hexToBytes("00000009000000034254430000000344414900000003444f5400000003455448000000044c494e4b00000005535553484900000003554e4900000004555344430000000455534454000000003b9aca00")
	result := hexToBytes("00000009000013faf3dd5340000000003b0233800000000157907d580000016f7567e2200000000181864f520000000042b99aa0000000015d6ea6a0000000003b0f2270000000003bd7d300")

	commonOutput := price.MustDecodeResult(calldata, result)
	require.EqualValues(t, expected, commonOutput)
}

func TestSuccessMustDecodeResult(t *testing.T) {
	expected := price.CommonOutput{
		Symbols:    []string{"BTC", "ETH"},
		Rates:      []uint64{21739584759800, 1538462293800},
		Multiplier: price.DefaultMultiplier,
	}

	// This instance was obtained from https://laozi-testnet6.cosmoscan.io/request/891387
	calldata := hexToBytes("00000002000000034254430000000345544803")
	result := hexToBytes("000000020000000342544300000013c5a43a27f8000000034554480000000166337f9f28")

	commonOutput := price.MustDecodeResult(calldata, result)

	require.EqualValues(t, expected, commonOutput)
}
