package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bandprotocol/chain/v3/x/tunnel/types"
)

func TestStringAxelarMemo(t *testing.T) {
	mockPayload := []byte{1, 2, 3}

	memo := types.NewAxelarMemo(
		"mock-chain",
		"0x75F01b3a2352bdc6e0D3983e40E09E9A8AAf4DF6",
		mockPayload,
		types.AxelarMessageTypeGeneralMessage,
		&types.AxelarFee{
			Amount:          "100",
			Recipient:       "axelar1aythygn6z5thymj6tmzfwekzh05ewg3l7d6y89",
			RefundRecipient: nil,
		},
	)
	memoStr, err := memo.String()
	require.NoError(t, err)
	require.Equal(
		t,
		`{"destination_chain":"mock-chain","destination_address":"0x75F01b3a2352bdc6e0D3983e40E09E9A8AAf4DF6","payload":"AQID","type":1,"fee":{"amount":"100","recipient":"axelar1aythygn6z5thymj6tmzfwekzh05ewg3l7d6y89","refund_recipient":null}}`,
		memoStr,
	)
}
