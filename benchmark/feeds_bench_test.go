package benchmark

import (
	"encoding/json"
	"fmt"
	"math/rand"
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
	var allResults []struct {
		Name           string `json:"sub_bench_name"`
		Vals           int    `json:"vals"`
		Feeds          uint64 `json:"feeds"`
		Prices         int    `json:"prices"`
		GasUsedFirstTx uint64 `json:"gas_used_first_tx"`
		N              int    `json:"b_n"`
		NsPerOp        int64  `json:"ns_per_op"`
	}

	// Schedule a one-time cleanup that runs after ALL sub-benchmarks finish
	b.Cleanup(func() {
		// Convert allResults to JSON
		data, _ := json.MarshalIndent(allResults, "", "  ")
		fmt.Println(string(data))
	})

	numValsList := []int{1, 10, 50, 90}
	numFeedsList := []uint64{1, 10, 100, 300, 1000}
	numPricesList := []int{1, 10, 100, 300, 1000}

	for _, numVals := range numValsList {
		for _, numFeeds := range numFeedsList {
			for _, numPrices := range numPricesList {
				// Skip invalid cases where numPrices > numFeeds
				if uint64(numPrices) > numFeeds {
					continue
				}

				// Name the sub-benchmark to reflect the current configuration
				name := fmt.Sprintf("Vals_%d_Feeds_%d_Prices_%d", numVals, numFeeds, numPrices)

				b.Run(name, func(subB *testing.B) {
					var gasUsedFirstTx uint64

					for i := 0; i < subB.N; i++ {
						// Stop timer during setup so it doesn't affect ns/op measurement.
						subB.StopTimer()

						ba := InitializeBenchmarkApp(subB, -1)

						vals, err := generateValidators(ba, numVals)
						require.NoError(subB, err)

						err = setupFeeds(ba, numFeeds)
						require.NoError(subB, err)

						err = setupValidatorPriceList(ba, vals)
						require.NoError(subB, err)

						// Gather feeds and pick numPrices
						allFeeds := ba.FeedsKeeper.GetCurrentFeeds(ba.Ctx).Feeds
						selectedFeeds := allFeeds[:numPrices]

						// Create transactions
						txs := make([][]byte, 0, len(vals))
						for _, val := range vals {
							msg := GenMsgSubmitSignalPrices(val, selectedFeeds, ba.Ctx.BlockTime().Unix())
							tx := GenSequenceOfTxs(
								ba.TxEncoder,
								ba.TxConfig,
								msg,
								val,
								1, // sequence
							)[0]
							txs = append(txs, tx)
						}

						// Start timer for the core operation
						subB.StartTimer()

						// Finalize block
						res, err := ba.FinalizeBlock(&abci.RequestFinalizeBlock{
							Txs:    txs,
							Height: ba.LastBlockHeight() + 1,
							Time:   ba.Ctx.BlockTime(),
						})
						subB.StopTimer()

						require.NoError(subB, err)
						for _, txRes := range res.TxResults {
							require.Equal(subB, uint32(0), txRes.Code,
								"Tx failed for %s; Log: %s", name, txRes.Log)
						}

						_, err = ba.Commit()
						require.NoError(subB, err)

						// Capture gas usage for the *first iteration only*
						if i == 0 && len(res.TxResults) > 0 {
							gasUsedFirstTx = uint64(res.TxResults[0].GasUsed)
							// Optionally print it:
							// fmt.Printf("[%s] - Gas used first tx: %d\n", name, gasUsedFirstTx)
						}
					}

					// Manually compute total elapsed time and approximate NsPerOp
					// totalElapsedNs := time.Since(startTime).Nanoseconds()
					// nsPerOp := totalElapsedNs / int64(subB.N)

					// Add a record to allResults
					allResults = append(allResults, struct {
						Name           string `json:"sub_bench_name"`
						Vals           int    `json:"vals"`
						Feeds          uint64 `json:"feeds"`
						Prices         int    `json:"prices"`
						GasUsedFirstTx uint64 `json:"gas_used_first_tx"`
						N              int    `json:"b_n"`
						NsPerOp        int64  `json:"ns_per_op"`
					}{
						Name:           name,
						Vals:           numVals,
						Feeds:          numFeeds,
						Prices:         numPrices,
						GasUsedFirstTx: gasUsedFirstTx,
						N:              subB.N,
						NsPerOp:        int64(subB.Elapsed()) / int64(subB.N),
					})
				})
			}
		}
	}
}

// benchmark test for endblock of feeds module
func BenchmarkFeedsEndBlock(b *testing.B) {
	ba := InitializeBenchmarkApp(b, -1)

	params, err := ba.StakingKeeper.GetParams(ba.Ctx)
	require.NoError(b, err)

	numVals := params.MaxValidators

	vals, err := generateValidators(ba, int(numVals))
	require.NoError(b, err)

	err = setupFeeds(ba, 300)
	require.NoError(b, err)

	err = setupValidatorPrices(ba, vals)
	require.NoError(b, err)

	b.ResetTimer()
	b.StopTimer()

	// benchmark endblock
	for i := 0; i < b.N; i++ {
		// process endblock
		b.StartTimer()
		res, err := ba.FinalizeBlock(
			&abci.RequestFinalizeBlock{
				Height: ba.LastBlockHeight() + 1,
				Time:   ba.Ctx.BlockTime(),
			},
		)
		b.StopTimer()

		require.NoError(b, err)
		for _, tx := range res.TxResults {
			require.Equal(b, uint32(0), tx.Code)
		}

		_, err = ba.Commit()
		require.NoError(b, err)
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
