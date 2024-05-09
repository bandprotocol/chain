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
