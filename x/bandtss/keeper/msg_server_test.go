package keeper_test

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/pkg/tss/testutil"
	bandtesting "github.com/bandprotocol/chain/v2/testing"
	"github.com/bandprotocol/chain/v2/x/bandtss/types"
	tsstypes "github.com/bandprotocol/chain/v2/x/tss/types"
)

type TestCase struct {
	Msg         string
	Malleate    func()
	PostTest    func()
	ExpectedErr error
}

func (s *KeeperTestSuite) TestFlowCreatingGroup() {
	s.SetupGroup(tsstypes.GROUP_STATUS_ACTIVE)
}

func (s *KeeperTestSuite) TestSuccessCreateGroup() {
	ctx, k := s.ctx, s.app.BandtssKeeper

	_, err := s.msgSrvr.CreateGroup(ctx, &types.MsgCreateGroup{
		Members:   []string{"band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs"},
		Threshold: 1,
		Authority: s.authority.String(),
	})

	s.Require().NoError(err)

	// Check if the group is created but not set as active.
	s.Require().Equal(tss.GroupID(0), k.GetCurrentGroupID(ctx))

	group, err := s.app.TSSKeeper.GetGroup(ctx, tss.GroupID(1))
	s.Require().NoError(err)
	s.Require().Equal(tsstypes.GROUP_STATUS_ROUND_1, group.Status)
	s.Require().Equal(types.ModuleName, group.ModuleOwner)
	s.Require().Equal(uint64(1), group.Threshold)
	s.Require().Equal(uint64(1), group.Size_)

	// check if the member is added into the group.
	actualMembers, err := s.app.TSSKeeper.GetGroupMembers(ctx, tss.GroupID(1))
	s.Require().NoError(err)
	s.Require().Equal(1, len(actualMembers))
	s.Require().Equal("band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs", actualMembers[0].Address)
}

func (s *KeeperTestSuite) TestSuccessCreateNewGroupAfterHavingCurrentGroup() {
	ctx, k := s.ctx, s.app.BandtssKeeper

	// provided that group is already created
	s.SetupGroup(tsstypes.GROUP_STATUS_ACTIVE)

	s.Require().Equal(tss.GroupID(1), s.app.BandtssKeeper.GetCurrentGroupID(s.ctx))

	members, err := s.app.TSSKeeper.GetGroupMembers(ctx, tss.GroupID(1))
	s.Require().NoError(err)

	// even member in current group is deactivated, it should not affect the new group.
	err = k.DeactivateMember(ctx, sdk.MustAccAddressFromBech32(members[0].Address))
	s.Require().NoError(err)

	expectedGroupID := tss.GroupID(s.app.TSSKeeper.GetGroupCount(ctx) + 1)

	_, err = s.msgSrvr.CreateGroup(ctx, &types.MsgCreateGroup{
		Members:   []string{sdk.MustAccAddressFromBech32(members[0].Address).String()},
		Threshold: 1,
		Authority: s.authority.String(),
	})
	s.Require().NoError(err)

	s.Require().Equal(uint64(expectedGroupID), s.app.TSSKeeper.GetGroupCount(ctx))

	// Check if the group is created but not impact current group ID in bandtss.
	s.Require().Equal(tss.GroupID(1), k.GetCurrentGroupID(ctx))

	group, err := s.app.TSSKeeper.GetGroup(ctx, expectedGroupID)
	s.Require().NoError(err)
	s.Require().Equal(tsstypes.GROUP_STATUS_ROUND_1, group.Status)
	s.Require().Equal(types.ModuleName, group.ModuleOwner)
	s.Require().Equal(uint64(1), group.Threshold)
	s.Require().Equal(uint64(1), group.Size_)

	// check if the member is added into the group.
	actualMembers, err := s.app.TSSKeeper.GetGroupMembers(ctx, expectedGroupID)
	s.Require().NoError(err)
	s.Require().Equal(1, len(actualMembers))
	s.Require().Equal(members[0].Address, actualMembers[0].Address)
}

func (s *KeeperTestSuite) TestFailCreateGroup() {
	ctx := s.ctx
	tssParams := s.app.TSSKeeper.GetParams(ctx)

	testCases := []struct {
		name        string
		input       *types.MsgCreateGroup
		preProcess  func()
		expectErr   error
		postProcess func()
	}{
		{
			name: "invalid authority",
			input: &types.MsgCreateGroup{
				Members:   []string{"band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs"},
				Threshold: 1,
				Authority: "band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs",
			},
			preProcess:  func() {},
			postProcess: func() {},
			expectErr:   govtypes.ErrInvalidSigner,
		},
		{
			name: "over max group size",
			input: &types.MsgCreateGroup{
				Members:   []string{"band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs", bandtesting.Alice.Address.String()},
				Threshold: 1,
				Authority: s.authority.String(),
			},
			preProcess: func() {
				err := s.app.TSSKeeper.SetParams(ctx, tsstypes.Params{
					MaxGroupSize:   1,
					MaxDESize:      tssParams.MaxDESize,
					CreatingPeriod: tssParams.CreatingPeriod,
					SigningPeriod:  tssParams.SigningPeriod,
				})
				s.Require().NoError(err)
			},
			postProcess: func() {
				err := s.app.TSSKeeper.SetParams(ctx, tssParams)
				s.Require().NoError(err)
			},
			expectErr: tsstypes.ErrGroupSizeTooLarge,
		},
		{
			name: "duplicate members",
			input: &types.MsgCreateGroup{
				Members:   []string{"band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs", "band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs", "band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs"},
				Threshold: 1,
				Authority: s.authority.String(),
			},
			preProcess:  func() {},
			postProcess: func() {},
			expectErr:   fmt.Errorf("duplicated member found within the list"),
		},
		{
			name: "threshold more than members length",
			input: &types.MsgCreateGroup{
				Members:   []string{"band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs"},
				Threshold: 10,
				Authority: s.authority.String(),
			},
			preProcess:  func() {},
			postProcess: func() {},
			expectErr:   fmt.Errorf("threshold must be less than or equal to the members but more than zero"),
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			tc.preProcess()

			err := tc.input.ValidateBasic()
			if err != nil {
				s.Require().ErrorContains(err, tc.expectErr.Error())
			} else {
				_, err := s.msgSrvr.CreateGroup(ctx, tc.input)
				s.Require().ErrorIs(err, tc.expectErr)
			}

			tc.postProcess()
		})
	}
}

func (s *KeeperTestSuite) TestFailedReplaceGroup() {
	ctx, msgSrvr, k := s.ctx, s.msgSrvr, s.app.TSSKeeper
	authority := authtypes.NewModuleAddress(govtypes.ModuleName)
	newGroupID := tss.GroupID(2)

	var req types.MsgReplaceGroup

	s.SetupGroup(tsstypes.GROUP_STATUS_ACTIVE)
	currentGroupID := s.app.BandtssKeeper.GetCurrentGroupID(ctx)
	currentGroup := k.MustGetGroup(ctx, currentGroupID)

	tcs := []TestCase{
		{
			"failure due to incorrect authority",
			func() {
				req = types.MsgReplaceGroup{
					Authority:  "band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs",
					NewGroupID: newGroupID,
					ExecTime:   time.Now().UTC(),
				}
			},
			func() {
			},
			govtypes.ErrInvalidSigner,
		},
		{
			"failure due to group is not active",
			func() {
				req = types.MsgReplaceGroup{
					Authority:  authority.String(),
					NewGroupID: newGroupID,
					ExecTime:   time.Now().UTC(),
				}
				currentGroup.Status = tsstypes.GROUP_STATUS_FALLEN
				k.SetGroup(ctx, currentGroup)
			},
			func() {
				currentGroup.Status = tsstypes.GROUP_STATUS_ACTIVE
				k.SetGroup(ctx, currentGroup)
			},
			tsstypes.ErrGroupIsNotActive,
		},
	}

	for _, tc := range tcs {
		s.Run(fmt.Sprintf("Case %s", tc.Msg), func() {
			tc.Malleate()

			_, err := msgSrvr.ReplaceGroup(ctx, &req)
			s.Require().ErrorIs(tc.ExpectedErr, err)

			tc.PostTest()
		})
	}
}

func (s *KeeperTestSuite) TestSuccessReplaceGroup() {
	ctx, msgSrvr, _ := s.ctx, s.msgSrvr, s.app.TSSKeeper

	s.SetupGroup(tsstypes.GROUP_STATUS_ACTIVE)

	now := time.Now()

	_, err := msgSrvr.ReplaceGroup(ctx, &types.MsgReplaceGroup{
		NewGroupID: 2,
		ExecTime:   now,
		Authority:  s.authority.String(),
	})

	s.Require().NoError(err)
	replacement_status := s.app.BandtssKeeper.GetReplacement(ctx).Status
	s.Require().Equal(types.REPLACEMENT_STATUS_WAITING_SIGN, replacement_status)
}

func (s *KeeperTestSuite) TestFailedRequestSignatureReq() {
	ctx, msgSrvr, k := s.ctx, s.msgSrvr, s.app.TSSKeeper

	s.SetupGroup(tsstypes.GROUP_STATUS_ACTIVE)

	var req *types.MsgRequestSignature
	var err error

	tcs := []TestCase{
		{
			"failure with no groupID",
			func() {
				s.app.BandtssKeeper.SetCurrentGroupID(ctx, 0)
				req, err = types.NewMsgRequestSignature(
					tsstypes.NewTextSignatureOrder([]byte("msg")),
					sdk.NewCoins(sdk.NewInt64Coin("uband", 100)),
					bandtesting.FeePayer.Address,
				)
				s.Require().NoError(err)
			},
			func() {
				s.app.BandtssKeeper.SetCurrentGroupID(ctx, 1)
			},
			types.ErrNoActiveGroup,
		},
		{
			"failure with inactive group",
			func() {
				inactiveGroup := tsstypes.Group{
					ID:        2,
					Size_:     5,
					Threshold: 3,
					PubKey:    nil,
					Status:    tsstypes.GROUP_STATUS_FALLEN,
				}
				k.SetGroup(ctx, inactiveGroup)
				s.app.BandtssKeeper.SetCurrentGroupID(ctx, 2)

				req, err = types.NewMsgRequestSignature(
					tsstypes.NewTextSignatureOrder([]byte("msg")),
					sdk.NewCoins(sdk.NewInt64Coin("uband", 100)),
					bandtesting.FeePayer.Address,
				)
				s.Require().NoError(err)
			},
			func() {
				s.app.BandtssKeeper.SetCurrentGroupID(ctx, 1)
			},
			types.ErrNoActiveGroup,
		},
		{
			"failure with fee is more than user's limit",
			func() {
				req, err = types.NewMsgRequestSignature(
					tsstypes.NewTextSignatureOrder([]byte("msg")),
					sdk.NewCoins(sdk.NewInt64Coin("uband", 10)),
					bandtesting.FeePayer.Address,
				)
			},
			func() {},
			types.ErrFeeExceedsLimit,
		},
	}

	for _, tc := range tcs {
		s.Run(fmt.Sprintf("Case %s", tc.Msg), func() {
			tc.Malleate()
			defer tc.PostTest()

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

func (s *KeeperTestSuite) TestSuccessRequestSignatureReq() {
	ctx, msgSrvr, k := s.ctx, s.msgSrvr, s.app.BandtssKeeper

	s.SetupGroup(tsstypes.GROUP_STATUS_ACTIVE)

	for _, tc := range testutil.TestCases {
		// Request signature for each member in the group
		s.Run(fmt.Sprintf("success %s", tc.Name), func() {
			for _, signing := range tc.Signings {
				k.SetCurrentGroupID(ctx, tc.Group.ID)

				balancesBefore := s.app.BankKeeper.GetAllBalances(ctx, bandtesting.FeePayer.Address)
				balancesModuleBefore := s.app.BankKeeper.GetAllBalances(
					ctx,
					s.app.BandtssKeeper.GetBandtssAccount(ctx).GetAddress(),
				)

				msg, err := types.NewMsgRequestSignature(
					tsstypes.NewTextSignatureOrder(signing.Data),
					sdk.NewCoins(sdk.NewInt64Coin("uband", 100)),
					bandtesting.FeePayer.Address,
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

				diff := k.GetParams(ctx).Fee.MulInt(sdk.NewInt(int64(len(signing.AssignedMembers))))
				s.Require().Equal(diff, balancesBefore.Sub(balancesAfter...))
				s.Require().Equal(diff, balancesModuleAfter.Sub(balancesModuleBefore...))
			}
		})
	}
}

func (s *KeeperTestSuite) TestActivateReq() {
	ctx, msgSrvr := s.ctx, s.msgSrvr
	s.SetupGroup(tsstypes.GROUP_STATUS_ACTIVE)

	for _, tc := range testutil.TestCases {
		s.Run(fmt.Sprintf("success %s", tc.Name), func() {
			s.app.BandtssKeeper.SetCurrentGroupID(ctx, tc.Group.ID)

			for _, m := range tc.Group.Members {
				addr := sdk.AccAddress(m.PubKey())
				existed := s.app.BandtssKeeper.HasMember(ctx, addr)
				if !existed {
					err := s.app.BandtssKeeper.AddNewMember(ctx, addr)
					s.Require().NoError(err)
				}

				existedMember, err := s.app.BandtssKeeper.GetMember(ctx, addr)
				s.Require().NoError(err)
				if existedMember.IsActive {
					err := s.app.BandtssKeeper.DeactivateMember(ctx, addr)
					s.Require().NoError(err)
				}
			}

			// skip time frame.
			activeDuration := s.app.BandtssKeeper.GetParams(ctx).ActiveDuration
			ctx = ctx.WithBlockTime(ctx.BlockTime().Add(activeDuration))

			for _, m := range tc.Group.Members {
				member, err := s.app.BandtssKeeper.GetMember(ctx, sdk.AccAddress(m.PubKey()))
				s.Require().NoError(err)
				// There are some test cases in which the members are using the same private key.
				if member.IsActive {
					continue
				}

				_, err = msgSrvr.Activate(ctx, &types.MsgActivate{
					Address: sdk.AccAddress(m.PubKey()).String(),
				})
				s.Require().NoError(err)
			}

			for _, m := range tc.Group.Members {
				s.app.BandtssKeeper.DeleteMember(ctx, sdk.AccAddress(m.PubKey()))
			}
		})
	}
}

func (s *KeeperTestSuite) TestHealthCheckReq() {
	ctx, msgSrvr := s.ctx, s.msgSrvr
	s.SetupGroup(tsstypes.GROUP_STATUS_ACTIVE)

	for _, tc := range testutil.TestCases {
		s.Run(fmt.Sprintf("success %s", tc.Name), func() {
			s.app.BandtssKeeper.SetCurrentGroupID(ctx, tc.Group.ID)
			for _, m := range tc.Group.Members {
				addr := sdk.AccAddress(m.PubKey())
				existed := s.app.BandtssKeeper.HasMember(ctx, addr)
				if !existed {
					err := s.app.BandtssKeeper.AddNewMember(ctx, addr)
					s.Require().NoError(err)
				}

				_, err := msgSrvr.HealthCheck(ctx, &types.MsgHealthCheck{
					Address: sdk.AccAddress(m.PubKey()).String(),
				})
				s.Require().NoError(err)
			}
		})
	}
}

func (s *KeeperTestSuite) TestUpdateParams() {
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
					ActiveDuration:          types.DefaultActiveDuration,
					RewardPercentage:        types.DefaultRewardPercentage,
					InactivePenaltyDuration: types.DefaultInactivePenaltyDuration,
					Fee:                     types.DefaultFee,
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
