package oracle_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/exported"
	"github.com/stretchr/testify/suite"

	ibctesting "github.com/bandprotocol/chain/testing"
	"github.com/bandprotocol/chain/x/oracle/testapp"
	"github.com/bandprotocol/chain/x/oracle/types"
)

type OracleTestSuite struct {
	suite.Suite

	coordinator *ibctesting.Coordinator

	// testing chains used for convenience and readability
	chainA *ibctesting.TestChain
	chainB *ibctesting.TestChain
	chainC *ibctesting.TestChain
}

func (suite *OracleTestSuite) SetupTest() {
	suite.coordinator = ibctesting.NewCoordinator(suite.T(), 3)
	suite.chainA = suite.coordinator.GetChain(ibctesting.GetChainID(0))
	suite.chainB = suite.coordinator.GetChain(ibctesting.GetChainID(1))
	suite.chainC = suite.coordinator.GetChain(ibctesting.GetChainID(2))
}

func (suite *OracleTestSuite) TestHandleMsgRequest() {
	// setup between chainA and chainB
	clientA, clientB, connA, connB := suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, exported.Tendermint)
	suite.Require().NotNil(clientA)
	suite.Require().NotNil(clientB)

	suite.Require().NotNil(connA)
	suite.Require().NotNil(connB)

	channelA, channelB := suite.coordinator.CreateTransferChannels(suite.chainA, suite.chainB, connA, connB, channeltypes.UNORDERED)

	suite.Require().NotNil(channelA)
	suite.Require().NotNil(channelB)

	msg := types.NewMsgRequestData(
		1,
		[]byte("beeb"),
		1,
		1,
		clientA,
		sdk.NewCoins(sdk.NewCoin("uband", sdk.NewInt(10))),
		50000,
		300000,
		suite.chainA.SenderAccount.GetAddress(),
	)

	err := suite.coordinator.SendMsg(suite.chainA, suite.chainB, clientB, msg)
	suite.Require().NoError(err) // message committed

	timeoutHeight := clienttypes.NewHeight(25, 1000)

	// send from A to B
	oracleRequestPacket := types.NewOracleRequestPacketData(
		clientA,
		1,
		[]byte("beeb"),
		1,
		1,
		sdk.NewCoins(sdk.NewCoin("uband", sdk.NewInt(10))),
		"beeb-request",
		testapp.TestDefaultPrepareGas,
		testapp.TestDefaultExecuteGas,
	)
	packet := channeltypes.NewPacket(oracleRequestPacket.GetBytes(), 1, channelA.PortID, channelA.ID, channelB.PortID, channelB.ID, timeoutHeight, 0)
	err = suite.coordinator.SendPacket(suite.chainA, suite.chainB, packet, clientB)
	suite.Require().NoError(err)

	suite.chainA.PendingSendPackets = append(suite.chainA.PendingSendPackets, packet)
	err = suite.coordinator.RelayAndAckPendingPackets(suite.chainA, suite.chainB, clientA, clientB)
	suite.Require().NoError(err)

	ack, ok := suite.chainB.App.IBCKeeper.ChannelKeeper.GetPacketAcknowledgement(suite.chainB.GetContext(), channelB.PortID, channelB.ID, 1)
	suite.Require().True(ok)

	ackBytes := channeltypes.NewResultAcknowledgement(types.ModuleCdc.MustMarshalJSON(types.NewOracleRequestPacketAcknowledgement(1))).GetBytes()
	expectAck := channeltypes.CommitAcknowledgement(ackBytes)
	suite.Require().Equal(expectAck, ack)

	// TODO: check request is correct
	// req := suite.chainB.App.OracleKeeper.MustGetRequest(suite.chainB.GetContext(), 1)
	// suite.Require().Equal(nil, req)

	reports := []types.RawReport{types.NewRawReport(1, 0, []byte("data1")), types.NewRawReport(2, 0, []byte("data2")), types.NewRawReport(3, 0, []byte("data3"))}
	suite.chainB.SendMsgs(types.NewMsgReportData(1, reports, testapp.Validators[0].ValAddress, testapp.Validators[0].Address))
	suite.Require().NoError(err)

	oracleResponsePacket := types.NewOracleResponsePacketData(clientA, 1, 1, 1577923405, 1577923465, types.RESOLVE_STATUS_SUCCESS, []byte("beeb"))
	responsePacket := channeltypes.NewPacket(oracleResponsePacket.GetBytes(), 1, channelB.PortID, channelB.ID, channelA.PortID, channelA.ID, clienttypes.NewHeight(0, 0), 1577924065000000000)
	expectCommitment := channeltypes.CommitPacket(suite.chainB.Codec, responsePacket)
	commitment := suite.chainB.App.IBCKeeper.ChannelKeeper.GetPacketCommitment(suite.chainB.GetContext(), channelB.PortID, channelB.ID, 1)
	suite.Equal(expectCommitment, commitment)

}

func TestOracleTestSuite(t *testing.T) {
	suite.Run(t, new(OracleTestSuite))
}
