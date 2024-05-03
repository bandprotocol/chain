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
	tsskeeper "github.com/bandprotocol/chain/v2/x/tss/keeper"
	tsstypes "github.com/bandprotocol/chain/v2/x/tss/types"
)

type TestCase struct {
	Msg         string
	Malleate    func()
	PostTest    func()
	ExpectedErr error
}

func (s *KeeperTestSuite) TestCreateGroupReq() {
	ctx, msgSrvr := s.ctx, s.msgSrvr
	tssMsgSrvr := tsskeeper.NewMsgServerImpl(s.app.TSSKeeper)

	members := []string{
		"band18gtd9xgw6z5fma06fxnhet7z2ctrqjm3z4k7ad",
		"band1s743ydr36t6p29jsmrxm064guklgthsn3t90ym",
		"band1p08slm6sv2vqy4j48hddkd6hpj8yp6vlw3pf8p",
		"band12jf07lcaj67mthsnklngv93qkeuphhmxst9mh8",
	}

	for _, m := range members {
		_, err := msgSrvr.Activate(ctx, &types.MsgActivate{
			Address: m,
		})
		s.Require().NoError(err)

		_, err = tssMsgSrvr.SubmitDEs(ctx, &tsstypes.MsgSubmitDEs{
			DEs: []tsstypes.DE{
				{
					PubD: testutil.HexDecode("dddd"),
					PubE: testutil.HexDecode("eeee"),
				},
			},
			Address: m,
		})
		s.Require().NoError(err)
	}

	s.Run("create group", func() {
		_, err := msgSrvr.CreateGroup(ctx, &types.MsgCreateGroup{
			Members:   members,
			Threshold: 3,
			Authority: s.authority.String(),
		})
		s.Require().NoError(err)
	})
}

func (s *KeeperTestSuite) TestFailedReplaceGroup() {
	ctx, msgSrvr, k := s.ctx, s.msgSrvr, s.app.TSSKeeper
	authority := authtypes.NewModuleAddress(govtypes.ModuleName)
	newGroupID := tss.GroupID(2)

	var req types.MsgReplaceGroup

	s.SetupGroup(tsstypes.GROUP_STATUS_ACTIVE)
	group := k.MustGetGroup(ctx, newGroupID)

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
				group.Status = tsstypes.GROUP_STATUS_FALLEN
				k.SetGroup(ctx, group)
			},
			func() {
				group.Status = tsstypes.GROUP_STATUS_ACTIVE
				k.SetGroup(ctx, group)
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
			"failure with invalid groupID",
			func() {
				req, err = types.NewMsgRequestSignature(
					tss.GroupID(999), // non-existent groupID
					tsstypes.NewTextSignatureOrder([]byte("msg")),
					sdk.NewCoins(sdk.NewInt64Coin("uband", 100)),
					bandtesting.FeePayer.Address,
				)
				s.Require().NoError(err)
			},
			func() {},
			tsstypes.ErrGroupNotFound,
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
				req, err = types.NewMsgRequestSignature(
					tss.GroupID(2), // inactive groupID
					tsstypes.NewTextSignatureOrder([]byte("msg")),
					sdk.NewCoins(sdk.NewInt64Coin("uband", 100)),
					bandtesting.FeePayer.Address,
				)
				s.Require().NoError(err)
			},
			func() {},
			tsstypes.ErrGroupIsNotActive,
		},
		{
			"failure with not enough fee",
			func() {
				req, err = types.NewMsgRequestSignature(
					tss.GroupID(1),
					tsstypes.NewTextSignatureOrder([]byte("msg")),
					sdk.NewCoins(sdk.NewInt64Coin("uband", 10)),
					bandtesting.FeePayer.Address,
				)
			},
			func() {},
			types.ErrNotEnoughFee,
		},
	}

	for _, tc := range tcs {
		s.Run(fmt.Sprintf("Case %s", tc.Msg), func() {
			tc.Malleate()

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
			s.Require().Equal(balancesAfter, balancesBefore)
			s.Require().Equal(balancesModuleAfter, balancesModuleBefore)

			tc.PostTest()
		})
	}
}

func (s *KeeperTestSuite) TestSuccessRequestSignatureReq() {
	ctx, msgSrvr := s.ctx, s.msgSrvr

	s.SetupGroup(tsstypes.GROUP_STATUS_ACTIVE)

	for _, tc := range testutil.TestCases {
		// Request signature for each member in the group
		s.Run(fmt.Sprintf("success %s", tc.Name), func() {
			for _, signing := range tc.Signings {
				balancesBefore := s.app.BankKeeper.GetAllBalances(ctx, bandtesting.FeePayer.Address)
				balancesModuleBefore := s.app.BankKeeper.GetAllBalances(
					ctx,
					s.app.BandtssKeeper.GetBandtssAccount(ctx).GetAddress(),
				)

				msg, err := types.NewMsgRequestSignature(
					tc.Group.ID,
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

				diff := sdk.NewCoins(sdk.NewInt64Coin("uband", int64(10*len(signing.AssignedMembers))))
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
			for _, m := range tc.Group.Members {
				_, err := msgSrvr.Activate(ctx, &types.MsgActivate{
					Address: sdk.AccAddress(m.PubKey()).String(),
				})
				s.Require().NoError(err)
			}
		})
	}
}

func (s *KeeperTestSuite) TestHealthCheckReq() {
	ctx, msgSrvr := s.ctx, s.msgSrvr
	s.SetupGroup(tsstypes.GROUP_STATUS_ACTIVE)

	for _, tc := range testutil.TestCases {
		s.Run(fmt.Sprintf("success %s", tc.Name), func() {
			for _, m := range tc.Group.Members {
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
