package types

import (
	"encoding/hex"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
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
