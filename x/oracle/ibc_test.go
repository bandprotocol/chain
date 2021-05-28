package oracle_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/types"
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
}

func (suite *OracleTestSuite) SetupTest() {
	suite.coordinator = ibctesting.NewCoordinator(suite.T(), 3)
	suite.chainA = suite.coordinator.GetChain(ibctesting.GetChainID(0))
	suite.chainB = suite.coordinator.GetChain(ibctesting.GetChainID(1))
}

func NewOraclePath(chainA, chainB *ibctesting.TestChain) *ibctesting.Path {
	path := ibctesting.NewPath(chainA, chainB)
	path.EndpointA.ChannelConfig.PortID = ibctesting.OraclePort
	path.EndpointB.ChannelConfig.PortID = ibctesting.OraclePort

	return path
}

func (suite *OracleTestSuite) checkChainBTreasuryBalances(path *ibctesting.Path, expect sdk.Coins) {
	treasuryBalances := suite.chainB.App.BankKeeper.GetAllBalances(suite.chainB.GetContext(), suite.chainB.Treasury)
	suite.Require().Equal(expect, treasuryBalances)
}

func (suite *OracleTestSuite) checkChainBPoolBalances(path *ibctesting.Path, requestKey string, expect sdk.Coins) {
	poolBalances := suite.chainB.App.OracleKeeper.GetRequetPoolBalances(suite.chainB.GetContext(), requestKey, path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID)
	suite.Require().Equal(expect, poolBalances)
}

// constructs a send from chainA to chainB on the established channel/connection
// and sends the same coin back from chainB to chainA.
func (suite *OracleTestSuite) TestHandleIBCRequestSuccess() {
	// setup between chainA and chainB
	path := NewOraclePath(suite.chainA, suite.chainB)
	suite.coordinator.Setup(path)

	err := suite.chainB.App.OracleKeeper.DepositRequestPool(
		suite.chainB.GetContext(),
		"beeb-request",
		path.EndpointB.ChannelConfig.PortID,
		path.EndpointB.ChannelID,
		sdk.NewCoins(sdk.NewCoin("uband", sdk.NewInt(10000000))),
		suite.chainB.SenderAccount.GetAddress(),
	)
	suite.Require().NoError(err)
	suite.Require().True(suite.chainB.App.BankKeeper.GetAllBalances(suite.chainB.GetContext(), suite.chainB.Treasury).Empty())
	suite.checkChainBPoolBalances(path, "beeb-request", sdk.NewCoins(sdk.NewCoin("uband", sdk.NewInt(10000000))))

	// send from A to B
	timeoutHeight := clienttypes.NewHeight(0, 110)
	oracleRequestPacket := types.NewOracleRequestPacketData(
		path.EndpointA.ClientID,
		1,
		[]byte("beeb"),
		1,
		1,
		sdk.NewCoins(sdk.NewCoin("uband", sdk.NewInt(3000000))),
		"beeb-request",
		testapp.TestDefaultPrepareGas,
		testapp.TestDefaultExecuteGas,
	)
	packet := channeltypes.NewPacket(
		oracleRequestPacket.GetBytes(),
		1,
		path.EndpointA.ChannelConfig.PortID,
		path.EndpointA.ChannelID,
		path.EndpointB.ChannelConfig.PortID,
		path.EndpointB.ChannelID,
		timeoutHeight,
		0,
	)
	err = path.EndpointA.SendPacket(packet)
	suite.Require().NoError(err)

	ack := channeltypes.NewResultAcknowledgement(types.ModuleCdc.MustMarshalJSON(types.NewOracleRequestPacketAcknowledgement(1)))
	err = path.RelayPacket(packet, ack.GetBytes())
	suite.Require().NoError(err) // relay committed

	suite.checkChainBTreasuryBalances(path, sdk.NewCoins(sdk.NewCoin("uband", sdk.NewInt(3000000))))
	suite.checkChainBPoolBalances(path, "beeb-request", sdk.NewCoins(sdk.NewCoin("uband", sdk.NewInt(7000000))))

	reports := []types.RawReport{types.NewRawReport(1, 0, []byte("data1")), types.NewRawReport(2, 0, []byte("data2")), types.NewRawReport(3, 0, []byte("data3"))}
	suite.chainB.SendMsgs(types.NewMsgReportData(1, reports, testapp.Validators[0].ValAddress, testapp.Validators[0].Address))
	suite.Require().NoError(err)

	oracleResponsePacket := types.NewOracleResponsePacketData(path.EndpointA.ClientID, 1, 1, 1577923380, 1577923400, types.RESOLVE_STATUS_SUCCESS, []byte("beeb"))
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
	commitment := suite.chainB.App.IBCKeeper.ChannelKeeper.GetPacketCommitment(suite.chainB.GetContext(), path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID, 1)
	suite.Equal(expectCommitment, commitment)
}

func TestTransferTestSuite(t *testing.T) {
	suite.Run(t, new(OracleTestSuite))
}
