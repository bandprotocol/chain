package benchmark

import (
	"math"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	band "github.com/bandprotocol/chain/v3/app"
	"github.com/bandprotocol/chain/v3/pkg/tss"
	tsstestutil "github.com/bandprotocol/chain/v3/pkg/tss/testutil"
	bandtesting "github.com/bandprotocol/chain/v3/testing"
	bandtsskeeper "github.com/bandprotocol/chain/v3/x/bandtss/keeper"
	bandtsstypes "github.com/bandprotocol/chain/v3/x/bandtss/types"
	"github.com/bandprotocol/chain/v3/x/oracle/keeper"
	oracletypes "github.com/bandprotocol/chain/v3/x/oracle/types"
	tsskeeper "github.com/bandprotocol/chain/v3/x/tss/keeper"
	tsstypes "github.com/bandprotocol/chain/v3/x/tss/types"
)

func init() {
	band.SetBech32AddressPrefixesAndBip44CoinTypeAndSeal(sdk.GetConfig())
	sdk.DefaultBondDenom = "uband"
}

type BenchmarkApp struct {
	*band.BandApp
	Sender         *Account
	Validator      *Account
	Oid            uint64
	Did            uint64
	TxConfig       client.TxConfig
	TxEncoder      sdk.TxEncoder
	TxDecoder      sdk.TxDecoder
	TB             testing.TB
	Ctx            sdk.Context
	Querier        keeper.Querier
	TSSMsgSrvr     tsstypes.MsgServer
	BandtssMsgSrvr bandtsstypes.MsgServer
	Authority      sdk.AccAddress
	PrivKeyStore   map[string]tss.Scalar
}

var (
	PrivD = tsstestutil.HexDecode("de6aedbe8ba688dd6d342881eb1e67c3476e825106477360148e2858a5eb565c")
	PrivE = tsstestutil.HexDecode("3ff4fb2beac0cee0ab230829a5ae0881310046282e79c978ca22f44897ea434a")
	PubD  = tss.Scalar(PrivD).Point()
	PubE  = tss.Scalar(PrivE).Point()
)

func GetDEs(deLen int) []tsstypes.DE {
	var delist []tsstypes.DE
	for i := 0; i < deLen; i++ {
		delist = append(delist, tsstypes.DE{PubD: PubD, PubE: PubE})
	}
	return delist
}

func InitializeBenchmarkApp(tb testing.TB, maxGasPerBlock int64) *BenchmarkApp {
	dir := testutil.GetTempDir(tb)
	app := bandtesting.SetupWithCustomHome(false, dir)

	ba := &BenchmarkApp{
		BandApp: app,
		Sender: &Account{
			Account: bandtesting.Treasury,
			Num:     1,
			Seq:     0,
		},
		Validator: &Account{
			Account: bandtesting.Validators[0],
			Num:     7,
			Seq:     0,
		},
	}

	ba.TB = tb
	ba.Ctx = ba.NewUncachedContext(false, cmtproto.Header{ChainID: bandtesting.ChainID}).
		WithBlockTime(time.Unix(1000000, 0))
	ba.TSSMsgSrvr = tsskeeper.NewMsgServerImpl(ba.TSSKeeper)
	ba.BandtssMsgSrvr = bandtsskeeper.NewMsgServerImpl(ba.BandtssKeeper)
	ba.Querier = keeper.Querier{
		Keeper: ba.OracleKeeper,
	}
	ba.TxConfig = ba.GetTxConfig()
	ba.TxEncoder = ba.TxConfig.TxEncoder()
	ba.TxDecoder = ba.TxConfig.TxDecoder()

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
		&abci.RequestFinalizeBlock{Txs: txs, Height: ba.LastBlockHeight() + 1, Time: ba.Ctx.BlockTime()},
	)
	require.NoError(tb, err)

	for _, tx := range res.TxResults {
		require.Equal(tb, uint32(0), tx.Code)
	}

	// get gov address
	ba.Authority = authtypes.NewModuleAddress(govtypes.ModuleName)
	ba.PrivKeyStore = make(map[string]tss.Scalar)

	_, err = ba.FinalizeBlock(
		&abci.RequestFinalizeBlock{Height: ba.LastBlockHeight() + 1, Time: ba.Ctx.BlockTime()},
	)
	require.NoError(tb, err)

	for _, tx := range res.TxResults {
		require.Equal(tb, uint32(0), tx.Code)
	}

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

func (ba *BenchmarkApp) SetupTSSGroup() {
	ctx, msgSrvr := ba.Ctx, ba.TSSMsgSrvr
	tssKeeper, bandtssKeeper := ba.TSSKeeper, ba.BandtssKeeper

	memberPubKey := tsstestutil.TestCases[0].Group.Members[0].PubKey()
	memberPrivKey := tsstestutil.TestCases[0].Group.Members[0].PrivKey
	groupPubKey := tsstestutil.TestCases[0].Group.PubKey
	dkg := tsstestutil.TestCases[0].Group.DKGContext
	gid := tsstestutil.TestCases[0].Group.ID

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
		DEs:    GetDEs(300),
		Sender: ba.Sender.Address.String(),
	})
	require.NoError(ba.TB, err)

	// CreateGroup
	tssKeeper.SetGroup(ctx, tsstypes.Group{
		ID:            gid,
		Size_:         1,
		Threshold:     1,
		PubKey:        groupPubKey,
		Status:        tsstypes.GROUP_STATUS_ACTIVE,
		CreatedHeight: 1,
	})
	tssKeeper.SetGroupCount(ctx, 1)
	tssKeeper.SetDKGContext(ctx, gid, dkg)

	// Set current group in bandtss module
	bandtssKeeper.SetCurrentGroup(ctx, bandtsstypes.NewCurrentGroup(gid, ctx.BlockTime()))
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

			tx, err := bandtesting.GenSignedMockTx(
				rand.New(rand.NewSource(time.Now().UnixNano())),
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

	msg, err := bandtsstypes.NewMsgRequestSignature(content, feeLimit, sender.Address.String())
	require.NoError(ba.TB, err)

	_, err = msgSrvr.RequestSignature(ctx, msg)
	require.NoError(ba.TB, err)
}

func (ba *BenchmarkApp) AddDEs(addr sdk.AccAddress) {
	ctx, msgSrvr, k := ba.Ctx, ba.TSSMsgSrvr, ba.TSSKeeper

	deQueue := k.GetDEQueue(ctx, addr)
	count := deQueue.Tail - deQueue.Head
	if count < uint64(30) {
		_, err := msgSrvr.SubmitDEs(ctx, &tsstypes.MsgSubmitDEs{
			DEs:    GetDEs(30),
			Sender: addr.String(),
		})
		require.NoError(ba.TB, err)
	}
}
