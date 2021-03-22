package types

import (
	"encoding/hex"
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/bandprotocol/chain/pkg/obi"
)

func mustDecodeString(hexstr string) []byte {
	b, err := hex.DecodeString(hexstr)
	if err != nil {
		panic(err)
	}
	return b
}

func TestGetBytesRequestPacket(t *testing.T) {
	req := OracleRequestPacketData{
		ClientID:       "test",
		OracleScriptID: 1,
		Calldata:       mustDecodeString("030000004254436400000000000000"),
		AskCount:       1,
		MinCount:       1,
		FeeLimit:       sdk.NewCoins(sdk.NewCoin("uband", sdk.NewInt(10000))),
		RequestKey:     "TEST_KEY",
		PrepareGas:     100,
		ExecuteGas:     100,
	}
	fmt.Println(string(req.GetBytes()))
	require.Equal(t,
		[]byte(`{"ask_count":"1","calldata":"AwAAAEJUQ2QAAAAAAAAA","client_id":"test","execute_gas":"100","fee_limit":[{"amount":"10000","denom":"uband"}],"min_count":"1","oracle_script_id":"1","prepare_gas":"100","request_key":"TEST_KEY"}`),
		req.GetBytes(),
	)
}

func TestGetBytesResponsePacket(t *testing.T) {
	res := OracleResponsePacketData{
		ClientID:      "test",
		RequestID:     1,
		AnsCount:      1,
		RequestTime:   1589535020,
		ResolveTime:   1589535022,
		ResolveStatus: ResolveStatus(1),
		Result:        mustDecodeString("4bb10e0000000000"),
	}
	require.Equal(t, []byte(`{"ans_count":"1","client_id":"test","request_id":"1","request_time":"1589535020","resolve_status":"RESOLVE_STATUS_SUCCESS","resolve_time":"1589535022","result":"S7EOAAAAAAA="}`), res.GetBytes())
}

func TestOBIEncodeResult(t *testing.T) {
	result := NewResult(
		"beeb",
		1,
		mustDecodeString("0000000342544300000000000003e8"),
		1,
		1,
		2,
		1,
		1591622616,
		1591622618,
		ResolveStatus(1),
		mustDecodeString("00000000009443ee"),
	)
	expectedEncodedResult := mustDecodeString("000000046265656200000000000000010000000f0000000342544300000000000003e80000000000000001000000000000000100000000000000020000000000000001000000005ede3bd8000000005ede3bda000000010000000800000000009443ee")
	require.Equal(t, expectedEncodedResult, obi.MustEncode(result))
}

func TestOBIEncodeResultOfEmptyClientID(t *testing.T) {
	result := NewResult(
		"",
		1,
		mustDecodeString("0000000342544300000000000003e8"),
		1,
		1,
		1,
		1,
		1591622426,
		1591622429,
		ResolveStatus(1),
		mustDecodeString("0000000000944387"),
	)
	expectedEncodedResult := mustDecodeString("0000000000000000000000010000000f0000000342544300000000000003e80000000000000001000000000000000100000000000000010000000000000001000000005ede3b1a000000005ede3b1d00000001000000080000000000944387")
	require.Equal(t, expectedEncodedResult, obi.MustEncode(result))
}
