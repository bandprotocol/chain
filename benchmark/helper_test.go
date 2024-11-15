package benchmark

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	types "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cmttypes "github.com/cometbft/cometbft/types"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	owasm "github.com/bandprotocol/go-owasm/api"

	"github.com/bandprotocol/chain/v3/pkg/obi"
	"github.com/bandprotocol/chain/v3/pkg/tss"
	bandtesting "github.com/bandprotocol/chain/v3/testing"
	bandtsstypes "github.com/bandprotocol/chain/v3/x/bandtss/types"
	feedstypes "github.com/bandprotocol/chain/v3/x/feeds/types"
	oracletypes "github.com/bandprotocol/chain/v3/x/oracle/types"
	tsstypes "github.com/bandprotocol/chain/v3/x/tss/types"
)

type Account struct {
	bandtesting.Account
	Num uint64
	Seq uint64
}

type BenchmarkCalldata struct {
	DataSourceID uint64
	Scenario     uint64
	Value        uint64
	Text         string
}

func GetBenchmarkWasm() ([]byte, error) {
	oCode, err := os.ReadFile("./testdata/benchmark-oracle-script.wasm")
	return oCode, err
}

func GenMsgRequestData(
	sender *Account,
	oracleScriptID uint64,
	dataSourceID uint64,
	scenario uint64,
	value uint64,
	stringLength int,
	prepareGas uint64,
	executeGas uint64,
) []sdk.Msg {
	msg := oracletypes.MsgRequestData{
		OracleScriptID: oracletypes.OracleScriptID(oracleScriptID),
		Calldata: obi.MustEncode(BenchmarkCalldata{
			DataSourceID: dataSourceID,
			Scenario:     scenario,
			Value:        value,
			Text:         strings.Repeat("#", stringLength),
		}),
		AskCount:   1,
		MinCount:   1,
		ClientID:   "",
		FeeLimit:   sdk.Coins{sdk.NewInt64Coin("uband", 1)},
		PrepareGas: prepareGas,
		ExecuteGas: executeGas,
		Sender:     sender.Address.String(),
	}

	return []sdk.Msg{&msg}
}

func GenMsgSubmitSignalPrices(
	sender *Account,
	feeds []feedstypes.Feed,
	timestamp int64,
) []sdk.Msg {
	prices := make([]feedstypes.SignalPrice, 0, len(feeds))
	for _, feed := range feeds {
		prices = append(prices, feedstypes.SignalPrice{
			Status:   feedstypes.SIGNAL_PRICE_STATUS_AVAILABLE,
			SignalID: feed.SignalID,
			Price:    60000,
		})
	}

	msg := feedstypes.NewMsgSubmitSignalPrices(sender.ValAddress.String(), timestamp, prices)

	return []sdk.Msg{msg}
}

func GenMsgSend(
	sender *Account,
	receiver *Account,
) []sdk.Msg {
	msg := banktypes.MsgSend{
		FromAddress: sender.Address.String(),
		ToAddress:   receiver.Address.String(),
		Amount:      sdk.Coins{sdk.NewInt64Coin("uband", 1)},
	}

	return []sdk.Msg{&msg}
}

func GenMsgCreateOracleScript(sender *Account, code []byte) []sdk.Msg {
	msg := oracletypes.MsgCreateOracleScript{
		Name:          "test",
		Description:   "test",
		Schema:        "test",
		SourceCodeURL: "test",
		Code:          code,
		Owner:         sender.Address.String(),
		Sender:        sender.Address.String(),
	}

	return []sdk.Msg{&msg}
}

func GenMsgCreateDataSource(sender *Account, code []byte) []sdk.Msg {
	msg := oracletypes.MsgCreateDataSource{
		Name:        "test",
		Description: "test",
		Executable:  code,
		Fee:         sdk.Coins{},
		Treasury:    sender.Address.String(),
		Owner:       sender.Address.String(),
		Sender:      sender.Address.String(),
	}

	return []sdk.Msg{&msg}
}

func GenMsgActivate(account *Account) []sdk.Msg {
	msg := oracletypes.MsgActivate{
		Validator: account.ValAddress.String(),
	}

	return []sdk.Msg{&msg}
}

func MockByte(n int) []byte {
	msg := make([]byte, n)
	for i := 0; i < n; i++ {
		msg[i] = 'a' + byte(i%26)
	}
	return msg
}

func GenMsgRequestSignature(
	sender *Account,
	content tsstypes.Content,
	feeLimit sdk.Coins,
) []sdk.Msg {
	msg, err := bandtsstypes.NewMsgRequestSignature(content, feeLimit, sender.Address.String())
	if err != nil {
		panic(err)
	}
	return []sdk.Msg{msg}
}

func GenMsgSubmitSignature(sid tss.SigningID, mid tss.MemberID, sig tss.Signature, member sdk.AccAddress) []sdk.Msg {
	msg := tsstypes.MsgSubmitSignature{
		SigningID: sid,
		MemberID:  mid,
		Signature: sig,
		Signer:    member.String(),
	}

	return []sdk.Msg{&msg}
}

func CreateSignature(
	mid tss.MemberID,
	signing tsstypes.Signing,
	assignedMembers tsstypes.AssignedMembers,
	groupPubKey tss.Point,
	ownPrivKey tss.Scalar,
) (tss.Signature, error) {
	// Compute Lagrange coefficient
	lgc, err := tss.ComputeLagrangeCoefficient(
		mid,
		assignedMembers.MemberIDs(),
	)
	if err != nil {
		return nil, err
	}

	for _, am := range assignedMembers {
		if am.MemberID == mid {
			// Compute private nonce
			pn, err := tss.ComputeOwnPrivNonce(PrivD, PrivE, am.BindingFactor)
			if err != nil {
				return nil, err
			}
			// Sign the message
			return tss.SignSigning(
				signing.GroupPubNonce,
				groupPubKey,
				signing.Message,
				lgc,
				pn,
				ownPrivKey,
			)
		}
	}

	return nil, fmt.Errorf("this member is not assigned members")
}

func GenSequenceOfTxs(
	txEncoder sdk.TxEncoder,
	txConfig client.TxConfig,
	msgs []sdk.Msg,
	account *Account,
	numTxs int,
) [][]byte {
	txs := make([][]byte, numTxs)

	for i := 0; i < numTxs; i++ {
		tx, _ := bandtesting.GenSignedMockTx(
			rand.New(rand.NewSource(time.Now().UnixNano())),
			txConfig,
			msgs,
			sdk.Coins{sdk.NewInt64Coin("uband", 1)},
			math.MaxInt64,
			bandtesting.ChainID,
			[]uint64{account.Num},
			[]uint64{account.Seq},
			account.PrivKey,
		)

		txs[i], _ = txEncoder(tx)
		account.Seq++
	}

	return txs
}

type Event struct {
	Type       string
	Attributes map[string]string
}

func DecodeEvents(events []types.Event) []Event {
	evs := []Event{}
	for _, event := range events {
		attrs := make(map[string]string, 0)
		for _, attributes := range event.Attributes {
			attrs[attributes.Key] = attributes.Value
		}
		evs = append(evs, Event{
			Type:       event.Type,
			Attributes: attrs,
		})
	}

	return evs
}

func GetFirstAttributeOfLastEventValue(events []types.Event) (int, error) {
	evt := events[len(events)-1]
	attr := evt.Attributes[0]
	value, err := strconv.Atoi(attr.Value)

	return value, err
}

func InitOwasmTestEnv(
	tb testing.TB,
	cacheSize uint32,
	scenario uint64,
	parameter uint64,
	stringLength int,
) (*owasm.Vm, []byte, oracletypes.Request) {
	// prepare owasm vm
	owasmVM, err := owasm.NewVm(cacheSize)
	require.NoError(tb, err)

	// prepare owasm code
	oCode, err := GetBenchmarkWasm()
	require.NoError(tb, err)
	compiledCode, err := owasmVM.Compile(oCode, oracletypes.MaxCompiledWasmCodeSize)
	require.NoError(tb, err)

	// prepare request
	req := oracletypes.NewRequest(
		1, obi.MustEncode(BenchmarkCalldata{
			DataSourceID: 1,
			Scenario:     scenario,
			Value:        parameter,
			Text:         strings.Repeat("#", stringLength),
		}), []sdk.ValAddress{[]byte{}}, 1,
		1, time.Now(), "", nil, nil, ExecuteGasLimit, 0,
		bandtesting.FeePayer.Address.String(),
		bandtesting.Coins100000000uband,
	)

	return owasmVM, compiledCode, req
}

func GetConsensusParams(maxGas int64) cmtproto.ConsensusParams {
	return cmtproto.ConsensusParams{
		Block: &cmtproto.BlockParams{
			MaxBytes: 200000,
			MaxGas:   maxGas,
		},
		Evidence: &cmtproto.EvidenceParams{
			MaxAgeNumBlocks: 302400,
			MaxAgeDuration:  504 * time.Hour,
		},
		Validator: &cmtproto.ValidatorParams{
			PubKeyTypes: []string{
				cmttypes.ABCIPubKeyTypeSecp256k1,
			},
		},
	}
}

func ChunkSlice(slice []uint64, chunkSize int) [][]uint64 {
	var chunks [][]uint64
	for i := 0; i < len(slice); i += chunkSize {
		end := i + chunkSize

		// necessary check to avoid slicing beyond
		// slice capacity
		if end > len(slice) {
			end = len(slice)
		}

		chunks = append(chunks, slice[i:end])
	}

	return chunks
}

func GenOracleReports() []oracletypes.Report {
	return []oracletypes.Report{
		{
			Validator:       "",
			InBeforeResolve: true,
			RawReports: []oracletypes.RawReport{
				{
					ExternalID: 0,
					ExitCode:   0,
					Data:       []byte{},
				},
			},
		},
	}
}

func GetSpanSize() uint64 {
	if oracletypes.DefaultMaxReportDataSize > oracletypes.DefaultMaxCalldataSize {
		return oracletypes.DefaultMaxReportDataSize
	}
	return oracletypes.DefaultMaxCalldataSize
}
