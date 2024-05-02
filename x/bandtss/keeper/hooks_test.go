package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/bandtss/testutil"
	"github.com/bandprotocol/chain/v2/x/bandtss/types"
	tsstypes "github.com/bandprotocol/chain/v2/x/tss/types"
)

func TestHookAfterSigningFailed(t *testing.T) {
	requestor := sdk.MustAccAddressFromBech32("band1t5x8hrmht463eq4m0xhfgz95h62dyvkq049eek")
	testCases := []struct {
		name       string
		input      tsstypes.Signing
		preProcess func(s *testutil.TestSuite)
		postCheck  func(s *testutil.TestSuite)
	}{
		{
			name: "signing currentGroup with fee 10ubands 2 members",
			input: tsstypes.Signing{
				ID:      1,
				GroupID: 1,
				AssignedMembers: []tsstypes.AssignedMember{
					{MemberID: 1},
					{MemberID: 2},
				},
			},
			preProcess: func(s *testutil.TestSuite) {
				s.Keeper.SetSigningIDMapping(s.Ctx, 1, 1)
				s.Keeper.SetSigning(s.Ctx, types.Signing{
					ID:                      1,
					Fee:                     sdk.NewCoins(sdk.NewInt64Coin("uband", 10)),
					Requester:               requestor.String(),
					CurrentGroupSigningID:   1,
					ReplacingGroupSigningID: 0,
				})
				s.MockBankKeeper.EXPECT().SendCoinsFromModuleToAccount(
					s.Ctx,
					types.ModuleName,
					requestor,
					sdk.NewCoins(sdk.NewInt64Coin("uband", 20)),
				)
			},
			postCheck: func(s *testutil.TestSuite) {
				require.Zero(s.T(), s.Keeper.GetSigningIDMapping(s.Ctx, 1))
			},
		},
		{
			name: "signing currentGroup with no fee 2 members",
			input: tsstypes.Signing{
				ID:      1,
				GroupID: 1,
				AssignedMembers: []tsstypes.AssignedMember{
					{MemberID: 1},
					{MemberID: 2},
				},
			},
			preProcess: func(s *testutil.TestSuite) {
				s.Keeper.SetSigningIDMapping(s.Ctx, 1, 1)
				s.Keeper.SetSigning(s.Ctx, types.Signing{
					ID:                      1,
					Fee:                     sdk.NewCoins(),
					Requester:               requestor.String(),
					CurrentGroupSigningID:   1,
					ReplacingGroupSigningID: 0,
				})
			},
			postCheck: func(s *testutil.TestSuite) {
				require.Zero(s.T(), s.Keeper.GetSigningIDMapping(s.Ctx, 1))
			},
		},
		{
			name: "signing currentGroup with fee 10ubands 3 members",
			input: tsstypes.Signing{
				ID:      1,
				GroupID: 1,
				AssignedMembers: []tsstypes.AssignedMember{
					{MemberID: 1},
					{MemberID: 2},
					{MemberID: 3},
				},
			},
			preProcess: func(s *testutil.TestSuite) {
				s.Keeper.SetSigningIDMapping(s.Ctx, 1, 1)
				s.Keeper.SetSigningIDMapping(s.Ctx, 2, 1)
				s.Keeper.SetSigning(s.Ctx, types.Signing{
					ID:                      1,
					Fee:                     sdk.NewCoins(sdk.NewInt64Coin("uband", 10)),
					Requester:               requestor.String(),
					CurrentGroupSigningID:   1,
					ReplacingGroupSigningID: 2,
				})
				s.MockBankKeeper.EXPECT().SendCoinsFromModuleToAccount(
					s.Ctx,
					types.ModuleName,
					requestor,
					sdk.NewCoins(sdk.NewInt64Coin("uband", 30)),
				)
			},
			postCheck: func(s *testutil.TestSuite) {
				require.Zero(s.T(), s.Keeper.GetSigningIDMapping(s.Ctx, 1))
				require.Equal(s.T(), types.SigningID(1), s.Keeper.GetSigningIDMapping(s.Ctx, 2))
			},
		},
		{
			name: "signing replacingGroup with fee 10ubands 3 members",
			input: tsstypes.Signing{
				ID:      2,
				GroupID: 2,
				AssignedMembers: []tsstypes.AssignedMember{
					{MemberID: 1},
					{MemberID: 2},
					{MemberID: 3},
				},
			},
			preProcess: func(s *testutil.TestSuite) {
				s.Keeper.SetSigningIDMapping(s.Ctx, 2, 1)
				s.Keeper.SetSigningIDMapping(s.Ctx, 1, 1)
				s.Keeper.SetSigning(s.Ctx, types.Signing{
					ID:                      1,
					Fee:                     sdk.NewCoins(sdk.NewInt64Coin("uband", 10)),
					Requester:               requestor.String(),
					CurrentGroupSigningID:   1,
					ReplacingGroupSigningID: 2,
				})
			},
			postCheck: func(s *testutil.TestSuite) {
				require.Zero(s.T(), s.Keeper.GetSigningIDMapping(s.Ctx, 2))
				require.Equal(s.T(), types.SigningID(1), s.Keeper.GetSigningIDMapping(s.Ctx, 1))
			},
		},
		{
			name: "no signingID mapping",
			input: tsstypes.Signing{
				ID:      2,
				GroupID: 1,
				AssignedMembers: []tsstypes.AssignedMember{
					{MemberID: 1},
					{MemberID: 2},
					{MemberID: 3},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := testutil.NewTestSuite(t)

			if tc.preProcess != nil {
				tc.preProcess(&s)
			}

			err := s.Hook.AfterSigningFailed(s.Ctx, tc.input)
			require.NoError(t, err)

			if tc.postCheck != nil {
				tc.postCheck(&s)
			}
		})
	}
}

func TestHookBeforeSigningExpired(t *testing.T) {
	requestor := sdk.MustAccAddressFromBech32("band1t5x8hrmht463eq4m0xhfgz95h62dyvkq049eek")
	penalizedMembers := []sdk.AccAddress{
		sdk.MustAccAddressFromBech32("band1t5x8hrmht463eq4m0xhfgz95h62dyvkq049eek"),
	}

	testCases := []struct {
		name       string
		input      tsstypes.Signing
		preProcess func(s *testutil.TestSuite)
		postCheck  func(s *testutil.TestSuite)
	}{
		{
			name: "signing currentGroup with fee 10ubands 2 members",
			input: tsstypes.Signing{
				ID:      1,
				GroupID: 1,
				AssignedMembers: []tsstypes.AssignedMember{
					{MemberID: 1},
					{MemberID: 2},
				},
			},
			preProcess: func(s *testutil.TestSuite) {
				s.Keeper.SetCurrentGroupID(s.Ctx, 1)
				s.Keeper.SetSigningIDMapping(s.Ctx, 1, 1)
				s.Keeper.SetMember(s.Ctx, types.Member{
					Address:    penalizedMembers[0].String(),
					IsActive:   true,
					Since:      s.Ctx.BlockTime(),
					LastActive: s.Ctx.BlockTime(),
				})

				s.Keeper.SetSigning(s.Ctx, types.Signing{
					ID:                      1,
					Fee:                     sdk.NewCoins(sdk.NewInt64Coin("uband", 10)),
					Requester:               requestor.String(),
					CurrentGroupSigningID:   1,
					ReplacingGroupSigningID: 0,
				})

				s.MockTSSKeeper.EXPECT().GetPenalizedMembersExpiredSigning(
					s.Ctx,
					tsstypes.Signing{
						ID:      1,
						GroupID: 1,
						AssignedMembers: []tsstypes.AssignedMember{
							{MemberID: 1},
							{MemberID: 2},
						},
					},
				).Return(penalizedMembers, nil)

				s.MockTSSKeeper.EXPECT().DeactivateMember(s.Ctx, tss.GroupID(1), penalizedMembers[0]).Return(nil)
				s.MockBankKeeper.EXPECT().SendCoinsFromModuleToAccount(
					s.Ctx,
					types.ModuleName,
					requestor,
					sdk.NewCoins(sdk.NewInt64Coin("uband", 20)),
				)
			},
			postCheck: func(s *testutil.TestSuite) {
				require.Zero(s.T(), s.Keeper.GetSigningIDMapping(s.Ctx, 1))

				member, err := s.Keeper.GetMember(s.Ctx, penalizedMembers[0])
				require.NoError(s.T(), err)
				require.False(s.T(), member.IsActive)
			},
		},
		{
			name: "signing currentGroup with no fee 2 members",
			input: tsstypes.Signing{
				ID:      1,
				GroupID: 1,
				AssignedMembers: []tsstypes.AssignedMember{
					{MemberID: 1},
					{MemberID: 2},
				},
			},
			preProcess: func(s *testutil.TestSuite) {
				s.Keeper.SetCurrentGroupID(s.Ctx, 1)
				s.Keeper.SetSigningIDMapping(s.Ctx, 1, 1)
				s.Keeper.SetMember(s.Ctx, types.Member{
					Address:    penalizedMembers[0].String(),
					IsActive:   true,
					Since:      s.Ctx.BlockTime(),
					LastActive: s.Ctx.BlockTime(),
				})

				s.Keeper.SetSigning(s.Ctx, types.Signing{
					ID:                      1,
					Fee:                     sdk.NewCoins(),
					Requester:               requestor.String(),
					CurrentGroupSigningID:   1,
					ReplacingGroupSigningID: 0,
				})

				s.MockTSSKeeper.EXPECT().GetPenalizedMembersExpiredSigning(
					s.Ctx,
					tsstypes.Signing{
						ID:      1,
						GroupID: 1,
						AssignedMembers: []tsstypes.AssignedMember{
							{MemberID: 1},
							{MemberID: 2},
						},
					},
				).Return(penalizedMembers, nil)

				s.MockTSSKeeper.EXPECT().DeactivateMember(s.Ctx, tss.GroupID(1), penalizedMembers[0]).Return(nil)
			},
			postCheck: func(s *testutil.TestSuite) {
				require.Zero(s.T(), s.Keeper.GetSigningIDMapping(s.Ctx, 1))

				member, err := s.Keeper.GetMember(s.Ctx, penalizedMembers[0])
				require.NoError(s.T(), err)
				require.False(s.T(), member.IsActive)
			},
		},
		{
			name: "signing currentGroup with fee 10ubands 3 members",
			input: tsstypes.Signing{
				ID:      1,
				GroupID: 1,
				AssignedMembers: []tsstypes.AssignedMember{
					{MemberID: 1},
					{MemberID: 2},
					{MemberID: 3},
				},
			},
			preProcess: func(s *testutil.TestSuite) {
				s.Keeper.SetCurrentGroupID(s.Ctx, 1)
				s.Keeper.SetSigningIDMapping(s.Ctx, 1, 1)
				s.Keeper.SetMember(s.Ctx, types.Member{
					Address:    penalizedMembers[0].String(),
					IsActive:   true,
					Since:      s.Ctx.BlockTime(),
					LastActive: s.Ctx.BlockTime(),
				})

				s.Keeper.SetSigning(s.Ctx, types.Signing{
					ID:                      1,
					Fee:                     sdk.NewCoins(sdk.NewInt64Coin("uband", 10)),
					Requester:               requestor.String(),
					CurrentGroupSigningID:   1,
					ReplacingGroupSigningID: 0,
				})

				s.MockTSSKeeper.EXPECT().GetPenalizedMembersExpiredSigning(
					s.Ctx,
					tsstypes.Signing{
						ID:      1,
						GroupID: 1,
						AssignedMembers: []tsstypes.AssignedMember{
							{MemberID: 1},
							{MemberID: 2},
							{MemberID: 3},
						},
					},
				).Return(penalizedMembers, nil)

				s.MockTSSKeeper.EXPECT().DeactivateMember(s.Ctx, tss.GroupID(1), penalizedMembers[0]).Return(nil)
				s.MockBankKeeper.EXPECT().SendCoinsFromModuleToAccount(
					s.Ctx,
					types.ModuleName,
					requestor,
					sdk.NewCoins(sdk.NewInt64Coin("uband", 30)),
				)
			},
			postCheck: func(s *testutil.TestSuite) {
				require.Zero(s.T(), s.Keeper.GetSigningIDMapping(s.Ctx, 1))

				member, err := s.Keeper.GetMember(s.Ctx, penalizedMembers[0])
				require.NoError(s.T(), err)
				require.False(s.T(), member.IsActive)
			},
		},
		{
			name: "signing replacingGroup with fee 10ubands 2 members",
			input: tsstypes.Signing{
				ID:      2,
				GroupID: 2,
				AssignedMembers: []tsstypes.AssignedMember{
					{MemberID: 1},
					{MemberID: 2},
				},
			},
			preProcess: func(s *testutil.TestSuite) {
				s.Keeper.SetCurrentGroupID(s.Ctx, 1)
				s.Keeper.SetSigningIDMapping(s.Ctx, 1, 1)
				s.Keeper.SetSigningIDMapping(s.Ctx, 2, 1)
				s.Keeper.SetMember(s.Ctx, types.Member{
					Address:    penalizedMembers[0].String(),
					IsActive:   true,
					Since:      s.Ctx.BlockTime(),
					LastActive: s.Ctx.BlockTime(),
				})

				s.Keeper.SetSigning(s.Ctx, types.Signing{
					ID:                      1,
					Fee:                     sdk.NewCoins(sdk.NewInt64Coin("uband", 10)),
					Requester:               requestor.String(),
					CurrentGroupSigningID:   1,
					ReplacingGroupSigningID: 2,
				})
			},
			postCheck: func(s *testutil.TestSuite) {
				require.Zero(s.T(), s.Keeper.GetSigningIDMapping(s.Ctx, 2))
				require.Equal(s.T(), types.SigningID(1), s.Keeper.GetSigningIDMapping(s.Ctx, 1))

				member, err := s.Keeper.GetMember(s.Ctx, penalizedMembers[0])
				require.NoError(s.T(), err)
				require.True(s.T(), member.IsActive)
			},
		},
		{
			name: "signing old currentGroup with fee 10ubands 2 members",
			input: tsstypes.Signing{
				ID:      1,
				GroupID: 1,
				AssignedMembers: []tsstypes.AssignedMember{
					{MemberID: 1},
					{MemberID: 2},
				},
			},
			preProcess: func(s *testutil.TestSuite) {
				s.Keeper.SetCurrentGroupID(s.Ctx, 2)
				s.Keeper.SetSigningIDMapping(s.Ctx, 1, 1)
				s.Keeper.SetSigningIDMapping(s.Ctx, 2, 1)
				s.Keeper.SetMember(s.Ctx, types.Member{
					Address:    penalizedMembers[0].String(),
					IsActive:   true,
					Since:      s.Ctx.BlockTime(),
					LastActive: s.Ctx.BlockTime(),
				})

				s.Keeper.SetSigning(s.Ctx, types.Signing{
					ID:                      1,
					Fee:                     sdk.NewCoins(sdk.NewInt64Coin("uband", 10)),
					Requester:               requestor.String(),
					CurrentGroupSigningID:   1,
					ReplacingGroupSigningID: 2,
				})
				s.MockBankKeeper.EXPECT().SendCoinsFromModuleToAccount(
					s.Ctx,
					types.ModuleName,
					requestor,
					sdk.NewCoins(sdk.NewInt64Coin("uband", 20)),
				)
			},
			postCheck: func(s *testutil.TestSuite) {
				require.Zero(s.T(), s.Keeper.GetSigningIDMapping(s.Ctx, 1))
				require.Equal(s.T(), types.SigningID(1), s.Keeper.GetSigningIDMapping(s.Ctx, 2))

				member, err := s.Keeper.GetMember(s.Ctx, penalizedMembers[0])
				require.NoError(s.T(), err)
				require.True(s.T(), member.IsActive)
			},
		},
		{
			name: "signing no ID mapping",
			input: tsstypes.Signing{
				ID:      3,
				GroupID: 2,
				AssignedMembers: []tsstypes.AssignedMember{
					{MemberID: 1},
					{MemberID: 2},
				},
			},
			preProcess: func(s *testutil.TestSuite) {
				s.Keeper.SetCurrentGroupID(s.Ctx, 1)
				s.Keeper.SetSigningIDMapping(s.Ctx, 1, 1)
				s.Keeper.SetMember(s.Ctx, types.Member{
					Address:    penalizedMembers[0].String(),
					IsActive:   true,
					Since:      s.Ctx.BlockTime(),
					LastActive: s.Ctx.BlockTime(),
				})

				s.Keeper.SetSigning(s.Ctx, types.Signing{
					ID:                      1,
					Fee:                     sdk.NewCoins(sdk.NewInt64Coin("uband", 10)),
					Requester:               requestor.String(),
					CurrentGroupSigningID:   1,
					ReplacingGroupSigningID: 0,
				})
			},
			postCheck: func(s *testutil.TestSuite) {
				require.Equal(s.T(), types.SigningID(1), s.Keeper.GetSigningIDMapping(s.Ctx, 1))

				member, err := s.Keeper.GetMember(s.Ctx, penalizedMembers[0])
				require.NoError(s.T(), err)
				require.True(s.T(), member.IsActive)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := testutil.NewTestSuite(t)

			if tc.preProcess != nil {
				tc.preProcess(&s)
			}

			err := s.Hook.BeforeSetSigningExpired(s.Ctx, tc.input)
			require.NoError(t, err)

			if tc.postCheck != nil {
				tc.postCheck(&s)
			}
		})
	}
}

func TestHookAfterSigningComplete(t *testing.T) {
	requestor := sdk.MustAccAddressFromBech32("band1t5x8hrmht463eq4m0xhfgz95h62dyvkq049eek")

	testCases := []struct {
		name       string
		input      tsstypes.Signing
		preProcess func(s *testutil.TestSuite)
		postCheck  func(s *testutil.TestSuite)
	}{

		{
			name: "signing currentGroup with fee 10ubands 2 members",
			input: tsstypes.Signing{
				ID:      1,
				GroupID: 1,
				AssignedMembers: []tsstypes.AssignedMember{
					{MemberID: 1, Address: "band1t5x8hrmht463eq4m0xhfgz95h62dyvkq049eek"},
					{MemberID: 2, Address: "band1a22hgwm4tz8gj82y6zad3de2dcg5dpymtj20m5"},
				},
			},
			preProcess: func(s *testutil.TestSuite) {
				s.Keeper.SetSigningIDMapping(s.Ctx, 1, 1)
				s.Keeper.SetSigning(s.Ctx, types.Signing{
					ID:                      1,
					Fee:                     sdk.NewCoins(sdk.NewInt64Coin("uband", 10)),
					Requester:               requestor.String(),
					CurrentGroupSigningID:   1,
					ReplacingGroupSigningID: 0,
				})
				s.MockBankKeeper.EXPECT().SendCoinsFromModuleToAccount(
					s.Ctx,
					types.ModuleName,
					sdk.MustAccAddressFromBech32("band1t5x8hrmht463eq4m0xhfgz95h62dyvkq049eek"),
					sdk.NewCoins(sdk.NewInt64Coin("uband", 10)),
				)
				s.MockBankKeeper.EXPECT().SendCoinsFromModuleToAccount(
					s.Ctx,
					types.ModuleName,
					sdk.MustAccAddressFromBech32("band1a22hgwm4tz8gj82y6zad3de2dcg5dpymtj20m5"),
					sdk.NewCoins(sdk.NewInt64Coin("uband", 10)),
				)
			},
			postCheck: func(s *testutil.TestSuite) {
				require.Zero(s.T(), s.Keeper.GetSigningIDMapping(s.Ctx, 1))
			},
		},
		{
			name: "signing currentGroup with no fee 2 members",
			input: tsstypes.Signing{
				ID:      1,
				GroupID: 1,
				AssignedMembers: []tsstypes.AssignedMember{
					{MemberID: 1},
					{MemberID: 2},
				},
			},
			preProcess: func(s *testutil.TestSuite) {
				s.Keeper.SetSigningIDMapping(s.Ctx, 1, 1)
				s.Keeper.SetSigning(s.Ctx, types.Signing{
					ID:                      1,
					Fee:                     sdk.NewCoins(),
					Requester:               requestor.String(),
					CurrentGroupSigningID:   1,
					ReplacingGroupSigningID: 0,
				})
			},
			postCheck: func(s *testutil.TestSuite) {
				require.Zero(s.T(), s.Keeper.GetSigningIDMapping(s.Ctx, 1))
			},
		},
		{
			name: "signing replacingGroup with fee 10ubands 3 members",
			input: tsstypes.Signing{
				ID:      2,
				GroupID: 2,
				AssignedMembers: []tsstypes.AssignedMember{
					{MemberID: 1},
					{MemberID: 2},
					{MemberID: 3},
				},
			},
			preProcess: func(s *testutil.TestSuite) {
				s.Keeper.SetSigningIDMapping(s.Ctx, 2, 1)
				s.Keeper.SetSigningIDMapping(s.Ctx, 1, 1)
				s.Keeper.SetSigning(s.Ctx, types.Signing{
					ID:                      1,
					Fee:                     sdk.NewCoins(sdk.NewInt64Coin("uband", 10)),
					Requester:               requestor.String(),
					CurrentGroupSigningID:   1,
					ReplacingGroupSigningID: 2,
				})
			},
			postCheck: func(s *testutil.TestSuite) {
				require.Zero(s.T(), s.Keeper.GetSigningIDMapping(s.Ctx, 2))
				require.Equal(s.T(), types.SigningID(1), s.Keeper.GetSigningIDMapping(s.Ctx, 1))
			},
		},
		{
			name: "no signingID mapping",
			input: tsstypes.Signing{
				ID:      2,
				GroupID: 1,
				AssignedMembers: []tsstypes.AssignedMember{
					{MemberID: 1},
					{MemberID: 2},
					{MemberID: 3},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := testutil.NewTestSuite(t)

			if tc.preProcess != nil {
				tc.preProcess(&s)
			}

			err := s.Hook.AfterSigningCompleted(s.Ctx, tc.input)
			require.NoError(t, err)

			if tc.postCheck != nil {
				tc.postCheck(&s)
			}
		})
	}
}

func TestHookAfterCreatingGroupComplete(t *testing.T) {
	members := []sdk.AccAddress{
		sdk.MustAccAddressFromBech32("band1t5x8hrmht463eq4m0xhfgz95h62dyvkq049eek"),
		sdk.MustAccAddressFromBech32("band1a22hgwm4tz8gj82y6zad3de2dcg5dpymtj20m5"),
	}

	testCases := []struct {
		name       string
		input      tsstypes.Group
		preProcess func(s *testutil.TestSuite)
		postCheck  func(s *testutil.TestSuite)
	}{
		{
			name: "no currentGroup",
			input: tsstypes.Group{
				ID:          1,
				ModuleOwner: types.ModuleName,
				Status:      tsstypes.GROUP_STATUS_ACTIVE,
			},
			preProcess: func(s *testutil.TestSuite) {
				s.MockTSSKeeper.EXPECT().MustGetMembers(s.Ctx, tss.GroupID(1)).Return([]tsstypes.Member{
					{ID: 1, Address: members[0].String(), GroupID: 1},
					{ID: 2, Address: members[1].String(), GroupID: 1},
				})
			},
			postCheck: func(s *testutil.TestSuite) {
				require.Equal(t, tss.GroupID(1), s.Keeper.GetCurrentGroupID(s.Ctx))

				for _, member_addr := range members {
					member, err := s.Keeper.GetMember(s.Ctx, member_addr)

					require.NoError(s.T(), err)
					require.True(s.T(), member.IsActive)
					require.Equal(s.T(), s.Ctx.BlockTime(), member.Since)
					require.Equal(s.T(), s.Ctx.BlockTime(), member.LastActive)
				}
			},
		},
		{
			name: "already set currentGroup",
			input: tsstypes.Group{
				ID:          2,
				ModuleOwner: types.ModuleName,
				Status:      tsstypes.GROUP_STATUS_ACTIVE,
			},
			preProcess: func(s *testutil.TestSuite) {
				s.Keeper.SetCurrentGroupID(s.Ctx, 1)
			},
			postCheck: func(s *testutil.TestSuite) {
				require.Equal(t, tss.GroupID(1), s.Keeper.GetCurrentGroupID(s.Ctx))
			},
		},
		{
			name: "group from another module",
			input: tsstypes.Group{
				ID:          2,
				ModuleOwner: tsstypes.ModuleName,
				Status:      tsstypes.GROUP_STATUS_ACTIVE,
			},
			preProcess: func(s *testutil.TestSuite) {
				s.Keeper.SetCurrentGroupID(s.Ctx, 1)
			},
			postCheck: func(s *testutil.TestSuite) {
				require.Equal(t, tss.GroupID(1), s.Keeper.GetCurrentGroupID(s.Ctx))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := testutil.NewTestSuite(t)

			if tc.preProcess != nil {
				tc.preProcess(&s)
			}

			err := s.Hook.AfterCreatingGroupCompleted(s.Ctx, tc.input)
			require.NoError(t, err)

			if tc.postCheck != nil {
				tc.postCheck(&s)
			}
		})
	}

}
