package benchmark

import (
	"encoding/json"
	"fmt"
	"os"
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
	// Define a record type for sub-bench data
	type benchRecord struct {
		Name       string `json:"sub_bench_name"` // e.g. "request_signature (byte_length: 200)"
		ByteLength int    `json:"byte_length"`
		GasUsed    uint64 `json:"gas_used"`
		B_N        int    `json:"b_n"`
		NsPerOp    int64  `json:"ns_per_op"`
	}

	var allResults []benchRecord

	// Print/write JSON after all sub-benchmarks complete
	b.Cleanup(func() {
		data, _ := json.MarshalIndent(allResults, "", "  ")
		// Print JSON to stdout
		fmt.Println(string(data))

		// Write JSON to a file
		err := os.WriteFile("request_sig_bench.json", data, 0o644)
		if err != nil {
			b.Logf("Error writing request_sig_bench.json: %v", err)
		} else {
			b.Logf("Wrote %d benchmark results to request_sig_bench.json", len(allResults))
		}
	})

	for name, tc := range RequestSignCases {
		for _, blen := range tc.byteLength {
			// Name each sub-benchmark
			subBenchName := fmt.Sprintf("%s (byte_length: %d)", name, blen)

			b.Run(subBenchName, func(subB *testing.B) {
				var gasUsed uint64

				for i := 0; i < subB.N; i++ {
					ba := InitializeBenchmarkApp(subB, -1)
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
						subB.N,
					)

					// Finalize an empty block first if needed
					res, err := ba.FinalizeBlock(
						&abci.RequestFinalizeBlock{Height: ba.LastBlockHeight() + 1, Hash: ba.LastCommitID().Hash},
					)
					require.NoError(subB, err)

					for _, tx := range res.TxResults {
						require.Equal(subB, uint32(0), tx.Code)
					}

					// Measure time
					subB.ResetTimer()
					subB.StopTimer()

					tx, err := ba.TxDecoder(txs[0])
					require.NoError(subB, err)

					// Start timing only for the critical part
					subB.StartTimer()
					gasInfo, _, err := ba.SimDeliver(ba.TxEncoder, tx)
					subB.StopTimer()

					if err != nil {
						fmt.Println("\tDeliver Error:", err.Error())
					} else {
						fmt.Println("\tCosmos Gas used:", gasInfo.GasUsed)
					}
					gasUsed += gasInfo.GasUsed
				}

				// Build one record
				allResults = append(allResults, benchRecord{
					Name:       subBenchName,
					ByteLength: blen,
					GasUsed:    gasUsed / uint64(subB.N),
					B_N:        subB.N,
					NsPerOp:    int64(subB.Elapsed()) / int64(subB.N),
				})
			})
		}
	}
}

func BenchmarkSubmitSignatureDeliver(b *testing.B) {
	type benchRecord struct {
		Name       string `json:"sub_bench_name"` // e.g. "request_signature (byte_length: 200)"
		ByteLength int    `json:"byte_length"`
		GasUsed    uint64 `json:"gas_used"`
		B_N        int    `json:"b_n"`
		NsPerOp    int64  `json:"ns_per_op"`
	}

	var allResults []benchRecord

	b.Cleanup(func() {
		data, _ := json.MarshalIndent(allResults, "", "  ")
		fmt.Println(string(data))

		err := os.WriteFile("submit_sig_bench.json", data, 0o644)
		if err != nil {
			b.Logf("Error writing submit_sig_bench.json: %v", err)
		} else {
			b.Logf("Wrote %d benchmark results to submit_sig_bench.json", len(allResults))
		}
	})

	for name, tc := range RequestSignCases {
		for _, blen := range tc.byteLength {
			subBenchName := fmt.Sprintf("%s (byte_length: %d)", name, blen)

			b.Run(subBenchName, func(subB *testing.B) {
				var gasUsed uint64

				// We'll run subB.N times
				for i := 0; i < subB.N; i++ {
					ba := InitializeBenchmarkApp(subB, -1)
					ba.SetupTSSGroup()

					// Optionally finalize an initial empty block
					res, err := ba.FinalizeBlock(
						&abci.RequestFinalizeBlock{Height: ba.LastBlockHeight() + 1, Hash: ba.LastCommitID().Hash},
					)
					require.NoError(subB, err)

					for _, tx := range res.TxResults {
						require.Equal(subB, uint32(0), tx.Code)
					}

					subB.ResetTimer()
					subB.StopTimer()

					msg := MockByte(blen)
					// gather the group ID
					gid := ba.BandtssKeeper.GetCurrentGroup(ba.Ctx).GroupID
					require.NotZero(subB, gid)

					// generate tx
					txs := ba.HandleGenPendingSignTxs(
						gid,
						tsstypes.NewTextSignatureOrder(msg),
						tc.feeLimit,
					)
					require.NotEmpty(subB, txs, "no pending sign TXs generated")

					subB.StartTimer()
					gasInfo, _, err := ba.SimDeliver(ba.TxEncoder, txs[0])
					subB.StopTimer()

					if err != nil {
						fmt.Println("\tDeliver Error:", err.Error())
					} else {
						fmt.Println("\tCosmos Gas used:", gasInfo.GasUsed)
					}

					gasUsed += gasInfo.GasUsed
				}

				allResults = append(allResults, benchRecord{
					Name:       subBenchName,
					ByteLength: blen,
					GasUsed:    gasUsed / uint64(subB.N),
					B_N:        subB.N,
					NsPerOp:    int64(subB.Elapsed()) / int64(subB.N),
				})
			})
		}
	}
}

func BenchmarkEndBlockHandleProcessSigning(b *testing.B) {
	type benchRecord struct {
		Name       string `json:"sub_bench_name"` // e.g. "request_signature (byte_length: 200)"
		ByteLength int    `json:"byte_length"`
		B_N        int    `json:"b_n"`
		NsPerOp    int64  `json:"ns_per_op"`
	}

	var allResults []benchRecord

	b.Cleanup(func() {
		data, _ := json.MarshalIndent(allResults, "", "  ")
		fmt.Println(string(data))

		err := os.WriteFile("endblock_sig_bench.json", data, 0o644)
		if err != nil {
			b.Logf("Error writing endblock_sig_bench.json: %v", err)
		} else {
			b.Logf("Wrote %d benchmark results to endblock_sig_bench.json", len(allResults))
		}
	})

	for name, tc := range RequestSignCases {
		for _, blen := range tc.byteLength {
			subBenchName := fmt.Sprintf("%s (byte_length: %d)", name, blen)

			b.Run(subBenchName, func(subB *testing.B) {
				ba := InitializeBenchmarkApp(subB, -1)
				ba.SetupTSSGroup()

				subB.ResetTimer()
				subB.StopTimer()

				msg := MockByte(blen)

				for i := 0; i < subB.N; i++ {
					// finalize a block (end block)
					res, err := ba.FinalizeBlock(
						&abci.RequestFinalizeBlock{
							Height: ba.LastBlockHeight() + 1,
							Hash:   ba.LastCommitID().Hash,
							Time:   time.Now(),
						},
					)
					require.NoError(subB, err)

					gid := ba.BandtssKeeper.GetCurrentGroup(ba.Ctx).GroupID
					require.NotZero(subB, gid)

					// generate request signature
					ba.RequestSignature(ba.Sender, tsstypes.NewTextSignatureOrder(msg), tc.feeLimit)

					// everyone submit signature
					txs := ba.GetPendingSignTxs(gid)
					for _, tx := range txs {
						_, _, err := ba.SimDeliver(ba.TxEncoder, tx)
						require.NoError(subB, err)
					}

					// Possibly add DEs to each member
					members := ba.TSSKeeper.MustGetMembers(ba.Ctx, gid)
					for _, m := range members {
						ba.AddDEs(sdk.MustAccAddressFromBech32(m.Address))
					}

					subB.StartTimer()
					// finalize next block, which triggers end-block logic
					res2, err := ba.FinalizeBlock(&abci.RequestFinalizeBlock{
						Height: ba.LastBlockHeight() + 1,
						Time:   time.Now(),
					})
					subB.StopTimer()

					require.NoError(subB, err)

					// check TX results
					for _, txr := range res.TxResults {
						require.Equal(subB, uint32(0), txr.Code)
					}
					for _, txr := range res2.TxResults {
						require.Equal(subB, uint32(0), txr.Code)
					}

					_, err = ba.Commit()
					require.NoError(subB, err)
				}

				// Append a record
				allResults = append(allResults, benchRecord{
					Name:       subBenchName,
					ByteLength: blen,
					B_N:        subB.N,
					NsPerOp:    int64(subB.Elapsed())/int64(subB.N) - 2000000,
				})
			})
		}
	}
}
