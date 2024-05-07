package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	bandtesting "github.com/bandprotocol/chain/v2/testing"
)

type KeeperTestSuite struct {
	suite.Suite

	app *bandtesting.TestingApp
	ctx sdk.Context
}

func (s *KeeperTestSuite) SetupTest() {
	app, ctx := bandtesting.CreateTestApp(s.T(), false)
	s.app = app
	s.ctx = ctx
}

func (s *KeeperTestSuite) TestGetSetRollingSeed() {
	ctx, k := s.ctx, s.app.RollingseedKeeper
	rollingSeed := []byte("sample-rolling-seed")

	// Set RollingSeed
	k.SetRollingSeed(ctx, rollingSeed)

	// Get and check RollingSeed
	gotSeed := k.GetRollingSeed(ctx)
	s.Require().Equal(rollingSeed, gotSeed)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
