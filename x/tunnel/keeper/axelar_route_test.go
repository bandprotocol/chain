package keeper_test

func (s *KeeperTestSuite) TestGetSetAxelarRouteCount() {
	ctx, k := s.ctx, s.feedsKeeper

	// Set axelar route count
	count := uint64(1)
	k.SetAxelarRouteCount(ctx, count)

	// Get axelar route count
	got := k.GetAxelarRouteCount(ctx)

	// Assert equality
	s.Require().Equal(count, got)
}
