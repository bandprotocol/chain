package benchmark

import (
	"fmt"
	"math"
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
	for cache, cacheSize := range CacheCases {
		for name, tc := range ExecuteCases {
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

						// call execute on new env
						for i := 0; i < b.N; i++ {
							env := oracletypes.NewExecuteEnv(
								req,
								GenOracleReports(),
								time.Now(),
								int64(GetSpanSize()),
							)

							b.StartTimer()
							res, err := owasmVM.Execute(
								compiledCode,
								oraclekeeper.ConvertToOwasmGas(ExecuteGasLimit),
								env,
							)
							b.StopTimer()
							if i == 0 {
								if err != nil {
									fmt.Println("\tEndblock Error:", err.Error())
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
										Time:   time.Now(),
									},
								)
								require.NoError(b, err)

								b.StopTimer()

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
	for name, tc := range ExecuteCases {
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
								ba := InitializeBenchmarkApp(b, BlockMaxGas)

								// add request
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

								_, err := ba.FinalizeBlock(
									&abci.RequestFinalizeBlock{
										Txs:    txs,
										Height: ba.LastBlockHeight() + 1,
										Time:   time.Now(),
									},
								)
								require.NoError(b, err)

								_, err = ba.Commit()
								require.NoError(b, err)

								// get pending requests
								pendingRequests := ba.GetAllPendingRequests(ba.Validator)

								// create msg report data
								tx := GenSequenceOfTxs(
									ba.TxEncoder,
									ba.TxConfig,
									ba.GenMsgReportData(ba.Validator, pendingRequests.RequestIDs),
									ba.Validator,
									1,
								)[0]

								b.StartTimer()

								res, err := ba.FinalizeBlock(
									&abci.RequestFinalizeBlock{
										Txs:    [][]byte{tx},
										Height: ba.LastBlockHeight() + 1,
										Time:   time.Now(),
									},
								)
								require.NoError(b, err)

								b.StopTimer()

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
