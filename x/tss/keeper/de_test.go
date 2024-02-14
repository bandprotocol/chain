package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/tss/types"
)

func (s *KeeperTestSuite) TestGetSetDEQueue() {
	ctx, k := s.ctx, s.app.TSSKeeper
	address := "band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs"
	accAddress := sdk.MustAccAddressFromBech32(address)
	deQueue := types.DEQueue{
		Address: address,
		Head:    1,
		Tail:    2,
	}

	// Set de queue
	k.SetDEQueue(ctx, deQueue)

	// Get de queue
	got := k.GetDEQueue(ctx, accAddress)

	s.Require().Equal(deQueue, got)
}

func (s *KeeperTestSuite) TestGetDEQueuesGenesis() {
	ctx, k := s.ctx, s.app.TSSKeeper

	before := k.GetDEQueues(ctx)
	deQueue := types.DEQueue{
		Address: "band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs",
		Head:    1,
		Tail:    2,
	}

	// Set de queue
	k.SetDEQueue(ctx, deQueue)

	// Get de queues with address
	after := k.GetDEQueues(ctx)

	s.Require().Equal(len(before)+1, len(after))
	for _, q := range after {
		if q.Address == deQueue.Address {
			s.Require().Equal(deQueue, q)
		}
	}
}

func (s *KeeperTestSuite) TestGetSetDE() {
	ctx, k := s.ctx, s.app.TSSKeeper
	address := sdk.MustAccAddressFromBech32("band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs")
	index := uint64(1)
	de := types.DE{
		PubD: []byte("D"),
		PubE: []byte("E"),
	}

	// Set DE
	k.SetDE(ctx, address, index, de)

	// Get DE
	got, err := k.GetDE(ctx, address, index)

	s.Require().NoError(err)
	s.Require().Equal(de, got)
}

func (s *KeeperTestSuite) TestDeleteDE() {
	ctx, k := s.ctx, s.app.TSSKeeper
	address := sdk.MustAccAddressFromBech32("band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs")
	index := uint64(1)
	de := types.DE{
		PubD: []byte("D"),
		PubE: []byte("E"),
	}

	// Set DE
	k.SetDE(ctx, address, index, de)

	// Get DE
	k.DeleteDE(ctx, address, index)

	// Try to get the deleted DE
	got, err := k.GetDE(ctx, address, index)

	s.Require().ErrorIs(types.ErrDENotFound, err)
	s.Require().Equal(types.DE{}, got)
}

func (s *KeeperTestSuite) TestGetDEsGenesis() {
	ctx, k := s.ctx, s.app.TSSKeeper
	address := sdk.MustAccAddressFromBech32("band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs")
	index := uint64(1)
	before := k.GetDEsGenesis(ctx)
	de := types.DE{
		PubD: []byte("D"),
		PubE: []byte("E"),
	}

	// Set DE
	k.SetDE(ctx, address, index, de)

	// Get des with address and index
	after := k.GetDEsGenesis(ctx)

	s.Require().Equal(len(before)+1, len(after))
	for _, q := range after {
		if q.Address == string(address) {
			s.Require().Equal(types.DEGenesis{
				Address: address.String(),
				Index:   index,
				DE:      de,
			}, q)
		}
	}
}

func (s *KeeperTestSuite) TestNextQueueValue() {
	ctx, k := s.ctx, s.app.TSSKeeper

	testCases := []struct {
		name     string
		value    uint64
		expValue uint64
	}{
		{
			"first value",
			0,
			1,
		},
		{
			"second value",
			1,
			2,
		},
		{
			"last value",
			99,
			0,
		},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.name), func() {
			nextVal := k.NextQueueValue(ctx, tc.value)
			s.Require().Equal(tc.expValue, nextVal)
		})
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

	// Get DEQueue
	deQueue := k.GetDEQueue(ctx, address)

	// Check that all DEs have been stored correctly
	s.Require().Equal(uint64(len(des)), deQueue.Tail)
	for i := uint64(0); i < deQueue.Tail; i++ {
		gotDE, err := k.GetDE(ctx, address, i)
		s.Require().NoError(err)
		s.Require().Equal(des[i], gotDE)
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
	index := uint64(1)

	// Set DE and DEQueue
	err := k.HandleSetDEs(ctx, address, des)
	s.Require().NoError(err)

	// Poll DE
	polledDE, err := k.PollDE(ctx, address)
	s.Require().NoError(err)

	// Ensure polled DE is equal to original DE
	s.Require().Equal(des[0], polledDE)

	// Attempt to get deleted DE
	deletedDE, err := k.GetDE(ctx, address, index)

	// Should return error
	s.Require().ErrorIs(types.ErrDENotFound, err)
	s.Require().Equal(types.DE{}, deletedDE)
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

	assignedMembers, err := k.HandleAssignedMembersPollDE(ctx, members)
	s.Require().NoError(err)
	s.Require().Equal(types.AssignedMembers{
		{
			MemberID:      1,
			Address:       members[0].Address,
			PubD:          des[0].PubD,
			PubE:          des[0].PubE,
			BindingFactor: nil,
			PubNonce:      nil,
		},
		{
			MemberID:      2,
			Address:       members[1].Address,
			PubD:          des[0].PubD,
			PubE:          des[0].PubE,
			BindingFactor: nil,
			PubNonce:      nil,
		},
	}, assignedMembers)
}
