package keeper_test

import (
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

func (s *KeeperTestSuite) TestGetSetParams() {
	ctx, k := s.ctx, s.app.TSSKeeper
	params := types.DefaultParams()

	err := k.SetParams(ctx, params)
	s.Require().NoError(err)

	s.Require().Equal(params, k.GetParams(ctx))
}
