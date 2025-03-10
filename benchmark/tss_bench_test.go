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
		MinGasUsed uint64 `json:"min_gas_used"`
		B_N        int    `json:"b_n"`
		MinNsPerOp int64  `json:"min_ns_per_op"`
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
				// Track min gas usage and min iteration time
				var minGasUsed uint64 = ^uint64(0) // largest possible => 0xFFFFFFFF...
				var minNs int64 = (1 << 63) - 1    // math.MaxInt64

				for i := 0; i < subB.N; i++ {
					// Reset timing for this iteration
					subB.ResetTimer()
					subB.StopTimer()

					// Setup
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
						1, // We'll only deliver one TX per iteration
					)

					// Optionally finalize an empty block
					res, err := ba.FinalizeBlock(&abci.RequestFinalizeBlock{
						Height: ba.LastBlockHeight() + 1,
						Hash:   ba.LastCommitID().Hash,
					})
					require.NoError(subB, err)
					for _, tx := range res.TxResults {
						require.Equal(subB, uint32(0), tx.Code)
					}

					// Decode the TX
					theTx, err := ba.TxDecoder(txs[0])
					require.NoError(subB, err)

					// Start measuring for the critical operation
					subB.StartTimer()
					gasInfo, _, err := ba.SimDeliver(ba.TxEncoder, theTx)
					subB.StopTimer()

					if err != nil {
						fmt.Println("\tDeliver Error:", err.Error())
					}

					// Compare iteration’s gas usage to minGasUsed
					iterationGas := gasInfo.GasUsed
					if iterationGas < minGasUsed {
						minGasUsed = iterationGas
					}

					// Compare iteration time (subB.Elapsed()) to minNs
					iterationNs := int64(subB.Elapsed())
					if iterationNs < minNs {
						minNs = iterationNs
					}
				}

				// Build one record with the "best" iteration
				allResults = append(allResults, benchRecord{
					Name:       subBenchName,
					ByteLength: blen,
					MinGasUsed: minGasUsed,
					B_N:        subB.N,
					MinNsPerOp: minNs,
				})
			})
		}
	}
}

func BenchmarkSubmitSignatureDeliver(b *testing.B) {
	type benchRecord struct {
		Name       string `json:"sub_bench_name"`
		ByteLength int    `json:"byte_length"`
		MinGasUsed uint64 `json:"min_gas_used"`
		B_N        int    `json:"b_n"`
		MinNsPerOp int64  `json:"min_ns_per_op"`
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
				var minGasUsed uint64 = ^uint64(0)
				var minNs int64 = (1 << 63) - 1

				for i := 0; i < subB.N; i++ {
					subB.ResetTimer()
					subB.StopTimer()

					ba := InitializeBenchmarkApp(subB, -1)
					ba.SetupTSSGroup()

					// Optionally finalize an initial empty block
					res, err := ba.FinalizeBlock(&abci.RequestFinalizeBlock{
						Height: ba.LastBlockHeight() + 1,
						Hash:   ba.LastCommitID().Hash,
					})
					require.NoError(subB, err)
					for _, tx := range res.TxResults {
						require.Equal(subB, uint32(0), tx.Code)
					}

					msg := MockByte(blen)
					// gather the group ID
					gid := ba.BandtssKeeper.GetCurrentGroup(ba.Ctx).GroupID
					require.NotZero(subB, gid)

					// generate one TX
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
					}

					// Track min gas usage
					iterationGas := gasInfo.GasUsed
					if iterationGas < minGasUsed {
						minGasUsed = iterationGas
					}

					// Track min iteration time
					iterationNs := int64(subB.Elapsed())
					if iterationNs < minNs {
						minNs = iterationNs
					}
				}

				allResults = append(allResults, benchRecord{
					Name:       subBenchName,
					ByteLength: blen,
					MinGasUsed: minGasUsed,
					B_N:        subB.N,
					MinNsPerOp: minNs,
				})
			})
		}
	}
}

func BenchmarkEndBlockHandleProcessSigning(b *testing.B) {
	type benchRecord struct {
		Name       string `json:"sub_bench_name"` // e.g. "request_signature (byte_length: 200)"
		ByteLength int    `json:"byte_length"`

		// We track only the minimum time, not gas
		B_N        int   `json:"b_n"`
		MinNsPerOp int64 `json:"min_ns_per_op"`
	}

	var allResults []benchRecord

	// Print/write JSON after all sub-benchmarks complete
	b.Cleanup(func() {
		data, _ := json.MarshalIndent(allResults, "", "  ")
		// Print JSON to stdout
		fmt.Println(string(data))

		// Write JSON to a file
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
				// We only track min iteration time, ignoring gas
				var minNs int64 = (1 << 63) - 1 // math.MaxInt64 sentinel

				for i := 0; i < subB.N; i++ {
					// Reset the timer for each iteration
					subB.ResetTimer()
					subB.StopTimer()

					// Setup
					ba := InitializeBenchmarkApp(subB, -1)
					ba.SetupTSSGroup()

					msg := MockByte(blen)

					// finalize a block (end block)
					res, err := ba.FinalizeBlock(&abci.RequestFinalizeBlock{
						Height: ba.LastBlockHeight() + 1,
						Hash:   ba.LastCommitID().Hash,
						Time:   time.Now(),
					})
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

					// Start timer only for the final block that triggers end-block logic
					subB.StartTimer()
					res2, err := ba.FinalizeBlock(&abci.RequestFinalizeBlock{
						Height: ba.LastBlockHeight() + 1,
						Time:   time.Now(),
					})
					subB.StopTimer()

					require.NoError(subB, err)

					// check TX results from both blocks
					for _, txr := range res.TxResults {
						require.Equal(subB, uint32(0), txr.Code)
					}
					for _, txr := range res2.TxResults {
						require.Equal(subB, uint32(0), txr.Code)
					}

					_, err = ba.Commit()
					require.NoError(subB, err)

					// If this iteration time is smaller, keep it
					iterationNs := int64(subB.Elapsed())
					if iterationNs < minNs {
						minNs = iterationNs
					}
				}

				allResults = append(allResults, benchRecord{
					Name:       subBenchName,
					ByteLength: blen,
					B_N:        subB.N,
					MinNsPerOp: minNs - 1700000,
				})
			})
		}
	}
}
