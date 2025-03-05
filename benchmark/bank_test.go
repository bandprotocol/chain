package benchmark

import (
	"encoding/json"
	"fmt"
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
			nsPerOp := int64(subB.Elapsed()) / int64(subB.N)

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
		Name    string `json:"sub_bench_name"`
		TxCount int    `json:"tx_count"`
		B_N     int    `json:"b_n"`
		NsPerOp int64  `json:"ns_per_op"`
	}

	// We accumulate all results in this slice and print them at the end.
	var allResults []benchRecord

	// One-time cleanup to print JSON after everything finishes.
	b.Cleanup(func() {
		data, _ := json.MarshalIndent(allResults, "", "  ")
		fmt.Println(string(data))
	})

	// Give a name to the sub-benchmark
	subBenchName := fmt.Sprintf("EmptyBlock")

	b.Run(subBenchName, func(subB *testing.B) {
		for i := 0; i < subB.N; i++ {
			// Stop timer during setup so it won't count toward NsPerOp.
			subB.StopTimer()

			// 1) Initialize the chain environment/app
			ba := InitializeBenchmarkApp(subB, -1)

			// 4) Start timer before the main operation (FinalizeBlock).
			subB.StartTimer()

			// 5) Finalize block with our send TXs
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
		}

		// Compute approximate NsPerOp using subB.Elapsed()
		// (available in Go 1.20+; otherwise measure time yourself)
		nsPerOp := int64(subB.Elapsed()) / int64(subB.N)

		// Append record
		allResults = append(allResults, benchRecord{
			Name:    subBenchName,
			B_N:     subB.N,
			NsPerOp: nsPerOp,
		})
	})
}
