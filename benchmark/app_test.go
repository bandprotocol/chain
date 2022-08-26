package benchmark

import (
	"strconv"
	"testing"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	testapp "github.com/bandprotocol/chain/v2/testing/testapp"
	"github.com/bandprotocol/chain/v2/x/oracle/keeper"
	oracletypes "github.com/bandprotocol/chain/v2/x/oracle/types"
)

type BenchmarkApp struct {
	*testapp.TestingApp
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

func InitializeBenchmarkApp(b testing.TB) *BenchmarkApp {
	ba := &BenchmarkApp{
		TestingApp: testapp.NewTestApp("", log.NewNopLogger()),
		Sender: &Account{
			Account: testapp.Owner,
			Num:     0,
			Seq:     0,
		},
		Validator: &Account{
			Account: testapp.Validators[0],
			Num:     5,
			Seq:     0,
		},
		TB: b,
	}
	ba.Ctx = ba.NewUncachedContext(false, tmproto.Header{})
	ba.Querier = keeper.Querier{
		Keeper: ba.OracleKeeper,
	}
	ba.TxConfig = ba.GetTxConfig()
	ba.TxEncoder = ba.TxConfig.TxEncoder()

	ba.Commit()
	ba.CallBeginBlock()

	// create oracle script
	oCode, err := GetBenchmarkWasm()
	require.NoError(b, err)
	oid := ba.SendMsgCreateOracleScript(ba.Sender, oCode)
	ba.Oid = uint64(oid)

	// create data source
	dCode := []byte("hello")
	did := ba.SendMsgCreateDataSource(ba.Sender, dCode)
	ba.Did = uint64(did)

	// activate oracle
	ba.SendMsgActivate(ba.Validator)

	ba.CallEndBlock()
	ba.Commit()

	return ba
}

func (ba *BenchmarkApp) SendMsgCreateOracleScript(account *Account, code []byte) int {
	tx := GenSequenceOfTxs(ba.TxConfig, GenMsgCreateOracleScript(account, code), account, 1)[0]
	_, res, err := ba.CallDeliver(tx)
	require.NoError(ba.TB, err)

	evt := res.Events[len(res.Events)-1]
	attr := evt.Attributes[0]
	oid, err := strconv.Atoi(string(attr.Value))
	require.NoError(ba.TB, err)

	return oid
}

func (ba *BenchmarkApp) SendMsgCreateDataSource(account *Account, code []byte) int {
	tx := GenSequenceOfTxs(ba.TxConfig, GenMsgCreateDataSource(account, code), account, 1)[0]
	_, res, err := ba.CallDeliver(tx)
	require.NoError(ba.TB, err)

	evt := res.Events[len(res.Events)-1]
	attr := evt.Attributes[0]
	did, err := strconv.Atoi(string(attr.Value))
	require.NoError(ba.TB, err)

	return did
}

func (ba *BenchmarkApp) SendMsgActivate(account *Account) {
	tx := GenSequenceOfTxs(ba.TxConfig, GenMsgActivate(account), account, 1)[0]
	_, _, err := ba.CallDeliver(tx)
	require.NoError(ba.TB, err)
}

func (ba *BenchmarkApp) CallBeginBlock() abci.ResponseBeginBlock {
	return ba.BeginBlock(
		abci.RequestBeginBlock{
			Header: tmproto.Header{Height: ba.LastBlockHeight() + 1},
			Hash:   ba.LastCommitID().Hash,
		},
	)
}

func (ba *BenchmarkApp) CallEndBlock() abci.ResponseEndBlock {
	return ba.EndBlock(abci.RequestEndBlock{Height: ba.LastBlockHeight() + 1})
}

func (ba *BenchmarkApp) CallDeliver(tx sdk.Tx) (sdk.GasInfo, *sdk.Result, error) {
	return ba.Deliver(ba.TxEncoder, tx)
}

func (ba *BenchmarkApp) SendAllPendingReports(account *Account) {
	// query all pending requests
	res, err := ba.Querier.PendingRequests(
		sdk.WrapSDKContext(ba.Ctx),
		&oracletypes.QueryPendingRequestsRequest{
			ValidatorAddress: account.ValAddress.String(),
		},
	)
	require.NoError(ba.TB, err)

	for _, rid := range res.RequestIDs {
		request, err := ba.OracleKeeper.GetRequest(ba.Ctx, oracletypes.RequestID(rid))

		// find  all external ids of the request
		eids := []int64{}
		for _, raw := range request.RawRequests {
			eids = append(eids, int64(raw.ExternalID))
		}
		require.NoError(ba.TB, err)

		tx := GenSequenceOfTxs(
			ba.TxConfig,
			GenMsgReportData(account, rid, eids),
			account,
			1,
		)[0]

		// deliver MsgReportData for the request
		_, _, err = ba.CallDeliver(tx)
		require.NoError(ba.TB, err)

	}
}
