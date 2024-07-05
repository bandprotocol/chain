package keeper_test

import (
	"github.com/bandprotocol/chain/v2/x/feeds/types"
)

func (suite *KeeperTestSuite) TestGetSetDelegatorSignals() {
	ctx := suite.ctx

	// set
	expSignals := types.DelegatorSignals{
		Delegator: ValidDelegator.String(),
		Signals: []types.Signal{
			{
				ID:    "crypto_price.bandusd",
				Power: 1e9,
			},
			{
				ID:    "crypto_price.btcusd",
				Power: 1e9,
			},
		},
	}
	suite.feedsKeeper.SetDelegatorSignals(ctx, expSignals)

	// get
	signals := suite.feedsKeeper.GetDelegatorSignals(ctx, ValidDelegator)
	suite.Require().Equal(expSignals.Signals, signals)
}

func (suite *KeeperTestSuite) TestGetSetDeleteSignalTotalPower() {
	ctx := suite.ctx

	// set
	expSignalTotalPower := types.Signal{
		ID:    "crypto_price.bandusd",
		Power: 1e9,
	}
	suite.feedsKeeper.SetSignalTotalPower(ctx, expSignalTotalPower)

	// get
	signal, err := suite.feedsKeeper.GetSignalTotalPower(ctx, expSignalTotalPower.ID)
	suite.Require().NoError(err)
	suite.Require().Equal(expSignalTotalPower, signal)

	// set with power 0
	SignalTotalPowerZero := types.Signal{
		ID:    "crypto_price.bandusd",
		Power: 0,
	}
	suite.feedsKeeper.SetSignalTotalPower(ctx, SignalTotalPowerZero)

	// get
	signal, err = suite.feedsKeeper.GetSignalTotalPower(ctx, SignalTotalPowerZero.ID)
	suite.Require().Error(err)
	suite.Require().Equal(types.Signal{}, signal)
}
