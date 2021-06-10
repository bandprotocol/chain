package emitter

import (
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/23-commitment/types"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/light-clients/07-tendermint/types"
	"github.com/stretchr/testify/suite"

	"github.com/bandprotocol/chain/hooks/common"
	ibctesting "github.com/bandprotocol/chain/testing"
	"github.com/bandprotocol/chain/testing/testapp"
	oracletypes "github.com/bandprotocol/chain/x/oracle/types"
)

var (
	SenderAddress   = sdk.AccAddress(genAddresFromString("Sender"))
	ValAddress      = sdk.ValAddress(genAddresFromString("Validator"))
	TreasuryAddress = sdk.AccAddress(genAddresFromString("Treasury"))
	OwnerAddress    = sdk.AccAddress(genAddresFromString("Owner"))
	ReporterAddress = sdk.AccAddress(genAddresFromString("Reporter"))
	Signer          = sdk.AccAddress(genAddresFromString("Signer"))

	clientHeight = clienttypes.NewHeight(0, 10)
)

type DecoderTestSuite struct {
	suite.Suite

	coordinator *ibctesting.Coordinator

	chainA *ibctesting.TestChain
	chainB *ibctesting.TestChain
}

func (suite *DecoderTestSuite) SetupTest() {
	suite.coordinator = ibctesting.NewCoordinator(suite.T(), 2)
	suite.chainA = suite.coordinator.GetChain(ibctesting.GetChainID(0))
	suite.chainB = suite.coordinator.GetChain(ibctesting.GetChainID(1))
}

func genAddresFromString(s string) []byte {
	var b [20]byte
	copy(b[:], s)
	return b[:]
}

func (suite *DecoderTestSuite) testCompareJson(msg common.JsDict, expect string) {
	res, _ := json.Marshal(msg)
	suite.Require().Equal(expect, string(res))
}

func (suite *DecoderTestSuite) TestDecodeMsgRequestData() {
	msgJson := make(common.JsDict)
	msg := oracletypes.NewMsgRequestData(1, []byte("calldata"), 1, 1, "cleint_id", testapp.Coins100000000uband, testapp.TestDefaultPrepareGas, testapp.TestDefaultExecuteGas, SenderAddress)
	decodeMsgRequestData(msg, msgJson)
	suite.testCompareJson(msgJson,
		"{\"ask_count\":1,\"calldata\":\"Y2FsbGRhdGE=\",\"client_id\":\"cleint_id\",\"execute_gas\":300000,\"fee_limit\":[{\"denom\":\"uband\",\"amount\":\"100000000\"}],\"min_count\":1,\"oracle_script_id\":1,\"prepare_gas\":40000,\"sender\":\"band12djkuer9wgqqqqqqqqqqqqqqqqqqqqqqck96t0\"}",
	)
}

func (suite *DecoderTestSuite) TestDecodeReportData() {
	msgJson := make(common.JsDict)
	msg := oracletypes.NewMsgReportData(1, []oracletypes.RawReport{{1, 1, []byte("data1")}, {2, 2, []byte("data2")}}, ValAddress, ReporterAddress)
	decodeMsgReportData(msg, msgJson)
	suite.testCompareJson(msgJson,
		"{\"raw_reports\":[{\"external_id\":1,\"exit_code\":1,\"data\":\"ZGF0YTE=\"},{\"external_id\":2,\"exit_code\":2,\"data\":\"ZGF0YTI=\"}],\"reporter\":\"band12fjhqmmjw3jhyqqqqqqqqqqqqqqqqqqqjfy83g\",\"request_id\":1,\"validator\":\"bandvaloper12eskc6tyv96x7usqqqqqqqqqqqqqqqqqw09xqg\"}",
	)
}

func (suite *DecoderTestSuite) TestDecodeMsgCreateDataSource() {
	msgJson := make(common.JsDict)
	msg := oracletypes.NewMsgCreateDataSource("name", "desc", []byte("exec"), testapp.Coins1000000uband, TreasuryAddress, OwnerAddress, SenderAddress)
	decodeMsgCreateDataSource(msg, msgJson)
	suite.testCompareJson(msgJson,
		"{\"description\":\"desc\",\"executable\":\"ZXhlYw==\",\"fee\":[{\"denom\":\"uband\",\"amount\":\"1000000\"}],\"name\":\"name\",\"owner\":\"band1famkuetjqqqqqqqqqqqqqqqqqqqqqqqqkzrxfg\",\"sender\":\"band12djkuer9wgqqqqqqqqqqqqqqqqqqqqqqck96t0\",\"treasury\":\"band123ex2ctnw4e8jqqqqqqqqqqqqqqqqqqqrmzwp0\"}",
	)
}

func (suite *DecoderTestSuite) TestDecodeCreateOracleScript() {
	msgJson := make(common.JsDict)
	msg := oracletypes.NewMsgCreateOracleScript("name", "desc", "schema", "url", []byte("code"), OwnerAddress, SenderAddress)
	decodeMsgCreateOracleScript(msg, msgJson)
	suite.testCompareJson(msgJson,
		"{\"code\":\"Y29kZQ==\",\"description\":\"desc\",\"name\":\"name\",\"owner\":\"band1famkuetjqqqqqqqqqqqqqqqqqqqqqqqqkzrxfg\",\"schema\":\"schema\",\"sender\":\"band12djkuer9wgqqqqqqqqqqqqqqqqqqqqqqck96t0\",\"source_code_url\":\"url\"}",
	)
}

func (suite *DecoderTestSuite) TestDecodeMsgEditDataSource() {
	msgJson := make(common.JsDict)
	msg := oracletypes.NewMsgEditDataSource(1, "name", "desc", []byte("exec"), testapp.Coins1000000uband, TreasuryAddress, OwnerAddress, SenderAddress)
	decodeMsgEditDataSource(msg, msgJson)
	suite.testCompareJson(msgJson,
		"{\"data_source_id\":1,\"description\":\"desc\",\"executable\":\"ZXhlYw==\",\"fee\":[{\"denom\":\"uband\",\"amount\":\"1000000\"}],\"name\":\"name\",\"owner\":\"band1famkuetjqqqqqqqqqqqqqqqqqqqqqqqqkzrxfg\",\"sender\":\"band12djkuer9wgqqqqqqqqqqqqqqqqqqqqqqck96t0\",\"treasury\":\"band123ex2ctnw4e8jqqqqqqqqqqqqqqqqqqqrmzwp0\"}",
	)
}

func (suite *DecoderTestSuite) TestDecodeMsgEditOracleScript() {
	msgJson := make(common.JsDict)
	msg := oracletypes.NewMsgEditOracleScript(1, "name", "desc", "schema", "url", []byte("code"), OwnerAddress, SenderAddress)
	decodeMsgEditOracleScript(msg, msgJson)
	suite.testCompareJson(msgJson,
		"{\"code\":\"Y29kZQ==\",\"description\":\"desc\",\"name\":\"name\",\"oracle_script_id\":1,\"owner\":\"band1famkuetjqqqqqqqqqqqqqqqqqqqqqqqqkzrxfg\",\"schema\":\"schema\",\"sender\":\"band12djkuer9wgqqqqqqqqqqqqqqqqqqqqqqck96t0\",\"source_code_url\":\"url\"}",
	)
}

func (suite *DecoderTestSuite) TestDecodeMsgAddReporter() {
	msgJson := make(common.JsDict)
	msg := oracletypes.NewMsgAddReporter(ValAddress, ReporterAddress)
	decodeMsgAddReporter(msg, msgJson)
	suite.testCompareJson(msgJson,
		"{\"reporter\":\"band12fjhqmmjw3jhyqqqqqqqqqqqqqqqqqqqjfy83g\",\"validator\":\"bandvaloper12eskc6tyv96x7usqqqqqqqqqqqqqqqqqw09xqg\"}",
	)
}

func (suite *DecoderTestSuite) TestDecodeMsgRemoveReporter() {
	msgJson := make(common.JsDict)
	msg := oracletypes.NewMsgRemoveReporter(ValAddress, ReporterAddress)
	decodeMsgRemoveReporter(msg, msgJson)
	suite.testCompareJson(msgJson,
		"{\"reporter\":\"band12fjhqmmjw3jhyqqqqqqqqqqqqqqqqqqqjfy83g\",\"validator\":\"bandvaloper12eskc6tyv96x7usqqqqqqqqqqqqqqqqqw09xqg\"}",
	)
}

func (suite *DecoderTestSuite) TestDecodeMsgActivate() {
	msgJson := make(common.JsDict)
	msg := oracletypes.NewMsgActivate(ValAddress)
	decodeMsgActivate(msg, msgJson)
	suite.testCompareJson(msgJson,
		"{\"validator\":\"bandvaloper12eskc6tyv96x7usqqqqqqqqqqqqqqqqqw09xqg\"}",
	)
}

func (suite *DecoderTestSuite) TestDecodeMsgCreateClient() {
	msgJson := make(common.JsDict)
	consensus := suite.chainA.CurrentTMClientHeader().ConsensusState()
	b64RootHash := b64.StdEncoding.EncodeToString(consensus.Root.Hash)
	tendermintClient := ibctmtypes.NewClientState(suite.chainA.ChainID, ibctesting.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, clientHeight, commitmenttypes.GetSDKSpecs(), ibctesting.UpgradePath, false, false)
	msg, _ := clienttypes.NewMsgCreateClient(tendermintClient, consensus, SenderAddress)
	decodeMsgCreateClient(msg, msgJson)
	suite.testCompareJson(msgJson,
		fmt.Sprintf(
			"{\"client_state\":{\"chain_id\":\"testchain0\",\"trust_level\":{\"numerator\":1,\"denominator\":3},\"trusting_period\":1209600000000000,\"unbonding_period\":1814400000000000,\"max_clock_drift\":10000000000,\"frozen_height\":{},\"latest_height\":{\"revision_height\":10},\"proof_specs\":[{\"leaf_spec\":{\"hash\":1,\"prehash_value\":1,\"length\":1,\"prefix\":\"AA==\"},\"inner_spec\":{\"child_order\":[0,1],\"child_size\":33,\"min_prefix_length\":4,\"max_prefix_length\":12,\"hash\":1}},{\"leaf_spec\":{\"hash\":1,\"prehash_value\":1,\"length\":1,\"prefix\":\"AA==\"},\"inner_spec\":{\"child_order\":[0,1],\"child_size\":32,\"min_prefix_length\":1,\"max_prefix_length\":1,\"hash\":1}}],\"upgrade_path\":[\"upgrade\",\"upgradedIBCState\"]},\"consensus_state\":{\"timestamp\":\"2020-01-02T00:00:00Z\",\"root\":{\"hash\":\"%s\"},\"next_validators_hash\":\"%s\"},\"signer\":\"band12djkuer9wgqqqqqqqqqqqqqqqqqqqqqqck96t0\"}",
			b64RootHash,
			consensus.NextValidatorsHash),
	)
	// msgCreateClient example
	// {"client_state":{"chain_id":"testchain0","trust_level":{"numerator":1,"denominator":3},"trusting_period":1209600000000000,"unbonding_period":1814400000000000,"max_clock_drift":10000000000,"frozen_height":{},"latest_height":{"revision_height":10},"proof_specs":[{"leaf_spec":{"hash":1,"prehash_value":1,"length":1,"prefix":"AA=="},"inner_spec":{"child_order":[0,1],"child_size":33,"min_prefix_length":4,"max_prefix_length":12,"hash":1}},{"leaf_spec":{"hash":1,"prehash_value":1,"length":1,"prefix":"AA=="},"inner_spec":{"child_order":[0,1],"child_size":32,"min_prefix_length":1,"max_prefix_length":1,"hash":1}}],"upgrade_path":["upgrade","upgradedIBCState"]},"consensus_state":{"timestamp":"2020-01-02T00:00:00Z","root":{"hash":"I0ofcG04FYhAyDFzygf8Q/6JEpBactgfhm68fSXwBro="},"next_validators_hash":"C8277795F71B45089E58F0994DCF4F88BECD5770C7E492A9A25B706888D6BF2F"},"signer":"band12djkuer9wgqqqqqqqqqqqqqqqqqqqqqqck96t0"}
}

func TestDecoderTestSuite(t *testing.T) {
	suite.Run(t, new(DecoderTestSuite))
}
