// 0.47 TODO: write this test file by importing testing directly from ibc
package oracle_test

import (
	"fmt"
	"strings"
	"testing"

	tmtypes "github.com/cometbft/cometbft/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	ibctesting "github.com/cosmos/ibc-go/v7/testing"
	"github.com/stretchr/testify/suite"

	"github.com/bandprotocol/chain/v2/pkg/obi"
	"github.com/bandprotocol/chain/v2/testing/bandibctesting"
	"github.com/bandprotocol/chain/v2/testing/testapp"
	"github.com/bandprotocol/chain/v2/x/oracle/types"
)

type OracleTestSuite struct {
	suite.Suite

	coordinator *ibctesting.Coordinator

	// testing chains used for convenience and readability
	chainA *bandibctesting.TestChain
	chainB *bandibctesting.TestChain

	pathAB *ibctesting.Path
}

func (suite *OracleTestSuite) SetupTest() {
	ibctesting.DefaultTestingAppInit = bandibctesting.SetupTestingApp
	suite.coordinator = bandibctesting.NewCoordinator(suite.T(), 2)
	suite.chainA = &bandibctesting.TestChain{
		TestChain: suite.coordinator.GetChain(ibctesting.GetChainID(1)),
	}
	suite.chainB = &bandibctesting.TestChain{
		TestChain: suite.coordinator.GetChain(ibctesting.GetChainID(2)),
	}
	err := suite.chainA.SetActiveValidators()
	suite.Require().NoError(err)
	err = suite.chainB.SetActiveValidators()
	suite.Require().NoError(err)
	err = suite.chainA.SendMoneyToValidators()
	suite.Require().NoError(err)
	err = suite.chainB.SendMoneyToValidators()
	suite.Require().NoError(err)
	suite.pathAB = NewOraclePath(suite.chainA, suite.chainB)
	suite.coordinator.Setup(suite.pathAB)
}

func NewOraclePath(chainA, chainB *bandibctesting.TestChain) *ibctesting.Path {
	path := ibctesting.NewPath(chainA.TestChain, chainB.TestChain)
	path.EndpointA.ChannelConfig.PortID = types.PortID
	path.EndpointB.ChannelConfig.PortID = types.PortID
	path.EndpointA.ChannelConfig.Version = types.Version
	path.EndpointB.ChannelConfig.Version = types.Version

	return path
}

type Chain int64

const (
	ChainA Chain = iota
	ChainB
)

func (suite *OracleTestSuite) GetChain(name Chain) *bandibctesting.TestChain {
	switch name {
	case ChainA:
		return suite.chainA
	case ChainB:
		return suite.chainB
	}
	return nil
}

func (suite *OracleTestSuite) sendReport(
	chain *bandibctesting.TestChain,
	rid types.RequestID,
	rawReps []types.RawReport,
	val *tmtypes.Validator,
) (*sdk.Result, error) {
	senderAccount := bandibctesting.ValSenders[val.Address.String()]

	b, err := chain.GetBandApp().BankKeeper.AllBalances(chain.GetContext(), &banktypes.QueryAllBalancesRequest{
		Address: sdk.AccAddress(val.Address).String(),
		Pagination: &query.PageRequest{
			Limit:      3,
			CountTotal: true,
		},
	})
	if err != nil {
		return nil, err
	}
	fmt.Printf("%+v", b.Balances)

	fmt.Printf("chain id %s, accNum: %d\n", chain.ChainID, senderAccount.GetAccountNumber())

	_, r, err := testapp.SignAndDeliver(
		suite.T(),
		chain.App.GetTxConfig(),
		chain.App.GetBaseApp(),
		chain.CurrentHeader,
		[]sdk.Msg{types.NewMsgReportData(rid, rawReps, sdk.ValAddress(val.Address))},
		chain.ChainID,
		[]uint64{16},
		[]uint64{senderAccount.GetSequence()},
		bandibctesting.ValSigners[val.Address.String()],
	)
	if err != nil {
		return nil, err
	}

	// SignAndDeliver calls app.Commit()
	chain.TestChain.NextBlock()

	chain.TestChain.NextBlock()
	// chain.Coordinator.CommitBlock()

	// increment sequence for successful transaction execution
	senderAccount.SetSequence(senderAccount.GetSequence() + 1)

	chain.Coordinator.IncrementTime()

	return r, nil
}

func (suite *OracleTestSuite) sendOracleRequestPacket(
	path *ibctesting.Path,
	seq uint64,
	oracleRequestPacket types.OracleRequestPacketData,
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

func (suite *OracleTestSuite) checkChainBTreasuryBalances(expect sdk.Coins) {
	treasuryBalances := suite.chainB.GetBandApp().BankKeeper.GetAllBalances(
		suite.chainB.GetContext(),
		testapp.Treasury.Address,
	)
	suite.Require().Equal(expect, treasuryBalances)
}

func (suite *OracleTestSuite) checkChainBSenderBalances(expect sdk.Coins) {
	b := suite.chainB.GetBandApp().BankKeeper.GetAllBalances(
		suite.chainB.GetContext(),
		suite.chainB.SenderAccount.GetAddress(),
	)
	suite.Require().Equal(expect, b)
}

// constructs a send from chainA to chainB on the established channel/connection
// and sends the same coin back from chainB to chainA.
func (suite *OracleTestSuite) TestHandleIBCRequestSuccess() {
	// fmt.Printf("aaa %+v\n", suite.chainB.GetBandApp().OracleKeeper.GetAllOracleScripts(suite.chainB.GetContext()))

	// send request from A to B
	timeoutHeight := clienttypes.NewHeight(1, 100)
	oracleRequestPacket := types.NewOracleRequestPacketData(
		suite.pathAB.EndpointA.ClientID,
		1,
		[]byte(""),
		2,
		2,
		sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(6000000))),
		testapp.TestDefaultPrepareGas,
		testapp.TestDefaultExecuteGas,
	)
	packet := suite.sendOracleRequestPacket(suite.pathAB, 1, oracleRequestPacket, timeoutHeight)

	err := suite.pathAB.RelayPacket(packet)
	suite.Require().NoError(err) // relay committed

	suite.checkChainBTreasuryBalances(sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(6000000))))
	suite.checkChainBSenderBalances(sdk.NewCoins(sdk.NewCoin("stake", sdk.NewIntFromUint64(9999999999993980000))))

	raws1 := []types.RawReport{
		types.NewRawReport(1, 0, []byte("data1")),
		types.NewRawReport(2, 0, []byte("data2")),
		types.NewRawReport(3, 0, []byte("data3")),
	}
	_, err = suite.sendReport(suite.chainB, 1, raws1, suite.chainB.Vals.Validators[0])
	suite.Require().NoError(err)

	raws2 := []types.RawReport{
		types.NewRawReport(1, 0, []byte("data1")),
		types.NewRawReport(2, 0, []byte("data2")),
		types.NewRawReport(3, 0, []byte("data3")),
	}
	_, err = suite.sendReport(suite.chainB, 1, raws2, suite.chainB.Vals.Validators[1])
	suite.Require().NoError(err)

	oracleResponsePacket := types.NewOracleResponsePacketData(
		suite.pathAB.EndpointA.ClientID,
		1,
		2,
		1577923380,
		1577923405,
		types.RESOLVE_STATUS_SUCCESS,
		[]byte("beeb"),
	)
	responsePacket := channeltypes.NewPacket(
		oracleResponsePacket.GetBytes(),
		1,
		suite.pathAB.EndpointB.ChannelConfig.PortID,
		suite.pathAB.EndpointB.ChannelID,
		suite.pathAB.EndpointA.ChannelConfig.PortID,
		suite.pathAB.EndpointA.ChannelID,
		clienttypes.ZeroHeight(),
		1577924005000000000,
	)
	expectCommitment := channeltypes.CommitPacket(suite.chainB.Codec, responsePacket)
	commitment := suite.chainB.GetBandApp().IBCKeeper.ChannelKeeper.GetPacketCommitment(
		suite.chainB.GetContext(),
		suite.pathAB.EndpointB.ChannelConfig.PortID,
		suite.pathAB.EndpointB.ChannelID,
		1,
	)
	suite.Equal(expectCommitment, commitment)
}

func (suite *OracleTestSuite) TestIBCPrepareValidateBasicFail() {
	path := suite.pathAB

	clientID := path.EndpointA.ClientID
	coins := sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(6000000)))

	oracleRequestPackets := []types.OracleRequestPacketData{
		types.NewOracleRequestPacketData(
			clientID,
			1,
			[]byte(strings.Repeat("beeb", 130)),
			1,
			1,
			coins,
			testapp.TestDefaultPrepareGas,
			testapp.TestDefaultExecuteGas,
		),
		types.NewOracleRequestPacketData(
			clientID,
			1,
			[]byte("beeb"),
			1,
			0,
			coins,
			testapp.TestDefaultPrepareGas,
			testapp.TestDefaultExecuteGas,
		),
		types.NewOracleRequestPacketData(
			clientID,
			1,
			[]byte("beeb"),
			1,
			2,
			coins,
			testapp.TestDefaultPrepareGas,
			testapp.TestDefaultExecuteGas,
		),
		types.NewOracleRequestPacketData(
			strings.Repeat(clientID, 9),
			1,
			[]byte("beeb"),
			1,
			1,
			coins,
			testapp.TestDefaultPrepareGas,
			testapp.TestDefaultExecuteGas,
		),
		types.NewOracleRequestPacketData(clientID, 1, []byte("beeb"), 1, 1, coins, 0, testapp.TestDefaultExecuteGas),
		types.NewOracleRequestPacketData(clientID, 1, []byte("beeb"), 1, 1, coins, testapp.TestDefaultPrepareGas, 0),
		types.NewOracleRequestPacketData(
			clientID,
			1,
			[]byte("beeb"),
			1,
			1,
			coins,
			types.MaximumOwasmGas,
			types.MaximumOwasmGas,
		),
		types.NewOracleRequestPacketData(
			clientID,
			1,
			[]byte("beeb"),
			1,
			1,
			testapp.BadCoins,
			testapp.TestDefaultPrepareGas,
			testapp.TestDefaultExecuteGas,
		),
	}

	timeoutHeight := clienttypes.NewHeight(1, 110)
	for i, requestPacket := range oracleRequestPackets {
		packet := suite.sendOracleRequestPacket(path, uint64(i)+1, requestPacket, timeoutHeight)

		err := path.RelayPacket(packet)
		suite.Require().NoError(err) // relay committed
	}
}

func (suite *OracleTestSuite) TestIBCPrepareRequestNotEnoughFund() {
	path := suite.pathAB

	// send request from A to B
	timeoutHeight := clienttypes.NewHeight(1, 110)
	oracleRequestPacket := types.NewOracleRequestPacketData(
		path.EndpointA.ClientID,
		1,
		[]byte("beeb"),
		1,
		1,
		sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(3000000))),
		testapp.TestDefaultPrepareGas,
		testapp.TestDefaultExecuteGas,
	)

	// Use Carol as a relayer
	carol := testapp.Carol
	carolExpectedBalance := sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(2500000)))
	suite.chainB.SendMsgs(banktypes.NewMsgSend(
		suite.chainB.SenderAccount.GetAddress(),
		carol.Address,
		sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(2500000))),
	))
	suite.chainB.SenderPrivKey = carol.PrivKey
	suite.chainB.SenderAccount = suite.chainB.GetBandApp().AccountKeeper.GetAccount(
		suite.chainB.GetContext(),
		carol.Address,
	)

	packet := suite.sendOracleRequestPacket(path, 1, oracleRequestPacket, timeoutHeight)

	err := path.RelayPacket(packet)
	suite.Require().NoError(err) // relay committed

	carolBalance := suite.chainB.GetBandApp().BankKeeper.GetAllBalances(suite.chainB.GetContext(), carol.Address)
	suite.Require().Equal(carolExpectedBalance, carolBalance)
}

func (suite *OracleTestSuite) TestIBCPrepareRequestNotEnoughFeeLimit() {
	path := suite.pathAB
	expectedBalance := suite.chainB.GetBandApp().BankKeeper.GetAllBalances(
		suite.chainB.GetContext(),
		suite.chainB.SenderAccount.GetAddress(),
	).Sub(sdk.NewCoin("stake", sdk.NewInt(7500)))

	// send request from A to B
	timeoutHeight := clienttypes.NewHeight(1, 110)
	oracleRequestPacket := types.NewOracleRequestPacketData(
		path.EndpointA.ClientID,
		1,
		[]byte("beeb"),
		1,
		1,
		sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(2000000))),
		testapp.TestDefaultPrepareGas,
		testapp.TestDefaultExecuteGas,
	)
	packet := suite.sendOracleRequestPacket(path, 1, oracleRequestPacket, timeoutHeight)

	err := path.RelayPacket(packet)
	suite.Require().NoError(err) // relay committed

	suite.checkChainBSenderBalances(expectedBalance)
}

func (suite *OracleTestSuite) TestIBCPrepareRequestInvalidCalldataSize() {
	path := suite.pathAB

	// send request from A to B
	timeoutHeight := clienttypes.NewHeight(1, 110)
	oracleRequestPacket := types.NewOracleRequestPacketData(
		path.EndpointA.ClientID,
		1,
		[]byte(strings.Repeat("beeb", 2000)),
		1,
		1,
		sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(3000000))),
		testapp.TestDefaultPrepareGas,
		testapp.TestDefaultExecuteGas,
	)
	packet := suite.sendOracleRequestPacket(path, 1, oracleRequestPacket, timeoutHeight)

	err := path.RelayPacket(packet)
	suite.Require().NoError(err) // relay committed
}

func (suite *OracleTestSuite) TestIBCPrepareRequestNotEnoughPrepareGas() {
	path := suite.pathAB
	// send request from A to B
	timeoutHeight := clienttypes.NewHeight(1, 110)
	oracleRequestPacket := types.NewOracleRequestPacketData(
		path.EndpointA.ClientID,
		1,
		[]byte("beeb"),
		1,
		1,
		sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(3000000))),
		1,
		testapp.TestDefaultExecuteGas,
	)
	packet := suite.sendOracleRequestPacket(path, 1, oracleRequestPacket, timeoutHeight)

	err := path.RelayPacket(packet)
	suite.Require().NoError(err) // relay committed
}

func (suite *OracleTestSuite) TestIBCPrepareRequestInvalidAskCountFail() {
	path := suite.pathAB

	// send request from A to B
	timeoutHeight := clienttypes.NewHeight(1, 110)
	oracleRequestPacket := types.NewOracleRequestPacketData(
		path.EndpointA.ClientID,
		1,
		[]byte("beeb"),
		17,
		1,
		sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(3000000))),
		testapp.TestDefaultPrepareGas,
		testapp.TestDefaultExecuteGas,
	)
	packet := suite.sendOracleRequestPacket(path, 1, oracleRequestPacket, timeoutHeight)

	err := path.RelayPacket(packet)
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

	err = path.RelayPacket(packet)
	suite.Require().NoError(err) // relay committed
}

func (suite *OracleTestSuite) TestIBCPrepareRequestBaseOwasmFeePanic() {
	path := suite.pathAB

	params := suite.chainB.GetBandApp().OracleKeeper.GetParams(suite.chainB.GetContext())
	params.BaseOwasmGas = 100000000
	params.PerValidatorRequestGas = 0
	suite.chainB.GetBandApp().OracleKeeper.SetParams(suite.chainB.GetContext(), params)

	// send request from A to B
	timeoutHeight := clienttypes.NewHeight(1, 110)
	oracleRequestPacket := types.NewOracleRequestPacketData(
		path.EndpointA.ClientID,
		1,
		[]byte("beeb"),
		1,
		1,
		sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(3000000))),
		testapp.TestDefaultPrepareGas,
		testapp.TestDefaultExecuteGas,
	)
	packet := suite.sendOracleRequestPacket(path, 1, oracleRequestPacket, timeoutHeight)

	// ConsumeGas panics due to insufficient gas, so ErrAcknowledgement is not created.
	err := path.RelayPacket(packet)
	suite.Require().Contains(err.Error(), "BASE_OWASM_FEE; gasWanted: 10000000")
}

func (suite *OracleTestSuite) TestIBCPrepareRequestPerValidatorRequestFeePanic() {
	path := suite.pathAB

	params := suite.chainB.GetBandApp().OracleKeeper.GetParams(suite.chainB.GetContext())
	params.PerValidatorRequestGas = 100000000
	suite.chainB.GetBandApp().OracleKeeper.SetParams(suite.chainB.GetContext(), params)

	// send request from A to B
	timeoutHeight := clienttypes.NewHeight(1, 110)
	oracleRequestPacket := types.NewOracleRequestPacketData(
		path.EndpointA.ClientID,
		1,
		[]byte("beeb"),
		1,
		1,
		sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(3000000))),
		testapp.TestDefaultPrepareGas,
		testapp.TestDefaultExecuteGas,
	)
	packet := suite.sendOracleRequestPacket(path, 1, oracleRequestPacket, timeoutHeight)

	// ConsumeGas panics due to insufficient gas, so ErrAcknowledgement is not created.
	err := path.RelayPacket(packet)
	suite.Require().Contains(err.Error(), "PER_VALIDATOR_REQUEST_FEE; gasWanted: 1000000")
}

func (suite *OracleTestSuite) TestIBCPrepareRequestOracleScriptNotFound() {
	path := suite.pathAB

	// send request from A to B
	timeoutHeight := clienttypes.NewHeight(1, 110)
	oracleRequestPacket := types.NewOracleRequestPacketData(
		path.EndpointA.ClientID,
		100,
		[]byte("beeb"),
		1,
		1,
		sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(3000000))),
		testapp.TestDefaultPrepareGas,
		testapp.TestDefaultExecuteGas,
	)
	packet := suite.sendOracleRequestPacket(path, 1, oracleRequestPacket, timeoutHeight)

	err := path.RelayPacket(packet)
	suite.Require().NoError(err) // relay committed
}

func (suite *OracleTestSuite) TestIBCPrepareRequestBadWasmExecutionFail() {
	path := suite.pathAB

	// send request from A to B
	timeoutHeight := clienttypes.NewHeight(1, 110)
	oracleRequestPacket := types.NewOracleRequestPacketData(
		path.EndpointA.ClientID,
		2,
		[]byte("beeb"),
		1,
		1,
		sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(3000000))),
		testapp.TestDefaultPrepareGas,
		testapp.TestDefaultExecuteGas,
	)
	packet := suite.sendOracleRequestPacket(path, 1, oracleRequestPacket, timeoutHeight)

	err := path.RelayPacket(packet)
	suite.Require().NoError(err) // relay committed
}

func (suite *OracleTestSuite) TestIBCPrepareRequestWithEmptyRawRequest() {
	path := suite.pathAB

	// send request from A to B
	timeoutHeight := clienttypes.NewHeight(1, 110)
	oracleRequestPacket := types.NewOracleRequestPacketData(
		path.EndpointA.ClientID,
		3,
		[]byte("beeb"),
		1,
		1,
		sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(3000000))),
		testapp.TestDefaultPrepareGas,
		testapp.TestDefaultExecuteGas,
	)
	packet := suite.sendOracleRequestPacket(path, 1, oracleRequestPacket, timeoutHeight)

	err := path.RelayPacket(packet)
	suite.Require().NoError(err) // relay committed
}

func (suite *OracleTestSuite) TestIBCPrepareRequestUnknownDataSource() {
	path := suite.pathAB

	// send request from A to B
	timeoutHeight := clienttypes.NewHeight(1, 110)
	oracleRequestPacket := types.NewOracleRequestPacketData(
		path.EndpointA.ClientID,
		4,
		[]byte("beeb"),
		1,
		1,
		sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(3000000))),
		testapp.TestDefaultPrepareGas,
		testapp.TestDefaultExecuteGas,
	)
	packet := suite.sendOracleRequestPacket(path, 1, oracleRequestPacket, timeoutHeight)

	err := path.RelayPacket(packet)
	suite.Require().NoError(err) // relay committed
}

func (suite *OracleTestSuite) TestIBCPrepareRequestInvalidDataSourceCount() {
	path := suite.pathAB

	params := suite.chainB.GetBandApp().OracleKeeper.GetParams(suite.chainB.GetContext())
	params.MaxRawRequestCount = 3
	suite.chainB.GetBandApp().OracleKeeper.SetParams(suite.chainB.GetContext(), params)

	// send request from A to B
	timeoutHeight := clienttypes.NewHeight(1, 110)
	oracleRequestPacket := types.NewOracleRequestPacketData(
		path.EndpointA.ClientID,
		4,
		obi.MustEncode(testapp.Wasm4Input{
			IDs:      []int64{1, 2, 3, 4},
			Calldata: "beeb",
		}),
		1,
		1,
		sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(4000000))),
		testapp.TestDefaultPrepareGas,
		testapp.TestDefaultExecuteGas,
	)
	packet := suite.sendOracleRequestPacket(path, 1, oracleRequestPacket, timeoutHeight)

	err := path.RelayPacket(packet)
	suite.Require().NoError(err) // relay committed
}

func (suite *OracleTestSuite) TestIBCPrepareRequestTooMuchWasmGas() {
	path := suite.pathAB

	// send request from A to B
	timeoutHeight := clienttypes.NewHeight(1, 110)
	oracleRequestPacket := types.NewOracleRequestPacketData(
		path.EndpointA.ClientID,
		6,
		[]byte("beeb"),
		1,
		1,
		sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(3000000))),
		testapp.TestDefaultPrepareGas,
		testapp.TestDefaultExecuteGas,
	)
	packet := suite.sendOracleRequestPacket(path, 1, oracleRequestPacket, timeoutHeight)

	err := path.RelayPacket(packet)
	suite.Require().NoError(err) // relay committed
}

func (suite *OracleTestSuite) TestIBCPrepareRequestTooLargeCalldata() {
	path := suite.pathAB
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

	err := path.RelayPacket(packet)
	suite.Require().NoError(err) // relay committed
}

func (suite *OracleTestSuite) TestIBCResolveRequestOutOfGas() {
	path := suite.pathAB

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

	err := path.RelayPacket(packet)
	suite.Require().NoError(err) // relay committed

	suite.checkChainBTreasuryBalances(sdk.NewCoins(sdk.NewCoin("uband", sdk.NewInt(6000000))))
	suite.checkChainBSenderBalances(sdk.NewCoins(sdk.NewCoin("uband", sdk.NewInt(3970000))))

	raws := []types.RawReport{
		types.NewRawReport(1, 0, []byte("data1")),
		types.NewRawReport(2, 0, []byte("data2")),
		types.NewRawReport(3, 0, []byte("data3")),
	}
	suite.sendReport(suite.chainB, 1, raws, suite.chainB.Vals.Validators[0])

	commitment := suite.chainB.GetBandApp().IBCKeeper.ChannelKeeper.GetPacketCommitment(
		suite.chainB.GetContext(),
		path.EndpointB.ChannelConfig.PortID,
		path.EndpointB.ChannelID,
		1,
	)

	oracleResponsePacket := types.NewOracleResponsePacketData(
		path.EndpointA.ClientID,
		1,
		1,
		1577923380,
		1577923400,
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
		1577924000000000000,
	)
	expectCommitment := channeltypes.CommitPacket(suite.chainB.Codec, responsePacket)
	suite.Equal(expectCommitment, commitment)
}

func (suite *OracleTestSuite) TestIBCResolveReadNilExternalData() {
	path := suite.pathAB

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

	err := path.RelayPacket(packet)
	suite.Require().NoError(err) // relay committed

	suite.checkChainBTreasuryBalances(sdk.NewCoins(sdk.NewCoin("uband", sdk.NewInt(4000000))))
	suite.checkChainBSenderBalances(sdk.NewCoins(sdk.NewCoin("uband", sdk.NewInt(5970000))))

	raws1 := []types.RawReport{types.NewRawReport(0, 0, nil), types.NewRawReport(1, 0, []byte("beebd2v1"))}
	suite.sendReport(suite.chainB, 1, raws1, suite.chainB.Vals.Validators[0])

	raws2 := []types.RawReport{types.NewRawReport(0, 0, []byte("beebd1v2")), types.NewRawReport(1, 0, nil)}
	suite.sendReport(suite.chainB, 1, raws2, suite.chainB.Vals.Validators[1])

	commitment := suite.chainB.GetBandApp().IBCKeeper.ChannelKeeper.GetPacketCommitment(
		suite.chainB.GetContext(),
		path.EndpointB.ChannelConfig.PortID,
		path.EndpointB.ChannelID,
		1,
	)

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
	path := suite.pathAB

	suite.chainB.GetBandApp().OracleKeeper.SetRequest(suite.chainB.GetContext(), 1, types.NewRequest(
		// 3rd Wasm - do nothing
		3,
		[]byte("beeb"),
		[]sdk.ValAddress{testapp.Validators[0].ValAddress, testapp.Validators[1].ValAddress},
		1,
		suite.chainB.GetContext().
			BlockHeight()-
			1,
		testapp.ParseTime(1577923380),
		path.EndpointA.ClientID,
		[]types.RawRequest{
			types.NewRawRequest(1, 1, []byte("beeb")),
		},
		&types.IBCChannel{PortId: path.EndpointB.ChannelConfig.PortID, ChannelId: path.EndpointB.ChannelID},
		0,
	))

	raws := []types.RawReport{types.NewRawReport(1, 0, []byte("beeb"))}
	suite.sendReport(suite.chainB, 1, raws, suite.chainB.Vals.Validators[0])

	commitment := suite.chainB.GetBandApp().IBCKeeper.ChannelKeeper.GetPacketCommitment(
		suite.chainB.GetContext(),
		path.EndpointB.ChannelConfig.PortID,
		path.EndpointB.ChannelID,
		1,
	)

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
	path := suite.pathAB

	suite.chainB.GetBandApp().OracleKeeper.SetRequest(suite.chainB.GetContext(), 1, types.NewRequest(
		// 6th Wasm - out-of-gas
		6,
		[]byte("beeb"),
		[]sdk.ValAddress{testapp.Validators[0].ValAddress, testapp.Validators[1].ValAddress},
		1,
		suite.chainB.GetContext().
			BlockHeight()-
			1,
		testapp.ParseTime(1577923380),
		path.EndpointA.ClientID,
		[]types.RawRequest{
			types.NewRawRequest(1, 1, []byte("beeb")),
		},
		&types.IBCChannel{PortId: path.EndpointB.ChannelConfig.PortID, ChannelId: path.EndpointB.ChannelID},
		testapp.TestDefaultExecuteGas,
	))

	raws := []types.RawReport{types.NewRawReport(1, 0, []byte("beeb"))}
	suite.sendReport(suite.chainB, 1, raws, suite.chainB.Vals.Validators[0])

	commitment := suite.chainB.GetBandApp().IBCKeeper.ChannelKeeper.GetPacketCommitment(
		suite.chainB.GetContext(),
		path.EndpointB.ChannelConfig.PortID,
		path.EndpointB.ChannelID,
		1,
	)

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

func (suite *OracleTestSuite) TestIBCResolveRequestCallReturnDataSeveralTimes() {
	path := suite.pathAB

	suite.chainB.GetBandApp().OracleKeeper.SetRequest(suite.chainB.GetContext(), 1, types.NewRequest(
		// 9th Wasm - set return data several times
		9,
		[]byte("beeb"),
		[]sdk.ValAddress{testapp.Validators[0].ValAddress, testapp.Validators[1].ValAddress},
		1,
		suite.chainB.GetContext().
			BlockHeight()-
			1,
		testapp.ParseTime(1577923380),
		path.EndpointA.ClientID,
		[]types.RawRequest{
			types.NewRawRequest(1, 1, []byte("beeb")),
		},
		&types.IBCChannel{PortId: path.EndpointB.ChannelConfig.PortID, ChannelId: path.EndpointB.ChannelID},
		testapp.TestDefaultExecuteGas,
	))

	raws := []types.RawReport{types.NewRawReport(1, 0, []byte("beeb"))}
	suite.sendReport(suite.chainB, 1, raws, suite.chainB.Vals.Validators[0])

	commitment := suite.chainB.GetBandApp().IBCKeeper.ChannelKeeper.GetPacketCommitment(
		suite.chainB.GetContext(),
		path.EndpointB.ChannelConfig.PortID,
		path.EndpointB.ChannelID,
		1,
	)

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

func TestOracleTestSuite(t *testing.T) {
	suite.Run(t, new(OracleTestSuite))
}
