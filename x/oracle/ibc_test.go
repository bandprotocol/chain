package oracle_test

import (
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	clienttypes "github.com/cosmos/ibc-go/v2/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v2/modules/core/04-channel/types"
	"github.com/stretchr/testify/suite"

	"github.com/bandprotocol/chain/v2/pkg/obi"

	ibctesting "github.com/bandprotocol/chain/v2/testing"
	"github.com/bandprotocol/chain/v2/testing/testapp"
	"github.com/bandprotocol/chain/v2/x/oracle/types"
)

type OracleTestSuite struct {
	suite.Suite

	coordinator *ibctesting.Coordinator

	// testing chains used for convenience and readability
	chainA *ibctesting.TestChain
	chainB *ibctesting.TestChain

	path *ibctesting.Path
}

func (suite *OracleTestSuite) SetupTest() {
	suite.coordinator = ibctesting.NewCoordinator(suite.T(), 3)
	suite.chainA = suite.coordinator.GetChain(ibctesting.GetChainID(0))
	suite.chainB = suite.coordinator.GetChain(ibctesting.GetChainID(1))

	suite.path = ibctesting.NewPath(suite.chainA, suite.chainB)
	suite.path.EndpointA.ChannelConfig.PortID = ibctesting.OraclePort
	suite.path.EndpointB.ChannelConfig.PortID = ibctesting.OraclePort

	suite.coordinator.Setup(suite.path)

}

func (suite *OracleTestSuite) sendOracleRequestPacket(path *ibctesting.Path, seq uint64, oracleRequestPacket types.OracleRequestPacketData, timeoutHeight clienttypes.Height) channeltypes.Packet {
	packet := channeltypes.NewPacket(
		oracleRequestPacket.GetBytes(),
		seq,
		path.EndpointA.ChannelConfig.PortID,
		path.EndpointA.ChannelID,
		path.EndpointB.ChannelConfig.PortID,
		path.EndpointB.ChannelID,
		timeoutHeight,
		0,
	)
	err := path.EndpointA.SendPacket(packet)
	suite.Require().NoError(err)
	return packet
}

func (suite *OracleTestSuite) checkChainBTreasuryBalances(expect sdk.Coins) {
	treasuryBalances := suite.chainB.App.BankKeeper.GetAllBalances(suite.chainB.GetContext(), suite.chainB.Treasury)
	suite.Require().Equal(expect, treasuryBalances)
}

func (suite *OracleTestSuite) checkChainBSenderBalances(expect sdk.Coins) {
	b := suite.chainB.App.BankKeeper.GetAllBalances(suite.chainB.GetContext(), suite.chainB.SenderAccount.GetAddress())
	suite.Require().Equal(expect, b)
}

// constructs a send from chainA to chainB on the established channel/connection
// and sends the same coin back from chainB to chainA.
func (suite *OracleTestSuite) TestHandleIBCRequestSuccess() {
	path := suite.path
	// send request from A to B
	timeoutHeight := clienttypes.NewHeight(0, 110)
	oracleRequestPacket := types.NewOracleRequestPacketData(
		path.EndpointA.ClientID,
		1,
		[]byte("beeb"),
		2,
		2,
		sdk.NewCoins(sdk.NewCoin("uband", sdk.NewInt(6000000))),
		testapp.TestDefaultPrepareGas,
		testapp.TestDefaultExecuteGas,
	)
	packet := suite.sendOracleRequestPacket(path, 1, oracleRequestPacket, timeoutHeight)

	ack := channeltypes.NewResultAcknowledgement(types.ModuleCdc.MustMarshalJSON(types.NewOracleRequestPacketAcknowledgement(1)))
	err := path.RelayPacket(packet, ack.Acknowledgement())
	suite.Require().NoError(err) // relay committed

	suite.checkChainBTreasuryBalances(sdk.NewCoins(sdk.NewCoin("uband", sdk.NewInt(6000000))))
	suite.checkChainBSenderBalances(sdk.NewCoins(sdk.NewCoin("uband", sdk.NewInt(4000000))))

	raws1 := []types.RawReport{types.NewRawReport(1, 0, []byte("data1")), types.NewRawReport(2, 0, []byte("data2")), types.NewRawReport(3, 0, []byte("data3"))}
	suite.chainB.SendReport(1, raws1, testapp.Validators[0])
	suite.Require().NoError(err)

	raws2 := []types.RawReport{types.NewRawReport(1, 0, []byte("data1")), types.NewRawReport(2, 0, []byte("data2")), types.NewRawReport(3, 0, []byte("data3"))}
	suite.chainB.SendReport(1, raws2, testapp.Validators[1])
	suite.Require().NoError(err)

	oracleResponsePacket := types.NewOracleResponsePacketData(path.EndpointA.ClientID, 1, 2, 1577923380, 1577923405, types.RESOLVE_STATUS_SUCCESS, []byte("beeb"))
	responsePacket := channeltypes.NewPacket(
		oracleResponsePacket.GetBytes(),
		1,
		path.EndpointB.ChannelConfig.PortID,
		path.EndpointB.ChannelID,
		path.EndpointA.ChannelConfig.PortID,
		path.EndpointA.ChannelID,
		clienttypes.ZeroHeight(),
		1577924005000000000,
	)
	expectCommitment := channeltypes.CommitPacket(suite.chainB.Codec, responsePacket)
	commitment := suite.chainB.App.IBCKeeper.ChannelKeeper.GetPacketCommitment(suite.chainB.GetContext(), path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID, 1)
	suite.Equal(expectCommitment, commitment)
}

func (suite *OracleTestSuite) TestIBCPrepareValidateBasicFail() {
	path := suite.path

	clientID := path.EndpointA.ClientID
	coins := sdk.NewCoins(sdk.NewCoin("uband", sdk.NewInt(6000000)))

	oracleRequestPackets := []types.OracleRequestPacketData{
		types.NewOracleRequestPacketData(clientID, 1, []byte(strings.Repeat("beeb", 65)), 1, 1, coins, testapp.TestDefaultPrepareGas, testapp.TestDefaultExecuteGas),
		types.NewOracleRequestPacketData(clientID, 1, []byte("beeb"), 1, 0, coins, testapp.TestDefaultPrepareGas, testapp.TestDefaultExecuteGas),
		types.NewOracleRequestPacketData(clientID, 1, []byte("beeb"), 1, 2, coins, testapp.TestDefaultPrepareGas, testapp.TestDefaultExecuteGas),
		types.NewOracleRequestPacketData(strings.Repeat(clientID, 9), 1, []byte("beeb"), 1, 1, coins, testapp.TestDefaultPrepareGas, testapp.TestDefaultExecuteGas),
		types.NewOracleRequestPacketData(clientID, 1, []byte("beeb"), 1, 1, coins, 0, testapp.TestDefaultExecuteGas),
		types.NewOracleRequestPacketData(clientID, 1, []byte("beeb"), 1, 1, coins, testapp.TestDefaultPrepareGas, 0),
		types.NewOracleRequestPacketData(clientID, 1, []byte("beeb"), 1, 1, coins, types.MaximumOwasmGas, types.MaximumOwasmGas),
		types.NewOracleRequestPacketData(clientID, 1, []byte("beeb"), 1, 1, testapp.BadCoins, testapp.TestDefaultPrepareGas, testapp.TestDefaultExecuteGas),
	}
	expectedErrs := []string{
		"got: 260, max: 256: too large calldata",
		"got: 0: invalid min count",
		"got: 1, min count: 2: invalid ask count",
		"got: 135, max: 128: too long client id",
		"invalid prepare gas: 0: invalid owasm gas",
		"invalid execute gas: 0: invalid owasm gas",
		"sum of prepare gas and execute gas (16000000) exceed 8000000: invalid owasm gas",
		"-1uband: invalid coins",
	}

	timeoutHeight := clienttypes.NewHeight(0, 110)
	for i, requestPacket := range oracleRequestPackets {
		packet := suite.sendOracleRequestPacket(path, uint64(i)+1, requestPacket, timeoutHeight)

		ack := channeltypes.NewErrorAcknowledgement(expectedErrs[i])
		err := path.RelayPacket(packet, ack.Acknowledgement())
		suite.Require().NoError(err) // relay committed
	}
}

func (suite *OracleTestSuite) TestIBCPrepareRequestNotEnoughFund() {
	path := suite.path

	// send request from A to B
	timeoutHeight := clienttypes.NewHeight(0, 110)
	oracleRequestPacket := types.NewOracleRequestPacketData(
		path.EndpointA.ClientID,
		1,
		[]byte("beeb"),
		1,
		1,
		sdk.NewCoins(sdk.NewCoin("uband", sdk.NewInt(3000000))),
		testapp.TestDefaultPrepareGas,
		testapp.TestDefaultExecuteGas,
	)

	// Use Carol as a relayer
	carol := testapp.Carol
	carolExpectedBalance := sdk.NewCoins(sdk.NewCoin("uband", sdk.NewInt(2500000)))
	suite.chainB.SendMsgs(banktypes.NewMsgSend(
		suite.chainB.SenderAccount.GetAddress(),
		carol.Address,
		carolExpectedBalance,
	))
	suite.chainB.SenderPrivKey = carol.PrivKey
	suite.chainB.SenderAccount = suite.chainB.App.AccountKeeper.GetAccount(suite.chainB.GetContext(), carol.Address)

	packet := suite.sendOracleRequestPacket(path, 1, oracleRequestPacket, timeoutHeight)

	ack := channeltypes.NewErrorAcknowledgement("500000uband is smaller than 1000000uband: insufficient funds")
	err := path.RelayPacket(packet, ack.Acknowledgement())
	suite.Require().NoError(err) // relay committed

	carolBalance := suite.chainB.App.BankKeeper.GetAllBalances(suite.chainB.GetContext(), carol.Address)
	suite.Require().Equal(carolExpectedBalance, carolBalance)
}

func (suite *OracleTestSuite) TestIBCPrepareRequestNotEnoughFeeLimit() {
	path := suite.path
	expectedBalance := suite.chainB.App.BankKeeper.GetAllBalances(suite.chainB.GetContext(), suite.chainB.SenderAccount.GetAddress())

	// send request from A to B
	timeoutHeight := clienttypes.NewHeight(0, 110)
	oracleRequestPacket := types.NewOracleRequestPacketData(
		path.EndpointA.ClientID,
		1,
		[]byte("beeb"),
		1,
		1,
		sdk.NewCoins(sdk.NewCoin("uband", sdk.NewInt(2000000))),
		testapp.TestDefaultPrepareGas,
		testapp.TestDefaultExecuteGas,
	)
	packet := suite.sendOracleRequestPacket(path, 1, oracleRequestPacket, timeoutHeight)

	ack := channeltypes.NewErrorAcknowledgement("require: 3000000uband, max: 2000000uband: not enough fee")
	err := path.RelayPacket(packet, ack.Acknowledgement())
	suite.Require().NoError(err) // relay committed

	suite.checkChainBSenderBalances(expectedBalance)
}

func (suite *OracleTestSuite) TestIBCPrepareRequestInvalidCalldataSize() {
	path := suite.path

	// send request from A to B
	timeoutHeight := clienttypes.NewHeight(0, 110)
	oracleRequestPacket := types.NewOracleRequestPacketData(
		path.EndpointA.ClientID,
		1,
		[]byte(strings.Repeat("beeb", 2000)),
		1,
		1,
		sdk.NewCoins(sdk.NewCoin("uband", sdk.NewInt(3000000))),
		testapp.TestDefaultPrepareGas,
		testapp.TestDefaultExecuteGas,
	)
	packet := suite.sendOracleRequestPacket(path, 1, oracleRequestPacket, timeoutHeight)

	ack := channeltypes.NewErrorAcknowledgement("got: 8000, max: 256: too large calldata")
	err := path.RelayPacket(packet, ack.Acknowledgement())
	suite.Require().NoError(err) // relay committed
}

func (suite *OracleTestSuite) TestIBCPrepareRequestNotEnoughPrepareGas() {
	path := suite.path
	// send request from A to B
	timeoutHeight := clienttypes.NewHeight(0, 110)
	oracleRequestPacket := types.NewOracleRequestPacketData(
		path.EndpointA.ClientID,
		1,
		[]byte("beeb"),
		1,
		1,
		sdk.NewCoins(sdk.NewCoin("uband", sdk.NewInt(3000000))),
		100,
		testapp.TestDefaultExecuteGas,
	)
	packet := suite.sendOracleRequestPacket(path, 1, oracleRequestPacket, timeoutHeight)

	ack := channeltypes.NewErrorAcknowledgement("out-of-gas while executing the wasm script: bad wasm execution")
	err := path.RelayPacket(packet, ack.Acknowledgement())
	suite.Require().NoError(err) // relay committed
}

func (suite *OracleTestSuite) TestIBCPrepareRequestInvalidAskCountFail() {
	path := suite.path

	// send request from A to B
	timeoutHeight := clienttypes.NewHeight(0, 110)
	oracleRequestPacket := types.NewOracleRequestPacketData(
		path.EndpointA.ClientID,
		1,
		[]byte("beeb"),
		17,
		1,
		sdk.NewCoins(sdk.NewCoin("uband", sdk.NewInt(3000000))),
		testapp.TestDefaultPrepareGas,
		testapp.TestDefaultExecuteGas,
	)
	packet := suite.sendOracleRequestPacket(path, 1, oracleRequestPacket, timeoutHeight)

	ack := channeltypes.NewErrorAcknowledgement("got: 17, max: 16: invalid ask count")
	err := path.RelayPacket(packet, ack.Acknowledgement())
	suite.Require().NoError(err) // relay committed

	oracleRequestPacket = types.NewOracleRequestPacketData(
		path.EndpointA.ClientID,
		1,
		[]byte("beeb"),
		3,
		1,
		sdk.NewCoins(sdk.NewCoin("uband", sdk.NewInt(3000000))),
		testapp.TestDefaultPrepareGas,
		testapp.TestDefaultExecuteGas,
	)
	packet = suite.sendOracleRequestPacket(path, 2, oracleRequestPacket, timeoutHeight)

	ack = channeltypes.NewErrorAcknowledgement("2 < 3: insufficient available validators")
	err = path.RelayPacket(packet, ack.Acknowledgement())
	suite.Require().NoError(err) // relay committed
}

func (suite *OracleTestSuite) TestIBCPrepareRequestBaseOwasmFeePanic() {
	path := suite.path

	params := suite.chainB.App.OracleKeeper.GetParams(suite.chainB.GetContext())
	params.BaseOwasmGas = 100000000
	params.PerValidatorRequestGas = 0
	suite.chainB.App.OracleKeeper.SetParams(suite.chainB.GetContext(), params)

	// send request from A to B
	timeoutHeight := clienttypes.NewHeight(0, 110)
	oracleRequestPacket := types.NewOracleRequestPacketData(
		path.EndpointA.ClientID,
		1,
		[]byte("beeb"),
		1,
		1,
		sdk.NewCoins(sdk.NewCoin("uband", sdk.NewInt(3000000))),
		testapp.TestDefaultPrepareGas,
		testapp.TestDefaultExecuteGas,
	)
	packet := suite.sendOracleRequestPacket(path, 1, oracleRequestPacket, timeoutHeight)

	// ConsumeGas panics due to insufficient gas, so ErrAcknowledgement is not created.
	ack := channeltypes.NewErrorAcknowledgement("")
	err := path.RelayPacket(packet, ack.Acknowledgement())
	suite.Require().Contains(err.Error(), "BASE_OWASM_FEE; gasWanted: 1000000")
}

func (suite *OracleTestSuite) TestIBCPrepareRequestPerValidatorRequestFeePanic() {
	path := suite.path

	params := suite.chainB.App.OracleKeeper.GetParams(suite.chainB.GetContext())
	params.PerValidatorRequestGas = 100000000
	suite.chainB.App.OracleKeeper.SetParams(suite.chainB.GetContext(), params)

	// send request from A to B
	timeoutHeight := clienttypes.NewHeight(0, 110)
	oracleRequestPacket := types.NewOracleRequestPacketData(
		path.EndpointA.ClientID,
		1,
		[]byte("beeb"),
		1,
		1,
		sdk.NewCoins(sdk.NewCoin("uband", sdk.NewInt(3000000))),
		testapp.TestDefaultPrepareGas,
		testapp.TestDefaultExecuteGas,
	)
	packet := suite.sendOracleRequestPacket(path, 1, oracleRequestPacket, timeoutHeight)

	// ConsumeGas panics due to insufficient gas, so ErrAcknowledgement is not created.
	ack := channeltypes.NewErrorAcknowledgement("")
	err := path.RelayPacket(packet, ack.Acknowledgement())
	suite.Require().Contains(err.Error(), "PER_VALIDATOR_REQUEST_FEE; gasWanted: 1000000")
}

func (suite *OracleTestSuite) TestIBCPrepareRequestOracleScriptNotFound() {
	path := suite.path

	// send request from A to B
	timeoutHeight := clienttypes.NewHeight(0, 110)
	oracleRequestPacket := types.NewOracleRequestPacketData(
		path.EndpointA.ClientID,
		100,
		[]byte("beeb"),
		1,
		1,
		sdk.NewCoins(sdk.NewCoin("uband", sdk.NewInt(3000000))),
		testapp.TestDefaultPrepareGas,
		testapp.TestDefaultExecuteGas,
	)
	packet := suite.sendOracleRequestPacket(path, 1, oracleRequestPacket, timeoutHeight)

	ack := channeltypes.NewErrorAcknowledgement("id: 100: oracle script not found")
	err := path.RelayPacket(packet, ack.Acknowledgement())
	suite.Require().NoError(err) // relay committed
}

func (suite *OracleTestSuite) TestIBCPrepareRequestBadWasmExecutionFail() {
	path := suite.path

	// send request from A to B
	timeoutHeight := clienttypes.NewHeight(0, 110)
	oracleRequestPacket := types.NewOracleRequestPacketData(
		path.EndpointA.ClientID,
		2,
		[]byte("beeb"),
		1,
		1,
		sdk.NewCoins(sdk.NewCoin("uband", sdk.NewInt(3000000))),
		testapp.TestDefaultPrepareGas,
		testapp.TestDefaultExecuteGas,
	)
	packet := suite.sendOracleRequestPacket(path, 1, oracleRequestPacket, timeoutHeight)

	ack := channeltypes.NewErrorAcknowledgement("OEI action to invoke is not available: bad wasm execution")
	err := path.RelayPacket(packet, ack.Acknowledgement())
	suite.Require().NoError(err) // relay committed
}

func (suite *OracleTestSuite) TestIBCPrepareRequestWithEmptyRawRequest() {
	path := suite.path

	// send request from A to B
	timeoutHeight := clienttypes.NewHeight(0, 110)
	oracleRequestPacket := types.NewOracleRequestPacketData(
		path.EndpointA.ClientID,
		3,
		[]byte("beeb"),
		1,
		1,
		sdk.NewCoins(sdk.NewCoin("uband", sdk.NewInt(3000000))),
		testapp.TestDefaultPrepareGas,
		testapp.TestDefaultExecuteGas,
	)
	packet := suite.sendOracleRequestPacket(path, 1, oracleRequestPacket, timeoutHeight)

	ack := channeltypes.NewErrorAcknowledgement("empty raw requests")
	err := path.RelayPacket(packet, ack.Acknowledgement())
	suite.Require().NoError(err) // relay committed
}

func (suite *OracleTestSuite) TestIBCPrepareRequestUnknownDataSource() {
	path := suite.path

	// send request from A to B
	timeoutHeight := clienttypes.NewHeight(0, 110)
	oracleRequestPacket := types.NewOracleRequestPacketData(
		path.EndpointA.ClientID,
		4,
		[]byte("beeb"),
		1,
		1,
		sdk.NewCoins(sdk.NewCoin("uband", sdk.NewInt(3000000))),
		testapp.TestDefaultPrepareGas,
		testapp.TestDefaultExecuteGas,
	)
	packet := suite.sendOracleRequestPacket(path, 1, oracleRequestPacket, timeoutHeight)

	ack := channeltypes.NewErrorAcknowledgement("runtime error while executing the Wasm script: bad wasm execution")
	err := path.RelayPacket(packet, ack.Acknowledgement())
	suite.Require().NoError(err) // relay committed
}

func (suite *OracleTestSuite) TestIBCPrepareRequestInvalidDataSourceCount() {
	path := suite.path

	params := suite.chainB.App.OracleKeeper.GetParams(suite.chainB.GetContext())
	params.MaxRawRequestCount = 3
	suite.chainB.App.OracleKeeper.SetParams(suite.chainB.GetContext(), params)

	// send request from A to B
	timeoutHeight := clienttypes.NewHeight(0, 110)
	oracleRequestPacket := types.NewOracleRequestPacketData(
		path.EndpointA.ClientID,
		4,
		obi.MustEncode(testapp.Wasm4Input{
			IDs:      []int64{1, 2, 3, 4},
			Calldata: "beeb",
		}),
		1,
		1,
		sdk.NewCoins(sdk.NewCoin("uband", sdk.NewInt(4000000))),
		testapp.TestDefaultPrepareGas,
		testapp.TestDefaultExecuteGas,
	)
	packet := suite.sendOracleRequestPacket(path, 1, oracleRequestPacket, timeoutHeight)

	ack := channeltypes.NewErrorAcknowledgement("too many external data requests: bad wasm execution")
	err := path.RelayPacket(packet, ack.Acknowledgement())
	suite.Require().NoError(err) // relay committed
}

func (suite *OracleTestSuite) TestIBCPrepareRequestTooMuchWasmGas() {
	path := suite.path

	// send request from A to B
	timeoutHeight := clienttypes.NewHeight(0, 110)
	oracleRequestPacket := types.NewOracleRequestPacketData(
		path.EndpointA.ClientID,
		6,
		[]byte("beeb"),
		1,
		1,
		sdk.NewCoins(sdk.NewCoin("uband", sdk.NewInt(3000000))),
		testapp.TestDefaultPrepareGas,
		testapp.TestDefaultExecuteGas,
	)
	packet := suite.sendOracleRequestPacket(path, 1, oracleRequestPacket, timeoutHeight)

	ack := channeltypes.NewErrorAcknowledgement("out-of-gas while executing the wasm script: bad wasm execution")
	err := path.RelayPacket(packet, ack.Acknowledgement())
	suite.Require().NoError(err) // relay committed
}

func (suite *OracleTestSuite) TestIBCPrepareRequestTooLargeCalldata() {
	path := suite.path
	// send request from A to B
	timeoutHeight := clienttypes.NewHeight(0, 110)
	oracleRequestPacket := types.NewOracleRequestPacketData(
		path.EndpointA.ClientID,
		8,
		[]byte("beeb"),
		1,
		1,
		sdk.NewCoins(sdk.NewCoin("uband", sdk.NewInt(3000000))),
		testapp.TestDefaultPrepareGas,
		testapp.TestDefaultExecuteGas,
	)
	packet := suite.sendOracleRequestPacket(path, 1, oracleRequestPacket, timeoutHeight)

	ack := channeltypes.NewErrorAcknowledgement("span to write is too small: bad wasm execution")
	err := path.RelayPacket(packet, ack.Acknowledgement())
	suite.Require().NoError(err) // relay committed
}

func (suite *OracleTestSuite) TestIBCResolveRequestOutOfGas() {
	path := suite.path

	// send request from A to B
	timeoutHeight := clienttypes.NewHeight(0, 110)
	oracleRequestPacket := types.NewOracleRequestPacketData(
		path.EndpointA.ClientID,
		1,
		[]byte("beeb"),
		2,
		1,
		sdk.NewCoins(sdk.NewCoin("uband", sdk.NewInt(6000000))),
		testapp.TestDefaultPrepareGas,
		1,
	)
	packet := suite.sendOracleRequestPacket(path, 1, oracleRequestPacket, timeoutHeight)

	ack := channeltypes.NewResultAcknowledgement(types.ModuleCdc.MustMarshalJSON(types.NewOracleRequestPacketAcknowledgement(1)))
	err := path.RelayPacket(packet, ack.Acknowledgement())
	suite.Require().NoError(err) // relay committed

	suite.checkChainBTreasuryBalances(sdk.NewCoins(sdk.NewCoin("uband", sdk.NewInt(6000000))))
	suite.checkChainBSenderBalances(sdk.NewCoins(sdk.NewCoin("uband", sdk.NewInt(4000000))))

	raws := []types.RawReport{types.NewRawReport(1, 0, []byte("data1")), types.NewRawReport(2, 0, []byte("data2")), types.NewRawReport(3, 0, []byte("data3"))}
	suite.chainB.SendReport(1, raws, testapp.Validators[0])

	commitment := suite.chainB.App.IBCKeeper.ChannelKeeper.GetPacketCommitment(suite.chainB.GetContext(), path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID, 1)

	oracleResponsePacket := types.NewOracleResponsePacketData(path.EndpointA.ClientID, 1, 1, 1577923380, 1577923400, types.RESOLVE_STATUS_FAILURE, []byte{})
	responsePacket := channeltypes.NewPacket(
		oracleResponsePacket.GetBytes(),
		1,
		path.EndpointB.ChannelConfig.PortID,
		path.EndpointB.ChannelID,
		path.EndpointA.ChannelConfig.PortID,
		path.EndpointA.ChannelID,
		clienttypes.ZeroHeight(),
		1577924000000000000,
	)
	expectCommitment := channeltypes.CommitPacket(suite.chainB.Codec, responsePacket)
	suite.Equal(expectCommitment, commitment)
}

func (suite *OracleTestSuite) TestIBCResolveReadNilExternalData() {
	path := suite.path

	// send request from A to B
	timeoutHeight := clienttypes.NewHeight(0, 110)
	oracleRequestPacket := types.NewOracleRequestPacketData(
		path.EndpointA.ClientID,
		4,
		obi.MustEncode(testapp.Wasm4Input{IDs: []int64{1, 2}, Calldata: string("beeb")}),
		2,
		2,
		sdk.NewCoins(sdk.NewCoin("uband", sdk.NewInt(4000000))),
		testapp.TestDefaultPrepareGas,
		testapp.TestDefaultExecuteGas,
	)
	packet := suite.sendOracleRequestPacket(path, 1, oracleRequestPacket, timeoutHeight)

	ack := channeltypes.NewResultAcknowledgement(types.ModuleCdc.MustMarshalJSON(types.NewOracleRequestPacketAcknowledgement(1)))
	err := path.RelayPacket(packet, ack.Acknowledgement())
	suite.Require().NoError(err) // relay committed

	suite.checkChainBTreasuryBalances(sdk.NewCoins(sdk.NewCoin("uband", sdk.NewInt(4000000))))
	suite.checkChainBSenderBalances(sdk.NewCoins(sdk.NewCoin("uband", sdk.NewInt(6000000))))

	raws1 := []types.RawReport{types.NewRawReport(0, 0, nil), types.NewRawReport(1, 0, []byte("beebd2v1"))}
	suite.chainB.SendReport(1, raws1, testapp.Validators[0])

	raws2 := []types.RawReport{types.NewRawReport(0, 0, []byte("beebd1v2")), types.NewRawReport(1, 0, nil)}
	suite.chainB.SendReport(1, raws2, testapp.Validators[1])

	commitment := suite.chainB.App.IBCKeeper.ChannelKeeper.GetPacketCommitment(suite.chainB.GetContext(), path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID, 1)

	oracleResponsePacket := types.NewOracleResponsePacketData(
		path.EndpointA.ClientID,
		1,
		2,
		1577923380,
		1577923405,
		types.RESOLVE_STATUS_SUCCESS,
		obi.MustEncode(testapp.Wasm4Output{Ret: "beebd1v2beebd2v1"}),
	)
	responsePacket := channeltypes.NewPacket(
		oracleResponsePacket.GetBytes(),
		1,
		path.EndpointB.ChannelConfig.PortID,
		path.EndpointB.ChannelID,
		path.EndpointA.ChannelConfig.PortID,
		path.EndpointA.ChannelID,
		clienttypes.ZeroHeight(),
		1577924005000000000,
	)
	expectCommitment := channeltypes.CommitPacket(suite.chainB.Codec, responsePacket)
	suite.Equal(expectCommitment, commitment)
}

func (suite *OracleTestSuite) TestIBCResolveRequestNoReturnData() {
	path := suite.path

	suite.chainB.App.OracleKeeper.SetRequest(suite.chainB.GetContext(), 1, types.NewRequest(
		// 3rd Wasm - do nothing
		3, []byte("beeb"), []sdk.ValAddress{testapp.Validators[0].ValAddress, testapp.Validators[1].ValAddress}, 1,
		suite.chainB.GetContext().BlockHeight()-1, testapp.ParseTime(1577923380), path.EndpointA.ClientID, []types.RawRequest{
			types.NewRawRequest(1, 1, []byte("beeb")),
		}, &types.IBCChannel{PortId: path.EndpointB.ChannelConfig.PortID, ChannelId: path.EndpointB.ChannelID}, 0,
	))

	raws := []types.RawReport{types.NewRawReport(1, 0, []byte("beeb"))}
	suite.chainB.SendReport(1, raws, testapp.Validators[0])

	commitment := suite.chainB.App.IBCKeeper.ChannelKeeper.GetPacketCommitment(suite.chainB.GetContext(), path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID, 1)

	oracleResponsePacket := types.NewOracleResponsePacketData(
		path.EndpointA.ClientID,
		1,
		1,
		1577923380,
		1577923355,
		types.RESOLVE_STATUS_FAILURE,
		[]byte{},
	)
	responsePacket := channeltypes.NewPacket(
		oracleResponsePacket.GetBytes(),
		1,
		path.EndpointB.ChannelConfig.PortID,
		path.EndpointB.ChannelID,
		path.EndpointA.ChannelConfig.PortID,
		path.EndpointA.ChannelID,
		clienttypes.ZeroHeight(),
		1577923955000000000,
	)
	expectCommitment := channeltypes.CommitPacket(suite.chainB.Codec, responsePacket)
	suite.Equal(expectCommitment, commitment)
}

func (suite *OracleTestSuite) TestIBCResolveRequestWasmFailure() {
	path := suite.path

	suite.chainB.App.OracleKeeper.SetRequest(suite.chainB.GetContext(), 1, types.NewRequest(
		// 6th Wasm - out-of-gas
		6, []byte("beeb"), []sdk.ValAddress{testapp.Validators[0].ValAddress, testapp.Validators[1].ValAddress}, 1,
		suite.chainB.GetContext().BlockHeight()-1, testapp.ParseTime(1577923380), path.EndpointA.ClientID, []types.RawRequest{
			types.NewRawRequest(1, 1, []byte("beeb")),
		}, &types.IBCChannel{PortId: path.EndpointB.ChannelConfig.PortID, ChannelId: path.EndpointB.ChannelID},
		testapp.TestDefaultExecuteGas,
	))

	raws := []types.RawReport{types.NewRawReport(1, 0, []byte("beeb"))}
	suite.chainB.SendReport(1, raws, testapp.Validators[0])

	commitment := suite.chainB.App.IBCKeeper.ChannelKeeper.GetPacketCommitment(suite.chainB.GetContext(), path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID, 1)

	oracleResponsePacket := types.NewOracleResponsePacketData(path.EndpointA.ClientID, 1, 1, 1577923380, 1577923355, types.RESOLVE_STATUS_FAILURE, []byte{})
	responsePacket := channeltypes.NewPacket(
		oracleResponsePacket.GetBytes(),
		1,
		path.EndpointB.ChannelConfig.PortID,
		path.EndpointB.ChannelID,
		path.EndpointA.ChannelConfig.PortID,
		path.EndpointA.ChannelID,
		clienttypes.ZeroHeight(),
		1577923955000000000,
	)
	expectCommitment := channeltypes.CommitPacket(suite.chainB.Codec, responsePacket)
	suite.Equal(expectCommitment, commitment)
}

func (suite *OracleTestSuite) TestIBCResolveRequestCallReturnDataSeveralTimes() {
	path := suite.path

	suite.chainB.App.OracleKeeper.SetRequest(suite.chainB.GetContext(), 1, types.NewRequest(
		// 9th Wasm - set return data several times
		9, []byte("beeb"), []sdk.ValAddress{testapp.Validators[0].ValAddress, testapp.Validators[1].ValAddress}, 1,
		suite.chainB.GetContext().BlockHeight()-1, testapp.ParseTime(1577923380), path.EndpointA.ClientID, []types.RawRequest{
			types.NewRawRequest(1, 1, []byte("beeb")),
		}, &types.IBCChannel{PortId: path.EndpointB.ChannelConfig.PortID, ChannelId: path.EndpointB.ChannelID},
		testapp.TestDefaultExecuteGas,
	))

	raws := []types.RawReport{types.NewRawReport(1, 0, []byte("beeb"))}
	suite.chainB.SendReport(1, raws, testapp.Validators[0])

	commitment := suite.chainB.App.IBCKeeper.ChannelKeeper.GetPacketCommitment(suite.chainB.GetContext(), path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID, 1)

	oracleResponsePacket := types.NewOracleResponsePacketData(path.EndpointA.ClientID, 1, 1, 1577923380, 1577923355, types.RESOLVE_STATUS_FAILURE, []byte{})
	responsePacket := channeltypes.NewPacket(
		oracleResponsePacket.GetBytes(),
		1,
		path.EndpointB.ChannelConfig.PortID,
		path.EndpointB.ChannelID,
		path.EndpointA.ChannelConfig.PortID,
		path.EndpointA.ChannelID,
		clienttypes.ZeroHeight(),
		1577923955000000000,
	)
	expectCommitment := channeltypes.CommitPacket(suite.chainB.Codec, responsePacket)
	suite.Equal(expectCommitment, commitment)
}

func TestOracleTestSuite(t *testing.T) {
	suite.Run(t, new(OracleTestSuite))
}
