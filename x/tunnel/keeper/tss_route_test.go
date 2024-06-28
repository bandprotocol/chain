package keeper_test

func (s *KeeperTestSuite) TestGetSetTSSRouteCount() {
	ctx, k := s.ctx, s.feedsKeeper

	// Set tss route count
	count := uint64(1)
	k.SetTSSRouteCount(ctx, count)

	// Get tss route count
	got := k.GetTSSRouteCount(ctx)

	// Assert equality
	s.Require().Equal(count, got)
}
