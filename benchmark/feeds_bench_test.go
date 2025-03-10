package benchmark

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"

	abci "github.com/cometbft/cometbft/abci/types"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	bandtesting "github.com/bandprotocol/chain/v3/testing"
	"github.com/bandprotocol/chain/v3/x/feeds/types"
	oracletypes "github.com/bandprotocol/chain/v3/x/oracle/types"
)

var (
	ValidValidator  = sdk.ValAddress("1000000001")
	ValidValidator2 = sdk.ValAddress("1000000002")
)

func BenchmarkSortMap(b *testing.B) {
	b.ResetTimer()
	b.StopTimer()
	ba := InitializeBenchmarkApp(b, -1)
	expValPrices := generateValidatorPrices(300, ba.Ctx.BlockTime().Unix())
	valPricesMap := make(map[string]types.ValidatorPrice)
	for _, valPrice := range expValPrices {
		valPricesMap[valPrice.SignalID] = valPrice
	}
	for i := 0; i < b.N; i++ {
		for j := 0; j < 2000; j++ {
			b.StartTimer()
			keys := make([]string, len(valPricesMap))
			k := int(0)
			for key := range valPricesMap {
				keys[k] = key
				k++
			}
			sort.Strings(keys)
			valPrices := make([]types.ValidatorPrice, len(valPricesMap))
			for idx, key := range keys {
				valPrices[idx] = valPricesMap[key]
			}
			b.StopTimer()
		}
	}
}

func BenchmarkSubmitSignalPricesDeliver(b *testing.B) {
	// We'll accumulate results for all sub-benchmarks here.
	type benchRecord struct {
		Name       string `json:"sub_bench_name"`
		Vals       int    `json:"vals"`
		Feeds      uint64 `json:"feeds"`
		Prices     int    `json:"prices"`
		MinGasUsed uint64 `json:"min_gas_used"`
		B_N        int    `json:"b_n"`
		MinNsPerOp int64  `json:"min_ns_per_op"`
	}

	var allResults []benchRecord

	// Once all sub-benchmarks finish, print & save the JSON.
	b.Cleanup(func() {
		data, _ := json.MarshalIndent(allResults, "", "  ")
		fmt.Println(string(data))

		err := os.WriteFile("feeds_submit_bench.json", data, 0o644)
		if err != nil {
			b.Logf("Error writing feeds_submit_bench.json: %v", err)
		} else {
			b.Logf("Wrote %d benchmark results to feeds_submit_bench.json", len(allResults))
		}
	})

	numValsList := []int{1, 10, 50, 70, 90}
	numFeedsList := []uint64{1, 10, 100, 300, 1000}
	numPricesList := []int{1, 10, 100, 300, 1000}

	for _, numVals := range numValsList {
		for _, numFeeds := range numFeedsList {
			for _, numPrices := range numPricesList {
				// Skip invalid cases where numPrices > numFeeds
				if uint64(numPrices) > numFeeds {
					continue
				}

				// Name each sub-benchmark
				name := fmt.Sprintf("Vals_%d_Feeds_%d_Prices_%d", numVals, numFeeds, numPrices)

				b.Run(name, func(subB *testing.B) {
					// Track MIN gas usage and MIN iteration time (ns) among all subB.N runs
					// Initialize them to "largest possible" so the first real measurement is kept.
					var minGasUsed uint64 = ^uint64(0) // 0xFFFF... to track minimum
					var minNs int64 = (1 << 63) - 1    // = math.MaxInt64

					for i := 0; i < subB.N; i++ {
						// This zeroes out the sub-benchmark's timing stats, so each sub-benchmark
						// measures only its own loops.
						subB.ResetTimer()
						// StopTimer while we do heavy setup (not counted in iteration time).
						subB.StopTimer()

						// 1) Re-initialize everything
						ba := InitializeBenchmarkApp(subB, -1)

						vals, err := generateValidators(ba, numVals)
						require.NoError(subB, err)

						err = setupFeeds(ba, numFeeds)
						require.NoError(subB, err)

						err = setupValidatorPriceList(ba, vals)
						require.NoError(subB, err)

						// Gather feeds & pick numPrices
						allFeeds := ba.FeedsKeeper.GetCurrentFeeds(ba.Ctx).Feeds
						selectedFeeds := allFeeds[:numPrices]

						// Build TXs
						txs := make([][]byte, 0, len(vals))
						for _, val := range vals {
							msg := GenMsgSubmitSignalPrices(val, selectedFeeds, ba.Ctx.BlockTime().Unix())
							tx := GenSequenceOfTxs(
								ba.TxEncoder,
								ba.TxConfig,
								msg,
								val,
								1,
							)[0]
							txs = append(txs, tx)
						}

						// 2) StartTimer for the core operation
						subB.StartTimer()

						// Finalize block with those TXs
						res, err := ba.FinalizeBlock(&abci.RequestFinalizeBlock{
							Txs:    txs,
							Height: ba.LastBlockHeight() + 1,
							Time:   ba.Ctx.BlockTime(),
						})

						// Stop the timer to exclude any post-op checks from iteration time
						subB.StopTimer()

						require.NoError(subB, err)
						for _, txRes := range res.TxResults {
							require.Equal(subB, uint32(0), txRes.Code,
								"Tx failed for %s; Log: %s", name, txRes.Log)
						}

						_, err = ba.Commit()
						require.NoError(subB, err)

						// if this iteration took less time, keep it
						if int64(subB.Elapsed()) < minNs {
							minNs = int64(subB.Elapsed())
						}

						// Sum the gas for this iteration
						var iterationGas uint64
						for _, txRes := range res.TxResults {
							iterationGas += uint64(txRes.GasUsed)
						}
						// If this iteration used less gas, keep it
						if iterationGas < minGasUsed {
							minGasUsed = iterationGas
						}
					}

					// After subB.N iterations, record the sub-benchmark's minimums
					allResults = append(allResults, benchRecord{
						Name:       name,
						Vals:       numVals,
						Feeds:      numFeeds,
						Prices:     numPrices,
						MinGasUsed: minGasUsed,
						B_N:        subB.N,
						MinNsPerOp: minNs - 2000000,
					})
				})
			}
		}
	}
}

// benchmark test for endblock of feeds module
func BenchmarkFeedsEndBlock(b *testing.B) {
	// We'll collect results for all sub-benchmarks in this slice.
	type benchRecord struct {
		Name       string `json:"sub_bench_name"`
		Vals       int    `json:"vals"`
		Feeds      uint64 `json:"feeds"`
		B_N        int    `json:"b_n"`
		MinNsPerOp int64  `json:"min_ns_per_op"`
	}

	var allResults []benchRecord

	// We'll write JSON after all sub-benchmarks finish.
	b.Cleanup(func() {
		data, _ := json.MarshalIndent(allResults, "", "  ")
		// Print to stdout
		fmt.Println(string(data))

		// Also write to a file, e.g. feeds_endblock_bench.json
		err := os.WriteFile("feeds_endblock_bench.json", data, 0o644)
		if err != nil {
			b.Logf("Error writing feeds_endblock_bench.json: %v", err)
		} else {
			b.Logf("Wrote %d benchmark results to feeds_endblock_bench.json", len(allResults))
		}
	})

	// Define the sets of (numVals, numFeeds) we want to benchmark.
	numValsList := []int{1, 10, 50, 70, 90}
	numFeedsList := []uint64{1, 10, 100, 300, 1000}

	// For each combo of numVals and numFeeds, run a sub-benchmark
	for _, valsCount := range numValsList {
		for _, feedsCount := range numFeedsList {

			// Make a name that indicates the parameters
			subBenchName := fmt.Sprintf("Vals_%d_Feeds_%d", valsCount, feedsCount)

			b.Run(subBenchName, func(subB *testing.B) {
				// Track the MIN gas usage & MIN iteration time across subB.N runs
				var minNs int64 = (1 << 63) - 1 // = math.MaxInt64

				for i := 0; i < subB.N; i++ {
					// Each iteration we do:
					// 1) subB.ResetTimer() to ensure we only measure *this iteration*
					// 2) subB.StopTimer() while doing heavy setup
					// 3) subB.StartTimer() just for the block finalization (the end-block logic)
					// 4) subB.StopTimer() again afterward, so we can measure iteration time

					// Step 1: Reset the timer stats for each iteration
					subB.ResetTimer()
					subB.StopTimer()

					// --- Heavy Setup ---
					ba := InitializeBenchmarkApp(subB, -1)

					// Generate validators
					vals, err := generateValidators(ba, valsCount)
					require.NoError(subB, err)

					// Setup feeds
					err = setupFeeds(ba, feedsCount)
					require.NoError(subB, err)

					// Setup validator prices
					err = setupValidatorPrices(ba, vals)
					require.NoError(subB, err)

					// Step 2: Start measuring the end-block
					subB.StartTimer()

					// Finalize block (triggers end-block logic)
					res, err := ba.FinalizeBlock(
						&abci.RequestFinalizeBlock{
							Height: ba.LastBlockHeight() + 1,
							Time:   ba.Ctx.BlockTime(),
						},
					)

					subB.StopTimer()

					require.NoError(subB, err)

					// If your end-block has any TxResults (unusual), sum their gas
					for _, txr := range res.TxResults {
						require.Equal(subB, uint32(0), txr.Code)
					}

					// Commit the block
					_, err = ba.Commit()
					require.NoError(subB, err)

					// If this iteration took less time, keep it
					iterationTimeNs := int64(subB.Elapsed())
					if iterationTimeNs < minNs {
						minNs = iterationTimeNs
					}
				}

				// Build a record with the “best” iteration metrics
				allResults = append(allResults, benchRecord{
					Name:       subBenchName,
					Vals:       valsCount,
					Feeds:      feedsCount,
					B_N:        subB.N,
					MinNsPerOp: minNs - 2000000,
				})
			})
		}
	}
}

func setupFeeds(ba *BenchmarkApp, numFeeds uint64) error {
	feeds := []types.Feed{}
	for i := uint64(0); i < numFeeds; i++ {
		feeds = append(feeds, types.Feed{
			SignalID: fmt.Sprintf("signal.%d", i),
			Interval: 60,
		})
	}
	ba.FeedsKeeper.SetCurrentFeeds(ba.Ctx, feeds)

	return nil
}

func setupValidatorPriceList(ba *BenchmarkApp, vals []*Account) error {
	sfs := ba.FeedsKeeper.GetCurrentFeeds(ba.Ctx)

	for valIdx, val := range vals {
		valPrices := []types.ValidatorPrice{}
		for _, feed := range sfs.Feeds {
			valPrices = append(valPrices, types.ValidatorPrice{
				SignalPriceStatus: types.SIGNAL_PRICE_STATUS_AVAILABLE,
				SignalID:          feed.SignalID,
				Price:             (10000 + uint64(valIdx)) * 10e9,
				Timestamp:         ba.Ctx.BlockTime().Unix() - 40,
			})
		}
		err := ba.FeedsKeeper.SetValidatorPriceList(ba.Ctx, val.ValAddress, valPrices)
		if err != nil {
			return err
		}
	}

	return nil
}

func setupValidatorPrices(ba *BenchmarkApp, vals []*Account) error {
	sfs := ba.FeedsKeeper.GetCurrentFeeds(ba.Ctx)

	for valIdx, val := range vals {
		valPrices := []types.ValidatorPrice{}
		for _, feed := range sfs.Feeds {
			valPrices = append(valPrices, types.ValidatorPrice{
				SignalPriceStatus: types.SIGNAL_PRICE_STATUS_AVAILABLE,
				SignalID:          feed.SignalID,
				Price:             (10000 + uint64(valIdx)) * 10e9,
				Timestamp:         ba.Ctx.BlockTime().Unix(),
			})
		}

		err := ba.FeedsKeeper.SetValidatorPriceList(ba.Ctx, val.ValAddress, valPrices)
		if err != nil {
			return err
		}
	}

	return nil
}

func generateValidators(ba *BenchmarkApp, num int) ([]*Account, error) {
	// transfer money
	vals := []bandtesting.Account{}
	txs := [][]byte{}
	for i := 0; i < num; i++ {
		r := rand.New(rand.NewSource(int64(i)))
		acc := bandtesting.CreateArbitraryAccount(r)
		vals = append(vals, acc)

		tx := GenSequenceOfTxs(
			ba.TxEncoder,
			ba.TxConfig,
			[]sdk.Msg{banktypes.NewMsgSend(ba.Sender.Address, acc.Address, bandtesting.Coins100band)},
			ba.Sender,
			1,
		)[0]

		txs = append(txs, tx)
	}

	res, err := ba.FinalizeBlock(
		&abci.RequestFinalizeBlock{
			Txs:    txs,
			Height: ba.LastBlockHeight() + 1,
			Time:   ba.Ctx.BlockTime(),
		},
	)
	if err != nil {
		return nil, err
	}

	for _, tx := range res.TxResults {
		if tx.Code != 0 {
			return nil, fmt.Errorf("transfer error: %s", tx.Log)
		}
	}

	_, err = ba.Commit()
	if err != nil {
		return nil, err
	}

	// apply to be a validator
	accs := []*Account{}
	txs = [][]byte{}
	for _, val := range vals {
		info := ba.AccountKeeper.GetAccount(ba.Ctx, val.Address)
		acc := &Account{
			Account: val,
			Num:     info.GetAccountNumber(),
			Seq:     info.GetSequence(),
		}
		accs = append(accs, acc)

		msgCreateVal, err := stakingtypes.NewMsgCreateValidator(
			val.ValAddress.String(),
			val.PubKey,
			sdk.NewCoin("uband", sdkmath.NewInt(1000000)),
			stakingtypes.NewDescription(val.Address.String(), val.Address.String(), "", "", ""),
			stakingtypes.NewCommissionRates(sdkmath.LegacyNewDec(1), sdkmath.LegacyNewDec(1), sdkmath.LegacyNewDec(1)),
			sdkmath.NewInt(1),
		)
		if err != nil {
			return nil, err
		}

		msgActivate := oracletypes.NewMsgActivate(val.ValAddress)

		tx := GenSequenceOfTxs(
			ba.TxEncoder,
			ba.TxConfig,
			[]sdk.Msg{msgCreateVal, msgActivate},
			acc,
			1,
		)[0]

		txs = append(txs, tx)
	}

	res, err = ba.FinalizeBlock(
		&abci.RequestFinalizeBlock{
			Txs:    txs,
			Height: ba.LastBlockHeight() + 1,
			Time:   ba.Ctx.BlockTime(),
		},
	)
	if err != nil {
		return nil, err
	}

	for _, tx := range res.TxResults {
		if tx.Code != 0 {
			return nil, fmt.Errorf("validator error: %s", tx.Log)
		}
	}

	_, err = ba.Commit()
	if err != nil {
		return nil, err
	}

	return accs, nil
}

// generateValidatorPrices generates a slice of ValidatorPrice with the specified number of elements.
func generateValidatorPrices(numElements int, timestamp int64) []types.ValidatorPrice {
	prices := make([]types.ValidatorPrice, numElements)

	for i := 0; i < numElements; i++ {
		prices[i] = types.ValidatorPrice{
			SignalID:  fmt.Sprintf("CS:BAND%d-USD", i),
			Price:     1e10,
			Timestamp: timestamp,
		}
	}
	return prices
}
