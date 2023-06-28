package keeper_test

import (
	"testing"

	"github.com/bandprotocol/chain/v2/testing/testapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
)

type KeeperTestSuite struct {
	suite.Suite

	app *testapp.TestingApp
	ctx sdk.Context
}

func (s *KeeperTestSuite) SetupTest() {
	app, ctx, _ := testapp.CreateTestInput(false)
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
