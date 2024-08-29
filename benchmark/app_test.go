package benchmark

import (
	"math"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/stretchr/testify/require"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/pkg/tss/testutil"
	bandtesting "github.com/bandprotocol/chain/v2/testing"
	bandtsskeeper "github.com/bandprotocol/chain/v2/x/bandtss/keeper"
	bandtsstypes "github.com/bandprotocol/chain/v2/x/bandtss/types"
	"github.com/bandprotocol/chain/v2/x/oracle/keeper"
	oracletypes "github.com/bandprotocol/chain/v2/x/oracle/types"
	tsskeeper "github.com/bandprotocol/chain/v2/x/tss/keeper"
	tsstypes "github.com/bandprotocol/chain/v2/x/tss/types"
)

type BenchmarkApp struct {
	*bandtesting.TestingApp
	Sender         *Account
	Validator      *Account
	Oid            uint64
	Did            uint64
	TxConfig       client.TxConfig
	TxEncoder      sdk.TxEncoder
	TB             testing.TB
	Ctx            sdk.Context
	Querier        keeper.Querier
	TSSMsgSrvr     tsstypes.MsgServer
	BandtssMsgSrvr bandtsstypes.MsgServer
	Authority      sdk.AccAddress
	PrivKeyStore   map[string]tss.Scalar
}

var (
	PrivD = testutil.HexDecode("de6aedbe8ba688dd6d342881eb1e67c3476e825106477360148e2858a5eb565c")
	PrivE = testutil.HexDecode("3ff4fb2beac0cee0ab230829a5ae0881310046282e79c978ca22f44897ea434a")
	PubD  = tss.Scalar(PrivD).Point()
	PubE  = tss.Scalar(PrivE).Point()

	DELen = 30
)

func GetDEs() []tsstypes.DE {
	var delist []tsstypes.DE
	for i := 0; i < DELen; i++ {
		delist = append(delist, tsstypes.DE{PubD: PubD, PubE: PubE})
	}
	return delist
}

func InitializeBenchmarkApp(tb testing.TB, maxGasPerBlock int64) *BenchmarkApp {
	app, _ := bandtesting.CreateTestApp(&testing.T{}, false)
	ba := &BenchmarkApp{
		TestingApp: app,
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
	ba.Ctx = ba.NewUncachedContext(false, tmproto.Header{ChainID: bandtesting.ChainID})
	ba.TSSMsgSrvr = tsskeeper.NewMsgServerImpl(ba.TestingApp.TSSKeeper)
	ba.BandtssMsgSrvr = bandtsskeeper.NewMsgServerImpl(ba.TestingApp.BandtssKeeper)
	ba.Querier = keeper.Querier{
		Keeper: ba.OracleKeeper,
	}
	ba.TxConfig = ba.GetTxConfig()
	ba.TxEncoder = ba.TxConfig.TxEncoder()

	ba.Commit()
	ba.CallBeginBlock()

	ba.StoreConsensusParams(ba.Ctx, GetConsensusParams(maxGasPerBlock))

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

	// get gov address
	ba.Authority = authtypes.NewModuleAddress(govtypes.ModuleName)
	ba.PrivKeyStore = make(map[string]tss.Scalar)

	ba.CallEndBlock()
	ba.Commit()

	return ba
}

func (ba *BenchmarkApp) DeliverMsg(account *Account, msgs []sdk.Msg) (sdk.GasInfo, *sdk.Result, error) {
	tx := GenSequenceOfTxs(ba.TxConfig, msgs, account, 1)[0]
	gas, res, err := ba.CallDeliver(tx)
	return gas, res, err
}

func (ba *BenchmarkApp) CallBeginBlock() abci.ResponseBeginBlock {
	return ba.BeginBlock(
		abci.RequestBeginBlock{
			Header: tmproto.Header{
				Height:  ba.LastBlockHeight() + 1,
				ChainID: bandtesting.ChainID,
			},
			Hash: ba.LastCommitID().Hash,
		},
	)
}

func (ba *BenchmarkApp) CallEndBlock() abci.ResponseEndBlock {
	return ba.EndBlock(abci.RequestEndBlock{Height: ba.LastBlockHeight() + 1})
}

func (ba *BenchmarkApp) CallDeliver(tx sdk.Tx) (sdk.GasInfo, *sdk.Result, error) {
	return ba.SimDeliver(ba.TxEncoder, tx)
}

func (ba *BenchmarkApp) AddMaxMsgRequests(msg []sdk.Msg) {
	// maximum of request blocks is only 20 because after that it will become report only block because of ante
	for block := 0; block < 10; block++ {
		ba.CallBeginBlock()

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

		ba.CallEndBlock()
		ba.Commit()
	}
}

func (ba *BenchmarkApp) GetAllPendingRequests(account *Account) *oracletypes.QueryPendingRequestsResponse {
	res, err := ba.Querier.PendingRequests(
		sdk.WrapSDKContext(ba.Ctx),
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

func (ba *BenchmarkApp) SetupTSSGroup() {
	ctx, msgSrvr := ba.Ctx, ba.TSSMsgSrvr
	tssKeeper, bandtssKeeper := ba.TestingApp.TSSKeeper, ba.TestingApp.BandtssKeeper

	memberPubKey := testutil.TestCases[0].Group.Members[0].PubKey()
	memberPrivKey := testutil.TestCases[0].Group.Members[0].PrivKey
	groupPubKey := testutil.TestCases[0].Group.PubKey
	dkg := testutil.TestCases[0].Group.DKGContext
	gid := testutil.TestCases[0].Group.ID

	// Set members and submit DEs
	tssKeeper.SetMember(ctx, tsstypes.Member{
		ID:          tss.MemberID(1),
		GroupID:     gid,
		Address:     ba.Sender.Address.String(),
		PubKey:      memberPubKey,
		IsMalicious: false,
		IsActive:    true,
	})
	_, err := msgSrvr.SubmitDEs(ctx, &tsstypes.MsgSubmitDEs{
		DEs:    GetDEs(),
		Sender: ba.Sender.Address.String(),
	})
	require.NoError(ba.TB, err)

	// CreateGroup
	tssKeeper.CreateNewGroup(ctx, tsstypes.Group{
		ID:            gid,
		Size_:         1,
		Threshold:     1,
		PubKey:        groupPubKey,
		Status:        tsstypes.GROUP_STATUS_ACTIVE,
		CreatedHeight: 1,
	})
	tssKeeper.SetDKGContext(ctx, gid, dkg)

	// Set current group in bandtss module
	bandtssKeeper.SetCurrentGroupID(ctx, gid)
	err = bandtssKeeper.AddMember(ctx, ba.Sender.Address, gid)
	require.NoError(ba.TB, err)

	// Set privKey Store
	ba.PrivKeyStore[memberPubKey.String()] = memberPrivKey
}

func (ba *BenchmarkApp) GetPendingSignTxs(gid tss.GroupID) []sdk.Tx {
	ctx, k := ba.Ctx, ba.TSSKeeper

	group := k.MustGetGroup(ctx, gid)

	var txs []sdk.Tx
	members := k.MustGetMembers(ctx, gid)

	for _, m := range members {
		sids := k.GetPendingSigningsByPubKey(ctx, m.PubKey)

		privKey, ok := ba.PrivKeyStore[m.PubKey.String()]
		require.True(ba.TB, ok)
		require.NotNil(ba.TB, privKey)

		for _, sid := range sids {
			signing := k.MustGetSigning(ctx, sid)
			signingAttempt := k.MustGetSigningAttempt(ctx, sid, signing.CurrentAttempt)
			assignedMembers := tsstypes.AssignedMembers(signingAttempt.AssignedMembers)

			sig, err := CreateSignature(m.ID, signing, assignedMembers, group.PubKey, privKey)
			require.NoError(ba.TB, err)

			tx, err := bandtesting.GenTx(
				ba.TxConfig,
				GenMsgSubmitSignature(sid, m.ID, sig, ba.Sender.Address),
				sdk.Coins{sdk.NewInt64Coin("uband", 1)},
				math.MaxInt64,
				bandtesting.ChainID,
				[]uint64{ba.Sender.Num},
				[]uint64{ba.Sender.Seq},
				ba.Sender.PrivKey,
			)
			require.NoError(ba.TB, err)

			ba.Sender.Seq += 1

			txs = append(txs, tx)
		}
	}
	return txs
}

func (ba *BenchmarkApp) HandleGenPendingSignTxs(
	gid tss.GroupID,
	content tsstypes.Content,
	feeLimit sdk.Coins,
) []sdk.Tx {
	txs := ba.GetPendingSignTxs(gid)
	if len(txs) > 0 {
		return txs
	}

	ba.RequestSignature(ba.Sender, content, feeLimit)
	ba.AddDEs(ba.Sender.Address)

	return ba.GetPendingSignTxs(gid)
}

func (ba *BenchmarkApp) RequestSignature(
	sender *Account,
	content tsstypes.Content,
	feeLimit sdk.Coins,
) {
	ctx, msgSrvr := ba.Ctx, ba.BandtssMsgSrvr

	msg, err := bandtsstypes.NewMsgRequestSignature(content, feeLimit, sender.Address)
	require.NoError(ba.TB, err)

	_, err = msgSrvr.RequestSignature(ctx, msg)
	require.NoError(ba.TB, err)
}

func (ba *BenchmarkApp) AddDEs(addr sdk.AccAddress) {
	ctx, msgSrvr, k := ba.Ctx, ba.TSSMsgSrvr, ba.TSSKeeper

	deQueue := k.GetDEQueue(ctx, addr)
	count := deQueue.Tail - deQueue.Head
	if count < uint64(DELen) {
		_, err := msgSrvr.SubmitDEs(ctx, &tsstypes.MsgSubmitDEs{
			DEs:    GetDEs(),
			Sender: addr.String(),
		})
		require.NoError(ba.TB, err)
	}
}
