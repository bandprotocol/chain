package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

func (s *KeeperTestSuite) TestGetSetDEQueue() {
	ctx, k := s.ctx, s.app.TSSKeeper
	address, _ := sdk.AccAddressFromBech32("band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs")
	deQueue := types.DEQueue{
		Head: 1,
		Tail: 2,
	}

	// Set de queue
	k.SetDEQueue(ctx, address, deQueue)

	// Get de queue
	got := k.GetDEQueue(ctx, address)

	s.Require().Equal(deQueue, got)
}

func (s *KeeperTestSuite) TestGetDEQueuesGenesis() {
	ctx, k := s.ctx, s.app.TSSKeeper
	address, _ := sdk.AccAddressFromBech32("band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs")
	deQueue := types.DEQueue{
		Head: 1,
		Tail: 2,
	}

	// Set de queue
	k.SetDEQueue(ctx, address, deQueue)

	// Get de queues with address
	got := k.GetDEQueuesGenesis(ctx)

	s.Require().Equal([]types.DEQueueGenesis{
		{
			Address: address,
			DEQueue: &deQueue,
		},
	}, got)
}

func (s *KeeperTestSuite) TestGetSetDE() {
	ctx, k := s.ctx, s.app.TSSKeeper
	address, _ := sdk.AccAddressFromBech32("band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs")
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
	address, _ := sdk.AccAddressFromBech32("band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs")
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
	address, _ := sdk.AccAddressFromBech32("band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs")
	index := uint64(1)
	de := types.DE{
		PubD: []byte("D"),
		PubE: []byte("E"),
	}

	// Set DE
	k.SetDE(ctx, address, index, de)

	// Get des with address and index
	got := k.GetDEsGenesis(ctx)

	s.Require().Equal([]types.DEGenesis{
		{
			Address: address,
			Index:   index,
			DE:      &de,
		},
	}, got)
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
	address, _ := sdk.AccAddressFromBech32("band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs")
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
	k.HandleSetDEs(ctx, address, des)

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
	address, _ := sdk.AccAddressFromBech32("band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs")
	des := []types.DE{
		{
			PubD: []byte("D"),
			PubE: []byte("E"),
		},
	}
	index := uint64(1)

	// Set DE and DEQueue
	k.HandleSetDEs(ctx, address, des)

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
			MemberID:    1,
			Address:     "band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs",
			PubKey:      tss.PublicKey(nil),
			IsMalicious: false,
		},
		{
			MemberID:    2,
			Address:     "band1p40yh3zkmhcv0ecqp3mcazy83sa57rgjp07dun",
			PubKey:      tss.PublicKey(nil),
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

	var accMembers []sdk.AccAddress
	for _, m := range members {
		acc, _ := sdk.AccAddressFromBech32(m.Address)
		k.HandleSetDEs(ctx, acc, des)
		accMembers = append(accMembers, acc)
	}

	assignedMembers, pubDs, pubEs, err := k.HandleAssignedMembersPollDE(ctx, members)
	s.Require().NoError(err)
	s.Require().Equal([]types.AssignedMember{
		{
			MemberID: 1,
			Member:   members[0].Address,
			PubD:     des[0].PubD,
			PubE:     des[0].PubE,
			PubNonce: nil,
		},
		{
			MemberID: 2,
			Member:   members[1].Address,
			PubD:     des[0].PubD,
			PubE:     des[0].PubE,
			PubNonce: nil,
		},
	}, assignedMembers)
	s.Require().Equal(tss.PublicKeys{[]byte("D1"), []byte("D1")}, pubDs)
	s.Require().Equal(tss.PublicKeys{[]byte("E1"), []byte("E1")}, pubEs)
}
