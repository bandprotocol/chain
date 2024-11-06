package keeper_test

import (
	"time"

	"go.uber.org/mock/gomock"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/pkg/tss"
	"github.com/bandprotocol/chain/v3/x/bandtss/types"
	tsstypes "github.com/bandprotocol/chain/v3/x/tss/types"
)

func (s *KeeperTestSuite) TestHandleGroupTransition() {
	acc1 := sdk.MustAccAddressFromBech32("band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs")
	acc2 := sdk.MustAccAddressFromBech32("band1p40yh3zkmhcv0ecqp3mcazy83sa57rgjp07dun")
	execTime := s.ctx.BlockTime().Add(-10 * time.Minute)

	type expectOut struct {
		status       types.TransitionStatus
		currentGroup types.CurrentGroup
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
			name: "transition with status WaitingExecution but no current group",
			preProcess: func(s *KeeperTestSuite) {
				s.keeper.SetGroupTransition(s.ctx, types.GroupTransition{
					SigningID:       tss.SigningID(1),
					CurrentGroupID:  tss.GroupID(0),
					IncomingGroupID: tss.GroupID(1),
					Status:          types.TRANSITION_STATUS_WAITING_EXECUTION,
					ExecTime:        s.ctx.BlockTime(),
				})
			},
			expectOut: expectOut{
				status: types.TRANSITION_STATUS_UNSPECIFIED,
				currentGroup: types.CurrentGroup{
					GroupID:    tss.GroupID(1),
					ActiveTime: s.ctx.BlockTime(),
				},
			},
		},
		{
			name: "force transition with status WaitingExecution; has a current group",
			preProcess: func(s *KeeperTestSuite) {
				s.keeper.SetGroupTransition(s.ctx, types.GroupTransition{
					SigningID:       tss.SigningID(0),
					CurrentGroupID:  tss.GroupID(1),
					IncomingGroupID: tss.GroupID(2),
					Status:          types.TRANSITION_STATUS_WAITING_EXECUTION,
					ExecTime:        execTime,
				})

				currentGroup := types.NewCurrentGroup(tss.GroupID(1), s.ctx.BlockTime().Add(-30*time.Minute))
				s.keeper.SetCurrentGroup(s.ctx, currentGroup)

				err := s.keeper.AddMember(s.ctx, acc1, tss.GroupID(1))
				s.Require().NoError(err)
				err = s.keeper.AddMember(s.ctx, acc2, tss.GroupID(1))
				s.Require().NoError(err)

				s.tssKeeper.EXPECT().MustGetMembers(gomock.Any(), tss.GroupID(1)).Return([]tsstypes.Member{
					{ID: 1, Address: acc1.String(), PubKey: []byte("test-pubkey-1")},
					{ID: 2, Address: acc2.String(), PubKey: []byte("test-pubkey-1")},
				})
			},
			expectOut: expectOut{
				status: types.TRANSITION_STATUS_UNSPECIFIED,
				currentGroup: types.CurrentGroup{
					GroupID:    tss.GroupID(2),
					ActiveTime: execTime,
				},
			},
			postProcess: func(s *KeeperTestSuite) {
				members := s.keeper.GetMembers(s.ctx)
				s.Require().Len(members, 0)
				s.Require().False(s.keeper.HasMember(s.ctx, acc1, tss.GroupID(1)))
				s.Require().False(s.keeper.HasMember(s.ctx, acc2, tss.GroupID(1)))
			},
		},
		{
			name: "transition with status WaitingExecution; has a current group",
			preProcess: func(s *KeeperTestSuite) {
				s.keeper.SetGroupTransition(s.ctx, types.GroupTransition{
					SigningID:       tss.SigningID(1),
					CurrentGroupID:  tss.GroupID(1),
					IncomingGroupID: tss.GroupID(2),
					Status:          types.TRANSITION_STATUS_WAITING_EXECUTION,
					ExecTime:        execTime,
				})

				currentGroup := types.NewCurrentGroup(tss.GroupID(1), s.ctx.BlockTime().Add(-30*time.Minute))
				s.keeper.SetCurrentGroup(s.ctx, currentGroup)

				err := s.keeper.AddMember(s.ctx, acc1, tss.GroupID(1))
				s.Require().NoError(err)
				err = s.keeper.AddMember(s.ctx, acc2, tss.GroupID(1))
				s.Require().NoError(err)

				s.tssKeeper.EXPECT().MustGetMembers(gomock.Any(), tss.GroupID(1)).Return([]tsstypes.Member{
					{ID: 1, Address: acc1.String(), PubKey: []byte("test-pubkey-1")},
					{ID: 2, Address: acc2.String(), PubKey: []byte("test-pubkey-1")},
				})
			},
			expectOut: expectOut{
				status: types.TRANSITION_STATUS_UNSPECIFIED,
				currentGroup: types.CurrentGroup{
					GroupID:    tss.GroupID(2),
					ActiveTime: execTime,
				},
			},
			postProcess: func(s *KeeperTestSuite) {
				members := s.keeper.GetMembers(s.ctx)
				s.Require().Len(members, 0)
				s.Require().False(s.keeper.HasMember(s.ctx, acc1, tss.GroupID(1)))
				s.Require().False(s.keeper.HasMember(s.ctx, acc2, tss.GroupID(1)))
			},
		},
		{
			name: "transition with status CreatingGroup; pass ExecTime",
			preProcess: func(s *KeeperTestSuite) {
				s.keeper.SetGroupTransition(s.ctx, types.GroupTransition{
					SigningID:       tss.SigningID(1),
					CurrentGroupID:  tss.GroupID(1),
					IncomingGroupID: tss.GroupID(2),
					Status:          types.TRANSITION_STATUS_CREATING_GROUP,
					ExecTime:        execTime,
				})

				currentGroup := types.NewCurrentGroup(tss.GroupID(1), s.ctx.BlockTime().Add(-30*time.Minute))
				s.keeper.SetCurrentGroup(s.ctx, currentGroup)
			},
			expectOut: expectOut{
				status: types.TRANSITION_STATUS_UNSPECIFIED,
				currentGroup: types.CurrentGroup{
					GroupID:    tss.GroupID(1),
					ActiveTime: s.ctx.BlockTime().Add(-30 * time.Minute),
				},
			},
		},
		{
			name: "transition with status WaitingSign; pass ExecTime",
			preProcess: func(s *KeeperTestSuite) {
				s.keeper.SetGroupTransition(s.ctx, types.GroupTransition{
					SigningID:       tss.SigningID(1),
					CurrentGroupID:  tss.GroupID(1),
					IncomingGroupID: tss.GroupID(2),
					Status:          types.TRANSITION_STATUS_WAITING_SIGN,
					ExecTime:        execTime,
				})

				currentGroup := types.NewCurrentGroup(tss.GroupID(1), s.ctx.BlockTime().Add(-30*time.Minute))
				s.keeper.SetCurrentGroup(s.ctx, currentGroup)
			},
			expectOut: expectOut{
				status: types.TRANSITION_STATUS_UNSPECIFIED,
				currentGroup: types.CurrentGroup{
					GroupID:    tss.GroupID(1),
					ActiveTime: s.ctx.BlockTime().Add(-30 * time.Minute),
				},
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			if tc.preProcess != nil {
				tc.preProcess(s)
			}

			if transition, ok := s.keeper.ShouldExecuteGroupTransition(s.ctx); ok {
				s.keeper.ExecuteGroupTransition(s.ctx, transition)
			}

			gt, found := s.keeper.GetGroupTransition(s.ctx)
			if tc.expectOut.status == types.TRANSITION_STATUS_UNSPECIFIED {
				s.Require().False(found)
			} else {
				s.Require().True(found)
				s.Require().Equal(tc.expectOut.status, gt.Status)
			}

			currentGroup := s.keeper.GetCurrentGroup(s.ctx)
			expect := tc.expectOut.currentGroup
			s.Require().Equal(expect.GroupID, currentGroup.GroupID)
			s.Require().Equal(expect.ActiveTime, currentGroup.ActiveTime)

			if tc.postProcess != nil {
				tc.postProcess(s)
			}
		})
	}
}
