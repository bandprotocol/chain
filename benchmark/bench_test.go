package benchmark

import (
	"fmt"
	"math"
	"testing"
	"time"

	oraclekeeper "github.com/bandprotocol/chain/v2/x/oracle/keeper"
	oracletypes "github.com/bandprotocol/chain/v2/x/oracle/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
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

var PrepareGasLimit uint64 = 7_500_000
var ExecuteGasLimit uint64 = 7_500_000
var BlockMaxGas int64 = 50_000_000
var GasRanges []int = []int{1, 1_000, 10_000, 100_000, 1_000_000, 7_900_000}
var NumRequestRanges []int = []int{0, 1, 5, 10, 20}

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

// benchmark test for delivering MsgRequestData
func BenchmarkRequestDataDeliver(b *testing.B) {
	for name, tc := range PrepareCases {
		for _, pm := range tc.parameters {
			for _, strlen := range tc.stringLength {
				b.Run(
					fmt.Sprintf(
						"%s (param: %d, strlen: %d)",
						name,
						pm,
						strlen,
					),
					func(b *testing.B) {
						ba := InitializeBenchmarkApp(b, -1)

						txs := GenSequenceOfTxs(
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
							b.N,
						)

						ba.CallBeginBlock()
						b.ResetTimer()
						b.StopTimer()

						// deliver MsgRequestData to the block
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
					},
				)
			}
		}
	}
}

// benchmark test for processing oracle scripts at endblock
func BenchmarkRequestDataEndBlock(b *testing.B) {
	for name, tc := range ExecuteCases {
		for _, pm := range tc.parameters {
			for _, strlen := range tc.stringLength {
				for _, nr := range []int{1, 5, 10, 20} {
					b.Run(
						fmt.Sprintf(
							"%s (param: %d, strlen: %d) - %d requests/block",
							name,
							pm,
							strlen,
							nr,
						),
						func(b *testing.B) {
							ba := InitializeBenchmarkApp(b, -1)

							txs := GenSequenceOfTxs(
								ba.TxConfig,
								GenMsgRequestData(
									ba.Sender,
									ba.Oid,
									ba.Did,
									tc.scenario,
									uint64(pm),
									strlen,
									10000,
									ExecuteGasLimit,
								),
								ba.Sender,
								b.N*nr,
							)

							b.ResetTimer()
							b.StopTimer()

							for i := 0; i < b.N; i++ {
								// deliver MsgRequestData to the first block
								ba.CallBeginBlock()

								for idx := 0; idx < nr; idx++ {
									_, _, err := ba.CallDeliver(txs[i*nr+idx])
									if i == 0 && idx == 0 && err != nil {
										fmt.Println("\tDeliver error:", err.Error())
									}
								}

								ba.CallEndBlock()
								ba.Commit()

								// deliver MsgReportData to the second block
								ba.CallBeginBlock()
								ba.SendAllPendingReports(ba.Validator)

								// process endblock
								b.StartTimer()
								ba.CallEndBlock()
								b.StopTimer()

								ba.Commit()
							}
						},
					)
				}
			}
		}
	}
}

func BenchmarkBlock(b *testing.B) {
	benchmarkBlockNormalMsg(b)
	benchmarkBlockReportMsg(b)
}

func benchmarkBlockNormalMsg(b *testing.B) {
	tmpApp := InitializeBenchmarkApp(b, BlockMaxGas)

	type caseType struct {
		name string
		msg  []sdk.Msg
	}

	// construct normal msg e.g. MsgSend of bank module
	cases := make([]caseType, 0)
	cases = append(cases, caseType{
		name: "bank_msg_send",
		msg: GenMsgSend(
			tmpApp.Sender,
			tmpApp.Validator,
		),
	})

	// add MsgRequestData of oracle for each parameter into test cases
	for name, tc := range PrepareCases {
		for _, prepareGas := range GasRanges {
			cases = append(cases, caseType{
				name: fmt.Sprintf(
					"oracle_msg_request_data - %s - %d prepare gas",
					name,
					prepareGas),
				msg: GenMsgRequestData(
					tmpApp.Sender,
					tmpApp.Oid,
					tmpApp.Did,
					tc.scenario,
					math.MaxUint64,
					1,
					uint64(prepareGas),
					1000,
				),
			})
		}
	}

	// use each msg to test full blocks
	for _, c := range cases {
		b.Run(c.name,
			func(b *testing.B) {
				b.ResetTimer()
				b.StopTimer()

				for i := 0; i < b.N; i++ {
					ba := InitializeBenchmarkApp(b, BlockMaxGas)

					b.StartTimer()
					ba.CallBeginBlock()
					b.StopTimer()

					var totalGas uint64 = 0
					for {
						tx := GenSequenceOfTxs(
							ba.TxConfig,
							c.msg,
							ba.Sender,
							1,
						)[0]

						b.StartTimer()
						gas, _, _ := ba.CallDeliver(tx)
						b.StopTimer()

						totalGas += gas.GasUsed
						if totalGas+gas.GasUsed >= uint64(BlockMaxGas) {
							break
						}
					}

					b.StartTimer()
					ba.CallEndBlock()
					ba.Commit()
					b.StopTimer()
				}
			},
		)
	}
}

func benchmarkBlockReportMsg(b *testing.B) {
	for name, tc := range ExecuteCases {
		for _, executeGas := range GasRanges {
			// reportSize is the number of MsgReportData in one tx
			// 1 means send one report per tx
			// Note: 1000 is the maximum number of MsgReportData in one tx that doesn't exceed MaxGas of block (50M)
			for _, reportSize := range []int{1, 100, 1000} {
				b.Run(
					fmt.Sprintf(
						"oracle_msg_report_data - %s - %d execute gas - %d report sizes",
						name,
						executeGas,
						reportSize,
					),
					func(b *testing.B) {
						b.ResetTimer()
						b.StopTimer()

						for i := 0; i < b.N; i++ {
							ba := InitializeBenchmarkApp(b, BlockMaxGas)
							ba.AddMaxMsgRequests(GenMsgRequestData(
								ba.Sender,
								ba.Oid,
								ba.Did,
								tc.scenario,
								math.MaxUint64,
								1,
								1000,
								uint64(executeGas),
							))

							b.StartTimer()
							ba.CallBeginBlock()
							b.StopTimer()

							res := ba.GetAllPendingRequests(ba.Validator)
							var totalGas uint64 = 0

							reqChunks := ChunkSlice(res.RequestIDs, reportSize)
							for _, reqChunk := range reqChunks {
								tx := GenSequenceOfTxs(
									ba.TxConfig,
									ba.GenMsgReportData(ba.Validator, reqChunk),
									ba.Validator,
									1,
								)[0]

								b.StartTimer()
								gas, _, err := ba.CallDeliver(tx)
								b.StopTimer()

								require.NoError(b, err)

								totalGas += gas.GasUsed
								// add 10% more because it will use more gas next time
								if totalGas+(gas.GasUsed*110/100) >= uint64(BlockMaxGas) {
									break
								}
							}

							b.StartTimer()
							ba.CallEndBlock()
							ba.Commit()
							b.StopTimer()
						}
					},
				)
			}
		}
	}
}
