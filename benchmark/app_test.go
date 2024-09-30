package benchmark

// TODO: Fix tests
// import (
// 	"testing"

// 	abci "github.com/cometbft/cometbft/abci/types"
// 	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
// 	"github.com/cosmos/cosmos-sdk/client"
// 	sdk "github.com/cosmos/cosmos-sdk/types"
// 	"github.com/stretchr/testify/require"

// 	bandtesting "github.com/bandprotocol/chain/v3/testing"
// 	"github.com/bandprotocol/chain/v3/x/oracle/keeper"
// 	oracletypes "github.com/bandprotocol/chain/v3/x/oracle/types"
// )

// type BenchmarkApp struct {
// 	*bandtesting.TestingApp
// 	Sender    *Account
// 	Validator *Account
// 	Oid       uint64
// 	Did       uint64
// 	TxConfig  client.TxConfig
// 	TxEncoder sdk.TxEncoder
// 	TB        testing.TB
// 	Ctx       sdk.Context
// 	Querier   keeper.Querier
// }

// func InitializeBenchmarkApp(tb testing.TB, maxGasPerBlock int64) *BenchmarkApp {
// 	app, _ := bandtesting.CreateTestApp(&testing.T{}, false)
// 	ba := &BenchmarkApp{
// 		TestingApp: app,
// 		Sender: &Account{
// 			Account: bandtesting.Owner,
// 			Num:     0,
// 			Seq:     0,
// 		},
// 		Validator: &Account{
// 			Account: bandtesting.Validators[0],
// 			Num:     5,
// 			Seq:     0,
// 		},
// 		TB: tb,
// 	}
// 	ba.Ctx = ba.NewUncachedContext(false, cmtproto.Header{ChainID: bandtesting.ChainID})
// 	ba.Querier = keeper.Querier{
// 		Keeper: ba.OracleKeeper,
// 	}
// 	ba.TxConfig = ba.GetTxConfig()
// 	ba.TxEncoder = ba.TxConfig.TxEncoder()

// 	ba.Commit()
// 	ba.CallBeginBlock()

// 	ba.StoreConsensusParams(ba.Ctx, GetConsensusParams(maxGasPerBlock))

// 	// create oracle script
// 	oCode, err := GetBenchmarkWasm()
// 	require.NoError(tb, err)
// 	_, res, err := ba.DeliverMsg(ba.Sender, GenMsgCreateOracleScript(ba.Sender, oCode))
// 	require.NoError(tb, err)
// 	oid, err := GetFirstAttributeOfLastEventValue(res.Events)
// 	require.NoError(tb, err)
// 	ba.Oid = uint64(oid)

// 	// create data source
// 	dCode := []byte("hello")
// 	_, res, err = ba.DeliverMsg(ba.Sender, GenMsgCreateDataSource(ba.Sender, dCode))
// 	require.NoError(tb, err)
// 	did, err := GetFirstAttributeOfLastEventValue(res.Events)
// 	require.NoError(tb, err)
// 	ba.Did = uint64(did)

// 	// activate oracle
// 	_, _, err = ba.DeliverMsg(ba.Validator, GenMsgActivate(ba.Validator))
// 	require.NoError(tb, err)

// 	ba.CallEndBlock()
// 	ba.Commit()

// 	return ba
// }

// func (ba *BenchmarkApp) DeliverMsg(account *Account, msgs []sdk.Msg) (sdk.GasInfo, *sdk.Result, error) {
// 	tx := GenSequenceOfTxs(ba.TxConfig, msgs, account, 1)[0]
// 	gas, res, err := ba.CallDeliver(tx)
// 	return gas, res, err
// }

// func (ba *BenchmarkApp) CallBeginBlock() abci.ResponseBeginBlock {
// 	return ba.FinalizeBlock(

// 		&abci.RequestFinalizeBlock{
// 			Height: ba.LastBlockHeight() + 1,
// 			Hash:   ba.LastCommitID().Hash,
// 		},
// 	)
// }

// func (ba *BenchmarkApp) CallEndBlock() abci.ResponseEndBlock {
// 	return ba.EndBlock(abci.RequestEndBlock{Height: ba.LastBlockHeight() + 1})
// }

// func (ba *BenchmarkApp) CallDeliver(tx sdk.Tx) (sdk.GasInfo, *sdk.Result, error) {
// 	return ba.SimDeliver(ba.TxEncoder, tx)
// }

// func (ba *BenchmarkApp) AddMaxMsgRequests(msg []sdk.Msg) {
// 	// maximum of request blocks is only 20 because after that it will become report only block because of ante
// 	for block := 0; block < 10; block++ {
// 		ba.CallBeginBlock()

// 		totalGas := uint64(0)
// 		for {
// 			tx := GenSequenceOfTxs(
// 				ba.TxConfig,
// 				msg,
// 				ba.Sender,
// 				1,
// 			)[0]

// 			gas, _, _ := ba.CallDeliver(tx)

// 			totalGas += gas.GasUsed
// 			if totalGas+gas.GasUsed >= uint64(BlockMaxGas) {
// 				break
// 			}
// 		}

// 		ba.CallEndBlock()
// 		ba.Commit()
// 	}
// }

// func (ba *BenchmarkApp) GetAllPendingRequests(account *Account) *oracletypes.QueryPendingRequestsResponse {
// 	res, err := ba.Querier.PendingRequests(
// 		sdk.WrapSDKContext(ba.Ctx),
// 		&oracletypes.QueryPendingRequestsRequest{
// 			ValidatorAddress: account.ValAddress.String(),
// 		},
// 	)
// 	require.NoError(ba.TB, err)

// 	return res
// }

// func (ba *BenchmarkApp) SendAllPendingReports(account *Account) {
// 	// query all pending requests
// 	res := ba.GetAllPendingRequests(account)

// 	for _, rid := range res.RequestIDs {
// 		_, _, err := ba.DeliverMsg(account, ba.GenMsgReportData(account, []uint64{rid}))
// 		require.NoError(ba.TB, err)
// 	}
// }

// func (ba *BenchmarkApp) GenMsgReportData(account *Account, rids []uint64) []sdk.Msg {
// 	msgs := make([]sdk.Msg, 0)

// 	for _, rid := range rids {
// 		request, err := ba.OracleKeeper.GetRequest(ba.Ctx, oracletypes.RequestID(rid))

// 		// find  all external ids of the request
// 		eids := []int64{}
// 		for _, raw := range request.RawRequests {
// 			eids = append(eids, int64(raw.ExternalID))
// 		}
// 		require.NoError(ba.TB, err)

// 		rawReports := []oracletypes.RawReport{}

// 		for _, eid := range eids {
// 			rawReports = append(rawReports, oracletypes.RawReport{
// 				ExternalID: oracletypes.ExternalID(eid),
// 				ExitCode:   0,
// 				Data:       []byte(""),
// 			})
// 		}

// 		msgs = append(msgs, &oracletypes.MsgReportData{
// 			RequestID:  oracletypes.RequestID(rid),
// 			RawReports: rawReports,
// 			Validator:  account.ValAddress.String(),
// 		})
// 	}

// 	return msgs
// }
