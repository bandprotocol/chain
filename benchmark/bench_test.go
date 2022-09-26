package benchmark

import (
	"fmt"
	"testing"

	"time"

	oraclekeeper "github.com/bandprotocol/chain/v2/x/oracle/keeper"
	oracletypes "github.com/bandprotocol/chain/v2/x/oracle/types"
	"github.com/stretchr/testify/require"
	// "github.com/stretchr/testify/require"
)

var PrepareCases = map[string]struct {
	scenario     uint64
	parameters   []uint64
	stringLength []int
}{
	"ask_external_data": {
		scenario:     1,
		parameters:   []uint64{1, 4, 8, 16},
		stringLength: []int{1, 10, 100, 200},
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
	"arithmatic_ops": {
		scenario:     102,
		parameters:   []uint64{1, 10, 1000, 10000, 100000},
		stringLength: []int{1},
	},
	"allocate_mem": {
		scenario:     103,
		parameters:   []uint64{1, 10, 1000, 10000, 100000},
		stringLength: []int{1},
	},
	"find_median": {
		scenario:     104,
		parameters:   []uint64{1, 10, 1000, 10000, 100000},
		stringLength: []int{1},
	},
	"finite_loop": {
		scenario:     105,
		parameters:   []uint64{1, 10, 1000, 10000, 100000},
		stringLength: []int{1},
	},
	"set_local_var": {
		scenario:     106,
		parameters:   []uint64{1, 10, 1000, 10000, 100000},
		stringLength: []int{1},
	},
	"get_ask_count": {
		scenario:     201,
		parameters:   []uint64{1, 10, 1000, 10000, 100000},
		stringLength: []int{1},
	},
	"get_min_count": {
		scenario:     202,
		parameters:   []uint64{1, 10, 1000, 10000, 100000},
		stringLength: []int{1},
	},
	"get_prepare_time": {
		scenario:     203,
		parameters:   []uint64{1, 10, 1000, 10000, 100000},
		stringLength: []int{1},
	},
	"get_execute_time": {
		scenario:     204,
		parameters:   []uint64{1, 10, 1000, 10000, 100000},
		stringLength: []int{1},
	},
	"get_ans_count": {
		scenario:     205,
		parameters:   []uint64{1, 10, 1000, 10000, 100000},
		stringLength: []int{1},
	},
	"get_calldata": {
		scenario:     206,
		parameters:   []uint64{1, 10, 1000, 10000, 100000},
		stringLength: []int{1, 10, 100, 200, 400},
	},
	"save_return_data": {
		scenario:     207,
		parameters:   []uint64{1, 10, 1000, 10000, 100000},
		stringLength: []int{1, 10, 100, 200, 400},
	},
	"get_external_data": {
		scenario:     208,
		parameters:   []uint64{1, 10, 1000, 10000, 100000},
		stringLength: []int{1, 10, 100, 200, 400},
	},
	"ecvrf_verify": {
		scenario:     209,
		parameters:   []uint64{1, 10, 1000, 10000, 100000},
		stringLength: []int{1},
	},
	"base_import": {
		scenario:     210,
		parameters:   []uint64{1},
		stringLength: []int{1},
	},
}

var CacheCases = map[string]uint32{
	"no_cache": 0,
	"cache":    1,
}

var PrepareGasLimit uint64 = 4_000_000
var ExecuteGasLimit uint64 = 4_000_000

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
								int64(oracletypes.DefaultMaxCalldataSize),
							)
							b.StartTimer()
							res, err := owasmVM.Prepare(
								compiledCode,
								oraclekeeper.ConvertToOwasmGas(PrepareGasLimit),
								env,
							)
							b.StopTimer()
							if i == b.N-1 {
								if err != nil {
									fmt.Println(err.Error())
									break
								} else {
									fmt.Println("	Owasm Gas used:", res.GasUsed)
								}
							}
						}
					})
				}
			}
		}
	}
}

func generateReports() []oracletypes.Report {
	return []oracletypes.Report{
		{
			Validator:       "",
			InBeforeResolve: true,
			RawReports: []oracletypes.RawReport{
				{
					ExternalID: 1,
					ExitCode:   0,
					Data:       []byte{},
				},
			},
		},
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
								generateReports(),
								time.Now(),
								int64(oracletypes.DefaultMaxReportDataSize),
							)

							b.StartTimer()
							res, err := owasmVM.Execute(
								compiledCode,
								oraclekeeper.ConvertToOwasmGas(ExecuteGasLimit),
								env,
							)
							b.StopTimer()
							if i == b.N-1 {
								if err != nil {
									fmt.Println(err.Error())
									break
								} else {
									fmt.Println("	Owasm Gas used:", res.GasUsed)
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
						ba := InitializeBenchmarkApp(b)

						txs := GenSequenceOfTxs(
							ba.TxConfig,
							GenMsgRequestData(ba.Sender, ba.Oid, ba.Did, tc.scenario, pm, strlen),
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
							require.NoError(b, err)
							b.StopTimer()
							if i == b.N-1 {
								if err != nil {
									fmt.Println(err.Error())
									break
								} else {
									fmt.Println("	Cosmos Gas used:", gasInfo.GasUsed)
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
				for _, nr := range []int{0, 1, 5, 10, 20} {
					b.Run(
						fmt.Sprintf(
							"%s (param: %d, strlen: %d) - %d requests/block",
							name,
							pm,
							strlen,
							nr,
						),
						func(b *testing.B) {
							ba := InitializeBenchmarkApp(b)

							txs := GenSequenceOfTxs(
								ba.TxConfig,
								GenMsgRequestData(ba.Sender, ba.Oid, ba.Did, tc.scenario, pm, strlen),
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
									require.NoError(b, err)
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
