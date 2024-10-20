package benchmark

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	abci "github.com/cometbft/cometbft/abci/types"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	tsstypes "github.com/bandprotocol/chain/v3/x/tss/types"
)

var RequestSignCases = map[string]struct {
	byteLength []int
	feeLimit   sdk.Coins
}{
	"request_signature": {
		byteLength: []int{1, 200, 400, 600},
		feeLimit:   sdk.NewCoins(sdk.NewCoin("uband", math.NewInt(10000))),
	},
}

func BenchmarkRequestSignatureDeliver(b *testing.B) {
	for name, tc := range RequestSignCases {
		for _, blen := range tc.byteLength {
			b.Run(fmt.Sprintf("%s (byte_length: %d)", name, blen), func(b *testing.B) {
				ba := InitializeBenchmarkApp(b, -1)
				ba.SetupTSSGroup()

				msg := MockByte(blen)
				txs := GenSequenceOfTxs(
					ba.TxEncoder,
					ba.TxConfig,
					GenMsgRequestSignature(
						ba.Sender,
						tsstypes.NewTextSignatureOrder(msg),
						tc.feeLimit,
					),
					ba.Sender,
					b.N,
				)

				_, err := ba.FinalizeBlock(
					&abci.RequestFinalizeBlock{Height: ba.LastBlockHeight() + 1, Hash: ba.LastCommitID().Hash},
				)
				require.NoError(b, err)

				b.ResetTimer()
				b.StopTimer()

				// deliver MsgRequestSignature to the block
				for i := 0; i < b.N; i++ {
					tx, err := ba.TxDecoder(txs[i])
					require.NoError(b, err)

					b.StartTimer()
					gasInfo, _, err := ba.SimDeliver(ba.TxEncoder, tx)
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
			b.Run(fmt.Sprintf("%s (byte_length: %d)", name, blen), func(b *testing.B) {
				ba := InitializeBenchmarkApp(b, -1)
				ba.SetupTSSGroup()

				_, err := ba.FinalizeBlock(
					&abci.RequestFinalizeBlock{Height: ba.LastBlockHeight() + 1, Hash: ba.LastCommitID().Hash},
				)
				require.NoError(b, err)

				b.ResetTimer()
				b.StopTimer()

				msg := MockByte(blen)

				// deliver MsgSubmitSignature to the block
				for i := 0; i < b.N; i++ {
					gid := ba.BandtssKeeper.GetCurrentGroupID(ba.Ctx)
					require.NotZero(b, gid)

					// generate tx
					txs := ba.HandleGenPendingSignTxs(
						gid,
						tsstypes.NewTextSignatureOrder(msg),
						tc.feeLimit,
					)

					b.StartTimer()
					gasInfo, _, err := ba.SimDeliver(ba.TxEncoder, txs[0])
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

func BenchmarkEndBlockHandleProcessSigning(b *testing.B) {
	for name, tc := range RequestSignCases {
		for _, blen := range tc.byteLength {
			b.Run(fmt.Sprintf("%s (byte_length: %d)", name, blen), func(b *testing.B) {
				ba := InitializeBenchmarkApp(b, -1)
				ba.SetupTSSGroup()

				b.ResetTimer()
				b.StopTimer()

				msg := MockByte(blen)

				// deliver MsgSubmitSignature to the block
				for i := 0; i < b.N; i++ {
					_, err := ba.FinalizeBlock(
						&abci.RequestFinalizeBlock{
							Height: ba.LastBlockHeight() + 1,
							Hash:   ba.LastCommitID().Hash,
						},
					)
					require.NoError(b, err)

					gid := ba.BandtssKeeper.GetCurrentGroupID(ba.Ctx)
					require.NotZero(b, gid)

					// generate tx
					ba.RequestSignature(ba.Sender, tsstypes.NewTextSignatureOrder(msg), tc.feeLimit)

					// everyone submit signature
					txs := ba.GetPendingSignTxs(gid)
					for _, tx := range txs {
						_, _, err := ba.SimDeliver(ba.TxEncode, tx)
						require.NoError(b, err)
					}

					members := ba.TSSKeeper.MustGetMembers(ba.Ctx, gid)
					for _, m := range members {
						ba.AddDEs(sdk.MustAccAddressFromBech32(m.Address))
					}

					b.StartTimer()
					_, err = ba.FinalizeBlock(
						&abci.RequestFinalizeBlock{
							Height: ba.LastBlockHeight() + 1,
							Time:   time.Now(),
						},
					)
					b.StopTimer()

					require.NoError(b, err)

					_, err = ba.Commit()
					require.NoError(b, err)
				}
			})
		}
	}
}
