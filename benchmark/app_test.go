package benchmark

import (
	"math"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/libs/log"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/stretchr/testify/require"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/pkg/tss/testutil"
	testapp "github.com/bandprotocol/chain/v2/testing/testapp"
	"github.com/bandprotocol/chain/v2/x/oracle/keeper"
	oracletypes "github.com/bandprotocol/chain/v2/x/oracle/types"
	tsskeeper "github.com/bandprotocol/chain/v2/x/tss/keeper"
	tsstypes "github.com/bandprotocol/chain/v2/x/tss/types"
)

type BenchmarkApp struct {
	*testapp.TestingApp
	Sender     *Account
	Validator  *Account
	Oid        uint64
	Did        uint64
	Gid        tss.GroupID
	TxConfig   client.TxConfig
	TxEncoder  sdk.TxEncoder
	TB         testing.TB
	Ctx        sdk.Context
	Querier    keeper.Querier
	TSSMsgSrvr tsstypes.MsgServer
	authority  sdk.AccAddress
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
		TB: tb,
	}
	ba.Ctx = ba.NewUncachedContext(false, tmproto.Header{})
	ba.TSSMsgSrvr = tsskeeper.NewMsgServerImpl(&ba.TestingApp.TSSKeeper)
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

	// set group ID
	ba.Gid = tss.GroupID(1)
	ba.authority = authtypes.NewModuleAddress(govtypes.ModuleName)

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
			Header: tmproto.Header{Height: ba.LastBlockHeight() + 1},
			Hash:   ba.LastCommitID().Hash,
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
	ctx, msgSrvr, k := ba.Ctx, ba.TSSMsgSrvr, ba.TestingApp.TSSKeeper

	// force address to owner
	owner := ba.Sender.Address.String()

	// Create group from testutil
	for _, tc := range testutil.TestCases {
		// Initialize members
		for i, m := range tc.Group.Members {
			k.SetMember(ctx, tsstypes.Member{
				ID:          tss.MemberID(i + 1),
				GroupID:     tc.Group.ID,
				Address:     owner,
				PubKey:      m.PubKey(),
				IsMalicious: false,
			})

			err := k.SetActiveStatus(ctx, ba.Sender.Address)
			require.NoError(ba.TB, err)
		}

		k.CreateNewGroup(ctx, tsstypes.Group{
			ID:            tc.Group.ID,
			Size_:         uint64(tc.Group.GetSize()),
			Threshold:     tc.Group.Threshold,
			PubKey:        tc.Group.PubKey,
			Status:        tsstypes.GROUP_STATUS_ACTIVE,
			Fee:           sdk.NewCoins(sdk.NewInt64Coin("uband", 10)),
			CreatedHeight: 1,
		})
		k.SetDKGContext(ctx, tc.Group.ID, tc.Group.DKGContext)

		// Submit DEs for each member
		_, err := msgSrvr.SubmitDEs(ctx, &tsstypes.MsgSubmitDEs{
			DEs:     GetDEs(),
			Address: ba.Sender.Address.String(),
		})
		require.NoError(ba.TB, err)
	}
}

func (ba *BenchmarkApp) GetPendingSignTxs(
	gid tss.GroupID,
	tcs []testutil.TestCase,
) []sdk.Tx {
	ctx, k := ba.Ctx, ba.TSSKeeper

	group := k.MustGetGroup(ctx, gid)

	var txs []sdk.Tx
	members := k.MustGetMembers(ctx, gid)

	for _, m := range members {
		addr := sdk.AccAddress(m.PubKey)

		sids := k.GetPendingSigningsByPubKey(ctx, m.PubKey)

		ownPrivkey := FindPrivateKey(tcs, gid, addr)
		require.NotNil(ba.TB, ownPrivkey)

		for _, sid := range sids {
			signing := k.MustGetSigning(ctx, tss.SigningID(sid))

			sig, err := CreateSignature(m.ID, signing, group.PubKey, ownPrivkey)
			require.NoError(ba.TB, err)

			tx, err := testapp.GenTx(
				ba.TxConfig,
				GenMsgSubmitSignature(tss.SigningID(sid), m.ID, sig, ba.Sender.Address),
				sdk.Coins{sdk.NewInt64Coin("uband", 1)},
				math.MaxInt64,
				"",
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
	tcs []testutil.TestCase,
) []sdk.Tx {
	txs := ba.GetPendingSignTxs(gid, tcs)
	if len(txs) > 0 {
		return txs
	}

	ba.RequestSignature(ba.Sender, gid, content, feeLimit)
	ba.AddDEs(ba.Gid)

	return ba.GetPendingSignTxs(gid, tcs)
}

func (ba *BenchmarkApp) RequestSignature(
	sender *Account,
	gid tss.GroupID,
	content tsstypes.Content,
	feeLimit sdk.Coins,
) {
	ctx, msgSrvr := ba.Ctx, ba.TSSMsgSrvr

	msg, err := tsstypes.NewMsgRequestSignature(gid, content, feeLimit, sender.Address)
	require.NoError(ba.TB, err)

	_, err = msgSrvr.RequestSignature(ctx, msg)
	require.NoError(ba.TB, err)
}

func (ba *BenchmarkApp) AddDEs(
	gid tss.GroupID,
) {
	ctx, msgSrvr, k := ba.Ctx, ba.TSSMsgSrvr, ba.TSSKeeper

	count := k.GetDECount(ctx, ba.Sender.Address)
	if count < uint64(DELen) {
		_, err := msgSrvr.SubmitDEs(ctx, &tsstypes.MsgSubmitDEs{
			DEs:     GetDEs(),
			Address: ba.Sender.Address.String(),
		})
		require.NoError(ba.TB, err)
	}
}
