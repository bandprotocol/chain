package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/x/tss/types"
)

func (s *KeeperTestSuite) TestGetSetDEQueue() {
	ctx, k := s.ctx, s.keeper

	address := "band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs"
	accAddress := sdk.MustAccAddressFromBech32(address)

	k.SetDEQueue(ctx, accAddress, types.DEQueue{Head: 1, Tail: 10})

	got := k.GetDEQueue(ctx, accAddress)
	s.Require().Equal(types.DEQueue{Head: 1, Tail: 10}, got)
}

func (s *KeeperTestSuite) TestGetSetDE() {
	ctx, k := s.ctx, s.keeper

	address := sdk.MustAccAddressFromBech32("band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs")
	de := types.DE{
		PubD: []byte("D"),
		PubE: []byte("E"),
	}
	k.SetDE(ctx, address, 1, de)

	got, err := k.GetDE(ctx, address, 1)
	s.Require().NoError(err)
	s.Require().Equal(de, got)
}

func (s *KeeperTestSuite) TestHandleSetDEs() {
	ctx, k := s.ctx, s.keeper

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
	err := k.EnqueueDEs(ctx, address, des)
	s.Require().NoError(err)

	// Get DEQueue
	cnt := k.GetDEQueue(ctx, address)
	s.Require().Equal(types.DEQueue{Head: 0, Tail: 2}, cnt)

	// Check that all DEs have been stored correctly
	existingDEs := []types.DE{}
	for i := 0; i < len(des); i++ {
		de, err := k.GetDE(ctx, address, uint64(i))
		s.Require().NoError(err)
		existingDEs = append(existingDEs, de)
	}

	s.Require().Equal(des, existingDEs)
}

func (s *KeeperTestSuite) TestPollDE() {
	ctx, k := s.ctx, s.keeper

	address := sdk.MustAccAddressFromBech32("band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs")
	des := []types.DE{
		{
			PubD: []byte("D"),
			PubE: []byte("E"),
		},
	}

	// Set DE
	err := k.EnqueueDEs(ctx, address, des)
	s.Require().NoError(err)

	// Poll DE
	polledDE, err := k.DequeueDE(ctx, address)
	s.Require().NoError(err)

	// Ensure polled DE is equal to original DE
	s.Require().Equal(des[0], polledDE)

	// Attempt to get deleted DE
	got, err := k.DequeueDE(ctx, address)

	// Should return error
	s.Require().ErrorIs(types.ErrDENotFound, err)
	s.Require().Equal(types.DE{}, got)
}

func (s *KeeperTestSuite) TestHandlePollDEForAssignedMembers() {
	ctx, k := s.ctx, s.keeper

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
		err := k.EnqueueDEs(ctx, accM, des)
		s.Require().NoError(err)
	}

	des, err := k.DequeueDEs(ctx, members)
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

func (s *KeeperTestSuite) TestResetDE() {
	ctx, k := s.ctx, s.keeper

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

	// Set DE
	err := k.EnqueueDEs(ctx, address, des)
	s.Require().NoError(err)
	deQueue := k.GetDEQueue(ctx, address)
	s.Require().Equal(types.DEQueue{Head: 0, Tail: 2}, deQueue)

	// Reset DE
	err = k.ResetDE(ctx, address)
	s.Require().NoError(err)

	// Ensure DEQueue is reset
	deQueue = k.GetDEQueue(ctx, address)
	s.Require().Equal(types.DEQueue{Head: 2, Tail: 2}, deQueue)

	// Attempt to get deleted DE; should return error
	_, err = k.DequeueDE(ctx, address)
	s.Require().ErrorIs(types.ErrDENotFound, err)
}
