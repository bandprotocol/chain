package benchmark

import (
	"testing"

	"github.com/stretchr/testify/require"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"

	band "github.com/bandprotocol/chain/v3/app"
	bandtesting "github.com/bandprotocol/chain/v3/testing"
	"github.com/bandprotocol/chain/v3/x/oracle/keeper"
	oracletypes "github.com/bandprotocol/chain/v3/x/oracle/types"
)

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
			Num:     5,
			Seq:     0,
		},
		TB: tb,
	}
	ba.Ctx = ba.NewUncachedContext(false, cmtproto.Header{ChainID: bandtesting.ChainID})
	ba.Querier = keeper.Querier{
		Keeper: ba.OracleKeeper,
	}
	ba.TxConfig = ba.GetTxConfig()
	ba.TxEncoder = ba.TxConfig.TxEncoder()

	_, err := ba.Commit()
	require.NoError(tb, err)

	_, err = ba.CallBeginBlock()
	require.NoError(tb, err)

	err = ba.StoreConsensusParams(ba.Ctx, GetConsensusParams(maxGasPerBlock))
	require.NoError(tb, err)

	// create oracle script
	oCode, err := GetBenchmarkWasm()
	require.NoError(tb, err)
	_, res, err := ba.DeliverMsg(ba.Sender, GenMsgCreateOracleScript(ba.Sender, oCode))
	require.NoError(tb, err)
	oid, err := GetFirstAttributeOfLastEventValue(res.Events)
	require.NoError(tb, err)
	ba.Oid = uint64(oid)

	// create data source
	dCode := []byte("hello")
	_, res, err = ba.DeliverMsg(ba.Sender, GenMsgCreateDataSource(ba.Sender, dCode))
	require.NoError(tb, err)
	did, err := GetFirstAttributeOfLastEventValue(res.Events)
	require.NoError(tb, err)
	ba.Did = uint64(did)

	// activate oracle
	_, _, err = ba.DeliverMsg(ba.Validator, GenMsgActivate(ba.Validator))
	require.NoError(tb, err)

	_, err = ba.CallEndBlock()
	require.NoError(tb, err)

	_, err = ba.Commit()
	require.NoError(tb, err)

	return ba
}

func (ba *BenchmarkApp) DeliverMsg(account *Account, msgs []sdk.Msg) (sdk.GasInfo, *sdk.Result, error) {
	tx := GenSequenceOfTxs(ba.TxConfig, msgs, account, 1)[0]
	gas, res, err := ba.CallDeliver(tx)
	return gas, res, err
}

func (ba *BenchmarkApp) CallBeginBlock() (sdk.BeginBlock, error) {
	return ba.BeginBlocker(ba.Ctx)
}

func (ba *BenchmarkApp) CallEndBlock() (sdk.EndBlock, error) {
	return ba.EndBlocker(ba.Ctx)
}

func (ba *BenchmarkApp) CallDeliver(tx sdk.Tx) (sdk.GasInfo, *sdk.Result, error) {
	return ba.SimDeliver(ba.TxEncoder, tx)
}

func (ba *BenchmarkApp) AddMaxMsgRequests(msg []sdk.Msg) {
	// maximum of request blocks is only 20 because after that it will become report only block because of ante
	for block := 0; block < 10; block++ {
		_, err := ba.CallBeginBlock()
		require.NoError(ba.TB, err)

		totalGas := uint64(0)
		for {
			tx := GenSequenceOfTxs(
				ba.TxConfig,
				msg,
				ba.Sender,
				1,
			)[0]

			gas, _, _ := ba.CallDeliver(tx)

			totalGas += gas.GasUsed
			if totalGas+gas.GasUsed >= uint64(BlockMaxGas) {
				break
			}
		}

		_, err = ba.CallEndBlock()
		require.NoError(ba.TB, err)

		_, err = ba.Commit()
		require.NoError(ba.TB, err)
	}
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

func (ba *BenchmarkApp) SendAllPendingReports(account *Account) {
	// query all pending requests
	res := ba.GetAllPendingRequests(account)

	for _, rid := range res.RequestIDs {
		_, _, err := ba.DeliverMsg(account, ba.GenMsgReportData(account, []uint64{rid}))
		require.NoError(ba.TB, err)
	}
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
