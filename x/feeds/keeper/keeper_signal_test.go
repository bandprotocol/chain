package keeper_test

import (
	"github.com/bandprotocol/chain/v3/x/feeds/types"
)

func (suite *KeeperTestSuite) TestGetSetVote() {
	ctx := suite.ctx

	// set
	expVote := types.Vote{
		Voter: ValidVoter.String(),
		Signals: []types.Signal{
			{
				ID:    "CS:BAND-USD",
				Power: 1e9,
			},
			{
				ID:    "CS:BTC-USD",
				Power: 1e9,
			},
		},
	}
	suite.feedsKeeper.SetVote(ctx, expVote)

	// get
	signals := suite.feedsKeeper.GetVoteSignals(ctx, ValidVoter)
	suite.Require().Equal(expVote.Signals, signals)
}

func (suite *KeeperTestSuite) TestGetSetDeleteSignalTotalPower() {
	ctx := suite.ctx

	// set
	expSignalTotalPower := types.Signal{
		ID:    "CS:BAND-USD",
		Power: 1e9,
	}
	suite.feedsKeeper.SetSignalTotalPower(ctx, expSignalTotalPower)

	// get
	signal, err := suite.feedsKeeper.GetSignalTotalPower(ctx, expSignalTotalPower.ID)
	suite.Require().NoError(err)
	suite.Require().Equal(expSignalTotalPower, signal)

	// set with power 0
	SignalTotalPowerZero := types.Signal{
		ID:    "CS:BAND-USD",
		Power: 0,
	}
	suite.feedsKeeper.SetSignalTotalPower(ctx, SignalTotalPowerZero)

	// get
	signal, err = suite.feedsKeeper.GetSignalTotalPower(ctx, SignalTotalPowerZero.ID)
	suite.Require().Error(err)
	suite.Require().Equal(types.Signal{}, signal)
}
