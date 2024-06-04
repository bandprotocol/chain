package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/tss/types"
)

func (s *KeeperTestSuite) TestGetSetDECount() {
	ctx, k := s.ctx, s.app.TSSKeeper
	address := "band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs"
	accAddress := sdk.MustAccAddressFromBech32(address)

	// Set DECount
	k.SetDECount(ctx, accAddress, 1)

	// Get DECount
	s.Require().Equal(uint64(1), k.GetDECount(ctx, accAddress))
}

func (s *KeeperTestSuite) TestGetSetDE() {
	ctx, k := s.ctx, s.app.TSSKeeper
	address := sdk.MustAccAddressFromBech32("band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs")
	de := types.DE{
		PubD: []byte("D"),
		PubE: []byte("E"),
	}

	// Set DE
	k.SetDE(ctx, address, de)

	s.Require().True(k.HasDE(ctx, address, de))

	got, err := k.GetFirstDE(ctx, address)
	s.Require().NoError(err)
	s.Require().Equal(de, got)
}

func (s *KeeperTestSuite) TestDeleteDE() {
	ctx, k := s.ctx, s.app.TSSKeeper
	address := sdk.MustAccAddressFromBech32("band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs")
	de := types.DE{
		PubD: []byte("D"),
		PubE: []byte("E"),
	}

	// Set DE
	k.SetDE(ctx, address, de)

	// Get DE
	k.DeleteDE(ctx, address, de)

	// Try to get the deleted DE
	s.Require().False(k.HasDE(ctx, address, de))

	got, err := k.GetFirstDE(ctx, address)
	s.Require().ErrorIs(types.ErrDENotFound, err)
	s.Require().Equal(types.DE{}, got)
}

func (s *KeeperTestSuite) TestGetDEsGenesis() {
	ctx, k := s.ctx, s.app.TSSKeeper
	address := sdk.MustAccAddressFromBech32("band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs")
	before := k.GetDEsGenesis(ctx)
	de := types.DE{
		PubD: []byte("D"),
		PubE: []byte("E"),
	}

	// Set DE
	k.SetDE(ctx, address, de)

	// Get des with address and index
	after := k.GetDEsGenesis(ctx)

	s.Require().Equal(len(before)+1, len(after))
	for _, q := range after {
		if q.Address == string(address) {
			s.Require().Equal(types.DEGenesis{
				Address: address.String(),
				DE:      de,
			}, q)
		}
	}
}

func (s *KeeperTestSuite) TestHandleSetDEs() {
	ctx, k := s.ctx, s.app.TSSKeeper
	address := sdk.MustAccAddressFromBech32("band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs")
	des := []types.DE{
		{
			PubD: []byte("D1"),
			PubE: []byte("E1"),
		},
		{
			PubD: []byte("D2"),
			PubE: []byte("E2"),
		},
	}

	// Handle setting DEs
	err := k.HandleSetDEs(ctx, address, des)
	s.Require().NoError(err)

	// Get DECount
	cnt := k.GetDECount(ctx, address)
	s.Require().Equal(uint64(len(des)), cnt)

	// Check that all DEs have been stored correctly
	for _, de := range des {
		s.Require().True(k.HasDE(ctx, address, de))
	}
}

func (s *KeeperTestSuite) TestPollDE() {
	ctx, k := s.ctx, s.app.TSSKeeper
	address := sdk.MustAccAddressFromBech32("band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs")
	des := []types.DE{
		{
			PubD: []byte("D"),
			PubE: []byte("E"),
		},
	}

	// Set DE
	err := k.HandleSetDEs(ctx, address, des)
	s.Require().NoError(err)

	// Poll DE
	polledDE, err := k.PollDE(ctx, address)
	s.Require().NoError(err)

	// Ensure polled DE is equal to original DE
	s.Require().Equal(des[0], polledDE)

	// Attempt to get deleted DE
	s.Require().False(k.HasDE(ctx, address, des[0]))
	got, err := k.GetFirstDE(ctx, address)

	// Should return error
	s.Require().ErrorIs(types.ErrDENotFound, err)
	s.Require().Equal(types.DE{}, got)
}

func (s *KeeperTestSuite) TestHandlePollDEForAssignedMembers() {
	ctx, k := s.ctx, s.app.TSSKeeper
	members := []types.Member{
		{
			ID:          1,
			Address:     "band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs",
			PubKey:      nil,
			IsMalicious: false,
		},
		{
			ID:          2,
			Address:     "band1p40yh3zkmhcv0ecqp3mcazy83sa57rgjp07dun",
			PubKey:      nil,
			IsMalicious: false,
		},
	}
	des := []types.DE{
		{
			PubD: []byte("D1"),
			PubE: []byte("E1"),
		},
		{
			PubD: []byte("D2"),
			PubE: []byte("E2"),
		},
	}

	for _, m := range members {
		accM := sdk.MustAccAddressFromBech32(m.Address)
		err := k.HandleSetDEs(ctx, accM, des)
		s.Require().NoError(err)
	}

	des, err := k.PollDEs(ctx, members)
	s.Require().NoError(err)
	s.Require().Equal([]types.DE{
		{
			PubD: des[0].PubD,
			PubE: des[0].PubE,
		},
		{
			PubD: des[0].PubD,
			PubE: des[0].PubE,
		},
	}, des)
}
