package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/bandprotocol/chain/v2/x/tss/types"
)

func TestGetSetDEQueue(t *testing.T) {
	s := NewKeeperTestSuite(t)
	ctx, k := s.Ctx, s.Keeper

	address := "band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs"
	accAddress := sdk.MustAccAddressFromBech32(address)

	k.SetDEQueue(ctx, accAddress, types.DEQueue{
		Head: 1, Tail: 10,
	})

	got := k.GetDEQueue(ctx, accAddress)
	require.Equal(t, types.DEQueue{Head: 1, Tail: 10}, got)
}

func TestGetSetDE(t *testing.T) {
	s := NewKeeperTestSuite(t)
	ctx, k := s.Ctx, s.Keeper

	address := sdk.MustAccAddressFromBech32("band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs")
	de := types.DE{
		PubD: []byte("D"),
		PubE: []byte("E"),
	}
	k.SetDE(ctx, address, 1, de)

	got, err := k.GetDE(ctx, address, 1)
	require.NoError(t, err)
	require.Equal(t, de, got)
}

func TestGetDEsGenesis(t *testing.T) {
	s := NewKeeperTestSuite(t)
	ctx, k := s.Ctx, s.Keeper

	address := sdk.MustAccAddressFromBech32("band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs")
	before := k.GetDEsGenesis(ctx)
	de := types.DE{
		PubD: []byte("D"),
		PubE: []byte("E"),
	}

	// Set DE
	err := k.HandleSetDEs(ctx, address, []types.DE{de})
	require.NoError(t, err)

	// Get des with address and index
	after := k.GetDEsGenesis(ctx)
	require.Equal(t, len(before)+1, len(after))
	for _, q := range after {
		if q.Address == string(address) {
			expected := types.DEGenesis{Address: address.String(), DE: de}
			require.Equal(t, expected, q)
		}
	}
}

func TestHandleSetDEs(t *testing.T) {
	s := NewKeeperTestSuite(t)
	ctx, k := s.Ctx, s.Keeper

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
	require.NoError(t, err)

	// Get DEQueue
	cnt := k.GetDEQueue(ctx, address)
	require.Equal(t, types.DEQueue{Head: 0, Tail: 2}, cnt)

	// Check that all DEs have been stored correctly
	existingDEs := []types.DE{}
	for i := 0; i < len(des); i++ {
		de, err := k.GetDE(ctx, address, uint64(i))
		require.NoError(t, err)
		existingDEs = append(existingDEs, de)
	}

	require.Equal(t, des, existingDEs)
}

func TestPollDE(t *testing.T) {
	s := NewKeeperTestSuite(t)
	ctx, k := s.Ctx, s.Keeper

	address := sdk.MustAccAddressFromBech32("band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs")
	des := []types.DE{
		{
			PubD: []byte("D"),
			PubE: []byte("E"),
		},
	}

	// Set DE
	err := k.HandleSetDEs(ctx, address, des)
	require.NoError(t, err)

	// Poll DE
	polledDE, err := k.PollDE(ctx, address)
	require.NoError(t, err)

	// Ensure polled DE is equal to original DE
	require.Equal(t, des[0], polledDE)

	// Attempt to get deleted DE
	got, err := k.PollDE(ctx, address)

	// Should return error
	require.ErrorIs(t, types.ErrDENotFound, err)
	require.Equal(t, types.DE{}, got)
}

func TestHandlePollDEForAssignedMembers(t *testing.T) {
	s := NewKeeperTestSuite(t)
	ctx, k := s.Ctx, s.Keeper

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
		require.NoError(t, err)
	}

	des, err := k.PollDEs(ctx, members)
	require.NoError(t, err)
	require.Equal(t, []types.DE{
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
