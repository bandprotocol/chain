package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	sdktestutil "github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"

	band "github.com/bandprotocol/chain/v3/app"
	bandtesting "github.com/bandprotocol/chain/v3/testing"
)

func init() {
	band.SetBech32AddressPrefixesAndBip44CoinTypeAndSeal(sdk.GetConfig())
}

type AppTestSuite struct {
	suite.Suite

	app *band.BandApp
	ctx sdk.Context
}

func (s *AppTestSuite) SetupTest() {
	dir := sdktestutil.GetTempDir(s.T())
	app := bandtesting.SetupWithCustomHome(false, dir)

	s.app = app
	s.ctx = s.app.BaseApp.NewUncachedContext(false, cmtproto.Header{ChainID: bandtesting.ChainID})
}

func (s *AppTestSuite) TestGetSetRollingSeed() {
	ctx, k := s.ctx, s.app.RollingseedKeeper
	rollingSeed := []byte("sample-rolling-seed")

	// Set RollingSeed
	k.SetRollingSeed(ctx, rollingSeed)

	// Get and check RollingSeed
	gotSeed := k.GetRollingSeed(ctx)
	s.Require().Equal(rollingSeed, gotSeed)
}

func TestAppTestSuite(t *testing.T) {
	suite.Run(t, new(AppTestSuite))
}
