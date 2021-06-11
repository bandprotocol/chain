package emitter

import (
	b64 "encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/23-commitment/types"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/light-clients/07-tendermint/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/suite"

	"github.com/bandprotocol/chain/hooks/common"
	ibctesting "github.com/bandprotocol/chain/testing"
	"github.com/bandprotocol/chain/testing/testapp"
	oracletypes "github.com/bandprotocol/chain/x/oracle/types"
)

var (
	SenderAddress    = sdk.AccAddress(genAddresFromString("Sender"))
	ValAddress       = sdk.ValAddress(genAddresFromString("Validator"))
	TreasuryAddress  = sdk.AccAddress(genAddresFromString("Treasury"))
	OwnerAddress     = sdk.AccAddress(genAddresFromString("Owner"))
	ReporterAddress  = sdk.AccAddress(genAddresFromString("Reporter"))
	Signer           = sdk.AccAddress(genAddresFromString("Signer"))
	DelegatorAddress = sdk.AccAddress(genAddresFromString("Delegator"))

	clientHeight = clienttypes.NewHeight(0, 10)

	content = govtypes.ContentFromProposalType("Title", "Desc", "Text")

	Delegation        = stakingtypes.NewDelegation(DelegatorAddress, ValAddress, sdk.NewDec(1))
	SelfDelegation    = sdk.NewCoin("uband", sdk.NewInt(1))
	MinSelfDelegation = sdk.NewInt(1)
	Description       = stakingtypes.NewDescription("moniker", "identity", "website", "securityContact", "details")
	CommissionRate    = stakingtypes.NewCommissionRates(sdk.NewDec(1), sdk.NewDec(5), sdk.NewDec(5))
	NewRate           = sdk.NewDec(1)
	PubKey            = newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB50")
	Amount            = sdk.NewCoin("uband", sdk.NewInt(1))
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

func newPubKey(pk string) (res cryptotypes.PubKey) {
	pkBytes, err := hex.DecodeString(pk)
	if err != nil {
		panic(err)
	}

	pubkey := &ed25519.PubKey{Key: pkBytes}

	return pubkey
}

func (suite *DecoderTestSuite) testCompareJson(msg common.JsDict, expect string) {
	res, _ := json.Marshal(msg)
	suite.Require().Equal(expect, string(res))
}

func (suite *DecoderTestSuite) TestDecodeMsgRequestData() {
	detail := make(common.JsDict)
	msg := oracletypes.NewMsgRequestData(1, []byte("calldata"), 1, 1, "cleint_id", testapp.Coins100000000uband, testapp.TestDefaultPrepareGas, testapp.TestDefaultExecuteGas, SenderAddress)
	decodeMsgRequestData(msg, detail)
	suite.testCompareJson(detail,
		"{\"ask_count\":1,\"calldata\":\"Y2FsbGRhdGE=\",\"client_id\":\"cleint_id\",\"execute_gas\":300000,\"fee_limit\":[{\"denom\":\"uband\",\"amount\":\"100000000\"}],\"min_count\":1,\"oracle_script_id\":1,\"prepare_gas\":40000,\"sender\":\"band12djkuer9wgqqqqqqqqqqqqqqqqqqqqqqck96t0\"}",
	)
}

func (suite *DecoderTestSuite) TestDecodeReportData() {
	detail := make(common.JsDict)
	msg := oracletypes.NewMsgReportData(1, []oracletypes.RawReport{{1, 1, []byte("data1")}, {2, 2, []byte("data2")}}, ValAddress, ReporterAddress)
	decodeMsgReportData(msg, detail)
	suite.testCompareJson(detail,
		"{\"raw_reports\":[{\"external_id\":1,\"exit_code\":1,\"data\":\"ZGF0YTE=\"},{\"external_id\":2,\"exit_code\":2,\"data\":\"ZGF0YTI=\"}],\"reporter\":\"band12fjhqmmjw3jhyqqqqqqqqqqqqqqqqqqqjfy83g\",\"request_id\":1,\"validator\":\"bandvaloper12eskc6tyv96x7usqqqqqqqqqqqqqqqqqw09xqg\"}",
	)
}

func (suite *DecoderTestSuite) TestDecodeMsgCreateDataSource() {
	detail := make(common.JsDict)
	msg := oracletypes.NewMsgCreateDataSource("name", "desc", []byte("exec"), testapp.Coins1000000uband, TreasuryAddress, OwnerAddress, SenderAddress)
	decodeMsgCreateDataSource(msg, detail)
	suite.testCompareJson(detail,
		"{\"description\":\"desc\",\"executable\":\"ZXhlYw==\",\"fee\":[{\"denom\":\"uband\",\"amount\":\"1000000\"}],\"name\":\"name\",\"owner\":\"band1famkuetjqqqqqqqqqqqqqqqqqqqqqqqqkzrxfg\",\"sender\":\"band12djkuer9wgqqqqqqqqqqqqqqqqqqqqqqck96t0\",\"treasury\":\"band123ex2ctnw4e8jqqqqqqqqqqqqqqqqqqqrmzwp0\"}",
	)
}

func (suite *DecoderTestSuite) TestDecodeCreateOracleScript() {
	detail := make(common.JsDict)
	msg := oracletypes.NewMsgCreateOracleScript("name", "desc", "schema", "url", []byte("code"), OwnerAddress, SenderAddress)
	decodeMsgCreateOracleScript(msg, detail)
	suite.testCompareJson(detail,
		"{\"code\":\"Y29kZQ==\",\"description\":\"desc\",\"name\":\"name\",\"owner\":\"band1famkuetjqqqqqqqqqqqqqqqqqqqqqqqqkzrxfg\",\"schema\":\"schema\",\"sender\":\"band12djkuer9wgqqqqqqqqqqqqqqqqqqqqqqck96t0\",\"source_code_url\":\"url\"}",
	)
}

func (suite *DecoderTestSuite) TestDecodeMsgEditDataSource() {
	detail := make(common.JsDict)
	msg := oracletypes.NewMsgEditDataSource(1, "name", "desc", []byte("exec"), testapp.Coins1000000uband, TreasuryAddress, OwnerAddress, SenderAddress)
	decodeMsgEditDataSource(msg, detail)
	suite.testCompareJson(detail,
		"{\"data_source_id\":1,\"description\":\"desc\",\"executable\":\"ZXhlYw==\",\"fee\":[{\"denom\":\"uband\",\"amount\":\"1000000\"}],\"name\":\"name\",\"owner\":\"band1famkuetjqqqqqqqqqqqqqqqqqqqqqqqqkzrxfg\",\"sender\":\"band12djkuer9wgqqqqqqqqqqqqqqqqqqqqqqck96t0\",\"treasury\":\"band123ex2ctnw4e8jqqqqqqqqqqqqqqqqqqqrmzwp0\"}",
	)
}

func (suite *DecoderTestSuite) TestDecodeMsgEditOracleScript() {
	detail := make(common.JsDict)
	msg := oracletypes.NewMsgEditOracleScript(1, "name", "desc", "schema", "url", []byte("code"), OwnerAddress, SenderAddress)
	decodeMsgEditOracleScript(msg, detail)
	suite.testCompareJson(detail,
		"{\"code\":\"Y29kZQ==\",\"description\":\"desc\",\"name\":\"name\",\"oracle_script_id\":1,\"owner\":\"band1famkuetjqqqqqqqqqqqqqqqqqqqqqqqqkzrxfg\",\"schema\":\"schema\",\"sender\":\"band12djkuer9wgqqqqqqqqqqqqqqqqqqqqqqck96t0\",\"source_code_url\":\"url\"}",
	)
}

func (suite *DecoderTestSuite) TestDecodeMsgAddReporter() {
	detail := make(common.JsDict)
	msg := oracletypes.NewMsgAddReporter(ValAddress, ReporterAddress)
	decodeMsgAddReporter(msg, detail)
	suite.testCompareJson(detail,
		"{\"reporter\":\"band12fjhqmmjw3jhyqqqqqqqqqqqqqqqqqqqjfy83g\",\"validator\":\"bandvaloper12eskc6tyv96x7usqqqqqqqqqqqqqqqqqw09xqg\"}",
	)
}

func (suite *DecoderTestSuite) TestDecodeMsgRemoveReporter() {
	detail := make(common.JsDict)
	msg := oracletypes.NewMsgRemoveReporter(ValAddress, ReporterAddress)
	decodeMsgRemoveReporter(msg, detail)
	suite.testCompareJson(detail,
		"{\"reporter\":\"band12fjhqmmjw3jhyqqqqqqqqqqqqqqqqqqqjfy83g\",\"validator\":\"bandvaloper12eskc6tyv96x7usqqqqqqqqqqqqqqqqqw09xqg\"}",
	)
}

func (suite *DecoderTestSuite) TestDecodeMsgActivate() {
	detail := make(common.JsDict)
	msg := oracletypes.NewMsgActivate(ValAddress)
	decodeMsgActivate(msg, detail)
	suite.testCompareJson(detail,
		"{\"validator\":\"bandvaloper12eskc6tyv96x7usqqqqqqqqqqqqqqqqqw09xqg\"}",
	)
}

func (suite *DecoderTestSuite) TestDecodeMsgCreateClient() {
	detail := make(common.JsDict)
	consensus := suite.chainA.CurrentTMClientHeader().ConsensusState()
	b64RootHash := b64.StdEncoding.EncodeToString(consensus.Root.Hash)
	tendermintClient := ibctmtypes.NewClientState(suite.chainA.ChainID, ibctesting.DefaultTrustLevel, ibctesting.TrustingPeriod, ibctesting.UnbondingPeriod, ibctesting.MaxClockDrift, clientHeight, commitmenttypes.GetSDKSpecs(), ibctesting.UpgradePath, false, false)
	msg, _ := clienttypes.NewMsgCreateClient(tendermintClient, consensus, SenderAddress)
	decodeMsgCreateClient(msg, detail)
	suite.testCompareJson(detail,
		fmt.Sprintf(
			"{\"client_state\":{\"chain_id\":\"testchain0\",\"trust_level\":{\"numerator\":1,\"denominator\":3},\"trusting_period\":1209600000000000,\"unbonding_period\":1814400000000000,\"max_clock_drift\":10000000000,\"frozen_height\":{},\"latest_height\":{\"revision_height\":10},\"proof_specs\":[{\"leaf_spec\":{\"hash\":1,\"prehash_value\":1,\"length\":1,\"prefix\":\"AA==\"},\"inner_spec\":{\"child_order\":[0,1],\"child_size\":33,\"min_prefix_length\":4,\"max_prefix_length\":12,\"hash\":1}},{\"leaf_spec\":{\"hash\":1,\"prehash_value\":1,\"length\":1,\"prefix\":\"AA==\"},\"inner_spec\":{\"child_order\":[0,1],\"child_size\":32,\"min_prefix_length\":1,\"max_prefix_length\":1,\"hash\":1}}],\"upgrade_path\":[\"upgrade\",\"upgradedIBCState\"]},\"consensus_state\":{\"timestamp\":\"2020-01-02T00:00:00Z\",\"root\":{\"hash\":\"%s\"},\"next_validators_hash\":\"%s\"},\"signer\":\"band12djkuer9wgqqqqqqqqqqqqqqqqqqqqqqck96t0\"}",
			b64RootHash,
			consensus.NextValidatorsHash),
	)
	// msgCreateClient example
	// {"client_state":{"chain_id":"testchain0","trust_level":{"numerator":1,"denominator":3},"trusting_period":1209600000000000,"unbonding_period":1814400000000000,"max_clock_drift":10000000000,"frozen_height":{},"latest_height":{"revision_height":10},"proof_specs":[{"leaf_spec":{"hash":1,"prehash_value":1,"length":1,"prefix":"AA=="},"inner_spec":{"child_order":[0,1],"child_size":33,"min_prefix_length":4,"max_prefix_length":12,"hash":1}},{"leaf_spec":{"hash":1,"prehash_value":1,"length":1,"prefix":"AA=="},"inner_spec":{"child_order":[0,1],"child_size":32,"min_prefix_length":1,"max_prefix_length":1,"hash":1}}],"upgrade_path":["upgrade","upgradedIBCState"]},"consensus_state":{"timestamp":"2020-01-02T00:00:00Z","root":{"hash":"I0ofcG04FYhAyDFzygf8Q/6JEpBactgfhm68fSXwBro="},"next_validators_hash":"C8277795F71B45089E58F0994DCF4F88BECD5770C7E492A9A25B706888D6BF2F"},"signer":"band12djkuer9wgqqqqqqqqqqqqqqqqqqqqqqck96t0"}
}

func (suite *DecoderTestSuite) TestDecodeMsgSubmitProposal() {
	detail := make(common.JsDict)
	msg, _ := govtypes.NewMsgSubmitProposal(content, testapp.Coins1000000uband, SenderAddress)
	decodeMsgSubmitProposal(msg, detail)
	suite.testCompareJson(detail,
		"{\"content\":{\"title\":\"Title\",\"description\":\"Desc\"},\"initial_deposit\":[{\"denom\":\"uband\",\"amount\":\"1000000\"}],\"proposer\":\"band12djkuer9wgqqqqqqqqqqqqqqqqqqqqqqck96t0\"}",
	)

}

func (suite *DecoderTestSuite) TestDecodeMsgDeposit() {
	detail := make(common.JsDict)
	msg := govtypes.NewMsgDeposit(SenderAddress, 1, testapp.Coins1000000uband)
	decodeMsgDeposit(msg, detail)
	suite.testCompareJson(detail,
		"{\"amount\":[{\"denom\":\"uband\",\"amount\":\"1000000\"}],\"depositor\":\"band12djkuer9wgqqqqqqqqqqqqqqqqqqqqqqck96t0\",\"proposal_id\":1}",
	)
}

func (suite *DecoderTestSuite) TestDecodeMsgVote() {
	detail := make(common.JsDict)
	msg := govtypes.NewMsgVote(SenderAddress, 1, 0)
	decodeMsgVote(msg, detail)
	suite.testCompareJson(detail,
		"{\"option\":0,\"proposal_id\":1,\"voter\":\"band12djkuer9wgqqqqqqqqqqqqqqqqqqqqqqck96t0\"}",
	)
}

func (suite *DecoderTestSuite) TestDecodeMsgCreateValidator() {
	detail := make(common.JsDict)
	msg, _ := stakingtypes.NewMsgCreateValidator(ValAddress, PubKey, SelfDelegation, Description, CommissionRate, MinSelfDelegation)

	decodeMsgCreateValidator(msg, detail)
	suite.testCompareJson(detail,
		"{\"commission_rates\":\"1.000000000000000000\",\"delegator_address\":\"band12eskc6tyv96x7usqqqqqqqqqqqqqqqqqzep99r\",\"description\":{\"moniker\":\"moniker\",\"identity\":\"identity\",\"website\":\"website\",\"security_contact\":\"securityContact\",\"details\":\"details\"},\"min_self_delegation\":\"1\",\"pubkey\":\"bandvalconspub1zcjduepqpdy9elqwanrpj3qyfppklr7fmaq9vmerd8njgqpgz32vk4f2ldgqjrvpk6\",\"validator_address\":\"bandvaloper12eskc6tyv96x7usqqqqqqqqqqqqqqqqqw09xqg\",\"value\":{\"denom\":\"uband\",\"amount\":\"1\"}}",
	)
}

func (suite *DecoderTestSuite) TestDecodeMsgEditValidator() {
	detail := make(common.JsDict)
	msg := stakingtypes.NewMsgEditValidator(ValAddress, Description, &NewRate, &MinSelfDelegation)

	decodeMsgEditValidator(msg, detail)
	suite.testCompareJson(detail,
		"{\"commission_rates\":\"1.000000000000000000\",\"description\":{\"moniker\":\"moniker\",\"identity\":\"identity\",\"website\":\"website\",\"security_contact\":\"securityContact\",\"details\":\"details\"},\"min_self_delegation\":\"1\",\"validator_address\":\"bandvaloper12eskc6tyv96x7usqqqqqqqqqqqqqqqqqw09xqg\"}",
	)
}

func (suite *DecoderTestSuite) TestDecodeMsgDelegate() {
	detail := make(common.JsDict)
	msg := stakingtypes.NewMsgDelegate(DelegatorAddress, ValAddress, Amount)

	decodeMsgDelegate(msg, detail)
	suite.testCompareJson(detail,
		"{\"amount\":{\"denom\":\"uband\",\"amount\":\"1\"},\"delegator_address\":\"band1g3jkcet8v96x7usqqqqqqqqqqqqqqqqqus6d5g\",\"validator_address\":\"bandvaloper12eskc6tyv96x7usqqqqqqqqqqqqqqqqqw09xqg\"}",
	)
}

func (suite *DecoderTestSuite) TestDecodeMsgUndelegate() {
	detail := make(common.JsDict)
	msg := stakingtypes.NewMsgUndelegate(DelegatorAddress, ValAddress, Amount)

	decodeMsgUndelegate(msg, detail)
	suite.testCompareJson(detail,
		"{\"amount\":{\"denom\":\"uband\",\"amount\":\"1\"},\"delegator_address\":\"band1g3jkcet8v96x7usqqqqqqqqqqqqqqqqqus6d5g\",\"validator_address\":\"bandvaloper12eskc6tyv96x7usqqqqqqqqqqqqqqqqqw09xqg\"}",
	)
}

func (suite *DecoderTestSuite) TestDecodeMsgBeginRedelegate() {
	detail := make(common.JsDict)
	msg := stakingtypes.NewMsgBeginRedelegate(DelegatorAddress, ValAddress, ValAddress, Amount)

	decodeMsgBeginRedelegate(msg, detail)
	suite.testCompareJson(detail,
		"{\"amount\":{\"denom\":\"uband\",\"amount\":\"1\"},\"delegator_address\":\"band1g3jkcet8v96x7usqqqqqqqqqqqqqqqqqus6d5g\",\"validator_dst_address\":\"bandvaloper12eskc6tyv96x7usqqqqqqqqqqqqqqqqqw09xqg\",\"validator_src_address\":\"bandvaloper12eskc6tyv96x7usqqqqqqqqqqqqqqqqqw09xqg\"}",
	)
}

func TestDecoderTestSuite(t *testing.T) {
	suite.Run(t, new(DecoderTestSuite))
}
