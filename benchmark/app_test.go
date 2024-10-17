package benchmark

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"

	band "github.com/bandprotocol/chain/v3/app"
	bandtesting "github.com/bandprotocol/chain/v3/testing"
	"github.com/bandprotocol/chain/v3/x/oracle/keeper"
	oracletypes "github.com/bandprotocol/chain/v3/x/oracle/types"
)

func init() {
	band.SetBech32AddressPrefixesAndBip44CoinTypeAndSeal(sdk.GetConfig())
	sdk.DefaultBondDenom = "uband"
}

type BenchmarkApp struct {
	*band.BandApp
	Sender    *Account
	Validator *Account
	Oid       uint64
	Did       uint64
	TxConfig  client.TxConfig
	TxEncoder sdk.TxEncoder
	TB        testing.TB
	Ctx       sdk.Context
	Querier   keeper.Querier
}

func InitializeBenchmarkApp(tb testing.TB, maxGasPerBlock int64) *BenchmarkApp {
	dir := testutil.GetTempDir(tb)
	app := bandtesting.SetupWithCustomHome(false, dir)

	ba := &BenchmarkApp{
		BandApp: app,
		Sender: &Account{
			Account: bandtesting.Owner,
			Num:     0,
			Seq:     0,
		},
		Validator: &Account{
			Account: bandtesting.Validators[0],
			Num:     6,
			Seq:     0,
		},
	}
	ba.TB = tb
	ba.Ctx = ba.NewUncachedContext(false, cmtproto.Header{ChainID: bandtesting.ChainID})
	ba.Querier = keeper.Querier{
		Keeper: ba.OracleKeeper,
	}
	ba.TxConfig = ba.GetTxConfig()
	ba.TxEncoder = ba.TxConfig.TxEncoder()

	err := ba.StoreConsensusParams(ba.Ctx, GetConsensusParams(maxGasPerBlock))
	require.NoError(tb, err)

	var txs [][]byte

	// create oracle script
	oCode, err := GetBenchmarkWasm()
	require.NoError(tb, err)
	txs = append(
		txs,
		GenSequenceOfTxs(ba.TxEncoder, ba.TxConfig, GenMsgCreateOracleScript(ba.Sender, oCode), ba.Sender, 1)[0],
	)

	// create data source
	dCode := []byte("hello")
	txs = append(
		txs,
		GenSequenceOfTxs(ba.TxEncoder, ba.TxConfig, GenMsgCreateDataSource(ba.Sender, dCode), ba.Sender, 1)[0],
	)

	// activate oracle
	txs = append(
		txs,
		GenSequenceOfTxs(ba.TxEncoder, ba.TxConfig, GenMsgActivate(ba.Validator), ba.Validator, 1)[0],
	)

	res, err := ba.FinalizeBlock(
		&abci.RequestFinalizeBlock{Txs: txs, Height: ba.LastBlockHeight() + 1, Time: time.Now()},
	)
	require.NoError(tb, err)

	_, err = ba.Commit()
	require.NoError(tb, err)

	oid, err := GetFirstAttributeOfLastEventValue(res.TxResults[0].Events)
	require.NoError(tb, err)
	ba.Oid = uint64(oid)

	did, err := GetFirstAttributeOfLastEventValue(res.TxResults[1].Events)
	require.NoError(tb, err)
	ba.Did = uint64(did)

	return ba
}

func (ba *BenchmarkApp) GetAllPendingRequests(account *Account) *oracletypes.QueryPendingRequestsResponse {
	res, err := ba.Querier.PendingRequests(
		ba.Ctx,
		&oracletypes.QueryPendingRequestsRequest{
			ValidatorAddress: account.ValAddress.String(),
		},
	)
	require.NoError(ba.TB, err)

	return res
}

func (ba *BenchmarkApp) GenMsgReportData(account *Account, rids []uint64) []sdk.Msg {
	msgs := make([]sdk.Msg, 0)

	for _, rid := range rids {
		request, err := ba.OracleKeeper.GetRequest(ba.Ctx, oracletypes.RequestID(rid))

		// find  all external ids of the request
		eids := []int64{}
		for _, raw := range request.RawRequests {
			eids = append(eids, int64(raw.ExternalID))
		}
		require.NoError(ba.TB, err)

		rawReports := []oracletypes.RawReport{}

		for _, eid := range eids {
			rawReports = append(rawReports, oracletypes.RawReport{
				ExternalID: oracletypes.ExternalID(eid),
				ExitCode:   0,
				Data:       []byte(""),
			})
		}

		msgs = append(msgs, &oracletypes.MsgReportData{
			RequestID:  oracletypes.RequestID(rid),
			RawReports: rawReports,
			Validator:  account.ValAddress.String(),
		})
	}

	return msgs
}
