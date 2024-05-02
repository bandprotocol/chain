package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/bandtss/testutil"
	"github.com/bandprotocol/chain/v2/x/bandtss/types"
	tsstypes "github.com/bandprotocol/chain/v2/x/tss/types"
)

func TestSuccessCreateGroupReplacement(t *testing.T) {
	s := testutil.NewTestSuite(t)
	ctx, k := s.Ctx, s.Keeper

	currentGroupID := tss.GroupID(1)
	newGroupID := tss.GroupID(2)
	execTime := time.Now().UTC().Add(10 * time.Minute)

	currentGroup := tsstypes.Group{
		ID:     currentGroupID,
		PubKey: []byte("test-pubkey-group1"),
	}
	newGroup := tsstypes.Group{
		ID:     newGroupID,
		PubKey: []byte("test-pubkey-group2"),
		Status: tsstypes.GROUP_STATUS_ACTIVE,
	}
	expectSigning := &tsstypes.Signing{
		ID: tss.SigningID(1),
	}

	k.SetCurrentGroupID(ctx, currentGroupID)
	s.MockTSSKeeper.EXPECT().GetGroup(ctx, currentGroupID).Return(currentGroup, nil).AnyTimes()
	s.MockTSSKeeper.EXPECT().GetGroup(ctx, newGroupID).Return(newGroup, nil).AnyTimes()
	s.MockTSSKeeper.EXPECT().HandleSigningContent(ctx, types.NewReplaceGroupSignatureOrder(newGroup.PubKey)).Return([]byte("test-msg"), nil)
	s.MockTSSKeeper.EXPECT().CreateSigning(ctx, currentGroup, []byte("test-msg")).Return(expectSigning, nil)

	signingID, err := k.CreateGroupReplacement(ctx, newGroupID, execTime)
	require.NoError(t, err)
	require.Equal(t, expectSigning.ID, signingID)

	expectReplacement := types.Replacement{
		SigningID:      expectSigning.ID,
		CurrentGroupID: currentGroupID,
		NewGroupID:     newGroupID,
		CurrentPubKey:  currentGroup.PubKey,
		NewPubKey:      newGroup.PubKey,
		Status:         types.REPLACEMENT_STATUS_WAITING_SIGNING,
		ExecTime:       execTime,
	}
	actualReplacement := k.GetReplacement(ctx)
	require.Equal(t, expectReplacement, actualReplacement)
}

func TestFailCreateGroupReplacement(t *testing.T) {
	currentGroupID := tss.GroupID(1)
	newGroupID := tss.GroupID(2)
	currentGroup := tsstypes.Group{
		ID:     currentGroupID,
		PubKey: []byte("test-pubkey-group1"),
	}

	type input struct {
		groupID      tss.GroupID
		waitDuration time.Duration
	}
	testCases := []struct {
		name       string
		preProcess func(s *testutil.TestSuite)
		input      input
		expectErr  error
	}{
		{
			name: "replacement in progress - waiting signing",
			preProcess: func(s *testutil.TestSuite) {
				s.Keeper.SetReplacement(s.Ctx, types.Replacement{
					Status: types.REPLACEMENT_STATUS_WAITING_SIGNING,
				})
			},
			input: input{
				groupID:      newGroupID,
				waitDuration: 10 * time.Minute,
			},
			expectErr: types.ErrReplacementInProgress,
		},
		{
			name: "replacement in progress - waiting replace",
			preProcess: func(s *testutil.TestSuite) {
				s.Keeper.SetReplacement(s.Ctx, types.Replacement{
					Status: types.REPLACEMENT_STATUS_WAITING_REPLACE,
				})
			},
			input: input{
				groupID:      newGroupID,
				waitDuration: 10 * time.Minute,
			},
			expectErr: types.ErrReplacementInProgress,
		},
		{
			name: "replacement exec time is in the past",
			input: input{
				groupID:      newGroupID,
				waitDuration: time.Duration(-10) * time.Hour,
			},
			expectErr: types.ErrInvalidExecTime,
		},
		{
			name: "no current group",
			preProcess: func(s *testutil.TestSuite) {
				s.Keeper.SetCurrentGroupID(s.Ctx, 0)
			},
			input: input{
				groupID:      newGroupID,
				waitDuration: 10 * time.Minute,
			},
			expectErr: types.ErrNoActiveGroup,
		},
		{
			name: "replacing group is inactive",
			preProcess: func(s *testutil.TestSuite) {
				s.MockTSSKeeper.EXPECT().GetGroup(s.Ctx, newGroupID).Return(tsstypes.Group{
					ID:     newGroupID,
					PubKey: []byte("test-pubkey-group2"),
					Status: tsstypes.GROUP_STATUS_FALLEN,
				}, nil)
			},
			input: input{
				groupID:      newGroupID,
				waitDuration: 10 * time.Minute,
			},
			expectErr: tsstypes.ErrGroupIsNotActive,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := testutil.NewTestSuite(t)
			ctx, k := s.Ctx, s.Keeper

			k.SetCurrentGroupID(ctx, currentGroupID)
			s.MockTSSKeeper.EXPECT().GetGroup(ctx, currentGroupID).Return(currentGroup, nil).AnyTimes()

			if tc.preProcess != nil {
				tc.preProcess(&s)
			}

			execTime := s.Ctx.BlockTime().Add(tc.input.waitDuration)
			_, err := k.CreateGroupReplacement(ctx, tc.input.groupID, execTime)
			require.ErrorIs(t, err, tc.expectErr)
		})
	}
}

func TestHandleReplaceGroup(t *testing.T) {
	currentGroupID := tss.GroupID(1)

	type expectOut struct {
		replacementStatus types.ReplacementStatus
		currentGroupID    tss.GroupID
	}

	testCases := []struct {
		name       string
		preProcess func(s *testutil.TestSuite)
		expectOut  expectOut
		postCheck  func(s *testutil.TestSuite)
	}{
		{
			name: "no replacement setup",
			expectOut: expectOut{
				replacementStatus: types.REPLACEMENT_STATUS_UNSPECIFIED,
				currentGroupID:    currentGroupID,
			},
		},
		{
			name: "have replacement but not signed",
			preProcess: func(s *testutil.TestSuite) {
				s.Keeper.SetReplacement(s.Ctx, types.Replacement{
					SigningID: tss.SigningID(1),
					Status:    types.REPLACEMENT_STATUS_WAITING_SIGNING,
					ExecTime:  s.Ctx.BlockTime().Add(10 * time.Minute),
				})
				s.MockTSSKeeper.EXPECT().GetSigning(s.Ctx, tss.SigningID(1)).Return(tsstypes.Signing{
					ID:     tss.SigningID(1),
					Status: tsstypes.SIGNING_STATUS_WAITING,
				}, nil)
			},
			expectOut: expectOut{
				replacementStatus: types.REPLACEMENT_STATUS_WAITING_SIGNING,
				currentGroupID:    currentGroupID,
			},
		},
		{
			name: "have replacement and signed but not meet exec time",
			preProcess: func(s *testutil.TestSuite) {
				s.Keeper.SetReplacement(s.Ctx, types.Replacement{
					SigningID: tss.SigningID(1),
					Status:    types.REPLACEMENT_STATUS_WAITING_SIGNING,
					ExecTime:  s.Ctx.BlockTime().Add(10 * time.Minute),
				})
				s.MockTSSKeeper.EXPECT().GetSigning(s.Ctx, tss.SigningID(1)).Return(tsstypes.Signing{
					ID:     tss.SigningID(1),
					Status: tsstypes.SIGNING_STATUS_SUCCESS,
				}, nil)
			},
			expectOut: expectOut{
				replacementStatus: types.REPLACEMENT_STATUS_WAITING_REPLACE,
				currentGroupID:    currentGroupID,
			},
		},
		{
			name: "have replacement and signing failed",
			preProcess: func(s *testutil.TestSuite) {
				s.Keeper.SetReplacement(s.Ctx, types.Replacement{
					SigningID: tss.SigningID(1),
					Status:    types.REPLACEMENT_STATUS_WAITING_SIGNING,
					ExecTime:  s.Ctx.BlockTime().Add(10 * time.Minute),
				})
				s.MockTSSKeeper.EXPECT().GetSigning(s.Ctx, tss.SigningID(1)).Return(tsstypes.Signing{
					ID:     tss.SigningID(1),
					Status: tsstypes.SIGNING_STATUS_FALLEN,
				}, nil)
			},
			expectOut: expectOut{
				replacementStatus: types.REPLACEMENT_STATUS_FALLEN,
				currentGroupID:    currentGroupID,
			},
		},
		{
			name: "have replacement and signing expired",
			preProcess: func(s *testutil.TestSuite) {
				s.Keeper.SetReplacement(s.Ctx, types.Replacement{
					SigningID: tss.SigningID(1),
					Status:    types.REPLACEMENT_STATUS_WAITING_SIGNING,
					ExecTime:  s.Ctx.BlockTime().Add(10 * time.Minute),
				})
				s.MockTSSKeeper.EXPECT().GetSigning(s.Ctx, tss.SigningID(1)).Return(tsstypes.Signing{
					ID:     tss.SigningID(1),
					Status: tsstypes.SIGNING_STATUS_EXPIRED,
				}, nil)
			},
			expectOut: expectOut{
				replacementStatus: types.REPLACEMENT_STATUS_FALLEN,
				currentGroupID:    currentGroupID,
			},
		},
		{
			name: "have replacement and signing is waiting but exec time is passed",
			preProcess: func(s *testutil.TestSuite) {
				s.Keeper.SetReplacement(s.Ctx, types.Replacement{
					SigningID: tss.SigningID(1),
					Status:    types.REPLACEMENT_STATUS_WAITING_SIGNING,
					ExecTime:  s.Ctx.BlockTime().Add(10 * time.Minute),
				})

				s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(11 * time.Minute))
				s.MockTSSKeeper.EXPECT().GetSigning(s.Ctx, tss.SigningID(1)).Return(tsstypes.Signing{
					ID:     tss.SigningID(1),
					Status: tsstypes.SIGNING_STATUS_WAITING,
				}, nil)
			},
			expectOut: expectOut{
				replacementStatus: types.REPLACEMENT_STATUS_FALLEN,
				currentGroupID:    currentGroupID,
			},
		},
		{
			name: "have replacement and signing is ready and passed exec time",
			preProcess: func(s *testutil.TestSuite) {
				s.Keeper.SetReplacement(s.Ctx, types.Replacement{
					SigningID:      tss.SigningID(1),
					Status:         types.REPLACEMENT_STATUS_WAITING_SIGNING,
					ExecTime:       s.Ctx.BlockTime().Add(10 * time.Minute),
					CurrentGroupID: currentGroupID,
					NewGroupID:     tss.GroupID(2),
				})
				s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(11 * time.Minute))

				s.MockTSSKeeper.EXPECT().GetSigning(s.Ctx, tss.SigningID(1)).Return(tsstypes.Signing{
					ID:     tss.SigningID(1),
					Status: tsstypes.SIGNING_STATUS_SUCCESS,
				}, nil)
				s.MockTSSKeeper.EXPECT().MustGetMembers(s.Ctx, tss.GroupID(1)).Return([]tsstypes.Member{
					{ID: 1, Address: "band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs"},
				})
				s.MockTSSKeeper.EXPECT().MustGetMembers(s.Ctx, tss.GroupID(2)).Return([]tsstypes.Member{
					{ID: 1, Address: "band1p40yh3zkmhcv0ecqp3mcazy83sa57rgjp07dun"},
				})
			},
			expectOut: expectOut{
				replacementStatus: types.REPLACEMENT_STATUS_SUCCESS,
				currentGroupID:    tss.GroupID(2),
			},
			postCheck: func(s *testutil.TestSuite) {
				members := s.Keeper.GetMembers(s.Ctx)
				require.Len(s.T(), members, 1)
				require.Equal(s.T(), "band1p40yh3zkmhcv0ecqp3mcazy83sa57rgjp07dun", members[0].Address)
			},
		},
		{
			name: "have replacement and only waiting to replace but not meet exec time",
			preProcess: func(s *testutil.TestSuite) {
				s.Keeper.SetReplacement(s.Ctx, types.Replacement{
					SigningID: tss.SigningID(1),
					Status:    types.REPLACEMENT_STATUS_WAITING_REPLACE,
					ExecTime:  s.Ctx.BlockTime().Add(10 * time.Minute),
				})
			},
			expectOut: expectOut{
				replacementStatus: types.REPLACEMENT_STATUS_WAITING_REPLACE,
				currentGroupID:    currentGroupID,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := testutil.NewTestSuite(t)
			s.Keeper.SetCurrentGroupID(s.Ctx, currentGroupID)

			if tc.preProcess != nil {
				tc.preProcess(&s)
			}

			err := s.Keeper.HandleReplaceGroup(s.Ctx, s.Ctx.BlockTime())
			require.NoError(t, err)

			require.Equal(t, tc.expectOut.replacementStatus, s.Keeper.GetReplacement(s.Ctx).Status)
			require.Equal(t, tc.expectOut.currentGroupID, s.Keeper.GetCurrentGroupID(s.Ctx))

			if tc.postCheck != nil {
				tc.postCheck(&s)
			}
		})
	}
}
