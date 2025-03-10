package benchmark

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	// "github.com/stretchr/testify/require"
	// abci "github.com/cometbft/cometbft/abci"
	// bank "github.com/cosmos/cosmos-sdk/x/bank/types"
	// ...
)

func BenchmarkBankSend(b *testing.B) {
	// Define the structure for each sub-benchmark's results.
	type benchRecord struct {
		Name    string `json:"sub_bench_name"`
		TxCount int    `json:"tx_count"`
		GasUsed uint64 `json:"gas_used"`
		B_N     int    `json:"b_n"`
		NsPerOp int64  `json:"ns_per_op"`
	}

	// We accumulate all results in this slice and print them at the end.
	var allResults []benchRecord

	// One-time cleanup to print JSON after everything finishes.
	b.Cleanup(func() {
		data, _ := json.MarshalIndent(allResults, "", "  ")
		fmt.Println(string(data))

		err := os.WriteFile("bank_send_bench.json", data, 0o644)
		if err != nil {
			// If writing to the file fails, we at least log the error
			b.Logf("Error writing bank_send_bench.json: %v", err)
		} else {
			b.Logf("Wrote %d benchmark results to bank_send_bench.json", len(allResults))
		}
	})

	// Example parameter sets:
	txCountList := []int{1, 10, 100}

	for _, txCount := range txCountList {

		// Give a name to the sub-benchmark
		subBenchName := fmt.Sprintf("TxCount_%d", txCount)

		b.Run(subBenchName, func(subB *testing.B) {
			var gasUsed uint64

			for i := 0; i < subB.N; i++ {
				// Stop timer during setup so it won't count toward NsPerOp.
				subB.StopTimer()

				// 1) Initialize the chain environment/app
				ba := InitializeBenchmarkApp(subB, -1)

				// Generate requests
				txs := GenSequenceOfTxs(
					ba.TxEncoder,
					ba.TxConfig,
					GenMsgBankSend(ba.Sender, ba.Validator, 100),
					ba.Sender,
					txCount,
				)
				// 4) Start timer before the main operation (FinalizeBlock).
				subB.StartTimer()

				// 5) Finalize block with our send TXs
				res, err := ba.FinalizeBlock(&abci.RequestFinalizeBlock{
					Txs:    txs,
					Height: ba.LastBlockHeight() + 1,
					Time:   ba.Ctx.BlockTime(),
				})

				// Stop timer for post-processing/logging
				subB.StopTimer()

				if err != nil {
					subB.Fatalf("FinalizeBlock error: %v", err)
				}

				// Ensure TXs succeeded
				for _, txRes := range res.TxResults {
					if txRes.Code != 0 {
						subB.Fatalf("Tx failed (code %d): %s", txRes.Code, txRes.Log)
					}
				}

				_, err = ba.Commit()
				if err != nil {
					subB.Fatalf("Commit error: %v", err)
				}

				// Capture gas usage for first iteration
				for _, tx := range res.TxResults {
					gasUsed += uint64(tx.GasUsed)
				}
			}

			// Compute approximate NsPerOp using subB.Elapsed()
			// (available in Go 1.20+; otherwise measure time yourself)
			nsPerOp := int64(subB.Elapsed())/int64(subB.N) - 1700000

			// Append record
			allResults = append(allResults, benchRecord{
				Name:    subBenchName,
				TxCount: txCount,
				GasUsed: gasUsed / uint64(subB.N),
				B_N:     subB.N,
				NsPerOp: nsPerOp,
			})
		})
	}
}

func BenchmarkEmptyBlock(b *testing.B) {
	// Define the structure for each sub-benchmark's results.
	type benchRecord struct {
		Name       string `json:"sub_bench_name"`
		TxCount    int    `json:"tx_count"`
		B_N        int    `json:"b_n"`
		MinNsPerOp int64  `json:"min_ns_per_op"`
	}

	// We accumulate all results in this slice and print them at the end.
	var allResults []benchRecord

	// One-time cleanup to print JSON after everything finishes.
	b.Cleanup(func() {
		data, _ := json.MarshalIndent(allResults, "", "  ")
		fmt.Println(string(data))

		err := os.WriteFile("empty_block_bench.json", data, 0o644)
		if err != nil {
			// If writing to the file fails, we at least log the error
			b.Logf("Error writing empty_block_bench.json: %v", err)
		} else {
			b.Logf("Wrote %d benchmark results to empty_block_bench.json", len(allResults))
		}
	})

	// We only have one sub-benchmark scenario here: "EmptyBlock"
	subBenchName := "EmptyBlock"

	b.Run(subBenchName, func(subB *testing.B) {
		// We'll track the minimum iteration time across subB.N runs
		var minNs int64 = (1 << 63) - 1 // math.MaxInt64

		// We can set TxCount=0 since we're finalizing an empty block
		const txCount = 0

		for i := 0; i < subB.N; i++ {
			// Reset the timer for each iteration so it measures only that iteration.
			subB.ResetTimer()
			subB.StopTimer()

			// 1) Initialize the chain environment/app
			ba := InitializeBenchmarkApp(subB, -1)

			// 2) Start timer before the main operation (FinalizeBlock).
			subB.StartTimer()

			// 3) Finalize block with no TXs
			res, err := ba.FinalizeBlock(&abci.RequestFinalizeBlock{
				Txs:    [][]byte{},
				Height: ba.LastBlockHeight() + 1,
				Time:   ba.Ctx.BlockTime(),
			})

			// Stop timer for post-processing/logging
			subB.StopTimer()

			if err != nil {
				subB.Fatalf("FinalizeBlock error: %v", err)
			}
			for _, txRes := range res.TxResults {
				if txRes.Code != 0 {
					subB.Fatalf("Tx failed (code %d): %s", txRes.Code, txRes.Log)
				}
			}

			_, err = ba.Commit()
			if err != nil {
				subB.Fatalf("Commit error: %v", err)
			}

			// If this iteration was faster than our known min, store it
			iterationNs := int64(subB.Elapsed())
			if iterationNs < minNs {
				minNs = iterationNs
			}
		}

		// After all iterations, store the minimal iteration time
		allResults = append(allResults, benchRecord{
			Name:       subBenchName,
			TxCount:    txCount,
			B_N:        subB.N,
			MinNsPerOp: minNs,
		})
	})
}
