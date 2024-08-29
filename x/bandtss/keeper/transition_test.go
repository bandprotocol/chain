package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/bandtss/types"
	tsstypes "github.com/bandprotocol/chain/v2/x/tss/types"
)

func TestHandleGroupTransition(t *testing.T) {
	acc1 := sdk.MustAccAddressFromBech32("band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs")
	acc2 := sdk.MustAccAddressFromBech32("band1p40yh3zkmhcv0ecqp3mcazy83sa57rgjp07dun")
	type expectOut struct {
		status         types.TransitionStatus
		currentGroupID tss.GroupID
	}

	testCases := []struct {
		name        string
		preProcess  func(s *KeeperTestSuite)
		expectOut   expectOut
		postProcess func(s *KeeperTestSuite)
	}{
		{
			name: "no transition setup and no current group",
			expectOut: expectOut{
				status: types.TRANSITION_STATUS_UNSPECIFIED,
			},
		},
		{
			name: "transition with status ForcedWaitingReplace but no current group",
			preProcess: func(s *KeeperTestSuite) {
				s.Keeper.SetGroupTransition(s.Ctx, types.GroupTransition{
					SigningID:       tss.SigningID(1),
					CurrentGroupID:  tss.GroupID(0),
					IncomingGroupID: tss.GroupID(1),
					Status:          types.TRANSITION_STATUS_WAITING_EXECUTION,
					ExecTime:        s.Ctx.BlockTime(),
				})
			},
			expectOut: expectOut{
				status:         types.TRANSITION_STATUS_UNSPECIFIED,
				currentGroupID: tss.GroupID(1),
			},
		},
		{
			name: "transition with status ForcedWaitingReplace; has a current group",
			preProcess: func(s *KeeperTestSuite) {
				s.Keeper.SetGroupTransition(s.Ctx, types.GroupTransition{
					SigningID:       tss.SigningID(0),
					CurrentGroupID:  tss.GroupID(1),
					IncomingGroupID: tss.GroupID(2),
					Status:          types.TRANSITION_STATUS_WAITING_EXECUTION,
					ExecTime:        s.Ctx.BlockTime().Add(-10 * time.Minute),
				})
				s.Keeper.SetCurrentGroupID(s.Ctx, tss.GroupID(1))
				err := s.Keeper.AddMember(s.Ctx, acc1, tss.GroupID(1))
				require.NoError(s.T(), err)
				err = s.Keeper.AddMember(s.Ctx, acc2, tss.GroupID(1))
				require.NoError(s.T(), err)

				s.MockTSSKeeper.EXPECT().MustGetMembers(gomock.Any(), tss.GroupID(1)).Return([]tsstypes.Member{
					{ID: 1, Address: acc1.String(), PubKey: []byte("test-pubkey-1")},
					{ID: 2, Address: acc2.String(), PubKey: []byte("test-pubkey-1")},
				})
			},
			expectOut: expectOut{
				status:         types.TRANSITION_STATUS_UNSPECIFIED,
				currentGroupID: tss.GroupID(2),
			},
			postProcess: func(s *KeeperTestSuite) {
				members := s.Keeper.GetMembers(s.Ctx)
				require.Len(s.T(), members, 0)
				require.False(s.T(), s.Keeper.HasMember(s.Ctx, acc1, tss.GroupID(1)))
				require.False(s.T(), s.Keeper.HasMember(s.Ctx, acc2, tss.GroupID(1)))
			},
		},
		{
			name: "transition with status ApprovedWaitingReplace; has a current group",
			preProcess: func(s *KeeperTestSuite) {
				s.Keeper.SetGroupTransition(s.Ctx, types.GroupTransition{
					SigningID:       tss.SigningID(1),
					CurrentGroupID:  tss.GroupID(1),
					IncomingGroupID: tss.GroupID(2),
					Status:          types.TRANSITION_STATUS_WAITING_EXECUTION,
					ExecTime:        s.Ctx.BlockTime().Add(-10 * time.Minute),
				})
				s.Keeper.SetCurrentGroupID(s.Ctx, tss.GroupID(1))
				err := s.Keeper.AddMember(s.Ctx, acc1, tss.GroupID(1))
				require.NoError(s.T(), err)
				err = s.Keeper.AddMember(s.Ctx, acc2, tss.GroupID(1))
				require.NoError(s.T(), err)

				s.MockTSSKeeper.EXPECT().MustGetMembers(gomock.Any(), tss.GroupID(1)).Return([]tsstypes.Member{
					{ID: 1, Address: acc1.String(), PubKey: []byte("test-pubkey-1")},
					{ID: 2, Address: acc2.String(), PubKey: []byte("test-pubkey-1")},
				})
			},
			expectOut: expectOut{
				status:         types.TRANSITION_STATUS_UNSPECIFIED,
				currentGroupID: tss.GroupID(2),
			},
			postProcess: func(s *KeeperTestSuite) {
				members := s.Keeper.GetMembers(s.Ctx)
				require.Len(s.T(), members, 0)
				require.False(s.T(), s.Keeper.HasMember(s.Ctx, acc1, tss.GroupID(1)))
				require.False(s.T(), s.Keeper.HasMember(s.Ctx, acc2, tss.GroupID(1)))
			},
		},
		{
			name: "transition with status CreatingGroup; pass ExecTime",
			preProcess: func(s *KeeperTestSuite) {
				s.Keeper.SetGroupTransition(s.Ctx, types.GroupTransition{
					SigningID:       tss.SigningID(1),
					CurrentGroupID:  tss.GroupID(1),
					IncomingGroupID: tss.GroupID(2),
					Status:          types.TRANSITION_STATUS_CREATING_GROUP,
					ExecTime:        s.Ctx.BlockTime().Add(-10 * time.Minute),
				})
				s.Keeper.SetCurrentGroupID(s.Ctx, tss.GroupID(1))
			},
			expectOut: expectOut{
				status:         types.TRANSITION_STATUS_UNSPECIFIED,
				currentGroupID: tss.GroupID(1),
			},
		},
		{
			name: "transition with status WaitingSign; pass ExecTime",
			preProcess: func(s *KeeperTestSuite) {
				s.Keeper.SetGroupTransition(s.Ctx, types.GroupTransition{
					SigningID:       tss.SigningID(1),
					CurrentGroupID:  tss.GroupID(1),
					IncomingGroupID: tss.GroupID(2),
					Status:          types.TRANSITION_STATUS_WAITING_SIGN,
					ExecTime:        s.Ctx.BlockTime().Add(-10 * time.Minute),
				})
				s.Keeper.SetCurrentGroupID(s.Ctx, tss.GroupID(1))
			},
			expectOut: expectOut{
				status:         types.TRANSITION_STATUS_UNSPECIFIED,
				currentGroupID: tss.GroupID(1),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewKeeperTestSuite(t)

			if tc.preProcess != nil {
				tc.preProcess(&s)
			}

			if transition, ok := s.Keeper.ShouldExecuteGroupTransition(s.Ctx); ok {
				s.Keeper.ExecuteGroupTransition(s.Ctx, transition)
			}

			gt, found := s.Keeper.GetGroupTransition(s.Ctx)
			if tc.expectOut.status == types.TRANSITION_STATUS_UNSPECIFIED {
				require.False(t, found)
			} else {
				require.True(t, found)
				require.Equal(t, tc.expectOut.status, gt.Status)
			}
			require.Equal(t, tc.expectOut.currentGroupID, s.Keeper.GetCurrentGroupID(s.Ctx))

			if tc.postProcess != nil {
				tc.postProcess(&s)
			}
		})
	}
}
