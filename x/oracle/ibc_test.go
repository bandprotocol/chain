package oracle_test

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	ibctesting "github.com/cosmos/ibc-go/v8/testing"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	band "github.com/bandprotocol/chain/v3/app"
	"github.com/bandprotocol/chain/v3/pkg/obi"
	bandtesting "github.com/bandprotocol/chain/v3/testing"
	"github.com/bandprotocol/chain/v3/testing/testdata"
	oracletypes "github.com/bandprotocol/chain/v3/x/oracle/types"
)

func init() {
	band.SetBech32AddressPrefixesAndBip44CoinTypeAndSeal(sdk.GetConfig())
	sdk.DefaultBondDenom = "uband"
}

type IBCTestSuite struct {
	suite.Suite

	coordinator *ibctesting.Coordinator

	// testing chains used for convenience and readability
	chainA *ibctesting.TestChain
	chainB *ibctesting.TestChain

	path *ibctesting.Path

	// shortcut to chainB (bandchain)
	bandApp *band.BandApp
}

func (suite *IBCTestSuite) SetupTest() {
	ibctesting.DefaultTestingAppInit = bandtesting.CreateTestingAppFn(suite.T())

	suite.coordinator = ibctesting.NewCoordinator(suite.T(), 2)
	suite.chainA = suite.coordinator.GetChain(ibctesting.GetChainID(1))
	suite.chainB = suite.coordinator.GetChain(ibctesting.GetChainID(2))

	suite.path = ibctesting.NewPath(suite.chainA, suite.chainB)
	suite.path.EndpointA.ChannelConfig.PortID = oracletypes.ModuleName
	suite.path.EndpointA.ChannelConfig.Version = oracletypes.Version
	suite.path.EndpointB.ChannelConfig.PortID = oracletypes.ModuleName
	suite.path.EndpointB.ChannelConfig.Version = oracletypes.Version

	suite.bandApp = suite.chainB.App.(*band.BandApp)

	suite.coordinator.Setup(suite.path)

	// Activate oracle validator on chain B (bandchain)
	for _, v := range suite.chainB.Vals.Validators {
		err := suite.bandApp.OracleKeeper.Activate(
			suite.chainB.GetContext(),
			sdk.ValAddress(v.Address),
		)
		suite.Require().NoError(err)
	}

	suite.coordinator.CommitBlock(suite.chainB)
}

func (suite *IBCTestSuite) sendReport(requestID oracletypes.RequestID, report oracletypes.Report, needToResolve bool) {
	suite.bandApp.OracleKeeper.SetReport(suite.chainB.GetContext(), requestID, report)
	if needToResolve {
		suite.bandApp.OracleKeeper.AddPendingRequest(suite.chainB.GetContext(), requestID)
	}

	suite.coordinator.CommitBlock(suite.chainB)
}

func (suite *IBCTestSuite) sendOracleRequestPacket(
	path *ibctesting.Path,
	seq uint64,
	oracleRequestPacket oracletypes.OracleRequestPacketData,
	timeoutHeight clienttypes.Height,
) channeltypes.Packet {
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
	_, err := path.EndpointA.SendPacket(timeoutHeight, 0, oracleRequestPacket.GetBytes())
	suite.Require().NoError(err)
	return packet
}

func (suite *IBCTestSuite) checkChainBTreasuryBalances(expect sdk.Coins) {
	treasuryBalances := suite.bandApp.BankKeeper.GetAllBalances(
		suite.chainB.GetContext(),
		bandtesting.Treasury.Address,
	)
	suite.Require().Equal(expect, treasuryBalances)
}

func (suite *IBCTestSuite) checkChainBSenderBalances(expect sdk.Coins) {
	b := suite.bandApp.BankKeeper.GetAllBalances(
		suite.chainB.GetContext(),
		suite.chainB.SenderAccount.GetAddress(),
	)
	suite.Require().Equal(expect, b)
}

// constructs a send from chainA to chainB on the established channel/connection
// and sends the same coin back from chainB to chainA.
func (suite *IBCTestSuite) TestHandleIBCRequestSuccess() {
	path := suite.path
	// send request from A to B
	timeoutHeight := clienttypes.NewHeight(10, 110)
	oracleRequestPacket := oracletypes.NewOracleRequestPacketData(
		path.EndpointA.ClientID,
		1,
		[]byte("test"),
		4,
		2,
		0,
		sdk.NewCoins(sdk.NewCoin("uband", math.NewInt(12000000))),
		bandtesting.TestDefaultPrepareGas,
		bandtesting.TestDefaultExecuteGas,
	)
	packet := suite.sendOracleRequestPacket(path, 1, oracleRequestPacket, timeoutHeight)

	err := path.RelayPacket(packet)
	suite.Require().NoError(err) // relay committed

	// Treasury get fees from relayer
	suite.checkChainBTreasuryBalances(sdk.NewCoins(sdk.NewCoin("uband", math.NewInt(12000000))))

	raws1 := []oracletypes.RawReport{
		oracletypes.NewRawReport(1, 0, []byte("data1")),
		oracletypes.NewRawReport(2, 0, []byte("data2")),
		oracletypes.NewRawReport(3, 0, []byte("data3")),
	}
	suite.sendReport(
		oracletypes.RequestID(1),
		oracletypes.NewReport(sdk.ValAddress(suite.chainB.Vals.Validators[0].Address), true, raws1),
		false,
	)

	raws2 := []oracletypes.RawReport{
		oracletypes.NewRawReport(1, 0, []byte("data1")),
		oracletypes.NewRawReport(2, 0, []byte("data2")),
		oracletypes.NewRawReport(3, 0, []byte("data3")),
	}
	suite.sendReport(
		oracletypes.RequestID(1),
		oracletypes.NewReport(sdk.ValAddress(suite.chainB.Vals.Validators[2].Address), true, raws2),
		true,
	)

	oracleResponsePacket := oracletypes.NewOracleResponsePacketData(
		path.EndpointA.ClientID,
		1,
		2,
		1577923360,
		1577923385,
		oracletypes.RESOLVE_STATUS_SUCCESS,
		[]byte("test"),
	)
	responsePacket := channeltypes.NewPacket(
		oracleResponsePacket.GetBytes(),
		1,
		path.EndpointB.ChannelConfig.PortID,
		path.EndpointB.ChannelID,
		path.EndpointA.ChannelConfig.PortID,
		path.EndpointA.ChannelID,
		clienttypes.ZeroHeight(),
		uint64(time.Unix(1577923385, 0).Add(10*time.Minute).UnixNano()),
	)
	expectCommitment := channeltypes.CommitPacket(suite.chainB.Codec, responsePacket)
	commitment := suite.bandApp.GetIBCKeeper().ChannelKeeper.GetPacketCommitment(
		suite.chainB.GetContext(),
		path.EndpointB.ChannelConfig.PortID,
		path.EndpointB.ChannelID,
		1,
	)
	suite.Equal(expectCommitment, commitment)
}

func (suite *IBCTestSuite) TestIBCPrepareValidateBasicFail() {
	path := suite.path

	clientID := path.EndpointA.ClientID
	coins := sdk.NewCoins(sdk.NewCoin("uband", math.NewInt(6000000)))

	oracleRequestPackets := []oracletypes.OracleRequestPacketData{
		oracletypes.NewOracleRequestPacketData(
			clientID,
			1,
			[]byte(strings.Repeat("test", 130)),
			1,
			1,
			0,
			coins,
			bandtesting.TestDefaultPrepareGas,
			bandtesting.TestDefaultExecuteGas,
		),
		oracletypes.NewOracleRequestPacketData(
			clientID,
			1,
			[]byte("test"),
			1,
			0,
			0,
			coins,
			bandtesting.TestDefaultPrepareGas,
			bandtesting.TestDefaultExecuteGas,
		),
		oracletypes.NewOracleRequestPacketData(
			clientID,
			1,
			[]byte("test"),
			1,
			2,
			0,
			coins,
			bandtesting.TestDefaultPrepareGas,
			bandtesting.TestDefaultExecuteGas,
		),
		oracletypes.NewOracleRequestPacketData(
			strings.Repeat(clientID, 9),
			1,
			[]byte("test"),
			1,
			1,
			0,
			coins,
			bandtesting.TestDefaultPrepareGas,
			bandtesting.TestDefaultExecuteGas,
		),
		oracletypes.NewOracleRequestPacketData(
			clientID,
			1,
			[]byte("test"),
			1,
			1,
			0,
			coins,
			0,
			bandtesting.TestDefaultExecuteGas,
		),
		oracletypes.NewOracleRequestPacketData(
			clientID,
			1,
			[]byte("test"),
			1,
			1,
			0,
			coins,
			bandtesting.TestDefaultPrepareGas,
			0,
		),
		oracletypes.NewOracleRequestPacketData(
			clientID,
			1,
			[]byte("test"),
			1,
			1,
			0,
			coins,
			0,
			bandtesting.TestDefaultExecuteGas,
		),
		oracletypes.NewOracleRequestPacketData(
			clientID,
			1,
			[]byte("test"),
			1,
			1,
			0,
			bandtesting.BadCoins,
			bandtesting.TestDefaultPrepareGas,
			bandtesting.TestDefaultExecuteGas,
		),
	}

	timeoutHeight := clienttypes.NewHeight(1, 110)
	for i, requestPacket := range oracleRequestPackets {
		packet := suite.sendOracleRequestPacket(path, uint64(i)+1, requestPacket, timeoutHeight)

		err := path.RelayPacket(packet)
		suite.Require().NoError(err) // relay committed
	}
}

func (suite *IBCTestSuite) TestIBCPrepareRequestNotEnoughFund() {
	path := suite.path

	// send request from A to B
	timeoutHeight := clienttypes.NewHeight(1, 110)
	oracleRequestPacket := oracletypes.NewOracleRequestPacketData(
		path.EndpointA.ClientID,
		1,
		[]byte("test"),
		1,
		1,
		0,
		sdk.NewCoins(sdk.NewCoin("uband", math.NewInt(3000000))),
		bandtesting.TestDefaultPrepareGas,
		bandtesting.TestDefaultExecuteGas,
	)

	// Use Carol as a relayer
	carol := bandtesting.Carol
	carolExpectedBalance := sdk.NewCoins(sdk.NewCoin("uband", math.NewInt(2500000)))
	_, err := suite.chainB.SendMsgs(banktypes.NewMsgSend(
		suite.chainB.SenderAccount.GetAddress(),
		carol.Address,
		sdk.NewCoins(sdk.NewCoin("uband", math.NewInt(2500000))),
	))
	suite.Require().NoError(err)

	suite.chainB.SenderPrivKey = carol.PrivKey
	suite.chainB.SenderAccount = suite.bandApp.AccountKeeper.GetAccount(
		suite.chainB.GetContext(),
		carol.Address,
	)

	packet := suite.sendOracleRequestPacket(path, 1, oracleRequestPacket, timeoutHeight)

	err = path.RelayPacket(packet)
	suite.Require().NoError(err) // relay committed

	carolBalance := suite.bandApp.BankKeeper.GetAllBalances(
		suite.chainB.GetContext(),
		carol.Address,
	)
	suite.Require().Equal(carolExpectedBalance, carolBalance)
}

func (suite *IBCTestSuite) TestIBCPrepareRequestNotEnoughFeeLimit() {
	path := suite.path
	expectedBalance := suite.bandApp.BankKeeper.GetAllBalances(
		suite.chainB.GetContext(),
		suite.chainB.SenderAccount.GetAddress(),
	)

	// send request from A to B
	timeoutHeight := clienttypes.NewHeight(1, 110)
	oracleRequestPacket := oracletypes.NewOracleRequestPacketData(
		path.EndpointA.ClientID,
		1,
		[]byte("test"),
		1,
		1,
		0,
		sdk.NewCoins(sdk.NewCoin("uband", math.NewInt(2000000))),
		bandtesting.TestDefaultPrepareGas,
		bandtesting.TestDefaultExecuteGas,
	)
	packet := suite.sendOracleRequestPacket(path, 1, oracleRequestPacket, timeoutHeight)

	err := path.RelayPacket(packet)
	suite.Require().NoError(err) // relay committed

	suite.checkChainBSenderBalances(expectedBalance)
}

func (suite *IBCTestSuite) TestIBCPrepareRequestInvalidCalldataSize() {
	path := suite.path

	// send request from A to B
	timeoutHeight := clienttypes.NewHeight(1, 110)
	oracleRequestPacket := oracletypes.NewOracleRequestPacketData(
		path.EndpointA.ClientID,
		1,
		[]byte(strings.Repeat("test", 2000)),
		1,
		1,
		0,
		sdk.NewCoins(sdk.NewCoin("uband", math.NewInt(3000000))),
		bandtesting.TestDefaultPrepareGas,
		bandtesting.TestDefaultExecuteGas,
	)
	packet := suite.sendOracleRequestPacket(path, 1, oracleRequestPacket, timeoutHeight)

	err := path.RelayPacket(packet)
	suite.Require().NoError(err) // relay committed
}

func (suite *IBCTestSuite) TestIBCPrepareRequestNotEnoughPrepareGas() {
	path := suite.path
	// send request from A to B
	timeoutHeight := clienttypes.NewHeight(1, 110)
	oracleRequestPacket := oracletypes.NewOracleRequestPacketData(
		path.EndpointA.ClientID,
		1,
		[]byte("test"),
		1,
		1,
		0,
		sdk.NewCoins(sdk.NewCoin("uband", math.NewInt(3000000))),
		1,
		bandtesting.TestDefaultExecuteGas,
	)
	packet := suite.sendOracleRequestPacket(path, 1, oracleRequestPacket, timeoutHeight)

	err := path.RelayPacket(packet)
	suite.Require().NoError(err) // relay committed
}

func (suite *IBCTestSuite) TestIBCPrepareRequestInvalidAskCountFail() {
	path := suite.path

	// send request from A to B
	timeoutHeight := clienttypes.NewHeight(1, 110)
	oracleRequestPacket := oracletypes.NewOracleRequestPacketData(
		path.EndpointA.ClientID,
		1,
		[]byte("test"),
		17,
		1,
		0,
		sdk.NewCoins(sdk.NewCoin("uband", math.NewInt(3000000))),
		bandtesting.TestDefaultPrepareGas,
		bandtesting.TestDefaultExecuteGas,
	)
	packet := suite.sendOracleRequestPacket(path, 1, oracleRequestPacket, timeoutHeight)

	err := path.RelayPacket(packet)
	suite.Require().NoError(err) // relay committed

	oracleRequestPacket = oracletypes.NewOracleRequestPacketData(
		path.EndpointA.ClientID,
		1,
		[]byte("test"),
		3,
		1,
		0,
		sdk.NewCoins(sdk.NewCoin("uband", math.NewInt(3000000))),
		bandtesting.TestDefaultPrepareGas,
		bandtesting.TestDefaultExecuteGas,
	)
	packet = suite.sendOracleRequestPacket(path, 2, oracleRequestPacket, timeoutHeight)

	err = path.RelayPacket(packet)
	suite.Require().NoError(err) // relay committed
}

func (suite *IBCTestSuite) TestIBCPrepareRequestBaseOwasmFeePanic() {
	path := suite.path

	params := suite.bandApp.OracleKeeper.GetParams(suite.chainB.GetContext())
	params.BaseOwasmGas = 100000000
	params.PerValidatorRequestGas = 0
	err := suite.bandApp.OracleKeeper.SetParams(suite.chainB.GetContext(), params)
	suite.Require().NoError(err)

	// send request from A to B
	timeoutHeight := clienttypes.NewHeight(1, 110)
	oracleRequestPacket := oracletypes.NewOracleRequestPacketData(
		path.EndpointA.ClientID,
		1,
		[]byte("test"),
		1,
		1,
		0,
		sdk.NewCoins(sdk.NewCoin("uband", math.NewInt(3000000))),
		bandtesting.TestDefaultPrepareGas,
		bandtesting.TestDefaultExecuteGas,
	)
	packet := suite.sendOracleRequestPacket(path, 1, oracleRequestPacket, timeoutHeight)

	// ConsumeGas panics due to insufficient gas, so ErrAcknowledgement is not created.
	err = path.RelayPacket(packet)
	suite.Require().Contains(err.Error(), "BASE_OWASM_FEE; gasWanted: 1000000")
}

func (suite *IBCTestSuite) TestIBCPrepareRequestPerValidatorRequestFeePanic() {
	path := suite.path

	params := suite.bandApp.OracleKeeper.GetParams(suite.chainB.GetContext())
	params.PerValidatorRequestGas = 100000000
	err := suite.bandApp.OracleKeeper.SetParams(suite.chainB.GetContext(), params)
	suite.Require().NoError(err)

	// send request from A to B
	timeoutHeight := clienttypes.NewHeight(1, 110)
	oracleRequestPacket := oracletypes.NewOracleRequestPacketData(
		path.EndpointA.ClientID,
		1,
		[]byte("test"),
		1,
		1,
		0,
		sdk.NewCoins(sdk.NewCoin("uband", math.NewInt(3000000))),
		bandtesting.TestDefaultPrepareGas,
		bandtesting.TestDefaultExecuteGas,
	)
	packet := suite.sendOracleRequestPacket(path, 1, oracleRequestPacket, timeoutHeight)

	// ConsumeGas panics due to insufficient gas, so ErrAcknowledgement is not created.
	err = path.RelayPacket(packet)
	suite.Require().Contains(err.Error(), "PER_VALIDATOR_REQUEST_FEE; gasWanted: 1000000")
}

func (suite *IBCTestSuite) TestIBCPrepareRequestOracleScriptNotFound() {
	path := suite.path

	// send request from A to B
	timeoutHeight := clienttypes.NewHeight(1, 110)
	oracleRequestPacket := oracletypes.NewOracleRequestPacketData(
		path.EndpointA.ClientID,
		100,
		[]byte("test"),
		1,
		1,
		0,
		sdk.NewCoins(sdk.NewCoin("uband", math.NewInt(3000000))),
		bandtesting.TestDefaultPrepareGas,
		bandtesting.TestDefaultExecuteGas,
	)
	packet := suite.sendOracleRequestPacket(path, 1, oracleRequestPacket, timeoutHeight)

	err := path.RelayPacket(packet)
	suite.Require().NoError(err) // relay committed
}

func (suite *IBCTestSuite) TestIBCPrepareRequestBadWasmExecutionFail() {
	path := suite.path

	// send request from A to B
	timeoutHeight := clienttypes.NewHeight(1, 110)
	oracleRequestPacket := oracletypes.NewOracleRequestPacketData(
		path.EndpointA.ClientID,
		2,
		[]byte("test"),
		1,
		1,
		0,
		sdk.NewCoins(sdk.NewCoin("uband", math.NewInt(3000000))),
		bandtesting.TestDefaultPrepareGas,
		bandtesting.TestDefaultExecuteGas,
	)
	packet := suite.sendOracleRequestPacket(path, 1, oracleRequestPacket, timeoutHeight)

	err := path.RelayPacket(packet)
	suite.Require().NoError(err) // relay committed
}

func (suite *IBCTestSuite) TestIBCPrepareRequestWithEmptyRawRequest() {
	path := suite.path

	// send request from A to B
	timeoutHeight := clienttypes.NewHeight(1, 110)
	oracleRequestPacket := oracletypes.NewOracleRequestPacketData(
		path.EndpointA.ClientID,
		3,
		[]byte("test"),
		1,
		1,
		0,
		sdk.NewCoins(sdk.NewCoin("uband", math.NewInt(3000000))),
		bandtesting.TestDefaultPrepareGas,
		bandtesting.TestDefaultExecuteGas,
	)
	packet := suite.sendOracleRequestPacket(path, 1, oracleRequestPacket, timeoutHeight)

	err := path.RelayPacket(packet)
	suite.Require().NoError(err) // relay committed
}

func (suite *IBCTestSuite) TestIBCPrepareRequestUnknownDataSource() {
	path := suite.path

	// send request from A to B
	timeoutHeight := clienttypes.NewHeight(1, 110)
	oracleRequestPacket := oracletypes.NewOracleRequestPacketData(
		path.EndpointA.ClientID,
		4,
		[]byte("test"),
		1,
		1,
		0,
		sdk.NewCoins(sdk.NewCoin("uband", math.NewInt(3000000))),
		bandtesting.TestDefaultPrepareGas,
		bandtesting.TestDefaultExecuteGas,
	)
	packet := suite.sendOracleRequestPacket(path, 1, oracleRequestPacket, timeoutHeight)

	err := path.RelayPacket(packet)
	suite.Require().NoError(err) // relay committed
}

func (suite *IBCTestSuite) TestIBCPrepareRequestInvalidDataSourceCount() {
	path := suite.path

	params := suite.bandApp.OracleKeeper.GetParams(suite.chainB.GetContext())
	params.MaxRawRequestCount = 3
	err := suite.bandApp.OracleKeeper.SetParams(suite.chainB.GetContext(), params)
	suite.Require().NoError(err)

	// send request from A to B
	timeoutHeight := clienttypes.NewHeight(1, 110)
	oracleRequestPacket := oracletypes.NewOracleRequestPacketData(
		path.EndpointA.ClientID,
		4,
		obi.MustEncode(testdata.Wasm4Input{
			IDs:      []int64{1, 2, 3, 4},
			Calldata: "test",
		}),
		1,
		1,
		0,
		sdk.NewCoins(sdk.NewCoin("uband", math.NewInt(4000000))),
		bandtesting.TestDefaultPrepareGas,
		bandtesting.TestDefaultExecuteGas,
	)
	packet := suite.sendOracleRequestPacket(path, 1, oracleRequestPacket, timeoutHeight)

	err = path.RelayPacket(packet)
	suite.Require().NoError(err) // relay committed
}

func (suite *IBCTestSuite) TestIBCPrepareRequestTooMuchWasmGas() {
	path := suite.path

	// send request from A to B
	timeoutHeight := clienttypes.NewHeight(1, 110)
	oracleRequestPacket := oracletypes.NewOracleRequestPacketData(
		path.EndpointA.ClientID,
		6,
		[]byte("test"),
		1,
		1,
		0,
		sdk.NewCoins(sdk.NewCoin("uband", math.NewInt(3000000))),
		bandtesting.TestDefaultPrepareGas,
		bandtesting.TestDefaultExecuteGas,
	)
	packet := suite.sendOracleRequestPacket(path, 1, oracleRequestPacket, timeoutHeight)

	err := path.RelayPacket(packet)
	suite.Require().NoError(err) // relay committed
}

func (suite *IBCTestSuite) TestIBCPrepareRequestTooLargeCalldata() {
	path := suite.path
	// send request from A to B
	timeoutHeight := clienttypes.NewHeight(1, 110)
	oracleRequestPacket := oracletypes.NewOracleRequestPacketData(
		path.EndpointA.ClientID,
		8,
		[]byte("test"),
		1,
		1,
		0,
		sdk.NewCoins(sdk.NewCoin("uband", math.NewInt(3000000))),
		bandtesting.TestDefaultPrepareGas,
		bandtesting.TestDefaultExecuteGas,
	)
	packet := suite.sendOracleRequestPacket(path, 1, oracleRequestPacket, timeoutHeight)

	err := path.RelayPacket(packet)
	suite.Require().NoError(err) // relay committed
}

func (suite *IBCTestSuite) TestIBCResolveRequestOutOfGas() {
	path := suite.path

	// send request from A to B
	timeoutHeight := clienttypes.NewHeight(1, 110)
	oracleRequestPacket := oracletypes.NewOracleRequestPacketData(
		path.EndpointA.ClientID,
		1,
		[]byte("test"),
		2,
		1,
		0,
		sdk.NewCoins(sdk.NewCoin("uband", math.NewInt(6000000))),
		bandtesting.TestDefaultPrepareGas,
		1,
	)
	expectedSenderBalance := suite.bandApp.BankKeeper.GetAllBalances(
		suite.chainB.GetContext(),
		suite.chainB.SenderAccount.GetAddress(),
	).Sub(sdk.NewCoin("uband", math.NewInt(6000000)))
	packet := suite.sendOracleRequestPacket(path, 1, oracleRequestPacket, timeoutHeight)

	err := path.RelayPacket(packet)
	suite.Require().NoError(err) // relay committed

	suite.checkChainBTreasuryBalances(sdk.NewCoins(sdk.NewCoin("uband", math.NewInt(6000000))))
	suite.checkChainBSenderBalances(expectedSenderBalance)

	raws := []oracletypes.RawReport{
		oracletypes.NewRawReport(1, 0, []byte("data1")),
		oracletypes.NewRawReport(2, 0, []byte("data2")),
		oracletypes.NewRawReport(3, 0, []byte("data3")),
	}
	suite.sendReport(
		oracletypes.RequestID(1),
		oracletypes.NewReport(sdk.ValAddress(suite.chainB.Vals.Validators[0].Address), true, raws),
		true,
	)

	commitment := suite.bandApp.IBCKeeper.ChannelKeeper.GetPacketCommitment(
		suite.chainB.GetContext(),
		path.EndpointB.ChannelConfig.PortID,
		path.EndpointB.ChannelID,
		1,
	)

	oracleResponsePacket := oracletypes.NewOracleResponsePacketData(
		path.EndpointA.ClientID,
		1,
		1,
		1577923360,
		1577923380,
		oracletypes.RESOLVE_STATUS_FAILURE,
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
		uint64(time.Unix(1577923380, 0).Add(10*time.Minute).UnixNano()),
	)
	expectCommitment := channeltypes.CommitPacket(suite.chainB.Codec, responsePacket)
	suite.Equal(expectCommitment, commitment)
}

func (suite *IBCTestSuite) TestIBCResolveReadNilExternalData() {
	path := suite.path

	// send request from A to B
	timeoutHeight := clienttypes.NewHeight(10, 110)
	oracleRequestPacket := oracletypes.NewOracleRequestPacketData(
		path.EndpointA.ClientID,
		4,
		obi.MustEncode(testdata.Wasm4Input{IDs: []int64{1, 2}, Calldata: string("test")}),
		2,
		2,
		0,
		sdk.NewCoins(sdk.NewCoin("uband", math.NewInt(4000000))),
		bandtesting.TestDefaultPrepareGas,
		bandtesting.TestDefaultExecuteGas,
	)
	expectedSenderBalance := suite.bandApp.BankKeeper.GetAllBalances(
		suite.chainB.GetContext(),
		suite.chainB.SenderAccount.GetAddress(),
	).Sub(sdk.NewCoin("uband", math.NewInt(4000000)))
	packet := suite.sendOracleRequestPacket(path, 1, oracleRequestPacket, timeoutHeight)

	err := path.RelayPacket(packet)
	suite.Require().NoError(err) // relay committed

	suite.checkChainBTreasuryBalances(sdk.NewCoins(sdk.NewCoin("uband", math.NewInt(4000000))))
	suite.checkChainBSenderBalances(expectedSenderBalance)

	raws1 := []oracletypes.RawReport{
		oracletypes.NewRawReport(0, 0, nil),
		oracletypes.NewRawReport(1, 0, []byte("testd2v1")),
	}
	suite.sendReport(
		oracletypes.RequestID(1),
		oracletypes.NewReport(sdk.ValAddress(suite.chainB.Vals.Validators[0].Address), true, raws1),
		false,
	)

	raws2 := []oracletypes.RawReport{
		oracletypes.NewRawReport(0, 0, []byte("testd1v2")),
		oracletypes.NewRawReport(1, 0, nil),
	}
	suite.sendReport(
		oracletypes.RequestID(1),
		oracletypes.NewReport(sdk.ValAddress(suite.chainB.Vals.Validators[2].Address), true, raws2),
		true,
	)

	commitment := suite.bandApp.GetIBCKeeper().ChannelKeeper.GetPacketCommitment(
		suite.chainB.GetContext(),
		path.EndpointB.ChannelConfig.PortID,
		path.EndpointB.ChannelID,
		1,
	)

	oracleResponsePacket := oracletypes.NewOracleResponsePacketData(
		path.EndpointA.ClientID,
		1,
		2,
		1577923360,
		1577923385,
		oracletypes.RESOLVE_STATUS_SUCCESS,
		obi.MustEncode(testdata.Wasm4Output{Ret: "testd1v2testd2v1"}),
	)
	responsePacket := channeltypes.NewPacket(
		oracleResponsePacket.GetBytes(),
		1,
		path.EndpointB.ChannelConfig.PortID,
		path.EndpointB.ChannelID,
		path.EndpointA.ChannelConfig.PortID,
		path.EndpointA.ChannelID,
		clienttypes.ZeroHeight(),
		uint64(time.Unix(1577923385, 0).Add(10*time.Minute).UnixNano()),
	)
	expectCommitment := channeltypes.CommitPacket(suite.bandApp.AppCodec(), responsePacket)
	suite.Equal(expectCommitment, commitment)
}

func (suite *IBCTestSuite) TestIBCResolveRequestNoReturnData() {
	path := suite.path

	suite.bandApp.OracleKeeper.SetRequest(suite.chainB.GetContext(), 1, oracletypes.NewRequest(
		// 3rd Wasm - do nothing
		3,
		[]byte("test"),
		[]sdk.ValAddress{
			sdk.ValAddress(suite.chainB.Vals.Validators[0].Address),
			sdk.ValAddress(suite.chainB.Vals.Validators[1].Address),
		},
		1,
		suite.chainB.GetContext().BlockHeight()-1,
		bandtesting.ParseTime(1577923380),
		path.EndpointA.ClientID,
		[]oracletypes.RawRequest{
			oracletypes.NewRawRequest(1, 1, []byte("test")),
		},
		&oracletypes.IBCChannel{PortId: path.EndpointB.ChannelConfig.PortID, ChannelId: path.EndpointB.ChannelID},
		0,
		0,
		bandtesting.FeePayer.Address.String(),
		bandtesting.Coins100000000uband,
	))

	raws := []oracletypes.RawReport{oracletypes.NewRawReport(1, 0, []byte("test"))}
	suite.sendReport(
		oracletypes.RequestID(1),
		oracletypes.NewReport(sdk.ValAddress(suite.chainB.Vals.Validators[0].Address), true, raws),
		true,
	)

	commitment := suite.bandApp.IBCKeeper.ChannelKeeper.GetPacketCommitment(
		suite.chainB.GetContext(),
		path.EndpointB.ChannelConfig.PortID,
		path.EndpointB.ChannelID,
		1,
	)

	oracleResponsePacket := oracletypes.NewOracleResponsePacketData(
		path.EndpointA.ClientID,
		1,
		1,
		1577923380,
		1577923335,
		oracletypes.RESOLVE_STATUS_FAILURE,
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
		uint64(time.Unix(1577923335, 0).Add(10*time.Minute).UnixNano()),
	)
	expectCommitment := channeltypes.CommitPacket(suite.chainB.Codec, responsePacket)
	suite.Equal(expectCommitment, commitment)
}

func (suite *IBCTestSuite) TestIBCResolveRequestWasmFailure() {
	path := suite.path

	suite.bandApp.OracleKeeper.SetRequest(suite.chainB.GetContext(), 1, oracletypes.NewRequest(
		// 6th Wasm - out-of-gas
		6,
		[]byte("test"),
		[]sdk.ValAddress{
			sdk.ValAddress(suite.chainB.Vals.Validators[0].Address),
			sdk.ValAddress(suite.chainB.Vals.Validators[1].Address),
		},
		1,
		suite.chainB.GetContext().BlockHeight()-1,
		bandtesting.ParseTime(1577923380),
		path.EndpointA.ClientID,
		[]oracletypes.RawRequest{
			oracletypes.NewRawRequest(1, 1, []byte("test")),
		},
		&oracletypes.IBCChannel{PortId: path.EndpointB.ChannelConfig.PortID, ChannelId: path.EndpointB.ChannelID},
		bandtesting.TestDefaultExecuteGas,
		0,
		bandtesting.FeePayer.Address.String(),
		bandtesting.Coins100000000uband,
	))

	raws := []oracletypes.RawReport{oracletypes.NewRawReport(1, 0, []byte("test"))}
	suite.sendReport(
		oracletypes.RequestID(1),
		oracletypes.NewReport(sdk.ValAddress(suite.chainB.Vals.Validators[0].Address), true, raws),
		true,
	)

	commitment := suite.bandApp.IBCKeeper.ChannelKeeper.GetPacketCommitment(
		suite.chainB.GetContext(),
		path.EndpointB.ChannelConfig.PortID,
		path.EndpointB.ChannelID,
		1,
	)

	oracleResponsePacket := oracletypes.NewOracleResponsePacketData(
		path.EndpointA.ClientID,
		1,
		1,
		1577923380,
		1577923335,
		oracletypes.RESOLVE_STATUS_FAILURE,
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
		uint64(time.Unix(1577923335, 0).Add(10*time.Minute).UnixNano()),
	)
	expectCommitment := channeltypes.CommitPacket(suite.chainB.Codec, responsePacket)
	suite.Equal(expectCommitment, commitment)
}

func (suite *IBCTestSuite) TestIBCResolveRequestCallReturnDataSeveralTimes() {
	path := suite.path

	suite.bandApp.OracleKeeper.SetRequest(suite.chainB.GetContext(), 1, oracletypes.NewRequest(
		// 9th Wasm - set return data several times
		9,
		[]byte("test"),
		[]sdk.ValAddress{
			sdk.ValAddress(suite.chainB.Vals.Validators[0].Address),
			sdk.ValAddress(suite.chainB.Vals.Validators[1].Address),
		},
		1,
		suite.chainB.GetContext().BlockHeight()-1,
		bandtesting.ParseTime(1577923360),
		path.EndpointA.ClientID,
		[]oracletypes.RawRequest{
			oracletypes.NewRawRequest(1, 1, []byte("test")),
		},
		&oracletypes.IBCChannel{PortId: path.EndpointB.ChannelConfig.PortID, ChannelId: path.EndpointB.ChannelID},
		bandtesting.TestDefaultExecuteGas,
		0,
		bandtesting.FeePayer.Address.String(),
		bandtesting.Coins100000000uband,
	))

	raws := []oracletypes.RawReport{oracletypes.NewRawReport(1, 0, []byte("test"))}
	suite.sendReport(
		oracletypes.RequestID(1),
		oracletypes.NewReport(sdk.ValAddress(suite.chainB.Vals.Validators[0].Address), true, raws),
		true,
	)

	commitment := suite.bandApp.IBCKeeper.ChannelKeeper.GetPacketCommitment(
		suite.chainB.GetContext(),
		path.EndpointB.ChannelConfig.PortID,
		path.EndpointB.ChannelID,
		1,
	)

	oracleResponsePacket := oracletypes.NewOracleResponsePacketData(
		path.EndpointA.ClientID,
		1,
		1,
		1577923360,
		1577923335,
		oracletypes.RESOLVE_STATUS_FAILURE,
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
		uint64(time.Unix(1577923335, 0).Add(10*time.Minute).UnixNano()),
	)
	expectCommitment := channeltypes.CommitPacket(suite.chainB.Codec, responsePacket)
	suite.Equal(expectCommitment, commitment)
}

func TestIBCTestSuite(t *testing.T) {
	suite.Run(t, new(IBCTestSuite))
}
