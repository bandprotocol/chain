package emitter_test

import (
	b64 "encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	connectiontypes "github.com/cosmos/ibc-go/v8/modules/core/03-connection/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	commitmenttypes "github.com/cosmos/ibc-go/v8/modules/core/23-commitment/types"
	ibctmtypes "github.com/cosmos/ibc-go/v8/modules/light-clients/07-tendermint"
	ibctesting "github.com/cosmos/ibc-go/v8/testing"

	"cosmossdk.io/math"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	band "github.com/bandprotocol/chain/v3/app"
	"github.com/bandprotocol/chain/v3/hooks/common"
	"github.com/bandprotocol/chain/v3/hooks/emitter"
	"github.com/bandprotocol/chain/v3/pkg/tss"
	bandtesting "github.com/bandprotocol/chain/v3/testing"
	bandtsstypes "github.com/bandprotocol/chain/v3/x/bandtss/types"
	feedstypes "github.com/bandprotocol/chain/v3/x/feeds/types"
	oracletypes "github.com/bandprotocol/chain/v3/x/oracle/types"
	restaketypes "github.com/bandprotocol/chain/v3/x/restake/types"
	tsstypes "github.com/bandprotocol/chain/v3/x/tss/types"
	tunneltypes "github.com/bandprotocol/chain/v3/x/tunnel/types"
)

const (
	TestDefaultPrepareGas uint64 = 40000
	TestDefaultExecuteGas uint64 = 300000
)

var (
	SenderAddress    = sdk.AccAddress(genAddressFromString("Sender"))
	ReceiverAddress  = sdk.AccAddress(genAddressFromString("Receiver"))
	ValAddress       = sdk.ValAddress(genAddressFromString("Validator"))
	TreasuryAddress  = sdk.AccAddress(genAddressFromString("Treasury"))
	OwnerAddress     = sdk.AccAddress(genAddressFromString("Owner"))
	ReporterAddress  = sdk.AccAddress(genAddressFromString("Reporter"))
	SignerAddress    = sdk.AccAddress(genAddressFromString("Signer"))
	DelegatorAddress = sdk.AccAddress(genAddressFromString("Delegator"))
	GranterAddress   = sdk.AccAddress(genAddressFromString("Granter"))
	GranteeAddress   = sdk.AccAddress(genAddressFromString("Grantee"))
	StakerAddress    = sdk.AccAddress(genAddressFromString("Staker"))
	AuthorityAddress = sdk.AccAddress(genAddressFromString("Authority"))
	CreatorAddress   = sdk.AccAddress(genAddressFromString("creator"))

	Coins1000000uband   = sdk.NewCoins(sdk.NewInt64Coin("uband", 1000000))
	Coins100000000uband = sdk.NewCoins(sdk.NewInt64Coin("uband", 100000000))

	clientHeight = clienttypes.NewHeight(0, 10)

	SelfDelegation    = sdk.NewInt64Coin("uband", 1)
	MinSelfDelegation = math.NewInt(1)
	Description       = stakingtypes.NewDescription("moniker", "identity", "website", "securityContact", "details")
	CommissionRate    = stakingtypes.NewCommissionRates(
		math.LegacyNewDec(1),
		math.LegacyNewDec(5),
		math.LegacyNewDec(5),
	)
	NewRate = math.LegacyNewDec(1)
	PubKey  = newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB50")
	Amount  = sdk.NewCoin("uband", math.NewInt(1))

	tssPoint          = tss.Point([]byte("point"))
	tssSignature      = tss.Signature([]byte("signature"))
	tssEncSecretShare = tss.EncSecretShare([]byte("encSecretShare"))

	content, _  = govv1beta1.ContentFromProposalType("Title", "Desc", "Text")
	proposalMsg *banktypes.MsgSend
)

func init() {
	band.SetBech32AddressPrefixesAndBip44CoinTypeAndSeal(sdk.GetConfig())
	sdk.DefaultBondDenom = "uband"

	proposalMsg = banktypes.NewMsgSend(SenderAddress, ReceiverAddress, sdk.Coins{Amount})
}

type DecoderTestSuite struct {
	suite.Suite

	coordinator *ibctesting.Coordinator

	chainA *ibctesting.TestChain
	chainB *ibctesting.TestChain
}

func (suite *DecoderTestSuite) SetupTest() {
	ibctesting.DefaultTestingAppInit = bandtesting.CreateTestingAppFn(suite.T())

	suite.coordinator = ibctesting.NewCoordinator(suite.T(), 2)
	suite.chainA = suite.coordinator.GetChain(ibctesting.GetChainID(1))
	suite.chainB = suite.coordinator.GetChain(ibctesting.GetChainID(2))
}

func NewOraclePath(chainA, chainB *ibctesting.TestChain) *ibctesting.Path {
	path := ibctesting.NewPath(chainA, chainB)
	path.EndpointA.ChannelConfig.PortID = oracletypes.ModuleName
	path.EndpointA.ChannelConfig.Version = oracletypes.Version
	path.EndpointB.ChannelConfig.PortID = oracletypes.ModuleName
	path.EndpointB.ChannelConfig.Version = oracletypes.Version

	return path
}

func genAddressFromString(s string) []byte {
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
	suite.Require().Equal(expect, string(res))
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
		banktypes.NewSendAuthorization(spendLimit, []sdk.AccAddress{}),
		&expiration,
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
		&expiration,
	)
	emitter.DecodeMsgGrant(genericMsg, detail)
	suite.testCompareJson(
		detail,
		"{\"grant\":{\"authorization\":{\"msg\":\"/band.oracle.v1.MsgReportData\"},\"expiration\":\"2020-01-02T00:00:00Z\"},\"grantee\":\"band1gaexzmn5v4jsqqqqqqqqqqqqqqqqqqqqwrdaed\",\"granter\":\"band1gaexzmn5v4eqqqqqqqqqqqqqqqqqqqqq3urue8\"}",
	)

	// TestStakeAuthorization
	stakeAuthorization, _ := stakingtypes.NewStakeAuthorization(
		[]sdk.ValAddress{ValAddress},
		[]sdk.ValAddress{},
		stakingtypes.AuthorizationType_AUTHORIZATION_TYPE_DELEGATE,
		&Amount,
	)
	stakeMsg, _ := authz.NewMsgGrant(GranterAddress, GranteeAddress, stakeAuthorization, &expiration)
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
		"{\"grantee\":\"band1gaexzmn5v4jsqqqqqqqqqqqqqqqqqqqqwrdaed\",\"granter\":\"band1gaexzmn5v4eqqqqqqqqqqqqqqqqqqqqq3urue8\",\"msg_type_url\":\"/band.oracle.v1.MsgReportData\"}",
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
		"{\"grantee\":\"band1gaexzmn5v4jsqqqqqqqqqqqqqqqqqqqqwrdaed\",\"msgs\":[{\"msg\":{\"amount\":[{\"denom\":\"uband\",\"amount\":\"1\"}],\"from_address\":\"band1gaexzmn5v4eqqqqqqqqqqqqqqqqqqqqq3urue8\",\"to_address\":\"band1gaexzmn5v4jsqqqqqqqqqqqqqqqqqqqqwrdaed\"},\"type\":\"/cosmos.bank.v1beta1.MsgSend\"},{\"msg\":{\"raw_reports\":null,\"request_id\":0,\"validator\":\"\"},\"type\":\"/band.oracle.v1.MsgReportData\"}]}",
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
		Coins100000000uband,
		TestDefaultPrepareGas,
		TestDefaultExecuteGas,
		SenderAddress,
		0,
	)
	emitter.DecodeMsgRequestData(msg, detail)
	suite.testCompareJson(
		detail,
		"{\"ask_count\":1,\"calldata\":\"Y2FsbGRhdGE=\",\"client_id\":\"cleint_id\",\"execute_gas\":300000,\"fee_limit\":[{\"denom\":\"uband\",\"amount\":\"100000000\"}],\"min_count\":1,\"oracle_script_id\":1,\"prepare_gas\":40000,\"sender\":\"band12djkuer9wgqqqqqqqqqqqqqqqqqqqqqqck96t0\",\"tss_encode_type\":0}",
	)
}

func (suite *DecoderTestSuite) TestDecodeReportData() {
	detail := make(common.JsDict)
	msg := oracletypes.NewMsgReportData(
		1,
		[]oracletypes.RawReport{{
			ExternalID: 1,
			ExitCode:   1,
			Data:       []byte("data1"),
		}, {
			ExternalID: 2,
			ExitCode:   2,
			Data:       []byte("data2"),
		}},
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
		Coins1000000uband,
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
		Coins1000000uband,
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
	)
	msg, _ := clienttypes.NewMsgCreateClient(tendermintClient, consensus, SenderAddress.String())
	emitter.DecodeMsgCreateClient(msg, detail)
	suite.testCompareJson(detail,
		fmt.Sprintf(
			"{\"client_state\":{\"chain_id\":\"testchain1-1\",\"trust_level\":{\"numerator\":1,\"denominator\":3},\"trusting_period\":1209600000000000,\"unbonding_period\":1814400000000000,\"max_clock_drift\":10000000000,\"frozen_height\":{},\"latest_height\":{\"revision_height\":10},\"proof_specs\":[{\"leaf_spec\":{\"hash\":1,\"prehash_value\":1,\"length\":1,\"prefix\":\"AA==\"},\"inner_spec\":{\"child_order\":[0,1],\"child_size\":33,\"min_prefix_length\":4,\"max_prefix_length\":12,\"hash\":1}},{\"leaf_spec\":{\"hash\":1,\"prehash_value\":1,\"length\":1,\"prefix\":\"AA==\"},\"inner_spec\":{\"child_order\":[0,1],\"child_size\":32,\"min_prefix_length\":1,\"max_prefix_length\":1,\"hash\":1}}],\"upgrade_path\":[\"upgrade\",\"upgradedIBCState\"]},\"consensus_state\":{\"timestamp\":\"2020-01-02T00:00:00Z\",\"root\":{\"hash\":\"%s\"},\"next_validators_hash\":\"%s\"},\"signer\":\"band12djkuer9wgqqqqqqqqqqqqqqqqqqqqqqck96t0\"}",
			b64RootHash,
			consensus.NextValidatorsHash,
		),
	)
}

func (suite *DecoderTestSuite) TestDecodeV1beta1MsgSubmitProposal() {
	detail := make(common.JsDict)
	msg, _ := govv1beta1.NewMsgSubmitProposal(content, Coins1000000uband, SenderAddress)
	emitter.DecodeV1beta1MsgSubmitProposal(msg, detail)
	suite.testCompareJson(
		detail,
		"{\"content\":{\"title\":\"Title\",\"description\":\"Desc\"},\"initial_deposit\":[{\"denom\":\"uband\",\"amount\":\"1000000\"}],\"proposer\":\"band12djkuer9wgqqqqqqqqqqqqqqqqqqqqqqck96t0\"}",
	)
}

func (suite *DecoderTestSuite) TestDecodeMsgSubmitProposal() {
	detail := make(common.JsDict)
	msg, _ := govv1.NewMsgSubmitProposal(
		[]sdk.Msg{banktypes.NewMsgSend(SenderAddress, ReceiverAddress, sdk.Coins{Amount})},
		Coins1000000uband,
		SenderAddress.String(),
		"metadata",
		"title",
		"summary",
		true,
	)
	emitter.DecodeMsgSubmitProposal(msg, detail)
	suite.testCompareJson(
		detail,
		"{\"initial_deposit\":[{\"denom\":\"uband\",\"amount\":\"1000000\"}],\"messages\":[{\"msg\":{\"amount\":[{\"denom\":\"uband\",\"amount\":\"1\"}],\"from_address\":\"band12djkuer9wgqqqqqqqqqqqqqqqqqqqqqqck96t0\",\"to_address\":\"band12fjkxetfwejhyqqqqqqqqqqqqqqqqqqqrhevnq\"},\"type\":\"/cosmos.bank.v1beta1.MsgSend\"}],\"metadata\":\"metadata\",\"proposer\":\"band12djkuer9wgqqqqqqqqqqqqqqqqqqqqqqck96t0\",\"summary\":\"summary\",\"title\":\"title\"}",
	)
}

func (suite *DecoderTestSuite) TestDecodeV1beta1MsgDeposit() {
	detail := make(common.JsDict)
	msg := govv1beta1.NewMsgDeposit(SenderAddress, 1, Coins1000000uband)
	emitter.DecodeV1beta1MsgDeposit(msg, detail)
	suite.testCompareJson(
		detail,
		"{\"amount\":[{\"denom\":\"uband\",\"amount\":\"1000000\"}],\"depositor\":\"band12djkuer9wgqqqqqqqqqqqqqqqqqqqqqqck96t0\",\"proposal_id\":1}",
	)
}

func (suite *DecoderTestSuite) TestDecodeMsgDeposit() {
	detail := make(common.JsDict)
	msg := govv1.NewMsgDeposit(SenderAddress, 1, Coins1000000uband)
	emitter.DecodeMsgDeposit(msg, detail)
	suite.testCompareJson(
		detail,
		"{\"amount\":[{\"denom\":\"uband\",\"amount\":\"1000000\"}],\"depositor\":\"band12djkuer9wgqqqqqqqqqqqqqqqqqqqqqqck96t0\",\"proposal_id\":1}",
	)
}

func (suite *DecoderTestSuite) TestDecodeV1beta1MsgVote() {
	detail := make(common.JsDict)
	msg := govv1beta1.NewMsgVote(SenderAddress, 1, 0)
	emitter.DecodeV1beta1MsgVote(msg, detail)
	suite.testCompareJson(detail,
		"{\"option\":0,\"proposal_id\":1,\"voter\":\"band12djkuer9wgqqqqqqqqqqqqqqqqqqqqqqck96t0\"}",
	)
}

func (suite *DecoderTestSuite) TestDecodeMsgVote() {
	detail := make(common.JsDict)
	msg := govv1.NewMsgVote(SenderAddress, 1, 0, "metadata")
	emitter.DecodeMsgVote(msg, detail)
	suite.testCompareJson(
		detail,
		"{\"metadata\":\"metadata\",\"option\":0,\"proposal_id\":1,\"voter\":\"band12djkuer9wgqqqqqqqqqqqqqqqqqqqqqqck96t0\"}",
	)
}

func (suite *DecoderTestSuite) TestDecodeMsgCreateValidator() {
	detail := make(common.JsDict)
	msg, _ := stakingtypes.NewMsgCreateValidator(
		ValAddress.String(),
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
	msg := stakingtypes.NewMsgEditValidator(ValAddress.String(), Description, &NewRate, &MinSelfDelegation)

	emitter.DecodeMsgEditValidator(msg, detail)
	suite.testCompareJson(
		detail,
		"{\"commission_rate\":\"1.000000000000000000\",\"description\":{\"details\":\"details\",\"identity\":\"identity\",\"moniker\":\"moniker\",\"security_contact\":\"securityContact\",\"website\":\"website\"},\"min_self_delegation\":\"1\",\"validator_address\":\"bandvaloper12eskc6tyv96x7usqqqqqqqqqqqqqqqqqw09xqg\"}",
	)
}

func (suite *DecoderTestSuite) TestDecodeMsgDelegate() {
	detail := make(common.JsDict)
	msg := stakingtypes.NewMsgDelegate(DelegatorAddress.String(), ValAddress.String(), Amount)

	emitter.DecodeMsgDelegate(msg, detail)
	suite.testCompareJson(
		detail,
		"{\"amount\":{\"denom\":\"uband\",\"amount\":\"1\"},\"delegator_address\":\"band1g3jkcet8v96x7usqqqqqqqqqqqqqqqqqus6d5g\",\"validator_address\":\"bandvaloper12eskc6tyv96x7usqqqqqqqqqqqqqqqqqw09xqg\"}",
	)
}

func (suite *DecoderTestSuite) TestDecodeMsgUndelegate() {
	detail := make(common.JsDict)
	msg := stakingtypes.NewMsgUndelegate(DelegatorAddress.String(), ValAddress.String(), Amount)

	emitter.DecodeMsgUndelegate(msg, detail)
	suite.testCompareJson(
		detail,
		"{\"amount\":{\"denom\":\"uband\",\"amount\":\"1\"},\"delegator_address\":\"band1g3jkcet8v96x7usqqqqqqqqqqqqqqqqqus6d5g\",\"validator_address\":\"bandvaloper12eskc6tyv96x7usqqqqqqqqqqqqqqqqqw09xqg\"}",
	)
}

func (suite *DecoderTestSuite) TestDecodeMsgBeginRedelegate() {
	detail := make(common.JsDict)
	msg := stakingtypes.NewMsgBeginRedelegate(
		DelegatorAddress.String(),
		ValAddress.String(),
		ValAddress.String(),
		Amount,
	)

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
		"{\"client_id\":\"tendermint\",\"header\":{\"signed_header\":{\"header\":{\"version\":{\"block\":11,\"app\":2},\"chain_id\":\"testchain1-1\",\"height\":2,\"time\":\"2020-01-02T00:00:00Z\",\"last_block_id\":{\"hash\":\"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=\",\"part_set_header\":{\"total\":10000,\"hash\":\"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=\"}}",
	)
}

func (suite *DecoderTestSuite) TestDecodeMsgUpgradeClient() {
	path := NewOraclePath(suite.chainA, suite.chainB)
	suite.coordinator.Setup(path)

	detail := make(common.JsDict)
	lastHeight := clienttypes.NewHeight(0, uint64(suite.chainB.GetContext().BlockHeight()+1))

	cs, found := suite.chainA.App.GetIBCKeeper().ClientKeeper.GetClientState(
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
	// "{\"client_id\":\"07-tendermint-0\",\"client_state\":{\"chain_id\":\"newChainId\",\"trust_level\":{\"numerator\":1,\"denominator\":3},\"trusting_period\":1209600000000000,\"unbonding_period\":3024000000000000,\"max_clock_drift\":10000000000,\"frozen_height\":{},\"latest_height\":{\"revision_number\":1,\"revision_height\":1},\"proof_specs\":[{\"leaf_spec\":{\"hash\":1,\"prehash_value\":1,\"length\":1,\"prefix\":\"AA==\"},\"inner_spec\":{\"child_order\":[0,1],\"child_size\":33,\"min_prefix_length\":4,\"max_prefix_length\":12,\"hash\":1}},{\"leaf_spec\":{\"hash\":1,\"prehash_value\":1,\"length\":1,\"prefix\":\"AA==\"},\"inner_spec\":{\"child_order\":[0,1],\"child_size\":32,\"min_prefix_length\":1,\"max_prefix_length\":1,\"hash\":1}}],\"upgrade_path\":[\"upgrade\",\"upgradedIBCState\"]},\"consensus_state\":{\"timestamp\":\"0001-01-01T00:00:00Z\",\"root\":{},\"next_validators_hash\":\"6E65787456616C7348617368\"},\"proof_upgrade_client\":\"CiYSJAoidXBncmFkZWRJQkNTdGF0ZS8xOC91cGdyYWRlZENsaWVudAquAQqrAQoHdXBncmFkZRIg47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFUaCQgBGAEgASoBACIlCAESIQG2RppbdEWeFVF5h90HmJZ/OIvuBr5jbE7mh/4a8ey+lSIlCAESIQHZm7f7BAECvMg69fhmRvif+axXjaVvh7wuDvibWJVoJiIlCAESIQGbHEApyKCI6yWJSWKQnvxTXX67FeS/avKzkttknO4VoA==\",\"proof_upgrade_consensus_state\":\"CikSJwoldXBncmFkZWRJQkNTdGF0ZS8xOC91cGdyYWRlZENvbnNTdGF0ZQquAQqrAQoHdXBncmFkZRIg47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFUaCQgBGAEgASoBACIlCAESIQG2RppbdEWeFVF5h90HmJZ/OIvuBr5jbE7mh/4a8ey+lSIlCAESIQHZm7f7BAECvMg69fhmRvif+axXjaVvh7wuDvibWJVoJiIlCAESIQGbHEApyKCI6yWJSWKQnvxTXX67FeS/avKzkttknO4VoA==\",\"signer\":\"band1ws6lm89d6xenm3cms264ejvxk8rurw55t4vpl9\"}" does not contain "{\"client_id\":\"tendermint\",\"header\":{\"signed_header\":{\"header\":{\"version\":{\"block\":11,\"app\":2},\"chain_id\":\"testchain1-1\",\"height\":3,\"time\":\"2020-01-02T00:00:00Z\",\"last_block_id\":{\"hash\":\"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=\",\"part_set_header\":{\"total\":10000,\"hash\":\"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=\"}}"
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
		"{\"client_id\":\"tendermint\",\"misbehaviour\":{\"client_id\":\"tendermint\",\"header_1\":{\"signed_header\":{\"header\":{\"version\":{\"block\":11,\"app\":2},\"chain_id\":\"testchain1-1\",\"height\":2,\"time\":\"2020-01-02T00:00:00Z\",\"last_block_id\":{\"hash\":\"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=\",\"part_set_header\":{\"total\":10000,\"hash\":\"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=\"}},\"last_commit_hash\":",
	)
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
	)
	msg := connectiontypes.NewMsgConnectionOpenTry(
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
		"{\"client_id\":\"07-tendermint-0\",\"client_state\":{\"chain_id\":\"testchain1-1\",\"trust_level\":{\"numerator\":1,\"denominator\":3},\"trusting_period\":1209600000000000,\"unbonding_period\":1814400000000000,\"max_clock_drift\":10000000000,\"frozen_height\":{},\"latest_height\":{\"revision_height\":10},\"proof_specs\":[{\"leaf_spec\":{\"hash\":1,\"prehash_value\":1,\"length\":1,\"prefix\":\"AA==\"},\"inner_spec\":{\"child_order\":[0,1],\"child_size\":33,\"min_prefix_length\":4,\"max_prefix_length\":12,\"hash\":1}},{\"leaf_spec\":{\"hash\":1,\"prehash_value\":1,\"length\":1,\"prefix\":\"AA==\"},\"inner_spec\":{\"child_order\":[0,1],\"child_size\":32,\"min_prefix_length\":1,\"max_prefix_length\":1,\"hash\":1}}],\"upgrade_path\":[\"upgrade\",\"upgradedIBCState\"]},\"consensus_height\":{\"revision_height\":10,\"revision_number\":0},\"counterparty\":{\"client_id\":\"07-tendermint-0\",\"connection_id\":\"connection-0\",\"prefix\":{\"key_prefix\":\"c3RvcmVQcmVmaXhLZXk=\"}},\"counterparty_versions\":[{\"identifier\":\"1\",\"features\":[\"ORDER_ORDERED\",\"ORDER_UNORDERED\"]}],\"delay_period\":500,\"previous_connection_id\":\"\",\"proof_client\":\"\",\"proof_consensus\":\"\",\"proof_height\":{\"revision_height\":10,\"revision_number\":0},\"proof_init\":\"\",\"signer\":\"band12d5kwmn9wgqqqqqqqqqqqqqqqqqqqqqqr057wh\"}",
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
		"{\"client_state\":{\"chain_id\":\"testchain1-1\",\"trust_level\":{\"numerator\":1,\"denominator\":3},\"trusting_period\":1209600000000000,\"unbonding_period\":1814400000000000,\"max_clock_drift\":10000000000,\"frozen_height\":{},\"latest_height\":{\"revision_height\":10},\"proof_specs\":[{\"leaf_spec\":{\"hash\":1,\"prehash_value\":1,\"length\":1,\"prefix\":\"AA==\"},\"inner_spec\":{\"child_order\":[0,1],\"child_size\":33,\"min_prefix_length\":4,\"max_prefix_length\":12,\"hash\":1}},{\"leaf_spec\":{\"hash\":1,\"prehash_value\":1,\"length\":1,\"prefix\":\"AA==\"},\"inner_spec\":{\"child_order\":[0,1],\"child_size\":32,\"min_prefix_length\":1,\"max_prefix_length\":1,\"hash\":1}}],\"upgrade_path\":[\"upgrade\",\"upgradedIBCState\"]},\"connection_id\":\"\",\"consensus_height\":{\"revision_height\":10,\"revision_number\":0},\"counterparty_connection_id\":\"\",\"proof_client\":\"\",\"proof_consensus\":\"\",\"proof_height\":{\"revision_height\":10,\"revision_number\":0},\"proof_try\":\"\",\"signer\":\"band12d5kwmn9wgqqqqqqqqqqqqqqqqqqqqqqr057wh\",\"version\":{\"identifier\":\"1\",\"features\":[\"ORDER_ORDERED\",\"ORDER_UNORDERED\"]}}",
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
		"{\"channel\":{\"state\":2,\"ordering\":2,\"counterparty\":{\"port_id\":\"oracle\",\"channel_id\":\"channel-0\"},\"connection_hops\":[\"connection-0\"],\"version\":\"bandchain-1\"},\"counterparty_version\":\"bandchain-1\",\"port_id\":\"oracle\",\"previous_channel_id\":\"\",\"proof_height\":{\"revision_height\":10,\"revision_number\":0},\"proof_init\":\"\",\"signer\":\"band12d5kwmn9wgqqqqqqqqqqqqqqqqqqqqqqr057wh\"}",
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

func (suite *DecoderTestSuite) TestDecodeFeedsMsgSubmitPrices() {
	detail := make(common.JsDict)

	msg := feedstypes.MsgSubmitSignalPrices{
		Validator: ValAddress.String(),
		Timestamp: 12345678,
		SignalPrices: []feedstypes.SignalPrice{
			{
				Status:   feedstypes.SignalPriceStatusAvailable,
				SignalID: "CS:ETH-USD",
				Price:    3500000000000,
			},
			{
				Status:   feedstypes.SignalPriceStatusUnavailable,
				SignalID: "CS:BTC-USD",
				Price:    0,
			},
		},
	}

	emitter.DecodeFeedsMsgSubmitSignalPrices(&msg, detail)
	suite.testCompareJson(
		detail,
		"{\"signal_prices\":[{\"status\":3,\"signal_id\":\"CS:ETH-USD\",\"price\":3500000000000},{\"status\":2,\"signal_id\":\"CS:BTC-USD\"}],\"timestamp\":12345678,\"validator\":\"bandvaloper12eskc6tyv96x7usqqqqqqqqqqqqqqqqqw09xqg\"}",
	)
}

func (suite *DecoderTestSuite) TestDecodeFeedsMsgSubmitSignals() {
	detail := make(common.JsDict)

	msg := feedstypes.MsgVote{
		Voter: DelegatorAddress.String(),
		Signals: []feedstypes.Signal{
			{
				ID:    "crypto_price.btcusd",
				Power: 30000000000,
			},
			{
				ID:    "crypto_price.ethusd",
				Power: 60000000000,
			},
		},
	}

	emitter.DecodeFeedsMsgVote(&msg, detail)
	suite.testCompareJson(
		detail,
		"{\"signals\":[{\"id\":\"crypto_price.btcusd\",\"power\":30000000000},{\"id\":\"crypto_price.ethusd\",\"power\":60000000000}],\"voter\":\"band1g3jkcet8v96x7usqqqqqqqqqqqqqqqqqus6d5g\"}",
	)
}

func (suite *DecoderTestSuite) TestDecodeFeedsMsgUpdatePriceService() {
	detail := make(common.JsDict)

	msg := feedstypes.MsgUpdateReferenceSourceConfig{
		Admin:                 OwnerAddress.String(),
		ReferenceSourceConfig: feedstypes.NewReferenceSourceConfig("testhash", "1.0.0"),
	}

	emitter.DecodeFeedsMsgUpdateReferenceSourceConfig(&msg, detail)
	suite.testCompareJson(
		detail,
		"{\"admin\":\"band1famkuetjqqqqqqqqqqqqqqqqqqqqqqqqkzrxfg\",\"reference_source_config\":{\"registry_ipfs_hash\":\"testhash\",\"registry_version\":\"1.0.0\"}}",
	)
}

func (suite *DecoderTestSuite) TestDecodeBandtssMsgTransitionGroup() {
	detail := make(common.JsDict)

	msg := bandtsstypes.MsgTransitionGroup{
		Members:   []string{"member1", "member2"},
		Threshold: 2,
		Authority: "some-authority-id",
		ExecTime:  time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC),
	}

	emitter.DecodeBandtssMsgTransitionGroup(&msg, detail)

	expectedJSON := `{"authority":"some-authority-id","exec_time":1577923200000000000,"members":["member1","member2"],"threshold":2}`
	suite.testCompareJson(detail, expectedJSON)
}

func (suite *DecoderTestSuite) TestDecodeGroupMsgReplaceGroup() {
	detail := make(common.JsDict)

	msg := bandtsstypes.MsgForceTransitionGroup{
		IncomingGroupID: 1,
		ExecTime:        time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC),
		Authority:       "authority123",
	}

	emitter.DecodeBandtssMsgForceTransitionGroup(&msg, detail)

	expectedJSON := `{"authority":"authority123","exec_time":1577923200000000000,"incoming_group_id":1}`
	suite.testCompareJson(detail, expectedJSON)
}

func (suite *DecoderTestSuite) TestDecodeBandtssMsgUpdateParams() {
	detail := make(common.JsDict)

	msg := bandtsstypes.MsgUpdateParams{
		Params: bandtsstypes.NewParams(10, 50, 50, sdk.Coins{Amount}),
	}

	emitter.DecodeBandtssMsgUpdateParams(&msg, detail)
	expectedJSON := `{"authority":"","fee":[{"denom":"uband","amount":"1"}],"inactive_penalty_duration":50,"max_transition_duration":50,"reward_percentage":10}`
	suite.testCompareJson(detail, expectedJSON)
}

func (suite *DecoderTestSuite) TestDecodeBandtssMsgActivate() {
	detail := make(common.JsDict)

	msg := bandtsstypes.MsgActivate{
		Sender:  "0x123",
		GroupID: 1,
	}

	emitter.DecodeBandtssMsgActivate(&msg, detail)
	expectedJSON := `{"group_id":1,"sender":"0x123"}`
	suite.testCompareJson(detail, expectedJSON)
}

func (suite *DecoderTestSuite) TestDecodeMsgSubmitDKGRound1() {
	detail := make(common.JsDict)

	msg := tsstypes.MsgSubmitDKGRound1{
		GroupID: 1,
		Round1Info: tsstypes.Round1Info{
			MemberID:           1,
			CoefficientCommits: tss.Points{tssPoint, tssPoint},
			OneTimePubKey:      tssPoint,
			A0Signature:        tssSignature,
			OneTimeSignature:   tssSignature,
		},
		Sender: "0x123",
	}

	emitter.DecodeMsgSubmitDKGRound1(&msg, detail)
	expectedJSON := "{\"group_id\":1,\"round1_info\":{\"member_id\":1,\"coefficient_commits\":[\"706F696E74\",\"706F696E74\"],\"one_time_pub_key\":\"706F696E74\",\"a0_signature\":\"7369676E6174757265\",\"one_time_signature\":\"7369676E6174757265\"},\"sender\":\"0x123\"}"
	suite.testCompareJson(detail, expectedJSON)
}

func (suite *DecoderTestSuite) TestDecodeMsgSubmitDKGRound2() {
	detail := make(common.JsDict)

	msg := tsstypes.MsgSubmitDKGRound2{
		GroupID: 1,
		Round2Info: tsstypes.Round2Info{
			MemberID:              1,
			EncryptedSecretShares: tss.EncSecretShares{tssEncSecretShare},
		},
		Sender: "0x456",
	}

	emitter.DecodeMsgSubmitDKGRound2(&msg, detail)
	expectedJSON := "{\"group_id\":1,\"round2_info\":{\"member_id\":1,\"encrypted_secret_shares\":[\"656E635365637265745368617265\"]},\"sender\":\"0x456\"}"
	suite.testCompareJson(detail, expectedJSON)
}

func (suite *DecoderTestSuite) TestDecodeFeedsMsgUpdateParams() {
	detail := make(common.JsDict)

	msg := feedstypes.MsgUpdateParams{
		Authority: OwnerAddress.String(),
		Params: feedstypes.Params{
			Admin:                         OwnerAddress.String(),
			AllowableBlockTimeDiscrepancy: 30,
			GracePeriod:                   30,
			MinInterval:                   60,
			MaxInterval:                   3600,
			PowerStepThreshold:            1_000_000_000,
			MaxCurrentFeeds:               100,
			CooldownTime:                  30,
			MinDeviationBasisPoint:        50,
			MaxDeviationBasisPoint:        3000,
		},
	}

	emitter.DecodeFeedsMsgUpdateParams(&msg, detail)
	suite.testCompareJson(
		detail,
		"{\"authority\":\"band1famkuetjqqqqqqqqqqqqqqqqqqqqqqqqkzrxfg\",\"params\":{\"admin\":\"band1famkuetjqqqqqqqqqqqqqqqqqqqqqqqqkzrxfg\",\"allowable_block_time_discrepancy\":30,\"grace_period\":30,\"min_interval\":60,\"max_interval\":3600,\"power_step_threshold\":1000000000,\"max_current_feeds\":100,\"cooldown_time\":30,\"min_deviation_basis_point\":50,\"max_deviation_basis_point\":3000}}",
	)
}

func (suite *DecoderTestSuite) TestDecodeRestakeMsgStake() {
	detail := make(common.JsDict)
	msg := restaketypes.NewMsgStake(
		StakerAddress,
		Coins1000000uband,
	)
	emitter.DecodeRestakeMsgStake(msg, detail)
	suite.testCompareJson(
		detail,
		"{\"coins\":[{\"denom\":\"uband\",\"amount\":\"1000000\"}],\"staker_address\":\"band12d6xz6m9wgqqqqqqqqqqqqqqqqqqqqqqtz8edw\"}",
	)
}

func (suite *DecoderTestSuite) TestDecodeRestakeMsgUnstake() {
	detail := make(common.JsDict)
	msg := restaketypes.NewMsgUnstake(
		StakerAddress,
		Coins1000000uband,
	)
	emitter.DecodeRestakeMsgUnstake(msg, detail)
	suite.testCompareJson(
		detail,
		"{\"coins\":[{\"denom\":\"uband\",\"amount\":\"1000000\"}],\"staker_address\":\"band12d6xz6m9wgqqqqqqqqqqqqqqqqqqqqqqtz8edw\"}",
	)
}

func (suite *DecoderTestSuite) TestDecodeRestakeMsgUpdateParams() {
	detail := make(common.JsDict)
	params := restaketypes.NewParams([]string{"stBand", "band"})
	msg := restaketypes.NewMsgUpdateParams(
		AuthorityAddress.String(),
		params,
	)
	emitter.DecodeRestakeMsgUpdateParams(msg, detail)
	suite.testCompareJson(
		detail,
		"{\"authority\":\"band1g96hg6r0wf5hg7gqqqqqqqqqqqqqqqqq4rjgsx\",\"params\":{\"allowed_denoms\":[\"stBand\",\"band\"]}}",
	)
}

func (suite *DecoderTestSuite) TestDecodeTunnelMsgCreateTunnel() {
	detail := make(common.JsDict)
	msg, err := tunneltypes.NewMsgCreateTSSTunnel(
		[]tunneltypes.SignalDeviation{
			tunneltypes.NewSignalDeviation("CS:BAND-USD", 10000, 10000),
		},
		60,
		"ethereum",
		"0xcabe9a5e6249c893a4b4fc263",
		feedstypes.ENCODER_TICK_ABI,
		sdk.NewCoins(sdk.NewCoin("uband", math.NewInt(5))),
		CreatorAddress,
	)
	suite.NoError(err)

	emitter.DecodeTunnelMsgCreateTunnel(msg, detail)
	suite.testCompareJson(
		detail,
		"{\"creator\":\"band1vdex2ct5daeqqqqqqqqqqqqqqqqqqqqqqgzyx3\",\"encoder\":2,\"initial_deposit\":[{\"denom\":\"uband\",\"amount\":\"5\"}],\"interval\":60,\"route\":{\"destination_chain_id\":\"ethereum\",\"destination_contract_address\":\"0xcabe9a5e6249c893a4b4fc263\"},\"route_type\":\"/band.tunnel.v1beta1.TSSRoute\",\"signal_deviations\":[{\"signal_id\":\"CS:BAND-USD\",\"soft_deviation_bps\":10000,\"hard_deviation_bps\":10000}]}",
	)
}

func (suite *DecoderTestSuite) TestDecodeTunnelMsgUpdateAndResetTunnel() {
	detail := make(common.JsDict)
	msg := tunneltypes.NewMsgUpdateAndResetTunnel(
		1,
		[]tunneltypes.SignalDeviation{
			tunneltypes.NewSignalDeviation("CS:BAND-USD", 10000, 10000),
		},
		60,
		CreatorAddress.String(),
	)

	emitter.DecodeTunnelMsgUpdateAndResetTunnel(msg, detail)
	suite.testCompareJson(
		detail,
		"{\"creator\":\"band1vdex2ct5daeqqqqqqqqqqqqqqqqqqqqqqgzyx3\",\"interval\":60,\"signal_deviations\":[{\"signal_id\":\"CS:BAND-USD\",\"soft_deviation_bps\":10000,\"hard_deviation_bps\":10000}],\"tunnel_id\":1}",
	)
}

func (suite *DecoderTestSuite) TestDecodeTunnelMsgActivate() {
	detail := make(common.JsDict)
	msg := tunneltypes.NewMsgActivate(
		1,
		CreatorAddress.String(),
	)

	emitter.DecodeTunnelMsgActivate(msg, detail)
	suite.testCompareJson(
		detail,
		"{\"creator\":\"band1vdex2ct5daeqqqqqqqqqqqqqqqqqqqqqqgzyx3\",\"tunnel_id\":1}",
	)
}

func (suite *DecoderTestSuite) TestDecodeTunnelMsgDeactivate() {
	detail := make(common.JsDict)
	msg := tunneltypes.NewMsgDeactivate(
		1,
		CreatorAddress.String(),
	)

	emitter.DecodeTunnelMsgDeactivate(msg, detail)
	suite.testCompareJson(
		detail,
		"{\"creator\":\"band1vdex2ct5daeqqqqqqqqqqqqqqqqqqqqqqgzyx3\",\"tunnel_id\":1}",
	)
}

func (suite *DecoderTestSuite) TestDecodeTunnelMsgTriggerTunnel() {
	detail := make(common.JsDict)
	msg := tunneltypes.NewMsgTriggerTunnel(
		1,
		CreatorAddress.String(),
	)

	emitter.DecodeTunnelMsgTriggerTunnel(msg, detail)
	suite.testCompareJson(
		detail,
		"{\"creator\":\"band1vdex2ct5daeqqqqqqqqqqqqqqqqqqqqqqgzyx3\",\"tunnel_id\":1}",
	)
}

func (suite *DecoderTestSuite) TestDecodeTunnelMsgDepositToTunnel() {
	detail := make(common.JsDict)
	msg := tunneltypes.NewMsgDepositToTunnel(
		1,
		sdk.NewCoins(sdk.NewCoin("uband", math.NewInt(5))),
		CreatorAddress.String(),
	)

	emitter.DecodeTunnelMsgDepositToTunnel(msg, detail)
	suite.testCompareJson(
		detail,
		"{\"amount\":[{\"denom\":\"uband\",\"amount\":\"5\"}],\"depositor\":\"band1vdex2ct5daeqqqqqqqqqqqqqqqqqqqqqqgzyx3\",\"tunnel_id\":1}",
	)
}

func (suite *DecoderTestSuite) TestDecodeTunnelMsgWithdrawFromTunnel() {
	detail := make(common.JsDict)
	msg := tunneltypes.NewMsgWithdrawFromTunnel(
		1,
		sdk.NewCoins(sdk.NewCoin("uband", math.NewInt(5))),
		CreatorAddress.String(),
	)

	emitter.DecodeTunnelMsgWithdrawFromTunnel(msg, detail)
	suite.testCompareJson(
		detail,
		"{\"amount\":[{\"denom\":\"uband\",\"amount\":\"5\"}],\"tunnel_id\":1,\"withdrawer\":\"band1vdex2ct5daeqqqqqqqqqqqqqqqqqqqqqqgzyx3\"}",
	)
}

func (suite *DecoderTestSuite) TestDecodeTunnelMsgUpdateParams() {
	detail := make(common.JsDict)
	msg := tunneltypes.NewMsgUpdateParams(
		AuthorityAddress.String(),
		tunneltypes.DefaultParams(),
	)

	emitter.DecodeTunnelMsgUpdateParams(msg, detail)
	suite.testCompareJson(
		detail,
		"{\"authority\":\"band1g96hg6r0wf5hg7gqqqqqqqqqqqqqqqqq4rjgsx\",\"params\":{\"min_deposit\":[{\"denom\":\"uband\",\"amount\":\"1000000000\"}],\"min_interval\":1,\"max_signals\":100,\"base_packet_fee\":[{\"denom\":\"uband\",\"amount\":\"10000\"}]}}",
	)
}

func TestDecoderTestSuite(t *testing.T) {
	suite.Run(t, new(DecoderTestSuite))
}
