package benchmark

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/bandprotocol/chain/v2/pkg/tss/testutil"
	tsstypes "github.com/bandprotocol/chain/v2/x/tss/types"
)

var RequestSignCases = map[string]struct {
	scenario   uint64
	byteLength []int
	feeLimit   sdk.Coins
}{
	"request_signature": {
		scenario:   1,
		byteLength: []int{1, 200, 400, 600},
		feeLimit:   sdk.NewCoins(sdk.NewCoin("uband", sdk.NewInt(10000))),
	},
}

func BenchmarkRequestSignatureDeliver(b *testing.B) {
	for name, tc := range RequestSignCases {
		for _, blen := range tc.byteLength {
			b.Run(fmt.Sprintf(
				"%s (byte_length: %d)",
				name,
				blen,
			), func(b *testing.B) {
				ba := InitializeBenchmarkApp(b, -1)

				ba.SetupGroup()

				msg := MockByte(blen)

				txs := GenSequenceOfTxs(
					ba.TxConfig,
					GenMsgRequestSignature(
						ba.Sender,
						ba.Gid,
						tsstypes.NewTextRequestSignature(msg),
						tc.feeLimit,
					),
					ba.Sender,
					b.N,
				)

				ba.CallBeginBlock()
				b.ResetTimer()
				b.StopTimer()

				// deliver MsgRequestSignature to the block
				for i := 0; i < b.N; i++ {
					b.StartTimer()
					gasInfo, _, err := ba.CallDeliver(txs[i])
					b.StopTimer()
					if i == 0 {
						if err != nil {
							fmt.Println("\tDeliver Error:", err.Error())
						} else {
							fmt.Println("\tCosmos Gas used:", gasInfo.GasUsed)
						}
					}
				}
			})
		}
	}
}

func BenchmarkSubmitSignatureDeliver(b *testing.B) {
	for name, tc := range RequestSignCases {
		for _, blen := range tc.byteLength {
			b.Run(fmt.Sprintf(
				"%s (byte_length: %d)",
				name,
				blen,
			), func(b *testing.B) {
				ba := InitializeBenchmarkApp(b, -1)

				ba.SetupGroup()

				ba.CallBeginBlock()
				b.ResetTimer()
				b.StopTimer()

				msg := MockByte(blen)

				// deliver MsgSubmitSignature to the block
				for i := 0; i < b.N; i++ {
					// generate tx
					txs := ba.HandleGenPendingSignTxs(
						ba.Gid,
						tsstypes.NewTextRequestSignature(msg),
						tc.feeLimit,
						testutil.TestCases,
					)

					b.StartTimer()
					gasInfo, _, err := ba.CallDeliver(txs[0])
					b.StopTimer()
					if i == 0 {
						if err != nil {
							fmt.Println("\tDeliver Error:", err.Error())
						} else {
							fmt.Println("\tCosmos Gas used:", gasInfo.GasUsed)
						}
					}
				}
			})
		}
	}
}

func BenchmarkBlockHandleProcessSigning(b *testing.B) {
	for name, tc := range RequestSignCases {
		for _, blen := range tc.byteLength {
			b.Run(fmt.Sprintf(
				"%s (byte_length: %d)",
				name,
				blen,
			), func(b *testing.B) {
				ba := InitializeBenchmarkApp(b, -1)

				ba.SetupGroup()

				b.ResetTimer()
				b.StopTimer()

				msg := MockByte(blen)

				// deliver MsgSubmitSignature to the block
				for i := 0; i < b.N; i++ {
					ba.CallBeginBlock()

					// generate tx
					ba.RequestSignature(ba.Sender, ba.Gid, tsstypes.NewTextRequestSignature(msg), tc.feeLimit)

					// everyone submit signature
					txs := ba.GetPendingSignTxs(ba.Gid, testutil.TestCases)
					for _, tx := range txs {
						_, _, err := ba.CallDeliver(tx)
						require.NoError(b, err)
					}
					ba.AddDEs(ba.Gid)

					b.StartTimer()
					ba.CallEndBlock()
					ba.Commit()
					b.StopTimer()
				}
			})
		}
	}
}
