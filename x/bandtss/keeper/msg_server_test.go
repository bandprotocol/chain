package keeper_test

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/pkg/tss/testutil"
	"github.com/bandprotocol/chain/v2/testing/testapp"
	"github.com/bandprotocol/chain/v2/x/bandtss/types"
	tsstypes "github.com/bandprotocol/chain/v2/x/tss/types"
)

type TestCase struct {
	Msg         string
	Malleate    func()
	PostTest    func()
	ExpectedErr error
}

func (s *KeeperTestSuite) TestCreateGroup() {
	s.SetupGroup(tsstypes.GROUP_STATUS_ACTIVE)
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
	s.Require().Equal(types.REPLACEMENT_STATUS_WAITING_SIGNING, replacement_status)
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
					testapp.FeePayer.Address,
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
					testapp.FeePayer.Address,
				)
				s.Require().NoError(err)
			},
			func() {
				s.app.BandtssKeeper.SetCurrentGroupID(ctx, 1)
			},
			tsstypes.ErrGroupIsNotActive,
		},
		{
			"failure with not enough fee",
			func() {
				req, err = types.NewMsgRequestSignature(
					tsstypes.NewTextSignatureOrder([]byte("msg")),
					sdk.NewCoins(sdk.NewInt64Coin("uband", 10)),
					testapp.FeePayer.Address,
				)
			},
			func() {},
			types.ErrNotEnoughFee,
		},
	}

	for _, tc := range tcs {
		s.Run(fmt.Sprintf("Case %s", tc.Msg), func() {
			tc.Malleate()

			balancesBefore := s.app.BankKeeper.GetAllBalances(ctx, testapp.FeePayer.Address)
			balancesModuleBefore := s.app.BankKeeper.GetAllBalances(
				ctx,
				s.app.BandtssKeeper.GetBandtssAccount(ctx).GetAddress(),
			)

			_, err := msgSrvr.RequestSignature(ctx, req)
			s.Require().ErrorIs(tc.ExpectedErr, err)

			balancesAfter := s.app.BankKeeper.GetAllBalances(ctx, testapp.FeePayer.Address)
			balancesModuleAfter := s.app.BankKeeper.GetAllBalances(
				ctx,
				s.app.BandtssKeeper.GetBandtssAccount(ctx).GetAddress(),
			)

			// Check if the balances of payer and module account doesn't change
			s.Require().Equal(balancesBefore, balancesAfter)
			s.Require().Equal(balancesModuleBefore, balancesModuleAfter)

			tc.PostTest()
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

				balancesBefore := s.app.BankKeeper.GetAllBalances(ctx, testapp.FeePayer.Address)
				balancesModuleBefore := s.app.BankKeeper.GetAllBalances(
					ctx,
					s.app.BandtssKeeper.GetBandtssAccount(ctx).GetAddress(),
				)

				msg, err := types.NewMsgRequestSignature(
					tsstypes.NewTextSignatureOrder(signing.Data),
					sdk.NewCoins(sdk.NewInt64Coin("uband", 100)),
					testapp.FeePayer.Address,
				)
				s.Require().NoError(err)

				_, err = msgSrvr.RequestSignature(ctx, msg)
				s.Require().NoError(err)

				// Fee should be paid after requesting signature
				balancesAfter := s.app.BankKeeper.GetAllBalances(ctx, testapp.FeePayer.Address)
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
					JailPenaltyDuration:     types.DefaultJailPenaltyDuration,
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
