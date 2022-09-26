package emitter_test

import (
	b64 "encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	clienttypes "github.com/cosmos/ibc-go/modules/core/02-client/types"
	connectiontypes "github.com/cosmos/ibc-go/modules/core/03-connection/types"
	channeltypes "github.com/cosmos/ibc-go/modules/core/04-channel/types"
	commitmenttypes "github.com/cosmos/ibc-go/modules/core/23-commitment/types"
	ibctmtypes "github.com/cosmos/ibc-go/modules/light-clients/07-tendermint/types"
	"github.com/stretchr/testify/suite"

	"github.com/bandprotocol/chain/v2/hooks/common"
	"github.com/bandprotocol/chain/v2/hooks/emitter"
	ibctesting "github.com/bandprotocol/chain/v2/testing"
	"github.com/bandprotocol/chain/v2/testing/testapp"
	oracletypes "github.com/bandprotocol/chain/v2/x/oracle/types"
)

var (
	SenderAddress    = sdk.AccAddress(genAddresFromString("Sender"))
	ValAddress       = sdk.ValAddress(genAddresFromString("Validator"))
	TreasuryAddress  = sdk.AccAddress(genAddresFromString("Treasury"))
	OwnerAddress     = sdk.AccAddress(genAddresFromString("Owner"))
	ReporterAddress  = sdk.AccAddress(genAddresFromString("Reporter"))
	SignerAddress    = sdk.AccAddress(genAddresFromString("Signer"))
	DelegatorAddress = sdk.AccAddress(genAddresFromString("Delegator"))
	GranterAddress   = sdk.AccAddress(genAddresFromString("Granter"))
	GranteeAddress   = sdk.AccAddress(genAddresFromString("Grantee"))

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

func NewOraclePath(chainA, chainB *ibctesting.TestChain) *ibctesting.Path {
	path := ibctesting.NewPath(chainA, chainB)
	path.EndpointA.ChannelConfig.PortID = ibctesting.OraclePort
	path.EndpointB.ChannelConfig.PortID = ibctesting.OraclePort

	return path
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

	pubkey := &secp256k1.PubKey{Key: pkBytes}

	return pubkey
}

func (suite *DecoderTestSuite) testCompareJson(msg common.JsDict, expect string) {
	res, _ := json.Marshal(msg)
	suite.Require().Equal(string(res), expect)
}

func (suite *DecoderTestSuite) testContains(msg common.JsDict, expect string) {
	res, _ := json.Marshal(msg)
	suite.Require().Contains(string(res), expect)
}

func (suite *DecoderTestSuite) TestDecodeMsgGrant() {
	detail := make(common.JsDict)
	expiration := suite.chainA.GetContext().BlockTime()

	// TestSendAuthorization
	spendLimit := sdk.NewCoins(Amount)
	sendMsg, _ := authz.NewMsgGrant(
		GranterAddress,
		GranteeAddress,
		banktypes.NewSendAuthorization(spendLimit),
		expiration,
	)
	emitter.DecodeMsgGrant(sendMsg, detail)
	suite.testCompareJson(
		detail,
		"{\"grant\":{\"authorization\":{\"spend_limit\":[{\"denom\":\"uband\",\"amount\":\"1\"}]},\"expiration\":\"2020-01-02T00:00:00Z\"},\"grantee\":\"band1gaexzmn5v4jsqqqqqqqqqqqqqqqqqqqqwrdaed\",\"granter\":\"band1gaexzmn5v4eqqqqqqqqqqqqqqqqqqqqq3urue8\"}",
	)

	// TestGenericAuthorization
	genericMsg, _ := authz.NewMsgGrant(
		GranterAddress,
		GranteeAddress,
		authz.NewGenericAuthorization(sdk.MsgTypeURL(&oracletypes.MsgReportData{})),
		expiration,
	)
	emitter.DecodeMsgGrant(genericMsg, detail)
	suite.testCompareJson(
		detail,
		"{\"grant\":{\"authorization\":{\"msg\":\"/oracle.v1.MsgReportData\"},\"expiration\":\"2020-01-02T00:00:00Z\"},\"grantee\":\"band1gaexzmn5v4jsqqqqqqqqqqqqqqqqqqqqwrdaed\",\"granter\":\"band1gaexzmn5v4eqqqqqqqqqqqqqqqqqqqqq3urue8\"}",
	)

	// TestStakeAuthorization
	stakeAuthorization, _ := stakingtypes.NewStakeAuthorization(
		[]sdk.ValAddress{ValAddress},
		[]sdk.ValAddress{},
		stakingtypes.AuthorizationType_AUTHORIZATION_TYPE_DELEGATE,
		&Amount,
	)
	stakeMsg, _ := authz.NewMsgGrant(GranterAddress, GranteeAddress, stakeAuthorization, expiration)
	emitter.DecodeMsgGrant(stakeMsg, detail)
	suite.testCompareJson(
		detail,
		"{\"grant\":{\"authorization\":{\"max_tokens\":{\"denom\":\"uband\",\"amount\":\"1\"},\"Validators\":{\"allow_list\":{\"address\":[\"bandvaloper12eskc6tyv96x7usqqqqqqqqqqqqqqqqqw09xqg\"]}},\"authorization_type\":1},\"expiration\":\"2020-01-02T00:00:00Z\"},\"grantee\":\"band1gaexzmn5v4jsqqqqqqqqqqqqqqqqqqqqwrdaed\",\"granter\":\"band1gaexzmn5v4eqqqqqqqqqqqqqqqqqqqqq3urue8\"}",
	)

}

func (suite *DecoderTestSuite) TestDecodeMsgRevoke() {
	detail := make(common.JsDict)
	msg := authz.NewMsgRevoke(GranterAddress, GranteeAddress, sdk.MsgTypeURL(&oracletypes.MsgReportData{}))
	emitter.DecodeMsgRevoke(&msg, detail)
	suite.testCompareJson(
		detail,
		"{\"grantee\":\"band1gaexzmn5v4jsqqqqqqqqqqqqqqqqqqqqwrdaed\",\"granter\":\"band1gaexzmn5v4eqqqqqqqqqqqqqqqqqqqqq3urue8\",\"msg_type_url\":\"/oracle.v1.MsgReportData\"}",
	)
}

func (suite *DecoderTestSuite) TestDecodeMsgExec() {
	detail := make(common.JsDict)
	msg := authz.NewMsgExec(GranteeAddress, []sdk.Msg{
		&banktypes.MsgSend{
			Amount:      sdk.NewCoins(Amount),
			FromAddress: GranterAddress.String(),
			ToAddress:   GranteeAddress.String(),
		},
		&oracletypes.MsgReportData{},
	})
	emitter.DecodeMsgExec(&msg, detail)
	suite.testCompareJson(
		detail,
		"{\"grantee\":\"band1gaexzmn5v4jsqqqqqqqqqqqqqqqqqqqqwrdaed\",\"msgs\":[{\"msg\":{\"amount\":[{\"denom\":\"uband\",\"amount\":\"1\"}],\"from_address\":\"band1gaexzmn5v4eqqqqqqqqqqqqqqqqqqqqq3urue8\",\"to_address\":\"band1gaexzmn5v4jsqqqqqqqqqqqqqqqqqqqqwrdaed\"},\"type\":\"/cosmos.bank.v1beta1.MsgSend\"},{\"msg\":{\"raw_reports\":null,\"request_id\":0,\"validator\":\"\"},\"type\":\"/oracle.v1.MsgReportData\"}]}",
	)

}

func (suite *DecoderTestSuite) TestDecodeMsgRequestData() {
	detail := make(common.JsDict)
	msg := oracletypes.NewMsgRequestData(
		1,
		[]byte("calldata"),
		1,
		1,
		"cleint_id",
		testapp.Coins100000000uband,
		testapp.TestDefaultPrepareGas,
		testapp.TestDefaultExecuteGas,
		SenderAddress,
	)
	emitter.DecodeMsgRequestData(msg, detail)
	suite.testCompareJson(
		detail,
		"{\"ask_count\":1,\"calldata\":\"Y2FsbGRhdGE=\",\"client_id\":\"cleint_id\",\"execute_gas\":300000,\"fee_limit\":[{\"denom\":\"uband\",\"amount\":\"100000000\"}],\"min_count\":1,\"oracle_script_id\":1,\"prepare_gas\":40000,\"sender\":\"band12djkuer9wgqqqqqqqqqqqqqqqqqqqqqqck96t0\"}",
	)
}

func (suite *DecoderTestSuite) TestDecodeReportData() {
	detail := make(common.JsDict)
	msg := oracletypes.NewMsgReportData(
		1,
		[]oracletypes.RawReport{{1, 1, []byte("data1")}, {2, 2, []byte("data2")}},
		ValAddress,
	)
	emitter.DecodeMsgReportData(msg, detail)
	suite.testCompareJson(
		detail,
		"{\"raw_reports\":[{\"external_id\":1,\"exit_code\":1,\"data\":\"ZGF0YTE=\"},{\"external_id\":2,\"exit_code\":2,\"data\":\"ZGF0YTI=\"}],\"request_id\":1,\"validator\":\"bandvaloper12eskc6tyv96x7usqqqqqqqqqqqqqqqqqw09xqg\"}",
	)
}

func (suite *DecoderTestSuite) TestDecodeMsgCreateDataSource() {
	detail := make(common.JsDict)
	msg := oracletypes.NewMsgCreateDataSource(
		"name",
		"desc",
		[]byte("exec"),
		testapp.Coins1000000uband,
		TreasuryAddress,
		OwnerAddress,
		SenderAddress,
	)
	emitter.DecodeMsgCreateDataSource(msg, detail)
	suite.testCompareJson(
		detail,
		"{\"description\":\"desc\",\"executable\":\"ZXhlYw==\",\"fee\":[{\"denom\":\"uband\",\"amount\":\"1000000\"}],\"name\":\"name\",\"owner\":\"band1famkuetjqqqqqqqqqqqqqqqqqqqqqqqqkzrxfg\",\"sender\":\"band12djkuer9wgqqqqqqqqqqqqqqqqqqqqqqck96t0\",\"treasury\":\"band123ex2ctnw4e8jqqqqqqqqqqqqqqqqqqqrmzwp0\"}",
	)
}

func (suite *DecoderTestSuite) TestDecodeCreateOracleScript() {
	detail := make(common.JsDict)
	msg := oracletypes.NewMsgCreateOracleScript(
		"name",
		"desc",
		"schema",
		"url",
		[]byte("code"),
		OwnerAddress,
		SenderAddress,
	)
	emitter.DecodeMsgCreateOracleScript(msg, detail)
	suite.testCompareJson(
		detail,
		"{\"code\":\"Y29kZQ==\",\"description\":\"desc\",\"name\":\"name\",\"owner\":\"band1famkuetjqqqqqqqqqqqqqqqqqqqqqqqqkzrxfg\",\"schema\":\"schema\",\"sender\":\"band12djkuer9wgqqqqqqqqqqqqqqqqqqqqqqck96t0\",\"source_code_url\":\"url\"}",
	)
}

func (suite *DecoderTestSuite) TestDecodeMsgEditDataSource() {
	detail := make(common.JsDict)
	msg := oracletypes.NewMsgEditDataSource(
		1,
		"name",
		"desc",
		[]byte("exec"),
		testapp.Coins1000000uband,
		TreasuryAddress,
		OwnerAddress,
		SenderAddress,
	)
	emitter.DecodeMsgEditDataSource(msg, detail)
	suite.testCompareJson(
		detail,
		"{\"data_source_id\":1,\"description\":\"desc\",\"executable\":\"ZXhlYw==\",\"fee\":[{\"denom\":\"uband\",\"amount\":\"1000000\"}],\"name\":\"name\",\"owner\":\"band1famkuetjqqqqqqqqqqqqqqqqqqqqqqqqkzrxfg\",\"sender\":\"band12djkuer9wgqqqqqqqqqqqqqqqqqqqqqqck96t0\",\"treasury\":\"band123ex2ctnw4e8jqqqqqqqqqqqqqqqqqqqrmzwp0\"}",
	)
}

func (suite *DecoderTestSuite) TestDecodeMsgEditOracleScript() {
	detail := make(common.JsDict)
	msg := oracletypes.NewMsgEditOracleScript(
		1,
		"name",
		"desc",
		"schema",
		"url",
		[]byte("code"),
		OwnerAddress,
		SenderAddress,
	)
	emitter.DecodeMsgEditOracleScript(msg, detail)
	suite.testCompareJson(
		detail,
		"{\"code\":\"Y29kZQ==\",\"description\":\"desc\",\"name\":\"name\",\"oracle_script_id\":1,\"owner\":\"band1famkuetjqqqqqqqqqqqqqqqqqqqqqqqqkzrxfg\",\"schema\":\"schema\",\"sender\":\"band12djkuer9wgqqqqqqqqqqqqqqqqqqqqqqck96t0\",\"source_code_url\":\"url\"}",
	)
}

func (suite *DecoderTestSuite) TestDecodeMsgActivate() {
	detail := make(common.JsDict)
	msg := oracletypes.NewMsgActivate(ValAddress)
	emitter.DecodeMsgActivate(msg, detail)
	suite.testCompareJson(detail,
		"{\"validator\":\"bandvaloper12eskc6tyv96x7usqqqqqqqqqqqqqqqqqw09xqg\"}",
	)
}

func (suite *DecoderTestSuite) TestDecodeMsgCreateClient() {
	detail := make(common.JsDict)
	consensus := suite.chainA.CurrentTMClientHeader().ConsensusState()
	b64RootHash := b64.StdEncoding.EncodeToString(consensus.Root.Hash)
	tendermintClient := ibctmtypes.NewClientState(
		suite.chainA.ChainID,
		ibctesting.DefaultTrustLevel,
		ibctesting.TrustingPeriod,
		ibctesting.UnbondingPeriod,
		ibctesting.MaxClockDrift,
		clientHeight,
		commitmenttypes.GetSDKSpecs(),
		ibctesting.UpgradePath,
		false,
		false,
	)
	msg, _ := clienttypes.NewMsgCreateClient(tendermintClient, consensus, SenderAddress.String())
	emitter.DecodeMsgCreateClient(msg, detail)
	suite.testCompareJson(detail,
		fmt.Sprintf(
			"{\"client_state\":{\"chain_id\":\"testchain0\",\"trust_level\":{\"numerator\":1,\"denominator\":3},\"trusting_period\":1209600000000000,\"unbonding_period\":1814400000000000,\"max_clock_drift\":10000000000,\"frozen_height\":{},\"latest_height\":{\"revision_height\":10},\"proof_specs\":[{\"leaf_spec\":{\"hash\":1,\"prehash_value\":1,\"length\":1,\"prefix\":\"AA==\"},\"inner_spec\":{\"child_order\":[0,1],\"child_size\":33,\"min_prefix_length\":4,\"max_prefix_length\":12,\"hash\":1}},{\"leaf_spec\":{\"hash\":1,\"prehash_value\":1,\"length\":1,\"prefix\":\"AA==\"},\"inner_spec\":{\"child_order\":[0,1],\"child_size\":32,\"min_prefix_length\":1,\"max_prefix_length\":1,\"hash\":1}}],\"upgrade_path\":[\"upgrade\",\"upgradedIBCState\"]},\"consensus_state\":{\"timestamp\":\"2020-01-02T00:00:00Z\",\"root\":{\"hash\":\"%s\"},\"next_validators_hash\":\"%s\"},\"signer\":\"band12djkuer9wgqqqqqqqqqqqqqqqqqqqqqqck96t0\"}",
			b64RootHash,
			consensus.NextValidatorsHash,
		),
	)
	// MsgCreateClient example
	// {"client_state":{"chain_id":"testchain0","trust_level":{"numerator":1,"denominator":3},"trusting_period":1209600000000000,"unbonding_period":1814400000000000,"max_clock_drift":10000000000,"frozen_height":{},"latest_height":{"revision_height":10},"proof_specs":[{"leaf_spec":{"hash":1,"prehash_value":1,"length":1,"prefix":"AA=="},"inner_spec":{"child_order":[0,1],"child_size":33,"min_prefix_length":4,"max_prefix_length":12,"hash":1}},{"leaf_spec":{"hash":1,"prehash_value":1,"length":1,"prefix":"AA=="},"inner_spec":{"child_order":[0,1],"child_size":32,"min_prefix_length":1,"max_prefix_length":1,"hash":1}}],"upgrade_path":["upgrade","upgradedIBCState"]},"consensus_state":{"timestamp":"2020-01-02T00:00:00Z","root":{"hash":"I0ofcG04FYhAyDFzygf8Q/6JEpBactgfhm68fSXwBro="},"next_validators_hash":"C8277795F71B45089E58F0994DCF4F88BECD5770C7E492A9A25B706888D6BF2F"},"signer":"band12djkuer9wgqqqqqqqqqqqqqqqqqqqqqqck96t0"}
}

func (suite *DecoderTestSuite) TestDecodeMsgSubmitProposal() {
	detail := make(common.JsDict)
	msg, _ := govtypes.NewMsgSubmitProposal(content, testapp.Coins1000000uband, SenderAddress)
	emitter.DecodeMsgSubmitProposal(msg, detail)
	suite.testCompareJson(
		detail,
		"{\"content\":{\"title\":\"Title\",\"description\":\"Desc\"},\"initial_deposit\":[{\"denom\":\"uband\",\"amount\":\"1000000\"}],\"proposer\":\"band12djkuer9wgqqqqqqqqqqqqqqqqqqqqqqck96t0\"}",
	)

}

func (suite *DecoderTestSuite) TestDecodeMsgDeposit() {
	detail := make(common.JsDict)
	msg := govtypes.NewMsgDeposit(SenderAddress, 1, testapp.Coins1000000uband)
	emitter.DecodeMsgDeposit(msg, detail)
	suite.testCompareJson(
		detail,
		"{\"amount\":[{\"denom\":\"uband\",\"amount\":\"1000000\"}],\"depositor\":\"band12djkuer9wgqqqqqqqqqqqqqqqqqqqqqqck96t0\",\"proposal_id\":1}",
	)
}

func (suite *DecoderTestSuite) TestDecodeMsgVote() {
	detail := make(common.JsDict)
	msg := govtypes.NewMsgVote(SenderAddress, 1, 0)
	emitter.DecodeMsgVote(msg, detail)
	suite.testCompareJson(detail,
		"{\"option\":0,\"proposal_id\":1,\"voter\":\"band12djkuer9wgqqqqqqqqqqqqqqqqqqqqqqck96t0\"}",
	)
}

func (suite *DecoderTestSuite) TestDecodeMsgCreateValidator() {
	detail := make(common.JsDict)
	msg, _ := stakingtypes.NewMsgCreateValidator(
		ValAddress,
		PubKey,
		SelfDelegation,
		Description,
		CommissionRate,
		MinSelfDelegation,
	)

	emitter.DecodeMsgCreateValidator(msg, detail)
	suite.testCompareJson(
		detail,
		"{\"commission\":{\"rate\":\"1.000000000000000000\",\"max_rate\":\"5.000000000000000000\",\"max_change_rate\":\"5.000000000000000000\"},\"delegator_address\":\"band12eskc6tyv96x7usqqqqqqqqqqqqqqqqqzep99r\",\"description\":{\"details\":\"details\",\"identity\":\"identity\",\"moniker\":\"moniker\",\"security_contact\":\"securityContact\",\"website\":\"website\"},\"min_self_delegation\":\"1\",\"pubkey\":\"0b485cfc0eecc619440448436f8fc9df40566f2369e72400281454cb552afb50\",\"validator_address\":\"bandvaloper12eskc6tyv96x7usqqqqqqqqqqqqqqqqqw09xqg\",\"value\":{\"denom\":\"uband\",\"amount\":\"1\"}}",
	)
}

func (suite *DecoderTestSuite) TestDecodeMsgEditValidator() {
	detail := make(common.JsDict)
	msg := stakingtypes.NewMsgEditValidator(ValAddress, Description, &NewRate, &MinSelfDelegation)

	emitter.DecodeMsgEditValidator(msg, detail)
	suite.testCompareJson(
		detail,
		"{\"commission_rate\":\"1.000000000000000000\",\"description\":{\"details\":\"details\",\"identity\":\"identity\",\"moniker\":\"moniker\",\"security_contact\":\"securityContact\",\"website\":\"website\"},\"min_self_delegation\":\"1\",\"validator_address\":\"bandvaloper12eskc6tyv96x7usqqqqqqqqqqqqqqqqqw09xqg\"}",
	)
}

func (suite *DecoderTestSuite) TestDecodeMsgDelegate() {
	detail := make(common.JsDict)
	msg := stakingtypes.NewMsgDelegate(DelegatorAddress, ValAddress, Amount)

	emitter.DecodeMsgDelegate(msg, detail)
	suite.testCompareJson(
		detail,
		"{\"amount\":{\"denom\":\"uband\",\"amount\":\"1\"},\"delegator_address\":\"band1g3jkcet8v96x7usqqqqqqqqqqqqqqqqqus6d5g\",\"validator_address\":\"bandvaloper12eskc6tyv96x7usqqqqqqqqqqqqqqqqqw09xqg\"}",
	)
}

func (suite *DecoderTestSuite) TestDecodeMsgUndelegate() {
	detail := make(common.JsDict)
	msg := stakingtypes.NewMsgUndelegate(DelegatorAddress, ValAddress, Amount)

	emitter.DecodeMsgUndelegate(msg, detail)
	suite.testCompareJson(
		detail,
		"{\"amount\":{\"denom\":\"uband\",\"amount\":\"1\"},\"delegator_address\":\"band1g3jkcet8v96x7usqqqqqqqqqqqqqqqqqus6d5g\",\"validator_address\":\"bandvaloper12eskc6tyv96x7usqqqqqqqqqqqqqqqqqw09xqg\"}",
	)
}

func (suite *DecoderTestSuite) TestDecodeMsgBeginRedelegate() {
	detail := make(common.JsDict)
	msg := stakingtypes.NewMsgBeginRedelegate(DelegatorAddress, ValAddress, ValAddress, Amount)

	emitter.DecodeMsgBeginRedelegate(msg, detail)
	suite.testCompareJson(
		detail,
		"{\"amount\":{\"denom\":\"uband\",\"amount\":\"1\"},\"delegator_address\":\"band1g3jkcet8v96x7usqqqqqqqqqqqqqqqqqus6d5g\",\"validator_dst_address\":\"bandvaloper12eskc6tyv96x7usqqqqqqqqqqqqqqqqqw09xqg\",\"validator_src_address\":\"bandvaloper12eskc6tyv96x7usqqqqqqqqqqqqqqqqqw09xqg\"}",
	)
}
func (suite *DecoderTestSuite) TestDecodeMsgUpdateClient() {
	detail := make(common.JsDict)
	msg, _ := clienttypes.NewMsgUpdateClient(
		"tendermint",
		suite.chainA.CurrentTMClientHeader(),
		SenderAddress.String(),
	)
	emitter.DecodeMsgUpdateClient(msg, detail)
	suite.testContains(
		detail,
		"{\"client_id\":\"tendermint\",\"header\":{\"signed_header\":{\"header\":{\"version\":{\"block\":11,\"app\":2},\"chain_id\":\"testchain0\",\"height\":3,\"time\":\"2020-01-02T00:00:00Z\",\"last_block_id\":{\"hash\":\"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=\",\"part_set_header\":{\"total\":10000,\"hash\":\"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=\"}}",
	)
	// MsgUpdateClient
	// "{\"client_id\":\"tendermint\",\"header\":{\"signed_header\":{\"header\":{\"version\":{\"block\":11,\"app\":2},\"chain_id\":\"testchain0\",\"height\":3,\"time\":\"2020-01-02T00:00:00Z\",\"last_block_id\":{\"hash\":\"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=\",\"part_set_header\":{\"total\":10000,\"hash\":\"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=\"}},\"last_commit_hash\":\"VnnIEw5Rphpyx5BgGrYlqa65CvjT8weLaOs/wbJaknQ=\",\"data_hash\":\"bW4ouLmLUycELqUKV91G5syFHHLlKL3qpu/e7v5moLg=\",\"validators_hash\":\"35jAWHlQWSshlZrerDcsJd5H8LuvI80BB4ezq6fHiJw=\",\"next_validators_hash\":\"35jAWHlQWSshlZrerDcsJd5H8LuvI80BB4ezq6fHiJw=\",\"consensus_hash\":\"5eVmxB7Vfj/4zBDxhBeHiLj6pgKwfPH0JSF72BefHyQ=\",\"app_hash\":\"VnnIEw5Rphpyx5BgGrYlqa65CvjT8weLaOs/wbJaknQ=\",\"last_results_hash\":\"CS4FhjAkftYAmGOhLu4RfSbNnQi1rcqrN/KrNdtHWjc=\",\"evidence_hash\":\"c4ZdsI9J1YQokF04mrTKS5bkWjIGx6adQ6Xcc3LmBxQ=\",\"proposer_address\":\"f/nWW2sIpnlCMZ1XYLa/jtNzVak=\"},\"commit\":{\"height\":3,\"round\":1,\"block_id\":{\"hash\":\"Vo4riCF+F1W/yPgGPEjyunesQNWSSMyp5nE8r12NQV0=\",\"part_set_header\":{\"total\":3,\"hash\":\"hwgKOc/jNqZj6lwNm97vSTq9wYt8Pj4MjmYTVMGDFDI=\"}},\"signatures\":[{\"block_id_flag\":2,\"validator_address\":\"f/nWW2sIpnlCMZ1XYLa/jtNzVak=\",\"timestamp\":\"2020-01-02T00:00:00Z\",\"signature\":\"fvGxOLWnEYK5HxqogNmQ63b037/zi1LT3wC6ES/msdMst6yBsIRg44StmbzNUsZlWMfBWVs39myGcQgTYzgkUg==\"}]}},\"validator_set\":{\"validators\":[{\"address\":\"f/nWW2sIpnlCMZ1XYLa/jtNzVak=\",\"pub_key\":{\"Sum\":{\"secp256k1\":\"Arn/2FLDO4dVHxEGAx6QsWKxjHj1HEpjgtW4asUV8lIy\"}},\"voting_power\":1}],\"proposer\":{\"address\":\"f/nWW2sIpnlCMZ1XYLa/jtNzVak=\",\"pub_key\":{\"Sum\":{\"secp256k1\":\"Arn/2FLDO4dVHxEGAx6QsWKxjHj1HEpjgtW4asUV8lIy\"}},\"voting_power\":1},\"total_voting_power\":1},\"trusted_height\":{}},\"signer\":\"band12djkuer9wgqqqqqqqqqqqqqqqqqqqqqqck96t0\"}"
}

func (suite *DecoderTestSuite) TestDecodeMsgUpgradeClient() {
	path := NewOraclePath(suite.chainA, suite.chainB)
	suite.coordinator.Setup(path)

	detail := make(common.JsDict)
	lastHeight := clienttypes.NewHeight(0, uint64(suite.chainB.GetContext().BlockHeight()+1))

	cs, found := suite.chainA.App.IBCKeeper.ClientKeeper.GetClientState(
		suite.chainA.GetContext(),
		path.EndpointA.ClientID,
	)
	suite.Require().True(found)

	newClientHeight := clienttypes.NewHeight(1, 1)
	upgradedClient := ibctmtypes.NewClientState(
		"newChainId",
		ibctmtypes.DefaultTrustLevel,
		ibctesting.TrustingPeriod,
		ibctesting.UnbondingPeriod+ibctesting.TrustingPeriod,
		ibctesting.MaxClockDrift,
		newClientHeight,
		commitmenttypes.GetSDKSpecs(),
		ibctesting.UpgradePath,
		false,
		false,
	)
	upgradedConsState := &ibctmtypes.ConsensusState{
		NextValidatorsHash: []byte("nextValsHash"),
	}

	proofUpgradeClient, _ := suite.chainB.QueryUpgradeProof(
		upgradetypes.UpgradedClientKey(int64(lastHeight.GetRevisionHeight())),
		cs.GetLatestHeight().GetRevisionHeight(),
	)
	proofUpgradedConsState, _ := suite.chainB.QueryUpgradeProof(
		upgradetypes.UpgradedConsStateKey(int64(lastHeight.GetRevisionHeight())),
		cs.GetLatestHeight().GetRevisionHeight(),
	)

	msg, err := clienttypes.NewMsgUpgradeClient(path.EndpointA.ClientID, upgradedClient, upgradedConsState,
		proofUpgradeClient, proofUpgradedConsState, suite.chainA.SenderAccount.GetAddress().String())
	suite.Require().NoError(err)

	emitter.DecodeMsgUpgradeClient(msg, detail)
	suite.testContains(
		detail,
		"{\"client_id\":\"07-tendermint-0\",\"client_state\":{\"chain_id\":\"newChainId\",\"trust_level\":{\"numerator\":1,\"denominator\":3},\"trusting_period\":1209600000000000,\"unbonding_period\":3024000000000000,\"max_clock_drift\":10000000000,\"frozen_height\":{},\"latest_height\":{\"revision_number\":1,\"revision_height\":1},\"proof_specs\":[{\"leaf_spec\":{\"hash\":1,\"prehash_value\":1,\"length\":1,\"prefix\":\"AA==\"},\"inner_spec\":{\"child_order\":[0,1],\"child_size\":33,\"min_prefix_length\":4,\"max_prefix_length\":12,\"hash\":1}},{\"leaf_spec\":{\"hash\":1,\"prehash_value\":1,\"length\":1,\"prefix\":\"AA==\"},\"inner_spec\":{\"child_order\":[0,1],\"child_size\":32,\"min_prefix_length\":1,\"max_prefix_length\":1,\"hash\":1}}],\"upgrade_path\":[\"upgrade\",\"upgradedIBCState\"]},\"consensus_state\":{\"timestamp\":\"0001-01-01T00:00:00Z\",\"root\":{},\"next_validators_hash\":\"6E65787456616C7348617368\"},",
	)
	// MsgUpgradeClient
	// "{\"client_id\":\"07-tendermint-0\",\"client_state\":{\"chain_id\":\"newChainId\",\"trust_level\":{\"numerator\":1,\"denominator\":3},\"trusting_period\":1209600000000000,\"unbonding_period\":3024000000000000,\"max_clock_drift\":10000000000,\"frozen_height\":{},\"latest_height\":{\"revision_number\":1,\"revision_height\":1},\"proof_specs\":[{\"leaf_spec\":{\"hash\":1,\"prehash_value\":1,\"length\":1,\"prefix\":\"AA==\"},\"inner_spec\":{\"child_order\":[0,1],\"child_size\":33,\"min_prefix_length\":4,\"max_prefix_length\":12,\"hash\":1}},{\"leaf_spec\":{\"hash\":1,\"prehash_value\":1,\"length\":1,\"prefix\":\"AA==\"},\"inner_spec\":{\"child_order\":[0,1],\"child_size\":32,\"min_prefix_length\":1,\"max_prefix_length\":1,\"hash\":1}}],\"upgrade_path\":[\"upgrade\",\"upgradedIBCState\"]},\"consensus_state\":{\"timestamp\":\"0001-01-01T00:00:00Z\",\"root\":{},\"next_validators_hash\":\"6E65787456616C7348617368\"},\"proof_upgrade_client\":\"CiYSJAoidXBncmFkZWRJQkNTdGF0ZS8xOC91cGdyYWRlZENsaWVudAquAQqrAQoHdXBncmFkZRIg47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFUaCQgBGAEgASoBACIlCAESIQG2RppbdEWeFVF5h90HmJZ/OIvuBr5jbE7mh/4a8ey+lSIlCAESIQHZm7f7BAECvMg69fhmRvif+axXjaVvh7wuDvibWJVoJiIlCAESIQGbHEApyKCI6yWJSWKQnvxTXX67FeS/avKzkttknO4VoA==\",\"proof_upgrade_consensus_state\":\"CikSJwoldXBncmFkZWRJQkNTdGF0ZS8xOC91cGdyYWRlZENvbnNTdGF0ZQquAQqrAQoHdXBncmFkZRIg47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFUaCQgBGAEgASoBACIlCAESIQG2RppbdEWeFVF5h90HmJZ/OIvuBr5jbE7mh/4a8ey+lSIlCAESIQHZm7f7BAECvMg69fhmRvif+axXjaVvh7wuDvibWJVoJiIlCAESIQGbHEApyKCI6yWJSWKQnvxTXX67FeS/avKzkttknO4VoA==\",\"signer\":\"band1ws6lm89d6xenm3cms264ejvxk8rurw55t4vpl9\"}" does not contain "{\"client_id\":\"tendermint\",\"header\":{\"signed_header\":{\"header\":{\"version\":{\"block\":11,\"app\":2},\"chain_id\":\"testchain0\",\"height\":3,\"time\":\"2020-01-02T00:00:00Z\",\"last_block_id\":{\"hash\":\"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=\",\"part_set_header\":{\"total\":10000,\"hash\":\"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=\"}}"
}

func (suite *DecoderTestSuite) TestDecodeMsgSubmitMisbehaviour() {
	detail := make(common.JsDict)

	height := clienttypes.NewHeight(0, uint64(suite.chainA.CurrentHeader.Height))
	heightMinus1 := clienttypes.NewHeight(0, uint64(suite.chainA.CurrentHeader.Height)-1)
	header1 := suite.chainA.CreateTMClientHeader(
		suite.chainA.ChainID,
		int64(height.RevisionHeight),
		heightMinus1,
		suite.chainA.CurrentHeader.Time,
		suite.chainA.Vals,
		suite.chainA.Vals,
		suite.chainA.Signers,
	)
	header2 := suite.chainA.CreateTMClientHeader(
		suite.chainA.ChainID,
		int64(height.RevisionHeight),
		heightMinus1,
		suite.chainA.CurrentHeader.Time.Add(time.Minute),
		suite.chainA.Vals,
		suite.chainA.Vals,
		suite.chainA.Signers,
	)

	misbehaviour := ibctmtypes.NewMisbehaviour("tendermint", header1, header2)
	msg, err := clienttypes.NewMsgSubmitMisbehaviour(
		"tendermint",
		misbehaviour,
		suite.chainA.SenderAccount.GetAddress().String(),
	)
	suite.Require().NoError(err)

	emitter.DecodeMsgSubmitMisbehaviour(msg, detail)
	suite.testContains(
		detail,
		"{\"client_id\":\"tendermint\",\"misbehaviour\":{\"client_id\":\"tendermint\",\"header_1\":{\"signed_header\":{\"header\":{\"version\":{\"block\":11,\"app\":2},\"chain_id\":\"testchain0\",\"height\":3,\"time\":\"2020-01-02T00:00:00Z\",\"last_block_id\":{\"hash\":\"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=\",\"part_set_header\":{\"total\":10000,\"hash\":\"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=\"}},\"last_commit_hash\":",
	)
	// MsgSubmitMisbehaviour
	// "{\"client_id\":\"tendermint\",\"misbehaviour\":{\"client_id\":\"tendermint\",\"header_1\":{\"signed_header\":{\"header\":{\"version\":{\"block\":11,\"app\":2},\"chain_id\":\"testchain0\",\"height\":3,\"time\":\"2020-01-02T00:00:00Z\",\"last_block_id\":{\"hash\":\"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=\",\"part_set_header\":{\"total\":10000,\"hash\":\"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=\"}},\"last_commit_hash\":\"Gy7QPczOYJhyvkFumSNmNOYFi0beSQP7K1T3U73ZPL0=\",\"data_hash\":\"bW4ouLmLUycELqUKV91G5syFHHLlKL3qpu/e7v5moLg=\",\"validators_hash\":\"UWHAFzvn3gBH0c928WeqdiwEY4ozNcuJsbO7i/ykGlI=\",\"next_validators_hash\":\"UWHAFzvn3gBH0c928WeqdiwEY4ozNcuJsbO7i/ykGlI=\",\"consensus_hash\":\"5eVmxB7Vfj/4zBDxhBeHiLj6pgKwfPH0JSF72BefHyQ=\",\"app_hash\":\"Gy7QPczOYJhyvkFumSNmNOYFi0beSQP7K1T3U73ZPL0=\",\"last_results_hash\":\"CS4FhjAkftYAmGOhLu4RfSbNnQi1rcqrN/KrNdtHWjc=\",\"evidence_hash\":\"c4ZdsI9J1YQokF04mrTKS5bkWjIGx6adQ6Xcc3LmBxQ=\",\"proposer_address\":\"H6sPOQrXCVy4QN7pv0ealpUP1zE=\"},\"commit\":{\"height\":3,\"round\":1,\"block_id\":{\"hash\":\"4BHQI7RdQzVdZjXlV5cFTWUX8FUUyZlRZlcJz57HDzU=\",\"part_set_header\":{\"total\":3,\"hash\":\"hwgKOc/jNqZj6lwNm97vSTq9wYt8Pj4MjmYTVMGDFDI=\"}},\"signatures\":[{\"block_id_flag\":2,\"validator_address\":\"H6sPOQrXCVy4QN7pv0ealpUP1zE=\",\"timestamp\":\"2020-01-02T00:00:00Z\",\"signature\":\"QBI8sEcCn1EQv3uDlWOatFxlyfKSj8Yq9eUdrbL8Y4Yfhr5+oByFD4D91N45Cg9GFPbYpLtlb3CvEsH7oyvSHg==\"}]}},\"validator_set\":{\"validators\":[{\"address\":\"H6sPOQrXCVy4QN7pv0ealpUP1zE=\",\"pub_key\":{\"Sum\":{\"secp256k1\":\"A6/xRIwBfvDbU2TkJs4rgKexroILGVJkTRUDkDMcbUX8\"}},\"voting_power\":1}],\"proposer\":{\"address\":\"H6sPOQrXCVy4QN7pv0ealpUP1zE=\",\"pub_key\":{\"Sum\":{\"secp256k1\":\"A6/xRIwBfvDbU2TkJs4rgKexroILGVJkTRUDkDMcbUX8\"}},\"voting_power\":1},\"total_voting_power\":1},\"trusted_height\":{\"revision_height\":2},\"trusted_validators\":{\"validators\":[{\"address\":\"H6sPOQrXCVy4QN7pv0ealpUP1zE=\",\"pub_key\":{\"Sum\":{\"secp256k1\":\"A6/xRIwBfvDbU2TkJs4rgKexroILGVJkTRUDkDMcbUX8\"}},\"voting_power\":1}],\"proposer\":{\"address\":\"H6sPOQrXCVy4QN7pv0ealpUP1zE=\",\"pub_key\":{\"Sum\":{\"secp256k1\":\"A6/xRIwBfvDbU2TkJs4rgKexroILGVJkTRUDkDMcbUX8\"}},\"voting_power\":1},\"total_voting_power\":1}},\"header_2\":{\"signed_header\":{\"header\":{\"version\":{\"block\":11,\"app\":2},\"chain_id\":\"testchain0\",\"height\":3,\"time\":\"2020-01-02T00:01:00Z\",\"last_block_id\":{\"hash\":\"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=\",\"part_set_header\":{\"total\":10000,\"hash\":\"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=\"}},\"last_commit_hash\":\"Gy7QPczOYJhyvkFumSNmNOYFi0beSQP7K1T3U73ZPL0=\",\"data_hash\":\"bW4ouLmLUycELqUKV91G5syFHHLlKL3qpu/e7v5moLg=\",\"validators_hash\":\"UWHAFzvn3gBH0c928WeqdiwEY4ozNcuJsbO7i/ykGlI=\",\"next_validators_hash\":\"UWHAFzvn3gBH0c928WeqdiwEY4ozNcuJsbO7i/ykGlI=\",\"consensus_hash\":\"5eVmxB7Vfj/4zBDxhBeHiLj6pgKwfPH0JSF72BefHyQ=\",\"app_hash\":\"Gy7QPczOYJhyvkFumSNmNOYFi0beSQP7K1T3U73ZPL0=\",\"last_results_hash\":\"CS4FhjAkftYAmGOhLu4RfSbNnQi1rcqrN/KrNdtHWjc=\",\"evidence_hash\":\"c4ZdsI9J1YQokF04mrTKS5bkWjIGx6adQ6Xcc3LmBxQ=\",\"proposer_address\":\"H6sPOQrXCVy4QN7pv0ealpUP1zE=\"},\"commit\":{\"height\":3,\"round\":1,\"block_id\":{\"hash\":\"OIlOUMldL7DwSF/CwxhzwvbCkB06ZIMKLn91cGqmye4=\",\"part_set_header\":{\"total\":3,\"hash\":\"hwgKOc/jNqZj6lwNm97vSTq9wYt8Pj4MjmYTVMGDFDI=\"}},\"signatures\":[{\"block_id_flag\":2,\"validator_address\":\"H6sPOQrXCVy4QN7pv0ealpUP1zE=\",\"timestamp\":\"2020-01-02T00:01:00Z\",\"signature\":\"2IrQF/dca6yjumwFw0BK7xbfxa5r3nxV2tpYh1my3IkDYRbTM/vmCyW6BiCRSCivuhM/9eoHKK/YAQAAZh8zcg==\"}]}},\"validator_set\":{\"validators\":[{\"address\":\"H6sPOQrXCVy4QN7pv0ealpUP1zE=\",\"pub_key\":{\"Sum\":{\"secp256k1\":\"A6/xRIwBfvDbU2TkJs4rgKexroILGVJkTRUDkDMcbUX8\"}},\"voting_power\":1}],\"proposer\":{\"address\":\"H6sPOQrXCVy4QN7pv0ealpUP1zE=\",\"pub_key\":{\"Sum\":{\"secp256k1\":\"A6/xRIwBfvDbU2TkJs4rgKexroILGVJkTRUDkDMcbUX8\"}},\"voting_power\":1},\"total_voting_power\":1},\"trusted_height\":{\"revision_height\":2},\"trusted_validators\":{\"validators\":[{\"address\":\"H6sPOQrXCVy4QN7pv0ealpUP1zE=\",\"pub_key\":{\"Sum\":{\"secp256k1\":\"A6/xRIwBfvDbU2TkJs4rgKexroILGVJkTRUDkDMcbUX8\"}},\"voting_power\":1}],\"proposer\":{\"address\":\"H6sPOQrXCVy4QN7pv0ealpUP1zE=\",\"pub_key\":{\"Sum\":{\"secp256k1\":\"A6/xRIwBfvDbU2TkJs4rgKexroILGVJkTRUDkDMcbUX8\"}},\"voting_power\":1},\"total_voting_power\":1}}},\"signer\":\"band1r74s7wg26uy4ewzqmm5m73u6j62sl4e38zpnws\"}"
}

func (suite *DecoderTestSuite) TestDecodeMsgConnectionOpenInit() {
	detail := make(common.JsDict)
	path := NewOraclePath(suite.chainA, suite.chainB)
	prefix := commitmenttypes.NewMerklePrefix([]byte("storePrefixKey"))
	msg := connectiontypes.NewMsgConnectionOpenInit(
		path.EndpointA.ConnectionID,
		path.EndpointB.ClientID,
		prefix,
		ibctesting.ConnectionVersion,
		ibctesting.DefaultDelayPeriod,
		SignerAddress.String(),
	)
	emitter.DecodeMsgConnectionOpenInit(msg, detail)
	suite.testCompareJson(
		detail,
		"{\"client_id\":\"\",\"counterparty\":{\"prefix\":{\"key_prefix\":\"c3RvcmVQcmVmaXhLZXk=\"}},\"delay_period\":0,\"signer\":\"band12d5kwmn9wgqqqqqqqqqqqqqqqqqqqqqqr057wh\",\"version\":{\"identifier\":\"1\",\"features\":[\"ORDER_ORDERED\",\"ORDER_UNORDERED\"]}}",
	)
}

func (suite *DecoderTestSuite) TestDecodeMsgConnectionOpenTry() {
	detail := make(common.JsDict)
	path := NewOraclePath(suite.chainA, suite.chainB)
	suite.coordinator.Setup(path)
	prefix := commitmenttypes.NewMerklePrefix([]byte("storePrefixKey"))
	clientState := ibctmtypes.NewClientState(
		suite.chainA.ChainID,
		ibctmtypes.DefaultTrustLevel,
		ibctesting.TrustingPeriod,
		ibctesting.UnbondingPeriod,
		ibctesting.MaxClockDrift,
		clientHeight,
		commitmenttypes.GetSDKSpecs(),
		ibctesting.UpgradePath,
		false,
		false,
	)
	msg := connectiontypes.NewMsgConnectionOpenTry(
		path.EndpointA.ConnectionID,
		path.EndpointA.ClientID,
		path.EndpointB.ConnectionID,
		path.EndpointB.ClientID,
		clientState,
		prefix,
		[]*connectiontypes.Version{ibctesting.ConnectionVersion},
		500,
		[]byte{},
		[]byte{},
		[]byte{},
		clientHeight,
		clientHeight,
		SignerAddress.String(),
	)
	emitter.DecodeMsgConnectionOpenTry(msg, detail)
	suite.testCompareJson(
		detail,
		"{\"client_id\":\"07-tendermint-0\",\"client_state\":{\"chain_id\":\"testchain0\",\"trust_level\":{\"numerator\":1,\"denominator\":3},\"trusting_period\":1209600000000000,\"unbonding_period\":1814400000000000,\"max_clock_drift\":10000000000,\"frozen_height\":{},\"latest_height\":{\"revision_height\":10},\"proof_specs\":[{\"leaf_spec\":{\"hash\":1,\"prehash_value\":1,\"length\":1,\"prefix\":\"AA==\"},\"inner_spec\":{\"child_order\":[0,1],\"child_size\":33,\"min_prefix_length\":4,\"max_prefix_length\":12,\"hash\":1}},{\"leaf_spec\":{\"hash\":1,\"prehash_value\":1,\"length\":1,\"prefix\":\"AA==\"},\"inner_spec\":{\"child_order\":[0,1],\"child_size\":32,\"min_prefix_length\":1,\"max_prefix_length\":1,\"hash\":1}}],\"upgrade_path\":[\"upgrade\",\"upgradedIBCState\"]},\"consensus_height\":{\"revision_height\":10,\"revision_number\":0},\"counterparty\":{\"client_id\":\"07-tendermint-0\",\"connection_id\":\"connection-0\",\"prefix\":{\"key_prefix\":\"c3RvcmVQcmVmaXhLZXk=\"}},\"counterparty_versions\":[{\"identifier\":\"1\",\"features\":[\"ORDER_ORDERED\",\"ORDER_UNORDERED\"]}],\"delay_period\":500,\"previous_connection_id\":\"connection-0\",\"proof_client\":\"\",\"proof_consensus\":\"\",\"proof_height\":{\"revision_height\":10,\"revision_number\":0},\"proof_init\":\"\",\"signer\":\"band12d5kwmn9wgqqqqqqqqqqqqqqqqqqqqqqr057wh\"}",
	)
}

func (suite *DecoderTestSuite) TestDecodeMsgConnectionOpenAck() {
	detail := make(common.JsDict)
	path := NewOraclePath(suite.chainA, suite.chainB)
	clientState := ibctmtypes.NewClientState(
		suite.chainA.ChainID,
		ibctmtypes.DefaultTrustLevel,
		ibctesting.TrustingPeriod,
		ibctesting.UnbondingPeriod,
		ibctesting.MaxClockDrift,
		clientHeight,
		commitmenttypes.GetSDKSpecs(),
		ibctesting.UpgradePath,
		false,
		false,
	)
	msg := connectiontypes.NewMsgConnectionOpenAck(
		path.EndpointA.ConnectionID,
		path.EndpointB.ConnectionID,
		clientState,
		[]byte{},
		[]byte{},
		[]byte{},
		clientHeight,
		clientHeight,
		ibctesting.ConnectionVersion,
		SignerAddress.String(),
	)
	emitter.DecodeMsgConnectionOpenAck(msg, detail)
	suite.testCompareJson(
		detail,
		"{\"client_state\":{\"chain_id\":\"testchain0\",\"trust_level\":{\"numerator\":1,\"denominator\":3},\"trusting_period\":1209600000000000,\"unbonding_period\":1814400000000000,\"max_clock_drift\":10000000000,\"frozen_height\":{},\"latest_height\":{\"revision_height\":10},\"proof_specs\":[{\"leaf_spec\":{\"hash\":1,\"prehash_value\":1,\"length\":1,\"prefix\":\"AA==\"},\"inner_spec\":{\"child_order\":[0,1],\"child_size\":33,\"min_prefix_length\":4,\"max_prefix_length\":12,\"hash\":1}},{\"leaf_spec\":{\"hash\":1,\"prehash_value\":1,\"length\":1,\"prefix\":\"AA==\"},\"inner_spec\":{\"child_order\":[0,1],\"child_size\":32,\"min_prefix_length\":1,\"max_prefix_length\":1,\"hash\":1}}],\"upgrade_path\":[\"upgrade\",\"upgradedIBCState\"]},\"connection_id\":\"\",\"consensus_height\":{\"revision_height\":10,\"revision_number\":0},\"counterparty_connection_id\":\"\",\"proof_client\":\"\",\"proof_consensus\":\"\",\"proof_height\":{\"revision_height\":10,\"revision_number\":0},\"proof_try\":\"\",\"signer\":\"band12d5kwmn9wgqqqqqqqqqqqqqqqqqqqqqqr057wh\",\"version\":{\"identifier\":\"1\",\"features\":[\"ORDER_ORDERED\",\"ORDER_UNORDERED\"]}}",
	)
}

func (suite *DecoderTestSuite) TestDecodeMsgConnectionOpenConfirm() {
	detail := make(common.JsDict)
	path := NewOraclePath(suite.chainA, suite.chainB)
	msg := connectiontypes.NewMsgConnectionOpenConfirm(
		path.EndpointA.ConnectionID,
		[]byte{},
		clientHeight,
		SignerAddress.String(),
	)
	emitter.DecodeMsgConnectionOpenConfirm(msg, detail)
	suite.testCompareJson(
		detail,
		"{\"connection_id\":\"\",\"proof_ack\":\"\",\"proof_height\":{\"revision_height\":10,\"revision_number\":0},\"signer\":\"band12d5kwmn9wgqqqqqqqqqqqqqqqqqqqqqqr057wh\"}",
	)
}

func (suite *DecoderTestSuite) TestDecodeMsgChannelOpenInit() {
	detail := make(common.JsDict)
	path := NewOraclePath(suite.chainA, suite.chainB)
	suite.coordinator.Setup(path)
	msg := channeltypes.NewMsgChannelOpenInit(
		path.EndpointA.ChannelConfig.PortID,
		path.EndpointA.ChannelConfig.Version,
		channeltypes.ORDERED,
		path.EndpointA.GetChannel().ConnectionHops,
		path.EndpointA.Counterparty.ChannelConfig.PortID,
		SignerAddress.String(),
	)
	emitter.DecodeMsgChannelOpenInit(msg, detail)
	suite.testCompareJson(
		detail,
		"{\"channel\":{\"state\":1,\"ordering\":2,\"counterparty\":{\"port_id\":\"oracle\"},\"connection_hops\":[\"connection-0\"],\"version\":\"bandchain-1\"},\"port_id\":\"oracle\",\"signer\":\"band12d5kwmn9wgqqqqqqqqqqqqqqqqqqqqqqr057wh\"}",
	)
}

func (suite *DecoderTestSuite) TestDecodeMsgChannelOpenTry() {
	detail := make(common.JsDict)
	path := NewOraclePath(suite.chainA, suite.chainB)
	suite.coordinator.Setup(path)
	msg := channeltypes.NewMsgChannelOpenTry(
		path.EndpointA.ChannelConfig.PortID,
		path.EndpointA.Counterparty.ChannelConfig.PortID,
		path.EndpointA.ChannelConfig.Version,
		channeltypes.ORDERED,
		path.EndpointA.GetChannel().ConnectionHops,
		path.EndpointA.Counterparty.ChannelConfig.PortID,
		path.EndpointA.Counterparty.ChannelID,
		path.EndpointA.Counterparty.ChannelConfig.Version,
		[]byte{},
		clientHeight,
		SignerAddress.String(),
	)
	emitter.DecodeMsgChannelOpenTry(msg, detail)
	suite.testCompareJson(
		detail,
		"{\"channel\":{\"state\":2,\"ordering\":2,\"counterparty\":{\"port_id\":\"oracle\",\"channel_id\":\"channel-0\"},\"connection_hops\":[\"connection-0\"],\"version\":\"bandchain-1\"},\"counterparty_version\":\"bandchain-1\",\"port_id\":\"oracle\",\"previous_channel_id\":\"oracle\",\"proof_height\":{\"revision_height\":10,\"revision_number\":0},\"proof_init\":\"\",\"signer\":\"band12d5kwmn9wgqqqqqqqqqqqqqqqqqqqqqqr057wh\"}",
	)
}

func (suite *DecoderTestSuite) TestDecodeMsgChannelOpenAck() {
	detail := make(common.JsDict)
	path := NewOraclePath(suite.chainA, suite.chainB)
	suite.coordinator.Setup(path)
	msg := channeltypes.NewMsgChannelOpenAck(
		path.EndpointA.ChannelConfig.PortID,
		path.EndpointA.ChannelID,
		path.EndpointA.Counterparty.ChannelID,
		"cpv",
		[]byte{},
		clientHeight,
		SignerAddress.String(),
	)
	emitter.DecodeMsgChannelOpenAck(msg, detail)
	suite.testCompareJson(
		detail,
		"{\"channel_id\":\"channel-0\",\"counterparty_channel_id\":\"channel-0\",\"counterparty_version\":\"cpv\",\"port_id\":\"oracle\",\"proof_height\":{\"revision_height\":10,\"revision_number\":0},\"proof_try\":\"\",\"signer\":\"band12d5kwmn9wgqqqqqqqqqqqqqqqqqqqqqqr057wh\"}",
	)
}

func (suite *DecoderTestSuite) TestDecodeMsgChannelOpenConfirm() {
	detail := make(common.JsDict)
	path := NewOraclePath(suite.chainA, suite.chainB)
	suite.coordinator.Setup(path)
	msg := channeltypes.NewMsgChannelOpenConfirm(
		path.EndpointA.ChannelConfig.PortID,
		path.EndpointA.ChannelID,
		[]byte{},
		clientHeight,
		SignerAddress.String(),
	)
	emitter.DecodeMsgChannelOpenConfirm(msg, detail)
	suite.testCompareJson(
		detail,
		"{\"channel_id\":\"channel-0\",\"port_id\":\"oracle\",\"proof_ack\":\"\",\"proof_height\":{\"revision_height\":10,\"revision_number\":0},\"signer\":\"band12d5kwmn9wgqqqqqqqqqqqqqqqqqqqqqqr057wh\"}",
	)
}

func (suite *DecoderTestSuite) TestDecodeMsgChannelCloseInit() {
	detail := make(common.JsDict)
	path := NewOraclePath(suite.chainA, suite.chainB)
	suite.coordinator.Setup(path)
	msg := channeltypes.NewMsgChannelCloseInit(
		path.EndpointA.ChannelConfig.PortID,
		path.EndpointA.ChannelID,
		SignerAddress.String(),
	)
	emitter.DecodeMsgChannelCloseInit(msg, detail)
	suite.testCompareJson(
		detail,
		"{\"channel_id\":\"channel-0\",\"port_id\":\"oracle\",\"signer\":\"band12d5kwmn9wgqqqqqqqqqqqqqqqqqqqqqqr057wh\"}",
	)
}

func (suite *DecoderTestSuite) TestDecodeMsgChannelCloseConfirm() {
	detail := make(common.JsDict)
	path := NewOraclePath(suite.chainA, suite.chainB)
	suite.coordinator.Setup(path)
	msg := channeltypes.NewMsgChannelCloseConfirm(
		path.EndpointA.ChannelConfig.PortID,
		path.EndpointA.ChannelID,
		[]byte{},
		clientHeight,
		SignerAddress.String(),
	)
	emitter.DecodeMsgChannelCloseConfirm(msg, detail)
	suite.testCompareJson(
		detail,
		"{\"channel_id\":\"channel-0\",\"port_id\":\"oracle\",\"proof_height\":{\"revision_height\":10,\"revision_number\":0},\"proof_init\":\"\",\"signer\":\"band12d5kwmn9wgqqqqqqqqqqqqqqqqqqqqqqr057wh\"}",
	)
}

func (suite *DecoderTestSuite) TestDecodeMsgRecvPacket() {
	detail := make(common.JsDict)
	path := NewOraclePath(suite.chainA, suite.chainB)
	suite.coordinator.Setup(path)
	packet := channeltypes.NewPacket(
		[]byte{},
		1,
		path.EndpointA.ChannelConfig.PortID,
		path.EndpointA.ChannelID,
		path.EndpointB.ChannelConfig.PortID,
		path.EndpointB.ChannelID, clientHeight, 0)
	msg := channeltypes.NewMsgRecvPacket(packet, []byte{}, clientHeight, SignerAddress.String())
	emitter.DecodeMsgRecvPacket(msg, detail)
	suite.testCompareJson(
		detail,
		"{\"packet\":{\"data\":\"\",\"destination_channel\":\"channel-0\",\"destination_port\":\"oracle\",\"sequence\":1,\"source_channel\":\"channel-0\",\"source_port\":\"oracle\",\"timeout_height\":{\"revision_height\":10,\"revision_number\":0},\"timeout_timestamp\":0},\"proof_commitment\":\"\",\"proof_height\":{\"revision_height\":10,\"revision_number\":0},\"signer\":\"band12d5kwmn9wgqqqqqqqqqqqqqqqqqqqqqqr057wh\"}",
	)
}

func (suite *DecoderTestSuite) TestDecodeMsgAcknowledgement() {
	detail := make(common.JsDict)
	path := NewOraclePath(suite.chainA, suite.chainB)
	suite.coordinator.Setup(path)
	packet := channeltypes.NewPacket(
		[]byte{},
		1,
		path.EndpointA.ChannelConfig.PortID,
		path.EndpointA.ChannelID,
		path.EndpointB.ChannelConfig.PortID,
		path.EndpointB.ChannelID, clientHeight, 0)
	msg := channeltypes.NewMsgAcknowledgement(
		packet,
		[]byte{},
		[]byte{},
		clientHeight,
		SignerAddress.String(),
	)
	emitter.DecodeMsgAcknowledgement(msg, detail)
	suite.testCompareJson(
		detail,
		"{\"acknowledgement\":\"\",\"packet\":{\"data\":\"\",\"destination_channel\":\"channel-0\",\"destination_port\":\"oracle\",\"sequence\":1,\"source_channel\":\"channel-0\",\"source_port\":\"oracle\",\"timeout_height\":{\"revision_height\":10,\"revision_number\":0},\"timeout_timestamp\":0},\"proof_acked\":\"\",\"proof_height\":{\"revision_height\":10,\"revision_number\":0},\"signer\":\"band12d5kwmn9wgqqqqqqqqqqqqqqqqqqqqqqr057wh\"}",
	)
}

func (suite *DecoderTestSuite) TestDecodeMsgTimeout() {
	detail := make(common.JsDict)
	path := NewOraclePath(suite.chainA, suite.chainB)
	suite.coordinator.Setup(path)
	packet := channeltypes.NewPacket(
		[]byte{},
		1,
		path.EndpointA.ChannelConfig.PortID,
		path.EndpointA.ChannelID,
		path.EndpointB.ChannelConfig.PortID,
		path.EndpointB.ChannelID, clientHeight, 0)
	msg := channeltypes.NewMsgTimeout(
		packet,
		1,
		[]byte{},
		clientHeight,
		SignerAddress.String(),
	)
	emitter.DecodeMsgTimeout(msg, detail)
	suite.testCompareJson(
		detail,
		"{\"next_sequence_recv\":1,\"packet\":{\"data\":\"\",\"destination_channel\":\"channel-0\",\"destination_port\":\"oracle\",\"sequence\":1,\"source_channel\":\"channel-0\",\"source_port\":\"oracle\",\"timeout_height\":{\"revision_height\":10,\"revision_number\":0},\"timeout_timestamp\":0},\"proof_height\":{\"revision_height\":10,\"revision_number\":0},\"proof_unreceived\":\"\",\"signer\":\"band12d5kwmn9wgqqqqqqqqqqqqqqqqqqqqqqr057wh\"}",
	)
}

func (suite *DecoderTestSuite) TestDecodeMsgTimeoutOnClose() {
	detail := make(common.JsDict)
	path := NewOraclePath(suite.chainA, suite.chainB)
	suite.coordinator.Setup(path)
	packet := channeltypes.NewPacket(
		[]byte{},
		1,
		path.EndpointA.ChannelConfig.PortID,
		path.EndpointA.ChannelID,
		path.EndpointB.ChannelConfig.PortID,
		path.EndpointB.ChannelID, clientHeight, 0)

	msg := channeltypes.NewMsgTimeoutOnClose(
		packet,
		1,
		[]byte{},
		[]byte{},
		clientHeight,
		SignerAddress.String(),
	)
	emitter.DecodeMsgTimeoutOnClose(msg, detail)
	suite.testCompareJson(
		detail,
		"{\"next_sequence_recv\":1,\"packet\":{\"data\":\"\",\"destination_channel\":\"channel-0\",\"destination_port\":\"oracle\",\"sequence\":1,\"source_channel\":\"channel-0\",\"source_port\":\"oracle\",\"timeout_height\":{\"revision_height\":10,\"revision_number\":0},\"timeout_timestamp\":0},\"proof_close\":\"\",\"proof_height\":{\"revision_height\":10,\"revision_number\":0},\"proof_unreceived\":\"\",\"signer\":\"band12d5kwmn9wgqqqqqqqqqqqqqqqqqqqqqqr057wh\"}",
	)
}

func TestDecoderTestSuite(t *testing.T) {
	suite.Run(t, new(DecoderTestSuite))
}
