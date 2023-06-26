package keeper_test

import (
	"testing"

	"github.com/bandprotocol/chain/v2/testing/testapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

type KeeperTestSuite struct {
	suite.Suite

	app *testapp.TestingApp
	ctx sdk.Context
}

func (s *KeeperTestSuite) SetupTest() {
	app := testapp.NewTestApp("BANDCHAIN", log.NewNopLogger())

	// Commit genesis for test get LastCommitHash in msg create group
	app.Commit()
	app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{
		Height:  app.LastBlockHeight() + 1,
		AppHash: []byte("app-hash sample"),
	}, Hash: []byte("app-hash sample")})

	ctx := app.NewContext(
		false,
		tmproto.Header{Height: app.LastBlockHeight(), LastCommitHash: []byte("app-hash sample")},
	)

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
