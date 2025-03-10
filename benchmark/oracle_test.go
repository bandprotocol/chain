package benchmark

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	abci "github.com/cometbft/cometbft/abci/types"

	oraclekeeper "github.com/bandprotocol/chain/v3/x/oracle/keeper"
	oracletypes "github.com/bandprotocol/chain/v3/x/oracle/types"
)

var PrepareCases = map[string]struct {
	scenario     uint64
	parameters   []uint64
	stringLength []int
}{
	"ask_external_data": {
		scenario:     1,
		parameters:   []uint64{1, 4, 8, 16},
		stringLength: []int{1, 200, 400, 600},
	},
	"infinite_loop": {
		scenario:     2,
		parameters:   []uint64{0},
		stringLength: []int{1},
	},
	"arithmetic_ops": {
		scenario:     3,
		parameters:   []uint64{1, 100, 10000, 1000000, math.MaxUint64},
		stringLength: []int{1},
	},
	"allocate_mem": {
		scenario:     4,
		parameters:   []uint64{1, 100, 10000, 1000000, math.MaxUint64},
		stringLength: []int{1},
	},
	"find_median": {
		scenario:     5,
		parameters:   []uint64{1, 100, 10000, 1000000, math.MaxUint64},
		stringLength: []int{1},
	},
	"finite_loop": {
		scenario:     6,
		parameters:   []uint64{1, 100, 10000, 1000000, math.MaxUint64},
		stringLength: []int{1},
	},
	"set_local_var": {
		scenario:     7,
		parameters:   []uint64{1, 100, 10000, 1000000, math.MaxUint64},
		stringLength: []int{1},
	},
}

var ExecuteCases = map[string]struct {
	scenario     uint64
	parameters   []uint64
	numRequests  []int
	stringLength []int
}{
	"nothing": {
		scenario:     0,
		parameters:   []uint64{0},
		stringLength: []int{1},
	},
	"infinite_loop": {
		scenario:     101,
		parameters:   []uint64{0},
		stringLength: []int{1},
	},
	"arithmetic_ops": {
		scenario:   102,
		parameters: []uint64{1, 100, 100000, math.MaxUint64},
	},
	"allocate_mem": {
		scenario:     103,
		parameters:   []uint64{1, 100, 100000, math.MaxUint64},
		stringLength: []int{1},
	},
	"find_median": {
		scenario:     104,
		parameters:   []uint64{1, 100, 100000, math.MaxUint64},
		stringLength: []int{1},
	},
	"finite_loop": {
		scenario:     105,
		parameters:   []uint64{1, 100, 100000, math.MaxUint64},
		stringLength: []int{1},
	},
	"set_local_var": {
		scenario:     106,
		parameters:   []uint64{1, 100, 100000, math.MaxUint64},
		stringLength: []int{1},
	},
	"get_ask_count": {
		scenario:     201,
		parameters:   []uint64{1, 100, 100000, math.MaxUint64},
		stringLength: []int{1},
	},
	"get_min_count": {
		scenario:     202,
		parameters:   []uint64{1, 100, 100000, math.MaxUint64},
		stringLength: []int{1},
	},
	"get_prepare_time": {
		scenario:     203,
		parameters:   []uint64{1, 100, 100000, math.MaxUint64},
		stringLength: []int{1},
	},
	"get_execute_time": {
		scenario:     204,
		parameters:   []uint64{1, 100, 100000, math.MaxUint64},
		stringLength: []int{1},
	},
	"get_ans_count": {
		scenario:     205,
		parameters:   []uint64{1, 100, 100000, math.MaxUint64},
		stringLength: []int{1},
	},
	"get_calldata": {
		scenario:     206,
		parameters:   []uint64{1, 100, 100000, math.MaxUint64},
		stringLength: []int{1, 200, 400, 600},
	},
	"save_return_data": {
		scenario:     207,
		parameters:   []uint64{1, 100, 100000, math.MaxUint64},
		stringLength: []int{1, 200, 400, 600},
	},
	"get_external_data": {
		scenario:     208,
		parameters:   []uint64{1, 100, 100000, math.MaxUint64},
		stringLength: []int{1, 200, 400, 600},
	},
	"ecvrf_verify": {
		scenario:     209,
		parameters:   []uint64{1, 100, 100000, math.MaxUint64},
		stringLength: []int{1},
	},
	"base_import": {
		scenario:     210,
		parameters:   []uint64{0},
		stringLength: []int{1},
	},
}

var CacheCases = map[string]uint32{
	"no_cache": 0,
	"cache":    1,
}

var (
	PrepareGasLimit  uint64 = 7_500_000
	ExecuteGasLimit  uint64 = 7_500_000
	BlockMaxGas      int64  = 50_000_000
	GasRanges               = []int{1, 1_000, 10_000, 100_000, 1_000_000, 7_900_000}
	NumRequestRanges        = []int{0, 1, 5, 10, 20}
)

// benchmark test for prepare function of owasm vm
func BenchmarkOwasmVMPrepare(b *testing.B) {
	for cache, cacheSize := range CacheCases {
		for name, tc := range PrepareCases {
			for _, pm := range tc.parameters {
				for _, strlen := range tc.stringLength {
					b.Run(fmt.Sprintf(
						"%s - %s (param: %d, strlen: %d)",
						cache,
						name,
						pm,
						strlen,
					), func(b *testing.B) {
						owasmVM, compiledCode, req := InitOwasmTestEnv(b, cacheSize, tc.scenario, pm, strlen)

						b.ResetTimer()
						b.StopTimer()

						// call prepare on new env
						for i := 0; i < b.N; i++ {
							env := oracletypes.NewPrepareEnv(
								req,
								int64(oracletypes.DefaultMaxCalldataSize),
								int64(oracletypes.DefaultMaxRawRequestCount),
								int64(GetSpanSize()),
							)
							b.StartTimer()
							res, err := owasmVM.Prepare(
								compiledCode,
								oraclekeeper.ConvertToOwasmGas(PrepareGasLimit),
								env,
							)
							b.StopTimer()
							if i == 0 {
								if err != nil {
									fmt.Println("\tDeliver Error:", err.Error())
								} else {
									fmt.Println("\tOwasm Gas used:", res.GasUsed)
								}
							}
						}
					})
				}
			}
		}
	}
}

// benchmark test for execute function of owasm vm
func BenchmarkOwasmVMExecute(b *testing.B) {
	// We define a struct to record each sub-benchmark result
	type benchRecord struct {
		Cache        string `json:"cache"`         // e.g., the cache key in CacheCases
		CaseName     string `json:"case_name"`     // e.g., "tc.scenario" or the sub-bench name
		Param        uint64 `json:"param"`         // each pm in tc.parameters
		StringLength int    `json:"string_length"` // each strlen in tc.stringLength
		GasUsedFirst uint64 `json:"gas_used_first_tx"`
		N            int    `json:"b_n"`
		NsPerOp      int64  `json:"ns_per_op"`
	}

	// We'll accumulate records for *all* sub-benchmarks here
	var allResults []benchRecord

	// We'll print the JSON after everything finishes
	b.Cleanup(func() {
		data, _ := json.MarshalIndent(allResults, "", "  ")
		fmt.Println(string(data))
	})

	for cache, cacheSize := range CacheCases {
		for name, tc := range ExecuteCases {
			for _, pm := range tc.parameters {
				for _, strlen := range tc.stringLength {

					// Define the sub-benchmark name
					subBenchName := fmt.Sprintf(
						"%s - %s (param: %d, strlen: %d)",
						cache,
						name,
						pm,
						strlen,
					)

					b.Run(subBenchName, func(subB *testing.B) {
						// We'll measure the total elapsed time for subB.N iterations
						startTime := time.Now()

						// Capture gas used in the first iteration
						var gasUsedFirstTx uint64

						// Construct or initialize everything just once outside the main loop
						owasmVM, compiledCode, req := InitOwasmTestEnv(subB, cacheSize, tc.scenario, pm, strlen)

						subB.ResetTimer()
						subB.StopTimer()

						// Perform subB.N iterations
						for i := 0; i < subB.N; i++ {
							// Build environment
							env := oracletypes.NewExecuteEnv(
								req,
								GenOracleReports(),
								time.Now(),
								int64(GetSpanSize()),
							)

							// Start timing the critical part
							subB.StartTimer()
							res, err := owasmVM.Execute(
								compiledCode,
								oraclekeeper.ConvertToOwasmGas(ExecuteGasLimit),
								env,
							)
							subB.StopTimer()

							if i == 0 {
								if err != nil {
									fmt.Println("\tEndblock Error:", err.Error())
								} else {
									fmt.Println("\tOwasm Gas used:", res.GasUsed)
									gasUsedFirstTx = res.GasUsed
								}
							}
						}

						// Manually measure total time and compute approximate ns/op
						totalElapsed := time.Since(startTime).Nanoseconds()
						nsPerOp := totalElapsed / int64(subB.N)

						// Save a record for this sub-benchmark
						allResults = append(allResults, benchRecord{
							Cache:        cache,
							CaseName:     name, // or use subBenchName if you like
							Param:        pm,
							StringLength: strlen,
							GasUsedFirst: gasUsedFirstTx,
							N:            subB.N,
							NsPerOp:      nsPerOp,
						})
					})
				}
			}
		}
	}
}

// BenchmarkBlockOracleMsgRequestData benchmarks MsgRequestData of oracle module
func BenchmarkBlockOracleMsgRequestData(b *testing.B) {
	for name, tc := range PrepareCases {
		for _, pm := range tc.parameters {
			for _, strlen := range tc.stringLength {
				for _, reqPerBlock := range []int{1, 5, 10, 20} {
					b.Run(
						fmt.Sprintf(
							"%s (param: %d, strlen: %d) - %d requests/block",
							name,
							pm,
							strlen,
							reqPerBlock,
						),
						func(b *testing.B) {
							b.ResetTimer()
							b.StopTimer()

							for i := 0; i < b.N; i++ {
								ba := InitializeBenchmarkApp(b, -1)

								txs := GenSequenceOfTxs(
									ba.TxEncoder,
									ba.TxConfig,
									GenMsgRequestData(
										ba.Sender,
										ba.Oid,
										ba.Did,
										tc.scenario,
										pm,
										strlen,
										PrepareGasLimit,
										1000,
									),
									ba.Sender,
									reqPerBlock,
								)

								b.StartTimer()

								res, err := ba.FinalizeBlock(
									&abci.RequestFinalizeBlock{
										Txs:    txs,
										Height: ba.LastBlockHeight() + 1,
										Time:   ba.Ctx.BlockTime(),
									},
								)
								b.StopTimer()

								require.NoError(b, err)

								if i == 0 {
									if res.TxResults[len(res.TxResults)-1].Code != 0 {
										fmt.Println("\tDeliver Error:", res.TxResults[0].Log)
									} else {
										fmt.Println("\tCosmos Gas used:", res.TxResults[0].GasUsed)
									}
								}
							}
						},
					)
				}
			}
		}
	}
}

// BenchmarkBlockOracleMsgReportData benchmarks MsgReportData of oracle module
func BenchmarkBlockOracleMsgReportData(b *testing.B) {
	// 1) Define a struct to hold the data we want in JSON.
	type benchRecord struct {
		Scenario     string `json:"scenario"`      // "name" from ExecuteCases
		Param        uint64 `json:"param"`         // pm
		StringLength int    `json:"string_length"` // strlen
		ReqPerBlock  int    `json:"req_per_block"`

		// We'll record the MIN gas usage and MIN time among subB.N iterations
		MinGasUsed uint64 `json:"min_gas_used"`
		B_N        int    `json:"b_n"`
		MinNsPerOp int64  `json:"min_ns_per_op"`
	}

	// 2) We'll accumulate results in this slice. We'll print them in a Cleanup callback.
	var allResults []benchRecord

	// 3) Schedule one-time cleanup to print JSON after all sub-benchmarks complete.
	b.Cleanup(func() {
		data, _ := json.MarshalIndent(allResults, "", "  ")
		fmt.Println(string(data))

		err := os.WriteFile("oracle_report_bench.json", data, 0o644)
		if err != nil {
			b.Logf("Error writing oracle_report_bench.json: %v", err)
		} else {
			b.Logf("Wrote %d benchmark results to oracle_report_bench.json", len(allResults))
		}
	})

	// Loop over your test configurations
	for name, tc := range ExecuteCases {
		for _, pm := range tc.parameters {
			for _, strlen := range tc.stringLength {
				for _, reqPerBlock := range []int{1, 5, 10, 20} {

					// Build a sub-benchmark name
					subBenchName := fmt.Sprintf(
						"%s (param: %d, strlen: %d) - %d requests/block",
						name, pm, strlen, reqPerBlock,
					)

					b.Run(subBenchName, func(subB *testing.B) {
						// Track the minimum gas usage and min iteration time
						var minGasUsed uint64 = ^uint64(0) // 0xFFFFFFFF... "max" for comparison
						var minNs int64 = (1 << 63) - 1    // = math.MaxInt64

						for i := 0; i < subB.N; i++ {
							// Start fresh for each iteration
							subB.ResetTimer()
							subB.StopTimer()

							ba := InitializeBenchmarkApp(subB, BlockMaxGas)

							// Generate requests
							txs := GenSequenceOfTxs(
								ba.TxEncoder,
								ba.TxConfig,
								GenMsgRequestData(
									ba.Sender,
									ba.Oid,
									ba.Did,
									tc.scenario,
									pm,
									strlen,
									10000,
									ExecuteGasLimit,
								),
								ba.Sender,
								reqPerBlock,
							)

							// Start measuring the "requests" finalization
							subB.StartTimer()
							res, err := ba.FinalizeBlock(&abci.RequestFinalizeBlock{
								Txs:    txs,
								Height: ba.LastBlockHeight() + 1,
								Time:   ba.Ctx.BlockTime(),
							})
							subB.StopTimer()
							require.NoError(subB, err)

							// Sum the gas usage for this iteration’s "requests" block
							var iterationGas uint64
							for _, tx := range res.TxResults {
								iterationGas += uint64(tx.GasUsed)
								require.Equal(subB, uint32(0), tx.Code, "Deliver Error: %s", tx.Log)
							}

							_, err = ba.Commit()
							require.NoError(subB, err)

							// Gather pending requests
							pendingRequests := ba.GetAllPendingRequests(ba.Validator)

							// Create msg report data
							tx := GenSequenceOfTxs(
								ba.TxEncoder,
								ba.TxConfig,
								ba.GenMsgReportData(ba.Validator, pendingRequests.RequestIDs),
								ba.Validator,
								1,
							)[0]

							// Measure "report data" finalization
							subB.StartTimer()
							res, err = ba.FinalizeBlock(&abci.RequestFinalizeBlock{
								Txs:    [][]byte{tx},
								Height: ba.LastBlockHeight() + 1,
								Time:   ba.Ctx.BlockTime(),
							})
							subB.StopTimer()
							require.NoError(subB, err)

							for _, tx := range res.TxResults {
								iterationGas += uint64(tx.GasUsed)
								require.Equal(subB, uint32(0), tx.Code, "Deliver Error: %s", tx.Log)
							}

							// If this iteration's total gas is smaller, store it
							if iterationGas < minGasUsed {
								minGasUsed = iterationGas
							}

							// Check final time
							iterationNs := int64(subB.Elapsed())
							if iterationNs < minNs {
								minNs = iterationNs
							}
						}

						// Append the min results for this sub-benchmark
						allResults = append(allResults, benchRecord{
							Scenario:     name,
							Param:        pm,
							StringLength: strlen,
							ReqPerBlock:  reqPerBlock,
							MinGasUsed:   minGasUsed,
							B_N:          subB.N,
							MinNsPerOp:   minNs - 1700000,
						})
					})
				}
			}
		}
	}
}
