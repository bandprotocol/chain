package benchmark

import (
	"io/ioutil"
	"testing"

	"github.com/bandprotocol/chain/v2/pkg/obi"
	"github.com/bandprotocol/chain/v2/testing/testapp"
	oracletypes "github.com/bandprotocol/chain/v2/x/oracle/types"
	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	types "github.com/tendermint/tendermint/abci/types"
)

type Account struct {
	testapp.Account
	Num uint64
	Seq uint64
}

type BenchmarkCalldata struct {
	DataSourceId uint64
	Scenario     uint64
	Value        uint64
}

func GetBenchmarkWasm() ([]byte, error) {
	oCode, err := ioutil.ReadFile("./testdata/benchmark-oracle-script.wasm")
	return oCode, err
}

func GenMsgRequestData(
	sender *Account,
	oracleScriptId uint64,
	dataSourceId uint64,
	scenario uint64,
	value uint64,
) []sdk.Msg {
	msg := oracletypes.MsgRequestData{
		OracleScriptID: oracletypes.OracleScriptID(oracleScriptId),
		Calldata: obi.MustEncode(BenchmarkCalldata{
			DataSourceId: dataSourceId,
			Scenario:     scenario,
			Value:        value,
		}),
		AskCount:   1,
		MinCount:   1,
		ClientID:   "",
		FeeLimit:   sdk.Coins{sdk.NewInt64Coin("uband", 1)},
		PrepareGas: GasLimit,
		ExecuteGas: GasLimit,
		Sender:     sender.Address.String(),
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

func GenMsgReportData(account *Account, rid uint64, eids []int64) []sdk.Msg {

	rawReports := []oracletypes.RawReport{}

	for _, eid := range eids {
		rawReports = append(rawReports, oracletypes.RawReport{
			ExternalID: oracletypes.ExternalID(eid),
			ExitCode:   0,
			Data:       []byte("empty"),
		})
	}

	msg := oracletypes.MsgReportData{
		RequestID:  oracletypes.RequestID(rid),
		RawReports: rawReports,
		Validator:  account.ValAddress.String(),
	}

	return []sdk.Msg{&msg}
}

func GenSequenceOfTxs(
	txConfig client.TxConfig,
	msgs []sdk.Msg,
	account *Account,
	numTxs int,
) []sdk.Tx {
	txs := make([]sdk.Tx, numTxs)

	for i := 0; i < numTxs; i++ {
		txs[i], _ = testapp.GenTx(
			txConfig,
			msgs,
			sdk.Coins{sdk.NewInt64Coin("uband", 1)},
			1000000000,
			"",
			[]uint64{account.Num},
			[]uint64{account.Seq},
			account.PrivKey,
		)
		account.Seq += 1
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
			attrs[string(attributes.Key)] = string(attributes.Value)
		}
		evs = append(evs, Event{
			Type:       event.Type,
			Attributes: attrs,
		})
	}

	return evs
}

func LogEvents(b testing.TB, events []types.Event) {
	evs := DecodeEvents(events)
	for i, ev := range evs {
		b.Logf("Event %d: %+v\n", i, ev)
	}

	if len(evs) == 0 {
		b.Logf("No Event")
	}
}
