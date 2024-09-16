package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

func TestGetSetMember(t *testing.T) {
	s := NewKeeperTestSuite(t)
	ctx, k := s.Ctx, s.Keeper
	groupID, memberID := tss.GroupID(1), tss.MemberID(1)
	member := types.Member{
		ID:          1,
		GroupID:     groupID,
		Address:     "band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs",
		PubKey:      nil,
		IsMalicious: false,
	}
	k.SetMember(ctx, member)

	got, err := k.GetMember(ctx, groupID, memberID)
	require.NoError(t, err)
	require.Equal(t, member, got)
}

func TestGetGroupMembers(t *testing.T) {
	s := NewKeeperTestSuite(t)
	ctx, k := s.Ctx, s.Keeper
	groupID := tss.GroupID(1)
	members := []types.Member{
		{
			ID:          1,
			GroupID:     groupID,
			Address:     "band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs",
			PubKey:      nil,
			IsMalicious: false,
		},
		{
			ID:          2,
			GroupID:     groupID,
			Address:     "band1p40yh3zkmhcv0ecqp3mcazy83sa57rgjp07dun",
			PubKey:      nil,
			IsMalicious: false,
		},
	}

	k.SetMembers(ctx, members)

	got, err := k.GetGroupMembers(ctx, groupID)
	require.NoError(t, err)
	require.Equal(t, members, got)
}

func TestGetSetMemberIsActive(t *testing.T) {
	s := NewKeeperTestSuite(t)
	ctx, k := s.Ctx, s.Keeper

	groupID := tss.GroupID(10)
	address := sdk.MustAccAddressFromBech32("band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs")
	k.SetMember(ctx, types.Member{
		ID:       tss.MemberID(1),
		GroupID:  groupID,
		Address:  address.String(),
		PubKey:   nil,
		IsActive: true,
	})

	// check when being set to active
	members, err := k.GetGroupMembers(ctx, groupID)
	require.NoError(t, err)
	require.Len(t, members, 1)

	for _, member := range members {
		require.True(t, member.IsActive)
	}

	err = k.SetMemberIsActive(ctx, groupID, address, false)
	require.NoError(t, err)

	members, err = k.GetGroupMembers(ctx, groupID)
	require.NoError(t, err)
	require.Len(t, members, 1)

	for _, member := range members {
		require.False(t, member.IsActive)
	}
}
