package keeper_test

import (
	"time"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/bandtss/types"
	tsstypes "github.com/bandprotocol/chain/v2/x/tss/types"
)

func (s *KeeperTestSuite) TestSuccessCreateGroupReplacement() {
	ctx, k := s.ctx, s.app.BandtssKeeper
	s.SetupGroup(tsstypes.GROUP_STATUS_ACTIVE)

	currentGroupID := s.app.BandtssKeeper.GetCurrentGroupID(ctx)
	newGroupID := tss.GroupID(2)
	currentGroup := s.app.TSSKeeper.MustGetGroup(ctx, currentGroupID)
	newGroup := s.app.TSSKeeper.MustGetGroup(ctx, newGroupID)
	execTime := time.Now().UTC().Add(10 * time.Minute)

	signingID, err := k.CreateGroupReplacement(ctx, newGroupID, execTime)
	s.Require().NoError(err)

	expectedReplacement := types.Replacement{
		SigningID:      signingID,
		CurrentGroupID: currentGroupID,
		NewGroupID:     newGroupID,
		CurrentPubKey:  currentGroup.PubKey,
		NewPubKey:      newGroup.PubKey,
		Status:         types.REPLACEMENT_STATUS_WAITING_SIGNING,
		ExecTime:       execTime,
	}

	resp, err := s.queryClient.Replacement(ctx, &types.QueryReplacementRequest{})
	s.Require().NoError(err)
	s.Require().Equal(expectedReplacement, resp.Replacement)
}

func (s *KeeperTestSuite) TestFailCreateGroupReplacement() {
	ctx, k := s.ctx, s.app.BandtssKeeper
	ctx = ctx.WithBlockTime(time.Now().UTC())

	s.SetupGroup(tsstypes.GROUP_STATUS_ACTIVE)
	currentGroupID := s.app.BandtssKeeper.GetCurrentGroupID(ctx)

	type input struct {
		groupID  tss.GroupID
		execTime time.Time
	}
	testcases := []struct {
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
				groupID:  tss.GroupID(2),
				execTime: time.Now().UTC().Add(10 * time.Minute),
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
				groupID:  tss.GroupID(2),
				execTime: time.Now().UTC().Add(10 * time.Minute),
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
				groupID:  tss.GroupID(2),
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
				groupID:  tss.GroupID(2),
				execTime: time.Now().UTC().Add(10 * time.Minute),
			},
			expectErr: types.ErrNoActiveGroup,
			postProcess: func() {
				k.SetCurrentGroupID(ctx, currentGroupID)
			},
		},
	}

	for _, tc := range testcases {
		s.Run(tc.name, func() {
			tc.preProcess()
			_, err := k.CreateGroupReplacement(ctx, tc.input.groupID, tc.input.execTime)
			s.Require().ErrorIs(err, tc.expectErr)
			tc.postProcess()
		})
	}
}

func (s *KeeperTestSuite) TestHandleReplaceGroup() {
	ctx, k := s.ctx, s.app.BandtssKeeper
	ctx = ctx.WithBlockTime(time.Now().UTC())

	s.SetupGroup(tsstypes.GROUP_STATUS_ACTIVE)
	currentGroupID := tss.GroupID(1)

	type expectOut struct {
		replacementStatus types.ReplacementStatus
		currentGroupID    tss.GroupID
	}

	testcases := []struct {
		name        string
		preProcess  func()
		expectOut   expectOut
		postProcess func()
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
					ExecTime:  time.Now().UTC().Add(10 * time.Minute),
				})
				s.app.TSSKeeper.SetSigning(ctx, tsstypes.Signing{
					ID:     tss.SigningID(1),
					Status: tsstypes.SIGNING_STATUS_WAITING,
				})
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
					ExecTime:  time.Now().UTC().Add(10 * time.Minute),
				})
				s.app.TSSKeeper.SetSigning(ctx, tsstypes.Signing{
					ID:     tss.SigningID(1),
					Status: tsstypes.SIGNING_STATUS_SUCCESS,
				})
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
					ExecTime:  time.Now().UTC().Add(10 * time.Minute),
				})
				s.app.TSSKeeper.SetSigning(ctx, tsstypes.Signing{
					ID:     tss.SigningID(1),
					Status: tsstypes.SIGNING_STATUS_FALLEN,
				})
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
					ExecTime:  time.Now().UTC().Add(10 * time.Minute),
				})
				s.app.TSSKeeper.SetSigning(ctx, tsstypes.Signing{
					ID:     tss.SigningID(1),
					Status: tsstypes.SIGNING_STATUS_EXPIRED,
				})
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
					ExecTime:  time.Now().UTC().Add(10 * time.Minute),
				})
				s.app.TSSKeeper.SetSigning(ctx, tsstypes.Signing{
					ID:     tss.SigningID(1),
					Status: tsstypes.SIGNING_STATUS_WAITING,
				})
				ctx = ctx.WithBlockTime(time.Now().UTC().Add(11 * time.Minute))
			},
			expectOut: expectOut{
				replacementStatus: types.REPLACEMENT_STATUS_FALLEN,
				currentGroupID:    currentGroupID,
			},
			postProcess: func() {
				ctx = ctx.WithBlockTime(time.Now().UTC())
			},
		},
		// TODO: remove this after fix test_case 2 (group 2; same user)
		// {
		// 	name: "have replacement and signing is ready and passed exec time",
		// 	preProcess: func() {
		// 		k.SetReplacement(ctx, types.Replacement{
		// 			SigningID:      tss.SigningID(1),
		// 			Status:         types.REPLACEMENT_STATUS_WAITING_SIGNING,
		// 			ExecTime:       time.Now().UTC().Add(10 * time.Minute),
		// 			CurrentGroupID: currentGroupID,
		// 			NewGroupID:     tss.GroupID(2),
		// 		})
		// 		s.app.TSSKeeper.SetSigning(ctx, tsstypes.Signing{
		// 			ID:     tss.SigningID(1),
		// 			Status: tsstypes.SIGNING_STATUS_SUCCESS,
		// 		})
		// 		ctx = ctx.WithBlockTime(time.Now().UTC().Add(11 * time.Minute))
		// 	},
		// 	expectOut: expectOut{
		// 		replacementStatus: types.REPLACEMENT_STATUS_SUCCESS,
		// 		currentGroupID:    tss.GroupID(2),
		// 	},
		// 	postProcess: func() {
		// 		ctx = ctx.WithBlockTime(time.Now().UTC())
		// 	},
		// },
		{
			name: "have replacement and only waiting to replace but not meet exec time",
			preProcess: func() {
				k.SetReplacement(ctx, types.Replacement{
					SigningID: tss.SigningID(1),
					Status:    types.REPLACEMENT_STATUS_WAITING_REPLACE,
					ExecTime:  time.Now().UTC().Add(10 * time.Minute),
				})
				s.app.TSSKeeper.SetSigning(ctx, tsstypes.Signing{
					ID:     tss.SigningID(1),
					Status: tsstypes.SIGNING_STATUS_SUCCESS,
				})
			},
			expectOut: expectOut{
				replacementStatus: types.REPLACEMENT_STATUS_WAITING_REPLACE,
				currentGroupID:    currentGroupID,
			},
			postProcess: func() {},
		},
	}

	for _, tc := range testcases {
		s.Run(tc.name, func() {
			s.Require().Equal(currentGroupID, k.GetCurrentGroupID(ctx))
			tc.preProcess()

			err := k.HandleReplaceGroup(ctx, ctx.BlockTime())
			s.Require().NoError(err)

			// s.Require().Equal(tc.expectOut.replacementStatus, k.GetReplacement(ctx).Status)
			s.Require().Equal(tc.expectOut.currentGroupID, k.GetCurrentGroupID(ctx))

			tc.postProcess()
		})
	}
}
