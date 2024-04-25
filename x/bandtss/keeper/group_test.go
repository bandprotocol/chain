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
	s := testutil.NewTestSuite(t)
	ctx, k := s.Ctx, s.Keeper

	currentGroupID := tss.GroupID(1)
	newGroupID := tss.GroupID(2)
	currentGroup := tsstypes.Group{
		ID:     currentGroupID,
		PubKey: []byte("test-pubkey-group1"),
	}

	k.SetCurrentGroupID(ctx, currentGroupID)
	s.MockTSSKeeper.EXPECT().GetGroup(ctx, currentGroupID).Return(currentGroup, nil).AnyTimes()

	type input struct {
		groupID  tss.GroupID
		execTime time.Time
	}
	testCases := []struct {
		name        string
		preProcess  func()
		input       input
		expectErr   error
		postProcess func()
	}{
		{
			name: "replacement in progress - waiting signing",
			preProcess: func() {
				k.SetReplacement(ctx, types.Replacement{
					Status: types.REPLACEMENT_STATUS_WAITING_SIGNING,
				})
			},
			input: input{
				groupID:  newGroupID,
				execTime: ctx.BlockTime().Add(10 * time.Minute),
			},
			expectErr: types.ErrReplacementInProgress,
			postProcess: func() {
				k.SetReplacement(ctx, types.Replacement{})
			},
		},
		{
			name: "replacement in progress - waiting replace",
			preProcess: func() {
				k.SetReplacement(ctx, types.Replacement{
					Status: types.REPLACEMENT_STATUS_WAITING_REPLACE,
				})
			},
			input: input{
				groupID:  newGroupID,
				execTime: ctx.BlockTime().Add(10 * time.Minute),
			},
			expectErr: types.ErrReplacementInProgress,
			postProcess: func() {
				k.SetReplacement(ctx, types.Replacement{})
			},
		},
		{
			name:       "replacement exec time is in the past",
			preProcess: func() {},
			input: input{
				groupID:  newGroupID,
				execTime: time.Now().UTC().Add(time.Duration(-10) * time.Hour),
			},
			expectErr:   types.ErrInvalidExecTime,
			postProcess: func() {},
		},
		{
			name: "no current group",
			preProcess: func() {
				k.SetCurrentGroupID(ctx, 0)
			},
			input: input{
				groupID:  newGroupID,
				execTime: time.Now().UTC().Add(10 * time.Minute),
			},
			expectErr: types.ErrNoActiveGroup,
			postProcess: func() {
				k.SetCurrentGroupID(ctx, currentGroupID)
			},
		},
		{
			name: "replacing group is inactive",
			preProcess: func() {
				s.MockTSSKeeper.EXPECT().GetGroup(ctx, newGroupID).Return(tsstypes.Group{
					ID:     newGroupID,
					PubKey: []byte("test-pubkey-group2"),
					Status: tsstypes.GROUP_STATUS_FALLEN,
				}, nil)
			},
			input: input{
				groupID:  newGroupID,
				execTime: time.Now().UTC().Add(10 * time.Minute),
			},
			expectErr:   tsstypes.ErrGroupIsNotActive,
			postProcess: func() {},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.preProcess()
			_, err := k.CreateGroupReplacement(ctx, tc.input.groupID, tc.input.execTime)
			require.ErrorIs(t, err, tc.expectErr)
			tc.postProcess()
		})
	}
}

func TestHandleReplaceGroup(t *testing.T) {
	s := testutil.NewTestSuite(t)
	ctx, k := s.Ctx, s.Keeper

	currentGroupID := tss.GroupID(1)

	k.SetCurrentGroupID(ctx, currentGroupID)

	type expectOut struct {
		replacementStatus types.ReplacementStatus
		currentGroupID    tss.GroupID
	}

	testCases := []struct {
		name        string
		preProcess  func()
		expectOut   expectOut
		postProcess func()
		postCheck   func(t *testing.T)
	}{
		{
			name:       "no replacement setup",
			preProcess: func() {},
			expectOut: expectOut{
				replacementStatus: types.REPLACEMENT_STATUS_UNSPECIFIED,
				currentGroupID:    currentGroupID,
			},
			postProcess: func() {},
		},
		{
			name: "have replacement but not signed",
			preProcess: func() {
				k.SetReplacement(ctx, types.Replacement{
					SigningID: tss.SigningID(1),
					Status:    types.REPLACEMENT_STATUS_WAITING_SIGNING,
					ExecTime:  ctx.BlockTime().Add(10 * time.Minute),
				})
				s.MockTSSKeeper.EXPECT().GetSigning(ctx, tss.SigningID(1)).Return(tsstypes.Signing{
					ID:     tss.SigningID(1),
					Status: tsstypes.SIGNING_STATUS_WAITING,
				}, nil)
			},
			expectOut: expectOut{
				replacementStatus: types.REPLACEMENT_STATUS_WAITING_SIGNING,
				currentGroupID:    currentGroupID,
			},
			postProcess: func() {},
		},
		{
			name: "have replacement and signed but not meet exec time",
			preProcess: func() {
				k.SetReplacement(ctx, types.Replacement{
					SigningID: tss.SigningID(1),
					Status:    types.REPLACEMENT_STATUS_WAITING_SIGNING,
					ExecTime:  ctx.BlockTime().Add(10 * time.Minute),
				})
				s.MockTSSKeeper.EXPECT().GetSigning(ctx, tss.SigningID(1)).Return(tsstypes.Signing{
					ID:     tss.SigningID(1),
					Status: tsstypes.SIGNING_STATUS_SUCCESS,
				}, nil)
			},
			expectOut: expectOut{
				replacementStatus: types.REPLACEMENT_STATUS_WAITING_REPLACE,
				currentGroupID:    currentGroupID,
			},
			postProcess: func() {},
		},
		{
			name: "have replacement and signing failed",
			preProcess: func() {
				k.SetReplacement(ctx, types.Replacement{
					SigningID: tss.SigningID(1),
					Status:    types.REPLACEMENT_STATUS_WAITING_SIGNING,
					ExecTime:  ctx.BlockTime().Add(10 * time.Minute),
				})
				s.MockTSSKeeper.EXPECT().GetSigning(ctx, tss.SigningID(1)).Return(tsstypes.Signing{
					ID:     tss.SigningID(1),
					Status: tsstypes.SIGNING_STATUS_FALLEN,
				}, nil)
			},
			expectOut: expectOut{
				replacementStatus: types.REPLACEMENT_STATUS_FALLEN,
				currentGroupID:    currentGroupID,
			},
			postProcess: func() {},
		},
		{
			name: "have replacement and signing expired",
			preProcess: func() {
				k.SetReplacement(ctx, types.Replacement{
					SigningID: tss.SigningID(1),
					Status:    types.REPLACEMENT_STATUS_WAITING_SIGNING,
					ExecTime:  ctx.BlockTime().Add(10 * time.Minute),
				})
				s.MockTSSKeeper.EXPECT().GetSigning(ctx, tss.SigningID(1)).Return(tsstypes.Signing{
					ID:     tss.SigningID(1),
					Status: tsstypes.SIGNING_STATUS_EXPIRED,
				}, nil)
			},
			expectOut: expectOut{
				replacementStatus: types.REPLACEMENT_STATUS_FALLEN,
				currentGroupID:    currentGroupID,
			},
			postProcess: func() {},
		},
		{
			name: "have replacement and signing is waiting but exec time is passed",
			preProcess: func() {
				k.SetReplacement(ctx, types.Replacement{
					SigningID: tss.SigningID(1),
					Status:    types.REPLACEMENT_STATUS_WAITING_SIGNING,
					ExecTime:  ctx.BlockTime().Add(10 * time.Minute),
				})

				ctx = ctx.WithBlockTime(ctx.BlockTime().Add(11 * time.Minute))
				s.MockTSSKeeper.EXPECT().GetSigning(ctx, tss.SigningID(1)).Return(tsstypes.Signing{
					ID:     tss.SigningID(1),
					Status: tsstypes.SIGNING_STATUS_WAITING,
				}, nil)
			},
			expectOut: expectOut{
				replacementStatus: types.REPLACEMENT_STATUS_FALLEN,
				currentGroupID:    currentGroupID,
			},
			postProcess: func() {},
		},
		{
			name: "have replacement and signing is ready and passed exec time",
			preProcess: func() {
				k.SetReplacement(ctx, types.Replacement{
					SigningID:      tss.SigningID(1),
					Status:         types.REPLACEMENT_STATUS_WAITING_SIGNING,
					ExecTime:       ctx.BlockTime().Add(10 * time.Minute),
					CurrentGroupID: currentGroupID,
					NewGroupID:     tss.GroupID(2),
				})
				ctx = ctx.WithBlockTime(ctx.BlockTime().Add(11 * time.Minute))

				s.MockTSSKeeper.EXPECT().GetSigning(ctx, tss.SigningID(1)).Return(tsstypes.Signing{
					ID:     tss.SigningID(1),
					Status: tsstypes.SIGNING_STATUS_SUCCESS,
				}, nil)
				s.MockTSSKeeper.EXPECT().MustGetMembers(ctx, tss.GroupID(1)).Return([]tsstypes.Member{
					{ID: 1, Address: "band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs"},
				})
				s.MockTSSKeeper.EXPECT().MustGetMembers(ctx, tss.GroupID(2)).Return([]tsstypes.Member{
					{ID: 1, Address: "band1p40yh3zkmhcv0ecqp3mcazy83sa57rgjp07dun"},
				})
			},
			expectOut: expectOut{
				replacementStatus: types.REPLACEMENT_STATUS_SUCCESS,
				currentGroupID:    tss.GroupID(2),
			},
			postCheck: func(t *testing.T) {
				members := k.GetMembers(ctx)
				require.Len(t, members, 1)
				require.Equal(t, "band1p40yh3zkmhcv0ecqp3mcazy83sa57rgjp07dun", members[0].Address)
			},
			postProcess: func() {
				k.SetCurrentGroupID(ctx, tss.GroupID(1))
			},
		},
		{
			name: "have replacement and only waiting to replace but not meet exec time",
			preProcess: func() {
				k.SetReplacement(ctx, types.Replacement{
					SigningID: tss.SigningID(1),
					Status:    types.REPLACEMENT_STATUS_WAITING_REPLACE,
					ExecTime:  ctx.BlockTime().Add(10 * time.Minute),
				})
			},
			expectOut: expectOut{
				replacementStatus: types.REPLACEMENT_STATUS_WAITING_REPLACE,
				currentGroupID:    currentGroupID,
			},
			postProcess: func() {},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, currentGroupID, k.GetCurrentGroupID(ctx))
			tc.preProcess()

			err := k.HandleReplaceGroup(ctx, ctx.BlockTime())
			require.NoError(t, err)

			require.Equal(t, tc.expectOut.replacementStatus, k.GetReplacement(ctx).Status)
			require.Equal(t, tc.expectOut.currentGroupID, k.GetCurrentGroupID(ctx))

			if tc.postCheck != nil {
				tc.postCheck(t)
			}

			tc.postProcess()
		})
	}
}
