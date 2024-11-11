package benchmark

import (
	"fmt"
	"math/rand"
	"sort"
	"testing"
	"time"

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
	expValPrices := generateValidatorPrices(300, ValidValidator.String(), ba.Ctx.BlockTime().Unix())
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

// benchmark test for delivering MsgSubmitSignalPrices
func BenchmarkSubmitSignalPricesDeliver(b *testing.B) {
	b.ResetTimer()
	b.StopTimer()

	for i := 0; i < b.N; i++ {
		ba := InitializeBenchmarkApp(b, -1)

		params, err := ba.StakingKeeper.GetParams(ba.Ctx)
		require.NoError(b, err)

		numVals := params.MaxValidators

		vals, err := generateValidators(ba, int(numVals))
		require.NoError(b, err)

		err = setupFeeds(ba)
		require.NoError(b, err)

		err = setupValidatorPriceList(ba, vals)
		require.NoError(b, err)

		txs := [][]byte{}
		for _, val := range vals {
			tx := GenSequenceOfTxs(
				ba.TxEncoder,
				ba.TxConfig,
				GenMsgSubmitSignalPrices(
					val,
					ba.FeedsKeeper.GetCurrentFeeds(ba.Ctx).Feeds,
					ba.Ctx.BlockTime().Unix(),
				),
				val,
				1,
			)[0]

			txs = append(txs, tx)
		}

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
		_, err = ba.Commit()
		require.NoError(b, err)

		if i == 0 {
			if res.TxResults[len(res.TxResults)-1].Code != 0 {
				fmt.Println("\tDeliver Error:", res.TxResults[0].Log)
			} else {
				fmt.Println("\tCosmos Gas used:", res.TxResults[0].GasUsed)
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

	err = setupFeeds(ba)
	require.NoError(b, err)

	err = setupValidatorPrices(ba, vals)
	require.NoError(b, err)

	b.ResetTimer()
	b.StopTimer()

	// benchmark endblock
	for i := 0; i < b.N; i++ {
		// process endblock
		b.StartTimer()
		_, err := ba.FinalizeBlock(
			&abci.RequestFinalizeBlock{
				Height: ba.LastBlockHeight() + 1,
				Time:   time.Now(),
			},
		)
		require.NoError(b, err)
		b.StopTimer()

		_, err = ba.Commit()
		require.NoError(b, err)
	}
}

func setupFeeds(ba *BenchmarkApp) error {
	numFeeds := ba.FeedsKeeper.GetParams(ba.Ctx).MaxCurrentFeeds

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
				SignalPriceStatus: types.SignalPriceStatusAvailable,
				Validator:         val.ValAddress.String(),
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
				SignalPriceStatus: types.SignalPriceStatusAvailable,
				Validator:         val.ValAddress.String(),
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
			[]sdk.Msg{banktypes.NewMsgSend(ba.Sender.Address, acc.Address, sdk.NewCoins(sdk.NewCoin("uband", sdkmath.NewInt(1))))},
			ba.Sender,
			1,
		)[0]

		txs = append(txs, tx)
	}

	_, err := ba.FinalizeBlock(
		&abci.RequestFinalizeBlock{
			Txs:    txs,
			Height: ba.LastBlockHeight() + 1,
			Time:   time.Now(),
		},
	)
	if err != nil {
		return nil, err
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
			sdk.NewCoin("uband", sdkmath.NewInt(150000000)),
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

	_, err = ba.FinalizeBlock(
		&abci.RequestFinalizeBlock{
			Txs:    txs,
			Height: ba.LastBlockHeight() + 1,
			Time:   time.Now(),
		},
	)
	if err != nil {
		return nil, err
	}

	_, err = ba.Commit()
	if err != nil {
		return nil, err
	}

	return accs, nil
}

// generateValidatorPrices generates a slice of ValidatorPrice with the specified number of elements.
func generateValidatorPrices(numElements int, validatorAddress string, timestamp int64) []types.ValidatorPrice {
	prices := make([]types.ValidatorPrice, numElements)

	for i := 0; i < numElements; i++ {
		prices[i] = types.ValidatorPrice{
			Validator: validatorAddress,
			SignalID:  fmt.Sprintf("CS:BAND%d-USD", i),
			Price:     1e10,
			Timestamp: timestamp,
		}
	}
	return prices
}
