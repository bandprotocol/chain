package benchmark

import (
	"fmt"
	"testing"

	"time"

	oraclekeeper "github.com/bandprotocol/chain/v2/x/oracle/keeper"
	oracletypes "github.com/bandprotocol/chain/v2/x/oracle/types"
	"github.com/stretchr/testify/require"
)

var PrepareCases = map[string]struct {
	scenario   uint64
	parameters []uint64
}{
	"ask_external_data": {
		scenario:   1,
		parameters: []uint64{1, 4, 8, 16},
	},
}

var ExecuteCases = map[string]struct {
	scenario    uint64
	parameters  []uint64
	numRequests []int
}{
	"infinite_loop": {
		scenario:   101,
		parameters: []uint64{0},
	},
	// "arithmatic_ops": {
	// 	scenario:   102,
	// 	parameters: []uint64{1, 10, 1000, 10000, 100000},
	// },
	// "allocate_mem": {
	// 	scenario:   103,
	// 	parameters: []uint64{1, 10, 1000, 10000, 100000},
	// },
	// "find_median": {
	// 	scenario:   104,
	// 	parameters: []uint64{0, 1, 10, 50, 100},
	// },
}

var CacheCases = map[string]uint32{
	"no_cache": 0,
	"cache":    1,
}

var PrepareGasLimit uint64 = 4000000
var ExecuteGasLimit uint64 = 4000000

// benchmark test for prepare function of owasm vm
func BenchmarkOwasmVMPrepare(b *testing.B) {
	for cache, cacheSize := range CacheCases {
		for name, tc := range PrepareCases {
			for _, pm := range tc.parameters {
				b.Run(fmt.Sprintf(
					"%s - %s (param: %d)",
					cache,
					name,
					pm,
				), func(b *testing.B) {
					owasmVM, compiledCode, req := InitOwasmTestEnv(b, cacheSize, tc.scenario, pm)

					b.ResetTimer()
					b.StopTimer()

					// call prepare on new env
					for i := 0; i < b.N; i++ {
						env := oracletypes.NewPrepareEnv(
							req,
							int64(oracletypes.DefaultMaxCalldataSize),
							int64(oracletypes.DefaultMaxRawRequestCount),
						)
						b.StartTimer()
						_, _ = owasmVM.Prepare(
							compiledCode,
							oraclekeeper.ConvertToOwasmGas(PrepareGasLimit),
							int64(oracletypes.DefaultMaxCalldataSize),
							env,
						)
						b.StopTimer()
					}
				})
			}
		}
	}
}

// benchmark test for execute function of owasm vm
func BenchmarkOwasmVMExecute(b *testing.B) {
	for cache, cacheSize := range CacheCases {
		for name, tc := range ExecuteCases {
			for _, pm := range tc.parameters {
				b.Run(fmt.Sprintf(
					"%s - %s (param: %d)",
					cache,
					name,
					pm,
				), func(b *testing.B) {
					owasmVM, compiledCode, req := InitOwasmTestEnv(b, cacheSize, tc.scenario, pm)

					b.ResetTimer()
					b.StopTimer()

					// call execute on new env
					for i := 0; i < b.N; i++ {
						env := oracletypes.NewExecuteEnv(req, []oracletypes.Report{}, time.Now())

						b.StartTimer()
						_, _ = owasmVM.Execute(
							compiledCode,
							oraclekeeper.ConvertToOwasmGas(ExecuteGasLimit),
							int64(oracletypes.DefaultMaxCalldataSize),
							env,
						)
						b.StopTimer()
					}
				})
			}
		}
	}
}

// benchmark test for delivering MsgRequestData
func BenchmarkRequestDataDeliver(b *testing.B) {
	for name, tc := range PrepareCases {
		for _, pm := range tc.parameters {
			b.Run(
				fmt.Sprintf(
					"%s (param: %d)",
					name,
					pm,
				),
				func(b *testing.B) {
					ba := InitializeBenchmarkApp(b)

					txs := GenSequenceOfTxs(
						ba.TxConfig,
						GenMsgRequestData(ba.Sender, ba.Oid, ba.Did, tc.scenario, pm),
						ba.Sender,
						b.N,
					)

					ba.CallBeginBlock()
					b.ResetTimer()

					// deliver MsgRequestData to the block
					for i := 0; i < b.N; i++ {
						_, _, err := ba.CallDeliver(txs[i])
						require.NoError(b, err)
					}
					b.StopTimer()
				},
			)
		}
	}
}

// benchmark test for processing oracle scripts at endblock
func BenchmarkRequestDataEndBlock(b *testing.B) {
	for name, tc := range ExecuteCases {
		for _, pm := range tc.parameters {
			for _, nr := range []int{0, 1, 2, 3, 4} {
				b.Run(
					fmt.Sprintf(
						"%s (param: %d) - %d requests/block",
						name,
						pm,
						nr,
					),
					func(b *testing.B) {
						ba := InitializeBenchmarkApp(b)

						txs := GenSequenceOfTxs(
							ba.TxConfig,
							GenMsgRequestData(ba.Sender, ba.Oid, ba.Did, tc.scenario, pm),
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
