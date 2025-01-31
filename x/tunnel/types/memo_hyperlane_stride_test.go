package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bandprotocol/chain/v3/x/tunnel/types"
)

func TestStringHyperlaneStrideMemo(t *testing.T) {
	memo := types.NewHyperlaneStrideMemo(
		"stride13x29w58q38vytq03jnseg7jcfq9nxhcc8p4dfamp7p07a067y32qqwfrul",
		984122,
		"0xfceE86F472d0C19FccdD3AEDB89aa9cC0A1fb0D1",
		"3798c7f2000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000001600000000000000000000000000000000000000000000000000000000066dc9036000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000040000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000034254430000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000345544800000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000d4310000000000000000000000000000000000000000000000000000000000000c8a",
	)
	memoStr, err := memo.String()
	require.NoError(t, err)
	require.Equal(
		t,
		`{"wasm":{"contract":"stride13x29w58q38vytq03jnseg7jcfq9nxhcc8p4dfamp7p07a067y32qqwfrul","msg":{"dispatch":{"dest_domain":984122,"recipient_addr":"0xfceE86F472d0C19FccdD3AEDB89aa9cC0A1fb0D1","msg_body":"3798c7f2000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000001600000000000000000000000000000000000000000000000000000000066dc9036000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000040000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000034254430000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000345544800000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000d4310000000000000000000000000000000000000000000000000000000000000c8a"}}}}`,
		memoStr,
	)
}
