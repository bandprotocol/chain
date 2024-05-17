package benchmark

import (
	"fmt"
	"math/rand"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"

	bandtesting "github.com/bandprotocol/chain/v2/testing"
	"github.com/bandprotocol/chain/v2/x/feeds/types"
	oracletypes "github.com/bandprotocol/chain/v2/x/oracle/types"
)

// benchmark test for delivering MsgSubmitPrices
func BenchmarkSubmitPricesDeliver(b *testing.B) {
	b.ResetTimer()
	b.StopTimer()

	for i := 0; i < b.N; i++ {
		ba := InitializeBenchmarkApp(b, -1)

		numVals := ba.StakingKeeper.GetParams(ba.Ctx).MaxValidators

		vals, err := generateValidators(ba, int(numVals))
		require.NoError(b, err)

		err = setupFeeds(ba)
		require.NoError(b, err)

		ba.CallBeginBlock()

		txs := []sdk.Tx{}
		for _, val := range vals {
			tx := GenSequenceOfTxs(
				ba.TxConfig,
				GenMsgSubmitPrices(
					val,
					ba.FeedsKeeper.GetSupportedFeeds(ba.Ctx).Feeds,
					ba.Ctx.BlockTime().Unix(),
				),
				val,
				1,
			)[0]

			txs = append(txs, tx)
		}

		for txIdx, tx := range txs {
			b.StartTimer()
			gasInfo, _, err := ba.CallDeliver(tx)
			b.StopTimer()
			if err != nil {
				require.NoError(b, err)
			}
			if i == 0 && txIdx == 0 {
				fmt.Println("\tCosmos Gas used:", gasInfo.GasUsed)
			}
		}

		ba.CallEndBlock()
		ba.Commit()
	}
}

// benchmark test for endblock of feeds module
func BenchmarkFeedsEndBlock(b *testing.B) {
	ba := InitializeBenchmarkApp(b, -1)

	numVals := ba.StakingKeeper.GetParams(ba.Ctx).MaxValidators

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
		ba.CallBeginBlock()

		// process endblock
		b.StartTimer()
		ba.CallEndBlock()
		b.StopTimer()

		ba.Commit()
	}
}

func setupFeeds(ba *BenchmarkApp) error {
	numFeeds := ba.FeedsKeeper.GetParams(ba.Ctx).MaxSupportedFeeds

	ba.CallBeginBlock()

	feeds := []types.Feed{}
	for i := int64(0); i < numFeeds; i++ {
		feeds = append(feeds, types.Feed{
			SignalID:              fmt.Sprintf("signal.%d", i),
			Interval:              60,
			DeviationInThousandth: 5,
		})
	}
	ba.FeedsKeeper.SetSupportedFeeds(ba.Ctx, feeds)

	ba.CallEndBlock()
	ba.Commit()

	return nil
}

func setupValidatorPrices(ba *BenchmarkApp, vals []*Account) error {
	sfs := ba.FeedsKeeper.GetSupportedFeeds(ba.Ctx)

	ba.CallBeginBlock()
	for _, feed := range sfs.Feeds {
		for valIdx, val := range vals {
			err := ba.FeedsKeeper.SetValidatorPrice(ba.Ctx, types.ValidatorPrice{
				PriceStatus: types.PriceStatusAvailable,
				Validator:   val.ValAddress.String(),
				SignalID:    feed.SignalID,
				Price:       (10000 + uint64(valIdx)) * 10e9,
				Timestamp:   ba.Ctx.BlockTime().Unix(),
			})
			if err != nil {
				return err
			}
		}
	}
	ba.CallEndBlock()
	ba.Commit()

	return nil
}

func generateValidators(ba *BenchmarkApp, num int) ([]*Account, error) {
	// transfer money
	ba.CallBeginBlock()

	vals := []bandtesting.Account{}
	for i := 0; i < num; i++ {
		r := rand.New(rand.NewSource(int64(i)))
		acc := bandtesting.CreateArbitraryAccount(r)
		vals = append(vals, acc)

		tx := GenSequenceOfTxs(
			ba.TxConfig,
			[]sdk.Msg{banktypes.NewMsgSend(ba.Sender.Address, acc.Address, sdk.NewCoins(sdk.NewCoin("uband", sdk.NewInt(200000000))))},
			ba.Sender,
			1,
		)[0]

		_, _, err := ba.CallDeliver(tx)
		if err != nil {
			return nil, err
		}
	}

	ba.CallEndBlock()
	ba.Commit()

	// apply to be a validator
	ba.CallBeginBlock()

	accs := []*Account{}
	for _, val := range vals {
		info := ba.AccountKeeper.GetAccount(ba.Ctx, val.Address)
		acc := &Account{
			Account: val,
			Num:     info.GetAccountNumber(),
			Seq:     info.GetSequence(),
		}
		accs = append(accs, acc)

		msgCreateVal, err := stakingtypes.NewMsgCreateValidator(
			val.ValAddress,
			val.PubKey,
			sdk.NewCoin("uband", sdk.NewInt(150000000)),
			stakingtypes.NewDescription(val.Address.String(), val.Address.String(), "", "", ""),
			stakingtypes.NewCommissionRates(sdk.NewDec(1), sdk.NewDec(1), sdk.NewDec(1)),
			sdk.NewInt(1),
		)
		if err != nil {
			return nil, err
		}

		msgActivate := oracletypes.NewMsgActivate(val.ValAddress)

		tx := GenSequenceOfTxs(
			ba.TxConfig,
			[]sdk.Msg{msgCreateVal, msgActivate},
			acc,
			1,
		)[0]

		_, _, err = ba.CallDeliver(tx)
		if err != nil {
			return nil, err
		}
	}

	ba.CallEndBlock()
	ba.Commit()

	return accs, nil
}
