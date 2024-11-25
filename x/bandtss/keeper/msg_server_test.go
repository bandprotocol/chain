package keeper_test

import (
	"fmt"
	"time"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/bandprotocol/chain/v3/pkg/tss"
	bandtesting "github.com/bandprotocol/chain/v3/testing"
	"github.com/bandprotocol/chain/v3/x/bandtss/types"
	tsstypes "github.com/bandprotocol/chain/v3/x/tss/types"
)

type TestCase struct {
	Name        string
	PreProcess  func()
	PostCheck   func()
	ExpectedErr error
}

func (s *AppTestSuite) TestSuccessTransitionGroupReqWithCurrentGroup() {
	ctx, k := s.ctx, s.app.BandtssKeeper
	members := []string{bandtesting.Alice.Address.String()}
	groupCtx := s.SetupNewGroup(5, 3)

	_, err := s.msgSrvr.TransitionGroup(ctx, &types.MsgTransitionGroup{
		Members:   members,
		Threshold: 1,
		ExecTime:  ctx.BlockTime().Add(10 * time.Second),
		Authority: s.authority.String(),
	})
	s.Require().NoError(err)

	// Check if the group is created but not impact current group ID in bandtss.
	s.Require().Equal(groupCtx.GroupID, k.GetCurrentGroup(ctx).GroupID)
	group := s.app.TSSKeeper.MustGetGroup(ctx, groupCtx.GroupID)
	transition, found := k.GetGroupTransition(ctx)
	expectedTransition := types.GroupTransition{
		Status:             types.TRANSITION_STATUS_CREATING_GROUP,
		CurrentGroupID:     groupCtx.GroupID,
		CurrentGroupPubKey: group.PubKey,
		IncomingGroupID:    groupCtx.GroupID + 1,
		ExecTime:           ctx.BlockTime().Add(10 * time.Second),
	}
	s.Require().True(found)
	s.Require().Equal(expectedTransition, transition)
}

func (s *AppTestSuite) TestSuccessTransitionGroupReqNoCurrentGroup() {
	ctx, k := s.ctx, s.app.BandtssKeeper
	members := []string{bandtesting.Alice.Address.String()}

	_, err := s.msgSrvr.TransitionGroup(ctx, &types.MsgTransitionGroup{
		Members:   members,
		Threshold: 1,
		ExecTime:  ctx.BlockTime().Add(10 * time.Second),
		Authority: s.authority.String(),
	})
	s.Require().NoError(err)

	// Check if the group is created but not impact current group ID in bandtss.
	s.Require().Equal(tss.GroupID(0), k.GetCurrentGroup(ctx).GroupID)
	transition, found := k.GetGroupTransition(ctx)
	expectedTransition := types.GroupTransition{
		Status:          types.TRANSITION_STATUS_CREATING_GROUP,
		CurrentGroupID:  tss.GroupID(0),
		IncomingGroupID: tss.GroupID(1),
		ExecTime:        ctx.BlockTime().Add(10 * time.Second),
	}
	s.Require().True(found)
	s.Require().Equal(expectedTransition, transition)
}

func (s *AppTestSuite) TestFailTransitionGroup() {
	ctx := s.ctx
	tssParams := s.app.TSSKeeper.GetParams(ctx)

	testCases := []struct {
		name        string
		input       *types.MsgTransitionGroup
		preProcess  func()
		expectErr   error
		postProcess func()
	}{
		{
			name: "invalid authority",
			input: &types.MsgTransitionGroup{
				Members:   []string{"band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs"},
				Threshold: 1,
				Authority: "band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs",
				ExecTime:  ctx.BlockTime().Add(10 * time.Second),
			},
			preProcess:  func() {},
			postProcess: func() {},
			expectErr:   govtypes.ErrInvalidSigner,
		},
		{
			name: "over max group size",
			input: &types.MsgTransitionGroup{
				Members:   []string{"band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs", bandtesting.Alice.Address.String()},
				Threshold: 1,
				Authority: s.authority.String(),
				ExecTime:  ctx.BlockTime().Add(10 * time.Second),
			},
			preProcess: func() {
				newParams := tssParams
				newParams.MaxGroupSize = 1
				err := s.app.TSSKeeper.SetParams(ctx, newParams)
				s.Require().NoError(err)
			},
			postProcess: func() {
				err := s.app.TSSKeeper.SetParams(ctx, tssParams)
				s.Require().NoError(err)
			},
			expectErr: tsstypes.ErrGroupCreationFailed,
		},
		{
			name: "duplicate members",
			input: &types.MsgTransitionGroup{
				Members: []string{
					"band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs",
					"band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs",
					"band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs",
				},
				Threshold: 1,
				Authority: s.authority.String(),
				ExecTime:  ctx.BlockTime().Add(10 * time.Second),
			},
			preProcess:  func() {},
			postProcess: func() {},
			expectErr:   fmt.Errorf("duplicated member found within the list"),
		},
		{
			name: "threshold more than members length",
			input: &types.MsgTransitionGroup{
				Members:   []string{"band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs"},
				Threshold: 10,
				Authority: s.authority.String(),
				ExecTime:  ctx.BlockTime().Add(10 * time.Second),
			},
			preProcess:  func() {},
			postProcess: func() {},
			expectErr:   types.ErrInvalidThreshold,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			tc.preProcess()

			err := tc.input.ValidateBasic()
			if err != nil {
				s.Require().ErrorContains(err, tc.expectErr.Error())
			} else {
				_, err := s.msgSrvr.TransitionGroup(ctx, tc.input)
				s.Require().ErrorIs(err, tc.expectErr)
			}

			tc.postProcess()
		})
	}
}

func (s *AppTestSuite) TestFailForceTransitionGroupInvalidExecTime() {
	_ = s.SetupNewGroup(5, 3)
	group2Ctx, err := s.CreateNewGroup(3, 2, s.ctx.BlockTime().Add(10*time.Minute))
	s.Require().NoError(err)

	maxTransitionDuration := s.app.BandtssKeeper.GetParams(s.ctx).MaxTransitionDuration
	execTime := s.ctx.BlockTime().Add(10 * time.Minute).Add(maxTransitionDuration)

	_, err = s.msgSrvr.ForceTransitionGroup(s.ctx, &types.MsgForceTransitionGroup{
		IncomingGroupID: group2Ctx.GroupID,
		ExecTime:        execTime,
		Authority:       s.authority.String(),
	})
	s.Require().ErrorIs(err, types.ErrInvalidExecTime)
}

func (s *AppTestSuite) TestFailForceTransitionGroupInvalidGroupStatus() {
	_ = s.SetupNewGroup(5, 3)
	group2Ctx, err := s.CreateNewGroup(3, 2, s.ctx.BlockTime().Add(10*time.Minute))
	s.Require().NoError(err)

	group2 := s.app.TSSKeeper.MustGetGroup(s.ctx, group2Ctx.GroupID)
	group2.Status = tsstypes.GROUP_STATUS_FALLEN
	s.app.TSSKeeper.SetGroup(s.ctx, group2)

	s.app.BandtssKeeper.DeleteGroupTransition(s.ctx)

	_, err = s.msgSrvr.ForceTransitionGroup(s.ctx, &types.MsgForceTransitionGroup{
		IncomingGroupID: group2Ctx.GroupID,
		ExecTime:        s.ctx.BlockTime().Add(10 * time.Minute),
		Authority:       s.authority.String(),
	})
	s.Require().ErrorIs(err, types.ErrInvalidIncomingGroup)
}

func (s *AppTestSuite) TestFailForceTransitionGroupInvalidGroupID() {
	ctx, msgSrvr, _ := s.ctx, s.msgSrvr, s.app.TSSKeeper

	group1Ctx := s.SetupNewGroup(5, 3)
	_, err := msgSrvr.ForceTransitionGroup(ctx, &types.MsgForceTransitionGroup{
		IncomingGroupID: group1Ctx.GroupID,
		ExecTime:        ctx.BlockTime().Add(10 * time.Minute),
		Authority:       s.authority.String(),
	})
	s.Require().ErrorIs(err, types.ErrInvalidGroupID)
}

func (s *AppTestSuite) TestFailForceTransitionGroupFromWaitingExecutionStatus() {
	group1Ctx := s.SetupNewGroup(5, 3)
	group2Ctx, err := s.CreateNewGroup(3, 2, s.ctx.BlockTime().Add(10*time.Minute))
	s.Require().NoError(err)

	s.app.BandtssKeeper.DeleteGroupTransition(s.ctx)

	group3Ctx, err := s.CreateNewGroup(3, 2, s.ctx.BlockTime().Add(10*time.Minute))
	s.Require().NoError(err)

	ctx, msgSrvr, _ := s.ctx, s.msgSrvr, s.app.TSSKeeper

	err = s.SignTransition(group1Ctx)
	s.Require().NoError(err)
	transition, found := s.app.BandtssKeeper.GetGroupTransition(ctx)
	s.Require().True(found)
	s.Require().Equal(types.TRANSITION_STATUS_WAITING_EXECUTION, transition.Status)

	_, err = msgSrvr.ForceTransitionGroup(ctx, &types.MsgForceTransitionGroup{
		IncomingGroupID: group2Ctx.GroupID,
		ExecTime:        ctx.BlockTime().Add(10 * time.Minute),
		Authority:       s.authority.String(),
	})
	s.Require().ErrorIs(err, types.ErrTransitionInProgress)

	transition, found = s.app.BandtssKeeper.GetGroupTransition(ctx)
	g1 := s.app.TSSKeeper.MustGetGroup(ctx, group1Ctx.GroupID)
	g3 := s.app.TSSKeeper.MustGetGroup(ctx, group3Ctx.GroupID)

	expectedTransition := types.GroupTransition{
		Status:              types.TRANSITION_STATUS_WAITING_EXECUTION,
		CurrentGroupID:      group1Ctx.GroupID,
		CurrentGroupPubKey:  g1.PubKey,
		IncomingGroupID:     group3Ctx.GroupID,
		IncomingGroupPubKey: g3.PubKey,
		ExecTime:            ctx.BlockTime().Add(10 * time.Minute),
		SigningID:           tss.SigningID(2),
	}
	s.Require().True(found)
	s.Require().Equal(expectedTransition, transition)

	for _, acc := range group1Ctx.Accounts {
		m, err := s.app.BandtssKeeper.GetMember(ctx, acc.Address, group1Ctx.GroupID)
		s.Require().NoError(err)
		s.Require().True(m.IsActive)
	}

	for _, acc := range group2Ctx.Accounts {
		ok := s.app.BandtssKeeper.HasMember(ctx, acc.Address, group2Ctx.GroupID)
		s.Require().False(ok)
	}

	for _, acc := range group3Ctx.Accounts {
		ok := s.app.BandtssKeeper.HasMember(ctx, acc.Address, group3Ctx.GroupID)
		s.Require().True(ok)
	}
}

func (s *AppTestSuite) TestSuccessForceTransitionGroupFromFallenStatus() {
	group1Ctx := s.SetupNewGroup(5, 3)
	group2Ctx, err := s.CreateNewGroup(3, 2, s.ctx.BlockTime().Add(10*time.Minute))
	s.Require().NoError(err)

	s.app.BandtssKeeper.DeleteGroupTransition(s.ctx)

	group3Ctx, err := s.CreateNewGroup(3, 2, s.ctx.BlockTime().Add(10*time.Minute))
	s.Require().NoError(err)

	s.app.BandtssKeeper.DeleteGroupTransition(s.ctx)

	_, err = s.msgSrvr.ForceTransitionGroup(s.ctx, &types.MsgForceTransitionGroup{
		IncomingGroupID: group2Ctx.GroupID,
		ExecTime:        s.ctx.BlockTime().Add(10 * time.Minute),
		Authority:       s.authority.String(),
	})
	s.Require().NoError(err)

	transition, found := s.app.BandtssKeeper.GetGroupTransition(s.ctx)
	g1 := s.app.TSSKeeper.MustGetGroup(s.ctx, group1Ctx.GroupID)
	g2 := s.app.TSSKeeper.MustGetGroup(s.ctx, group2Ctx.GroupID)

	expectedTransition := types.GroupTransition{
		Status:              types.TRANSITION_STATUS_WAITING_EXECUTION,
		CurrentGroupID:      group1Ctx.GroupID,
		CurrentGroupPubKey:  g1.PubKey,
		IncomingGroupID:     group2Ctx.GroupID,
		IncomingGroupPubKey: g2.PubKey,
		ExecTime:            s.ctx.BlockTime().Add(10 * time.Minute),
		SigningID:           tss.SigningID(0),
		IsForceTransition:   true,
	}
	s.Require().True(found)
	s.Require().Equal(expectedTransition, transition)

	for _, acc := range group1Ctx.Accounts {
		m, err := s.app.BandtssKeeper.GetMember(s.ctx, acc.Address, group1Ctx.GroupID)
		s.Require().NoError(err)
		s.Require().True(m.IsActive)
	}

	for _, acc := range group2Ctx.Accounts {
		m, err := s.app.BandtssKeeper.GetMember(s.ctx, acc.Address, group2Ctx.GroupID)
		s.Require().NoError(err)
		s.Require().True(m.IsActive)
	}

	for _, acc := range group3Ctx.Accounts {
		ok := s.app.BandtssKeeper.HasMember(s.ctx, acc.Address, group3Ctx.GroupID)
		s.Require().False(ok)
	}
}

func (s *AppTestSuite) TestFailedRequestSignatureReq() {
	ctx, msgSrvr := s.ctx, s.msgSrvr
	groupCtx := s.SetupNewGroup(5, 3)
	s.app.BandtssKeeper.SetCurrentGroup(ctx, types.NewCurrentGroup(groupCtx.GroupID, s.ctx.BlockTime()))

	var req *types.MsgRequestSignature
	var err error

	tcs := []TestCase{
		{
			Name: "failure with no groupID",
			PreProcess: func() {
				s.app.BandtssKeeper.SetCurrentGroup(ctx, types.NewCurrentGroup(0, time.Time{}))
				req, err = types.NewMsgRequestSignature(
					tsstypes.NewTextSignatureOrder([]byte("msg")),
					sdk.NewCoins(sdk.NewInt64Coin("uband", 100)),
					bandtesting.FeePayer.Address.String(),
				)
				s.Require().NoError(err)
			},
			PostCheck: func() {
				s.app.BandtssKeeper.SetCurrentGroup(
					ctx,
					types.NewCurrentGroup(groupCtx.GroupID, s.ctx.BlockTime()),
				)
			},
			ExpectedErr: types.ErrNoActiveGroup,
		},
		{
			Name: "failure with fee is more than user's limit",
			PreProcess: func() {
				req, err = types.NewMsgRequestSignature(
					tsstypes.NewTextSignatureOrder([]byte("msg")),
					sdk.NewCoins(sdk.NewInt64Coin("uband", 10)),
					bandtesting.FeePayer.Address.String(),
				)
			},
			PostCheck:   func() {},
			ExpectedErr: types.ErrFeeExceedsLimit,
		},
	}

	for _, tc := range tcs {
		s.Run(fmt.Sprintf("Case %s", tc.Name), func() {
			if tc.PreProcess != nil {
				tc.PreProcess()
			}
			if tc.PostCheck != nil {
				defer tc.PostCheck()
			}

			balancesBefore := s.app.BankKeeper.GetAllBalances(ctx, bandtesting.FeePayer.Address)
			balancesModuleBefore := s.app.BankKeeper.GetAllBalances(
				ctx,
				s.app.BandtssKeeper.GetBandtssAccount(ctx).GetAddress(),
			)

			_, err := msgSrvr.RequestSignature(ctx, req)
			s.Require().ErrorIs(tc.ExpectedErr, err)

			balancesAfter := s.app.BankKeeper.GetAllBalances(ctx, bandtesting.FeePayer.Address)
			balancesModuleAfter := s.app.BankKeeper.GetAllBalances(
				ctx,
				s.app.BandtssKeeper.GetBandtssAccount(ctx).GetAddress(),
			)

			// Check if the balances of payer and module account doesn't change
			s.Require().Equal(balancesBefore, balancesAfter)
			s.Require().Equal(balancesModuleBefore, balancesModuleAfter)
		})
	}
}

func (s *AppTestSuite) TestSuccessRequestSignatureOnCurrentGroup() {
	ctx, msgSrvr, k := s.ctx, s.msgSrvr, s.app.BandtssKeeper

	groupCtx := s.SetupNewGroup(5, 3)
	k.DeleteGroupTransition(ctx)

	balancesBefore := s.app.BankKeeper.GetAllBalances(ctx, bandtesting.FeePayer.Address)
	balancesModuleBefore := s.app.BankKeeper.GetAllBalances(
		ctx,
		s.app.BandtssKeeper.GetBandtssAccount(ctx).GetAddress(),
	)

	msg, err := types.NewMsgRequestSignature(
		tsstypes.NewTextSignatureOrder([]byte("msg")),
		sdk.NewCoins(sdk.NewInt64Coin("uband", 100)),
		bandtesting.FeePayer.Address.String(),
	)
	s.Require().NoError(err)

	_, err = msgSrvr.RequestSignature(ctx, msg)
	s.Require().NoError(err)

	// Fee should be paid after requesting signature
	balancesAfter := s.app.BankKeeper.GetAllBalances(ctx, bandtesting.FeePayer.Address)
	balancesModuleAfter := s.app.BankKeeper.GetAllBalances(
		ctx,
		s.app.BandtssKeeper.GetBandtssAccount(ctx).GetAddress(),
	)

	group, err := s.app.TSSKeeper.GetGroup(ctx, groupCtx.GroupID)
	s.Require().NoError(err)

	diff := k.GetParams(ctx).FeePerSigner.MulInt(math.NewInt(int64(group.Threshold)))
	s.Require().Equal(diff, balancesBefore.Sub(balancesAfter...))
	s.Require().Equal(diff, balancesModuleAfter.Sub(balancesModuleBefore...))

	bandtssSigningID := types.SigningID(s.app.BandtssKeeper.GetSigningCount(ctx))
	tssSigningID := tss.SigningID(s.app.TSSKeeper.GetSigningCount(ctx))

	bandtssSigning, err := s.app.BandtssKeeper.GetSigning(ctx, bandtssSigningID)
	s.Require().NoError(err)
	s.Require().Equal(tssSigningID, bandtssSigning.CurrentGroupSigningID)
	s.Require().Equal(tss.SigningID(0), bandtssSigning.IncomingGroupSigningID)

	bandtssSigningIDMapping := s.app.BandtssKeeper.GetSigningIDMapping(ctx, tssSigningID)
	s.Require().Equal(bandtssSigningID, bandtssSigningIDMapping)
}

func (s *AppTestSuite) TestFailRequestSignatureInternalMessage() {
	ctx, msgSrvr, k := s.ctx, s.msgSrvr, s.app.BandtssKeeper

	_ = s.SetupNewGroup(5, 3)
	k.DeleteGroupTransition(ctx)

	msg, err := types.NewMsgRequestSignature(
		types.NewGroupTransitionSignatureOrder([]byte("msg"), time.Now()),
		sdk.NewCoins(sdk.NewInt64Coin("uband", 100)),
		bandtesting.FeePayer.Address.String(),
	)
	s.Require().NoError(err)

	_, err = msgSrvr.RequestSignature(ctx, msg)
	s.Require().ErrorIs(err, types.ErrContentNotAllowed)
}

func (s *AppTestSuite) TestSuccessRequestSignatureOnIncomingGroup() {
	ctx, msgSrvr := s.ctx, s.msgSrvr

	_, err := s.CreateNewGroup(5, 3, ctx.BlockTime().Add(10*time.Minute))
	s.Require().NoError(err)

	balancesBefore := s.app.BankKeeper.GetAllBalances(ctx, bandtesting.FeePayer.Address)
	balancesModuleBefore := s.app.BankKeeper.GetAllBalances(
		ctx,
		s.app.BandtssKeeper.GetBandtssAccount(ctx).GetAddress(),
	)

	msg, err := types.NewMsgRequestSignature(
		tsstypes.NewTextSignatureOrder([]byte("msg")),
		sdk.NewCoins(sdk.NewInt64Coin("uband", 100)),
		bandtesting.FeePayer.Address.String(),
	)
	s.Require().NoError(err)

	_, err = msgSrvr.RequestSignature(ctx, msg)
	s.Require().NoError(err)

	balancesAfter := s.app.BankKeeper.GetAllBalances(ctx, bandtesting.FeePayer.Address)
	balancesModuleAfter := s.app.BankKeeper.GetAllBalances(
		ctx,
		s.app.BandtssKeeper.GetBandtssAccount(ctx).GetAddress(),
	)

	s.Require().Equal(sdk.NewCoins(), balancesBefore.Sub(balancesAfter...))
	s.Require().Equal(sdk.NewCoins(), balancesModuleAfter.Sub(balancesModuleBefore...))

	bandtssSigningID := types.SigningID(s.app.BandtssKeeper.GetSigningCount(ctx))
	tssSigningID := tss.SigningID(s.app.TSSKeeper.GetSigningCount(ctx))

	bandtssSigning, err := s.app.BandtssKeeper.GetSigning(ctx, bandtssSigningID)
	s.Require().NoError(err)
	s.Require().Equal(tss.SigningID(0), bandtssSigning.CurrentGroupSigningID)
	s.Require().Equal(tssSigningID, bandtssSigning.IncomingGroupSigningID)

	bandtssSigningIDMapping := s.app.BandtssKeeper.GetSigningIDMapping(ctx, tssSigningID)
	s.Require().Equal(bandtssSigningID, bandtssSigningIDMapping)
}

func (s *AppTestSuite) TestSuccessRequestSignatureOnBothGroups() {
	group1Ctx, err := s.CreateNewGroup(5, 3, s.ctx.BlockTime().Add(10*time.Minute))
	s.Require().NoError(err)
	err = s.ExecuteReplaceGroup()
	s.Require().NoError(err)

	_, err = s.CreateNewGroup(3, 2, s.ctx.BlockTime().Add(10*time.Minute))
	s.Require().NoError(err)
	err = s.SignTransition(group1Ctx)
	s.Require().NoError(err)

	ctx, msgSrvr := s.ctx, s.msgSrvr

	balancesBefore := s.app.BankKeeper.GetAllBalances(s.ctx, bandtesting.FeePayer.Address)
	balancesModuleBefore := s.app.BankKeeper.GetAllBalances(
		s.ctx,
		s.app.BandtssKeeper.GetBandtssAccount(s.ctx).GetAddress(),
	)

	msg, err := types.NewMsgRequestSignature(
		tsstypes.NewTextSignatureOrder([]byte("msg")),
		sdk.NewCoins(sdk.NewInt64Coin("uband", 100)),
		bandtesting.FeePayer.Address.String(),
	)
	s.Require().NoError(err)

	_, err = msgSrvr.RequestSignature(ctx, msg)
	s.Require().NoError(err)

	balancesAfter := s.app.BankKeeper.GetAllBalances(ctx, bandtesting.FeePayer.Address)
	balancesModuleAfter := s.app.BankKeeper.GetAllBalances(
		ctx,
		s.app.BandtssKeeper.GetBandtssAccount(ctx).GetAddress(),
	)

	group1 := s.app.TSSKeeper.MustGetGroup(ctx, group1Ctx.GroupID)
	diff := s.app.BandtssKeeper.GetParams(ctx).FeePerSigner.MulInt(math.NewInt(int64(group1.Threshold)))
	s.Require().Equal(diff, balancesBefore.Sub(balancesAfter...))
	s.Require().Equal(diff, balancesModuleAfter.Sub(balancesModuleBefore...))

	bandtssSigningID := types.SigningID(s.app.BandtssKeeper.GetSigningCount(ctx))
	tssSigningID := tss.SigningID(s.app.TSSKeeper.GetSigningCount(ctx))

	bandtssSigning, err := s.app.BandtssKeeper.GetSigning(ctx, bandtssSigningID)
	s.Require().NoError(err)
	s.Require().Equal(tssSigningID-1, bandtssSigning.CurrentGroupSigningID)
	s.Require().Equal(tssSigningID, bandtssSigning.IncomingGroupSigningID)

	bandtssSigningIDMapping := s.app.BandtssKeeper.GetSigningIDMapping(ctx, tssSigningID-1)
	s.Require().Equal(bandtssSigningID, bandtssSigningIDMapping)
	bandtssSigningIDMapping = s.app.BandtssKeeper.GetSigningIDMapping(ctx, tssSigningID)
	s.Require().Equal(bandtssSigningID, bandtssSigningIDMapping)
}

func (s *AppTestSuite) TestActivateReq() {
	ctx, msgSrvr := s.ctx, s.msgSrvr
	groupCtx := s.SetupNewGroup(5, 3)

	for _, acc := range groupCtx.Accounts {
		err := s.app.BandtssKeeper.DeactivateMember(ctx, acc.Address, groupCtx.GroupID)
		s.Require().NoError(err)
	}

	// skip time frame.
	inactivePenaltyDuration := s.app.BandtssKeeper.GetParams(ctx).InactivePenaltyDuration
	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(inactivePenaltyDuration))

	for _, acc := range groupCtx.Accounts {
		_, err := msgSrvr.Activate(ctx, &types.MsgActivate{
			Sender:  acc.Address.String(),
			GroupID: groupCtx.GroupID,
		})
		s.Require().NoError(err)
	}
}

func (s *AppTestSuite) TestFailActivateNotPassDuration() {
	ctx, msgSrvr := s.ctx, s.msgSrvr
	groupCtx := s.SetupNewGroup(5, 3)

	for _, acc := range groupCtx.Accounts {
		err := s.app.BandtssKeeper.DeactivateMember(ctx, acc.Address, groupCtx.GroupID)
		s.Require().NoError(err)
	}

	for _, acc := range groupCtx.Accounts {
		_, err := msgSrvr.Activate(ctx, &types.MsgActivate{
			Sender:  acc.Address.String(),
			GroupID: groupCtx.GroupID,
		})
		s.Require().ErrorIs(err, types.ErrPenaltyDurationNotElapsed)
	}
}

func (s *AppTestSuite) TestFailActivateIncorrectGroupID() {
	ctx, msgSrvr := s.ctx, s.msgSrvr
	groupCtx := s.SetupNewGroup(5, 3)

	for _, acc := range groupCtx.Accounts {
		err := s.app.BandtssKeeper.DeactivateMember(ctx, acc.Address, groupCtx.GroupID)
		s.Require().NoError(err)
	}

	// skip time frame.
	inactivePenaltyDuration := s.app.BandtssKeeper.GetParams(ctx).InactivePenaltyDuration
	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(inactivePenaltyDuration))

	for _, acc := range groupCtx.Accounts {
		_, err := msgSrvr.Activate(ctx, &types.MsgActivate{
			Sender:  acc.Address.String(),
			GroupID: tss.GroupID(300),
		})
		s.Require().ErrorIs(err, types.ErrMemberNotFound)
	}
}

func (s *AppTestSuite) TestFailActivateMemberIsActive() {
	ctx, msgSrvr := s.ctx, s.msgSrvr
	groupCtx := s.SetupNewGroup(5, 3)

	for _, acc := range groupCtx.Accounts {
		_, err := msgSrvr.Activate(ctx, &types.MsgActivate{
			Sender:  acc.Address.String(),
			GroupID: groupCtx.GroupID,
		})
		s.Require().ErrorIs(err, types.ErrMemberAlreadyActive)
	}
}

func (s *AppTestSuite) TestUpdateParams() {
	k, msgSrvr := s.app.TSSKeeper, s.msgSrvr

	testCases := []struct {
		name         string
		request      *types.MsgUpdateParams
		expectErr    bool
		expectErrStr string
	}{
		{
			name: "set invalid authority",
			request: &types.MsgUpdateParams{
				Authority: "foo",
			},
			expectErr:    true,
			expectErrStr: "invalid authority;",
		},
		{
			name: "set full valid params",
			request: &types.MsgUpdateParams{
				Authority: k.GetAuthority(),
				Params: types.Params{
					RewardPercentage:        types.DefaultRewardPercentage,
					InactivePenaltyDuration: types.DefaultInactivePenaltyDuration,
					MinTransitionDuration:   types.DefaultMinTransitionDuration,
					MaxTransitionDuration:   types.DefaultMaxTransitionDuration,
					FeePerSigner:            types.DefaultFeePerSigner,
				},
			},
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			_, err := msgSrvr.UpdateParams(s.ctx, tc.request)
			if tc.expectErr {
				s.Require().Error(err)
				s.Require().ErrorContains(err, tc.expectErrStr)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}
