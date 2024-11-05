package keeper_test

import (
	"time"

	"go.uber.org/mock/gomock"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/pkg/tss"
	"github.com/bandprotocol/chain/v3/x/bandtss/types"
	tsstypes "github.com/bandprotocol/chain/v3/x/tss/types"
)

func (s *KeeperTestSuite) TestCallbackOnSignFailed() {
	requestor := sdk.MustAccAddressFromBech32("band1t5x8hrmht463eq4m0xhfgz95h62dyvkq049eek")
	testCases := []struct {
		name       string
		input      tss.SigningID
		preProcess func(s *KeeperTestSuite)
		postCheck  func(s *KeeperTestSuite)
	}{
		{
			name:  "signing_currentGroup",
			input: 1,
			preProcess: func(s *KeeperTestSuite) {
				s.keeper.SetSigningIDMapping(s.ctx, 1, 1)
				s.keeper.SetSigning(s.ctx, types.Signing{
					ID:                     1,
					FeePerSigner:           sdk.NewCoins(),
					Requester:              requestor.String(),
					CurrentGroupSigningID:  1,
					IncomingGroupSigningID: 0,
				})
				s.keeper.SetGroupTransition(s.ctx, types.GroupTransition{
					SigningID: tss.SigningID(3),
					Status:    types.TRANSITION_STATUS_WAITING_SIGN,
				})
			},
			postCheck: func(s *KeeperTestSuite) {
				s.Require().Zero(s.keeper.GetSigningIDMapping(s.ctx, 1))

				transition, found := s.keeper.GetGroupTransition(s.ctx)
				s.Require().True(found)
				s.Require().Equal(types.TRANSITION_STATUS_WAITING_SIGN, transition.Status)
			},
		},
		{
			name:  "signing incomingGroup",
			input: 2,
			preProcess: func(s *KeeperTestSuite) {
				s.keeper.SetSigningIDMapping(s.ctx, 2, 1)
				s.keeper.SetSigningIDMapping(s.ctx, 1, 1)
				s.keeper.SetSigning(s.ctx, types.Signing{
					ID:                     1,
					FeePerSigner:           sdk.NewCoins(),
					Requester:              requestor.String(),
					CurrentGroupSigningID:  1,
					IncomingGroupSigningID: 2,
				})
				s.keeper.SetGroupTransition(s.ctx, types.GroupTransition{
					SigningID: tss.SigningID(3),
					Status:    types.TRANSITION_STATUS_WAITING_SIGN,
				})
			},
			postCheck: func(s *KeeperTestSuite) {
				s.Require().Zero(s.keeper.GetSigningIDMapping(s.ctx, 2))
				s.Require().Equal(types.SigningID(1), s.keeper.GetSigningIDMapping(s.ctx, 1))

				transition, found := s.keeper.GetGroupTransition(s.ctx)
				s.Require().True(found)
				s.Require().Equal(types.TRANSITION_STATUS_WAITING_SIGN, transition.Status)
			},
		},
		{
			name:  "no signingID mapping",
			input: 4,
			preProcess: func(s *KeeperTestSuite) {
				s.keeper.SetSigningIDMapping(s.ctx, 2, 1)
				s.keeper.SetSigningIDMapping(s.ctx, 1, 1)
				s.keeper.SetSigning(s.ctx, types.Signing{
					ID:                     1,
					FeePerSigner:           sdk.NewCoins(),
					Requester:              requestor.String(),
					CurrentGroupSigningID:  1,
					IncomingGroupSigningID: 2,
				})
				s.keeper.SetGroupTransition(s.ctx, types.GroupTransition{
					SigningID: tss.SigningID(3),
					Status:    types.TRANSITION_STATUS_WAITING_SIGN,
				})
			},
			postCheck: func(s *KeeperTestSuite) {
				s.Require().Equal(types.SigningID(1), s.keeper.GetSigningIDMapping(s.ctx, 2))
				s.Require().Equal(types.SigningID(1), s.keeper.GetSigningIDMapping(s.ctx, 1))

				transition, found := s.keeper.GetGroupTransition(s.ctx)
				s.Require().True(found)
				s.Require().Equal(types.TRANSITION_STATUS_WAITING_SIGN, transition.Status)
			},
		},
		{
			name:  "signing on group transition message",
			input: 3,
			preProcess: func(s *KeeperTestSuite) {
				s.keeper.SetSigningIDMapping(s.ctx, 2, 1)
				s.keeper.SetSigningIDMapping(s.ctx, 1, 1)
				s.keeper.SetSigning(s.ctx, types.Signing{
					ID:                     1,
					FeePerSigner:           sdk.NewCoins(),
					Requester:              requestor.String(),
					CurrentGroupSigningID:  1,
					IncomingGroupSigningID: 2,
				})
				s.keeper.SetGroupTransition(s.ctx, types.GroupTransition{
					SigningID:      tss.SigningID(3),
					CurrentGroupID: tss.GroupID(1),
					Status:         types.TRANSITION_STATUS_WAITING_SIGN,
				})
			},
			postCheck: func(s *KeeperTestSuite) {
				s.Require().Equal(types.SigningID(1), s.keeper.GetSigningIDMapping(s.ctx, 2))
				s.Require().Equal(types.SigningID(1), s.keeper.GetSigningIDMapping(s.ctx, 1))

				_, found := s.keeper.GetGroupTransition(s.ctx)
				s.Require().False(found)
			},
		},
		{
			name:  "signing on group transition message; but transition undefined",
			input: 3,
			preProcess: func(s *KeeperTestSuite) {
				s.keeper.SetSigningIDMapping(s.ctx, 2, 1)
				s.keeper.SetSigningIDMapping(s.ctx, 1, 1)
				s.keeper.SetSigning(s.ctx, types.Signing{
					ID:                     1,
					FeePerSigner:           sdk.NewCoins(),
					Requester:              requestor.String(),
					CurrentGroupSigningID:  1,
					IncomingGroupSigningID: 2,
				})
			},
			postCheck: func(s *KeeperTestSuite) {
				s.Require().Equal(types.SigningID(1), s.keeper.GetSigningIDMapping(s.ctx, 2))
				s.Require().Equal(types.SigningID(1), s.keeper.GetSigningIDMapping(s.ctx, 1))

				_, found := s.keeper.GetGroupTransition(s.ctx)
				s.Require().False(found)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			if tc.preProcess != nil {
				tc.preProcess(s)
			}

			s.tssCallback.OnSigningFailed(s.ctx, tc.input)

			if tc.postCheck != nil {
				tc.postCheck(s)
			}
		})
	}
}

func (s *KeeperTestSuite) TestCallbackOnSignTimeout() {
	requestor := sdk.MustAccAddressFromBech32("band1t5x8hrmht463eq4m0xhfgz95h62dyvkq049eek")
	penalizedMembers := []sdk.AccAddress{
		sdk.MustAccAddressFromBech32("band1t5x8hrmht463eq4m0xhfgz95h62dyvkq049eek"),
	}

	type input struct {
		signingID   tss.SigningID
		idleMembers []sdk.AccAddress
	}

	testCases := []struct {
		name       string
		input      input
		preProcess func(s *KeeperTestSuite)
		postCheck  func(s *KeeperTestSuite)
	}{
		{
			name:  "signing currentGroup",
			input: input{1, penalizedMembers},
			preProcess: func(s *KeeperTestSuite) {
				s.tssKeeper.EXPECT().MustGetSigning(gomock.Any(), tss.SigningID(1)).
					Return(tsstypes.Signing{
						ID:      1,
						GroupID: 1,
					})
				s.keeper.SetCurrentGroup(s.ctx, types.NewCurrentGroup(1, s.ctx.BlockTime()))
				s.keeper.SetSigningIDMapping(s.ctx, 1, 1)
				s.keeper.SetMember(s.ctx, types.Member{
					Address:    penalizedMembers[0].String(),
					GroupID:    tss.GroupID(1),
					IsActive:   true,
					Since:      s.ctx.BlockTime(),
					LastActive: s.ctx.BlockTime(),
				})

				s.keeper.SetSigning(s.ctx, types.Signing{
					ID:                     1,
					FeePerSigner:           sdk.NewCoins(sdk.NewInt64Coin("uband", 10)),
					Requester:              requestor.String(),
					CurrentGroupSigningID:  1,
					IncomingGroupSigningID: 0,
				})
				s.keeper.SetGroupTransition(s.ctx, types.GroupTransition{
					SigningID:       tss.SigningID(3),
					Status:          types.TRANSITION_STATUS_WAITING_SIGN,
					CurrentGroupID:  tss.GroupID(1),
					IncomingGroupID: tss.GroupID(0),
				})
				s.tssKeeper.EXPECT().
					DeactivateMember(gomock.Any(), tss.GroupID(1), penalizedMembers[0]).
					Return(nil)
			},
			postCheck: func(s *KeeperTestSuite) {
				s.Require().Equal(types.SigningID(1), s.keeper.GetSigningIDMapping(s.ctx, 1))

				member, err := s.keeper.GetMember(s.ctx, penalizedMembers[0], tss.GroupID(1))
				s.Require().NoError(err)
				s.Require().False(member.IsActive)

				transition, found := s.keeper.GetGroupTransition(s.ctx)
				s.Require().True(found)
				s.Require().Equal(types.TRANSITION_STATUS_WAITING_SIGN, transition.Status)
			},
		},
		{
			name:  "signing incomingGroup",
			input: input{2, penalizedMembers},
			preProcess: func(s *KeeperTestSuite) {
				s.tssKeeper.EXPECT().MustGetSigning(gomock.Any(), tss.SigningID(2)).
					Return(tsstypes.Signing{
						ID:      2,
						GroupID: 2,
					})
				s.keeper.SetCurrentGroup(s.ctx, types.NewCurrentGroup(1, s.ctx.BlockTime()))
				s.keeper.SetSigningIDMapping(s.ctx, 1, 1)
				s.keeper.SetSigningIDMapping(s.ctx, 2, 1)
				s.keeper.SetMember(s.ctx, types.Member{
					Address:    penalizedMembers[0].String(),
					GroupID:    tss.GroupID(2),
					IsActive:   true,
					Since:      s.ctx.BlockTime(),
					LastActive: s.ctx.BlockTime(),
				})
				s.keeper.SetMember(s.ctx, types.Member{
					Address:    penalizedMembers[0].String(),
					GroupID:    tss.GroupID(1),
					IsActive:   true,
					Since:      s.ctx.BlockTime(),
					LastActive: s.ctx.BlockTime(),
				})
				s.keeper.SetGroupTransition(s.ctx, types.GroupTransition{
					SigningID:       tss.SigningID(3),
					Status:          types.TRANSITION_STATUS_WAITING_SIGN,
					CurrentGroupID:  tss.GroupID(1),
					IncomingGroupID: tss.GroupID(2),
				})

				s.keeper.SetSigning(s.ctx, types.Signing{
					ID:                     1,
					FeePerSigner:           sdk.NewCoins(sdk.NewInt64Coin("uband", 10)),
					Requester:              requestor.String(),
					CurrentGroupSigningID:  1,
					IncomingGroupSigningID: 2,
				})

				s.tssKeeper.EXPECT().
					DeactivateMember(gomock.Any(), tss.GroupID(2), penalizedMembers[0]).
					Return(nil)
			},
			postCheck: func(s *KeeperTestSuite) {
				s.Require().Equal(types.SigningID(1), s.keeper.GetSigningIDMapping(s.ctx, 2))
				s.Require().Equal(types.SigningID(1), s.keeper.GetSigningIDMapping(s.ctx, 1))

				member, err := s.keeper.GetMember(s.ctx, penalizedMembers[0], tss.GroupID(1))
				s.Require().NoError(err)
				s.Require().True(member.IsActive)

				member, err = s.keeper.GetMember(s.ctx, penalizedMembers[0], tss.GroupID(2))
				s.Require().NoError(err)
				s.Require().False(member.IsActive)

				transition, found := s.keeper.GetGroupTransition(s.ctx)
				s.Require().True(found)
				s.Require().Equal(types.TRANSITION_STATUS_WAITING_SIGN, transition.Status)
			},
		},
		{
			name:  "no signingID mapping",
			input: input{4, penalizedMembers},
			preProcess: func(s *KeeperTestSuite) {
				s.tssKeeper.EXPECT().MustGetSigning(gomock.Any(), tss.SigningID(4)).
					Return(
						tsstypes.Signing{
							ID:      4,
							GroupID: 3,
						})
				s.keeper.SetSigningIDMapping(s.ctx, 2, 1)
				s.keeper.SetSigningIDMapping(s.ctx, 1, 1)
				s.keeper.SetSigning(s.ctx, types.Signing{
					ID:                     1,
					FeePerSigner:           sdk.NewCoins(),
					Requester:              requestor.String(),
					CurrentGroupSigningID:  1,
					IncomingGroupSigningID: 2,
				})
				s.keeper.SetGroupTransition(s.ctx, types.GroupTransition{
					SigningID: tss.SigningID(3),
					Status:    types.TRANSITION_STATUS_WAITING_SIGN,
				})
				s.keeper.SetMember(s.ctx, types.Member{
					Address:    penalizedMembers[0].String(),
					GroupID:    tss.GroupID(2),
					IsActive:   true,
					Since:      s.ctx.BlockTime(),
					LastActive: s.ctx.BlockTime(),
				})
				s.keeper.SetMember(s.ctx, types.Member{
					Address:    penalizedMembers[0].String(),
					GroupID:    tss.GroupID(1),
					IsActive:   true,
					Since:      s.ctx.BlockTime(),
					LastActive: s.ctx.BlockTime(),
				})
			},
			postCheck: func(s *KeeperTestSuite) {
				s.Require().Equal(types.SigningID(1), s.keeper.GetSigningIDMapping(s.ctx, 2))
				s.Require().Equal(types.SigningID(1), s.keeper.GetSigningIDMapping(s.ctx, 1))

				transition, found := s.keeper.GetGroupTransition(s.ctx)
				s.Require().True(found)
				s.Require().Equal(types.TRANSITION_STATUS_WAITING_SIGN, transition.Status)

				member, err := s.keeper.GetMember(s.ctx, penalizedMembers[0], tss.GroupID(1))
				s.Require().NoError(err)
				s.Require().True(member.IsActive)

				member, err = s.keeper.GetMember(s.ctx, penalizedMembers[0], tss.GroupID(2))
				s.Require().NoError(err)
				s.Require().True(member.IsActive)
			},
		},
		{
			name:  "signing on group transition message",
			input: input{3, penalizedMembers},
			preProcess: func(s *KeeperTestSuite) {
				s.tssKeeper.EXPECT().MustGetSigning(gomock.Any(), tss.SigningID(3)).
					Return(tsstypes.Signing{
						ID:      3,
						GroupID: 1,
					})
				s.keeper.SetSigningIDMapping(s.ctx, 2, 1)
				s.keeper.SetSigningIDMapping(s.ctx, 1, 1)
				s.keeper.SetSigning(s.ctx, types.Signing{
					ID:                     1,
					FeePerSigner:           sdk.NewCoins(),
					Requester:              requestor.String(),
					CurrentGroupSigningID:  1,
					IncomingGroupSigningID: 2,
				})
				s.keeper.SetMember(s.ctx, types.Member{
					Address:    penalizedMembers[0].String(),
					GroupID:    tss.GroupID(2),
					IsActive:   true,
					Since:      s.ctx.BlockTime(),
					LastActive: s.ctx.BlockTime(),
				})
				s.keeper.SetMember(s.ctx, types.Member{
					Address:    penalizedMembers[0].String(),
					GroupID:    tss.GroupID(1),
					IsActive:   true,
					Since:      s.ctx.BlockTime(),
					LastActive: s.ctx.BlockTime(),
				})

				s.keeper.SetGroupTransition(s.ctx, types.GroupTransition{
					SigningID: tss.SigningID(3),
					Status:    types.TRANSITION_STATUS_WAITING_SIGN,
				})
				s.tssKeeper.EXPECT().
					DeactivateMember(gomock.Any(), tss.GroupID(1), penalizedMembers[0]).
					Return(nil)
			},
			postCheck: func(s *KeeperTestSuite) {
				s.Require().Equal(types.SigningID(1), s.keeper.GetSigningIDMapping(s.ctx, 2))
				s.Require().Equal(types.SigningID(1), s.keeper.GetSigningIDMapping(s.ctx, 1))

				transition, found := s.keeper.GetGroupTransition(s.ctx)
				s.Require().True(found)
				s.Require().Equal(types.TRANSITION_STATUS_WAITING_SIGN, transition.Status)

				member, err := s.keeper.GetMember(s.ctx, penalizedMembers[0], tss.GroupID(1))
				s.Require().NoError(err)
				s.Require().False(member.IsActive)

				member, err = s.keeper.GetMember(s.ctx, penalizedMembers[0], tss.GroupID(2))
				s.Require().NoError(err)
				s.Require().True(member.IsActive)
			},
		},
		{
			name:  "signing on group transition message; but transition already expired",
			input: input{3, penalizedMembers},
			preProcess: func(s *KeeperTestSuite) {
				s.tssKeeper.EXPECT().MustGetSigning(gomock.Any(), tss.SigningID(3)).
					Return(
						tsstypes.Signing{
							ID:      3,
							GroupID: 1,
						})
				s.keeper.SetSigningIDMapping(s.ctx, 2, 1)
				s.keeper.SetSigningIDMapping(s.ctx, 1, 1)
				s.keeper.SetSigning(s.ctx, types.Signing{
					ID:                     1,
					FeePerSigner:           sdk.NewCoins(),
					Requester:              requestor.String(),
					CurrentGroupSigningID:  1,
					IncomingGroupSigningID: 2,
				})
				s.keeper.SetMember(s.ctx, types.Member{
					Address:    penalizedMembers[0].String(),
					GroupID:    tss.GroupID(2),
					IsActive:   true,
					Since:      s.ctx.BlockTime(),
					LastActive: s.ctx.BlockTime(),
				})
				s.keeper.SetMember(s.ctx, types.Member{
					Address:    penalizedMembers[0].String(),
					GroupID:    tss.GroupID(1),
					IsActive:   true,
					Since:      s.ctx.BlockTime(),
					LastActive: s.ctx.BlockTime(),
				})
				s.tssKeeper.EXPECT().
					DeactivateMember(gomock.Any(), tss.GroupID(1), penalizedMembers[0]).
					Return(nil)
			},
			postCheck: func(s *KeeperTestSuite) {
				s.Require().Equal(types.SigningID(1), s.keeper.GetSigningIDMapping(s.ctx, 2))
				s.Require().Equal(types.SigningID(1), s.keeper.GetSigningIDMapping(s.ctx, 1))

				_, found := s.keeper.GetGroupTransition(s.ctx)
				s.Require().False(found)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			if tc.preProcess != nil {
				tc.preProcess(s)
			}

			s.tssCallback.OnSigningTimeout(s.ctx, tc.input.signingID, tc.input.idleMembers)

			if tc.postCheck != nil {
				tc.postCheck(s)
			}
		})
	}
}

func (s *KeeperTestSuite) TestCallbackOnSignCompleted() {
	requestor := sdk.MustAccAddressFromBech32("band1t5x8hrmht463eq4m0xhfgz95h62dyvkq049eek")
	group2Members := []tsstypes.Member{
		{Address: "band1t5x8hrmht463eq4m0xhfgz95h62dyvkq049eek", GroupID: 2, IsActive: true, IsMalicious: false},
		{Address: "band1a22hgwm4tz8gj82y6zad3de2dcg5dpymtj20m5", GroupID: 2, IsActive: true, IsMalicious: false},
	}

	type input struct {
		signingID       tss.SigningID
		assignedMembers []sdk.AccAddress
	}

	testCases := []struct {
		name       string
		input      input
		preProcess func(s *KeeperTestSuite)
		postCheck  func(s *KeeperTestSuite)
	}{
		{
			name: "normal signing with fee 10uband 2 members",
			input: input{
				signingID: 1,
				assignedMembers: []sdk.AccAddress{
					sdk.MustAccAddressFromBech32("band1t5x8hrmht463eq4m0xhfgz95h62dyvkq049eek"),
					sdk.MustAccAddressFromBech32("band1a22hgwm4tz8gj82y6zad3de2dcg5dpymtj20m5"),
				},
			},
			preProcess: func(s *KeeperTestSuite) {
				s.keeper.SetSigningIDMapping(s.ctx, 1, 1)
				s.keeper.SetSigning(s.ctx, types.Signing{
					ID:                     1,
					FeePerSigner:           sdk.NewCoins(sdk.NewInt64Coin("uband", 10)),
					Requester:              requestor.String(),
					CurrentGroupSigningID:  1,
					IncomingGroupSigningID: 0,
				})
				s.keeper.SetCurrentGroup(s.ctx, types.NewCurrentGroup(1, s.ctx.BlockTime()))
				s.keeper.SetGroupTransition(s.ctx, types.GroupTransition{
					SigningID:       tss.SigningID(3),
					Status:          types.TRANSITION_STATUS_WAITING_SIGN,
					IncomingGroupID: tss.GroupID(2),
					ExecTime:        s.ctx.BlockTime().Add(10 * time.Minute),
				})

				s.bankKeeper.EXPECT().SendCoinsFromModuleToAccount(
					gomock.Any(),
					types.ModuleName,
					sdk.MustAccAddressFromBech32("band1t5x8hrmht463eq4m0xhfgz95h62dyvkq049eek"),
					sdk.NewCoins(sdk.NewInt64Coin("uband", 10)),
				)
				s.bankKeeper.EXPECT().SendCoinsFromModuleToAccount(
					gomock.Any(),
					types.ModuleName,
					sdk.MustAccAddressFromBech32("band1a22hgwm4tz8gj82y6zad3de2dcg5dpymtj20m5"),
					sdk.NewCoins(sdk.NewInt64Coin("uband", 10)),
				)
			},
			postCheck: func(s *KeeperTestSuite) {
				s.Require().Zero(s.keeper.GetSigningIDMapping(s.ctx, 1))
				transition, found := s.keeper.GetGroupTransition(s.ctx)
				s.Require().True(found)
				s.Require().Equal(types.TRANSITION_STATUS_WAITING_SIGN, transition.Status)
			},
		},
		{
			name: "signing currentGroup with no fee 2 members",
			input: input{
				signingID: 1,
				assignedMembers: []sdk.AccAddress{
					sdk.MustAccAddressFromBech32("band1t5x8hrmht463eq4m0xhfgz95h62dyvkq049eek"),
					sdk.MustAccAddressFromBech32("band1a22hgwm4tz8gj82y6zad3de2dcg5dpymtj20m5"),
				},
			},
			preProcess: func(s *KeeperTestSuite) {
				s.keeper.SetSigningIDMapping(s.ctx, 1, 1)
				s.keeper.SetSigning(s.ctx, types.Signing{
					ID:                     1,
					FeePerSigner:           sdk.NewCoins(),
					Requester:              requestor.String(),
					CurrentGroupSigningID:  1,
					IncomingGroupSigningID: 0,
				})
				s.keeper.SetCurrentGroup(s.ctx, types.NewCurrentGroup(1, s.ctx.BlockTime()))
				s.keeper.SetGroupTransition(s.ctx, types.GroupTransition{
					SigningID:       tss.SigningID(3),
					Status:          types.TRANSITION_STATUS_WAITING_SIGN,
					IncomingGroupID: tss.GroupID(2),
					ExecTime:        s.ctx.BlockTime().Add(10 * time.Minute),
				})
			},
			postCheck: func(s *KeeperTestSuite) {
				s.Require().Zero(s.keeper.GetSigningIDMapping(s.ctx, 1))
				transition, found := s.keeper.GetGroupTransition(s.ctx)
				s.Require().True(found)
				s.Require().Equal(types.TRANSITION_STATUS_WAITING_SIGN, transition.Status)
			},
		},
		{
			name: "normal signing with fee 10uband 2 members from previous group",
			input: input{
				signingID: 1,
				assignedMembers: []sdk.AccAddress{
					sdk.MustAccAddressFromBech32("band1t5x8hrmht463eq4m0xhfgz95h62dyvkq049eek"),
					sdk.MustAccAddressFromBech32("band1a22hgwm4tz8gj82y6zad3de2dcg5dpymtj20m5"),
				},
			},
			preProcess: func(s *KeeperTestSuite) {
				s.keeper.SetSigningIDMapping(s.ctx, 1, 1)
				s.keeper.SetSigning(s.ctx, types.Signing{
					ID:                     1,
					FeePerSigner:           sdk.NewCoins(sdk.NewInt64Coin("uband", 10)),
					Requester:              requestor.String(),
					CurrentGroupSigningID:  1,
					IncomingGroupSigningID: 0,
				})
				s.keeper.SetCurrentGroup(s.ctx, types.NewCurrentGroup(2, s.ctx.BlockTime()))

				s.bankKeeper.EXPECT().SendCoinsFromModuleToAccount(
					gomock.Any(),
					types.ModuleName,
					sdk.MustAccAddressFromBech32("band1t5x8hrmht463eq4m0xhfgz95h62dyvkq049eek"),
					sdk.NewCoins(sdk.NewInt64Coin("uband", 10)),
				)
				s.bankKeeper.EXPECT().SendCoinsFromModuleToAccount(
					gomock.Any(),
					types.ModuleName,
					sdk.MustAccAddressFromBech32("band1a22hgwm4tz8gj82y6zad3de2dcg5dpymtj20m5"),
					sdk.NewCoins(sdk.NewInt64Coin("uband", 10)),
				)
			},
			postCheck: func(s *KeeperTestSuite) {
				s.Require().Zero(s.keeper.GetSigningIDMapping(s.ctx, 1))
				_, found := s.keeper.GetGroupTransition(s.ctx)
				s.Require().False(found)
			},
		},
		{
			name: "normal signing on incomingGroup",
			input: input{
				signingID: 2,
				assignedMembers: []sdk.AccAddress{
					sdk.MustAccAddressFromBech32("band1t5x8hrmht463eq4m0xhfgz95h62dyvkq049eek"),
					sdk.MustAccAddressFromBech32("band1a22hgwm4tz8gj82y6zad3de2dcg5dpymtj20m5"),
				},
			},
			preProcess: func(s *KeeperTestSuite) {
				s.keeper.SetSigningIDMapping(s.ctx, 2, 1)
				s.keeper.SetSigningIDMapping(s.ctx, 1, 1)
				s.keeper.SetSigning(s.ctx, types.Signing{
					ID:                     1,
					FeePerSigner:           sdk.NewCoins(sdk.NewInt64Coin("uband", 10)),
					Requester:              requestor.String(),
					CurrentGroupSigningID:  1,
					IncomingGroupSigningID: 2,
				})
				s.keeper.SetGroupTransition(s.ctx, types.GroupTransition{
					SigningID:       tss.SigningID(3),
					Status:          types.TRANSITION_STATUS_WAITING_SIGN,
					IncomingGroupID: tss.GroupID(2),
					ExecTime:        s.ctx.BlockTime().Add(10 * time.Minute),
				})
			},
			postCheck: func(s *KeeperTestSuite) {
				s.Require().Zero(s.keeper.GetSigningIDMapping(s.ctx, 2))
				s.Require().Equal(types.SigningID(1), s.keeper.GetSigningIDMapping(s.ctx, 1))
				transition, found := s.keeper.GetGroupTransition(s.ctx)
				s.Require().True(found)
				s.Require().Equal(types.TRANSITION_STATUS_WAITING_SIGN, transition.Status)
			},
		},
		{
			name: "normal signing transition",
			input: input{
				signingID: 3,
				assignedMembers: []sdk.AccAddress{
					sdk.MustAccAddressFromBech32("band1t5x8hrmht463eq4m0xhfgz95h62dyvkq049eek"),
					sdk.MustAccAddressFromBech32("band1a22hgwm4tz8gj82y6zad3de2dcg5dpymtj20m5"),
				},
			},
			preProcess: func(s *KeeperTestSuite) {
				s.keeper.SetSigningIDMapping(s.ctx, 2, 1)
				s.keeper.SetSigningIDMapping(s.ctx, 1, 1)
				s.keeper.SetSigning(s.ctx, types.Signing{
					ID:                     1,
					CurrentGroupSigningID:  1,
					IncomingGroupSigningID: 2,
				})
				s.keeper.SetGroupTransition(s.ctx, types.GroupTransition{
					SigningID:           tss.SigningID(3),
					Status:              types.TRANSITION_STATUS_WAITING_SIGN,
					CurrentGroupID:      tss.GroupID(1),
					CurrentGroupPubKey:  tss.Point([]byte("pubkey-1")),
					IncomingGroupID:     tss.GroupID(2),
					IncomingGroupPubKey: tss.Point([]byte("pubkey-2")),
					ExecTime:            s.ctx.BlockTime().Add(10 * time.Minute),
				})

				s.tssKeeper.EXPECT().MustGetMembers(gomock.Any(), tss.GroupID(2)).Return(group2Members)
				s.tssKeeper.EXPECT().GetSigningResult(gomock.Any(), tss.SigningID(3)).Return(
					&tsstypes.SigningResult{
						EVMSignature: &tsstypes.EVMSignature{
							RAddress:  []byte("raddress"),
							Signature: []byte("sig"),
						},
					}, nil,
				)
			},

			postCheck: func(s *KeeperTestSuite) {
				s.Require().Equal(types.SigningID(1), s.keeper.GetSigningIDMapping(s.ctx, 2))
				s.Require().Equal(types.SigningID(1), s.keeper.GetSigningIDMapping(s.ctx, 1))
				transition, found := s.keeper.GetGroupTransition(s.ctx)
				s.Require().True(found)
				s.Require().Equal(types.TRANSITION_STATUS_WAITING_EXECUTION, transition.Status)
				for _, m := range group2Members {
					member, err := s.keeper.GetMember(s.ctx, sdk.MustAccAddressFromBech32(m.Address), tss.GroupID(2))
					s.Require().NoError(err)
					s.Require().True(member.IsActive)
				}
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			if tc.preProcess != nil {
				tc.preProcess(s)
			}

			s.tssCallback.OnSigningCompleted(s.ctx, tc.input.signingID, tc.input.assignedMembers)

			if tc.postCheck != nil {
				tc.postCheck(s)
			}
		})
	}
}

func (s *KeeperTestSuite) TestCallbackOnGroupCreationComplete() {
	addrs := []sdk.AccAddress{
		sdk.MustAccAddressFromBech32("band1t5x8hrmht463eq4m0xhfgz95h62dyvkq049eek"),
		sdk.MustAccAddressFromBech32("band1a22hgwm4tz8gj82y6zad3de2dcg5dpymtj20m5"),
	}
	members := make([]tsstypes.Member, len(addrs))
	for i, addr := range addrs {
		members[i] = tsstypes.Member{
			Address:     addr.String(),
			GroupID:     1,
			IsActive:    true,
			IsMalicious: false,
		}
	}

	testCases := []struct {
		name       string
		input      tss.GroupID
		preProcess func(s *KeeperTestSuite)
		postCheck  func(s *KeeperTestSuite)
	}{
		{
			name:  "transition status unspecified",
			input: 2,
			preProcess: func(s *KeeperTestSuite) {
				s.keeper.SetCurrentGroup(s.ctx, types.NewCurrentGroup(1, s.ctx.BlockTime()))
			},
			postCheck: func(s *KeeperTestSuite) {
				_, found := s.keeper.GetGroupTransition(s.ctx)
				s.Require().False(found)
				s.Require().Equal(tss.GroupID(1), s.keeper.GetCurrentGroup(s.ctx).GroupID)
			},
		},
		{
			name:  "transition exec time is already expired",
			input: 2,
			preProcess: func(s *KeeperTestSuite) {
				s.keeper.SetGroupTransition(s.ctx, types.GroupTransition{
					SigningID:       tss.SigningID(1),
					Status:          types.TRANSITION_STATUS_CREATING_GROUP,
					IncomingGroupID: tss.GroupID(2),
					ExecTime:        s.ctx.BlockTime().Add(-10 * time.Minute),
				})
				s.keeper.SetCurrentGroup(s.ctx, types.NewCurrentGroup(1, s.ctx.BlockTime()))
			},
			postCheck: func(s *KeeperTestSuite) {
				transition, found := s.keeper.GetGroupTransition(s.ctx)
				s.Require().True(found)
				s.Require().Equal(types.TRANSITION_STATUS_CREATING_GROUP, transition.Status)
				s.Require().Equal(tss.GroupID(1), s.keeper.GetCurrentGroup(s.ctx).GroupID)
			},
		},
		{
			name:  "transition group ID does not match",
			input: 2,
			preProcess: func(s *KeeperTestSuite) {
				s.keeper.SetGroupTransition(s.ctx, types.GroupTransition{
					SigningID:       tss.SigningID(1),
					Status:          types.TRANSITION_STATUS_CREATING_GROUP,
					IncomingGroupID: tss.GroupID(3),
					ExecTime:        s.ctx.BlockTime().Add(10 * time.Minute),
				})
				s.keeper.SetCurrentGroup(s.ctx, types.NewCurrentGroup(1, s.ctx.BlockTime()))
			},
			postCheck: func(s *KeeperTestSuite) {
				transition, found := s.keeper.GetGroupTransition(s.ctx)
				s.Require().True(found)
				s.Require().Equal(types.TRANSITION_STATUS_CREATING_GROUP, transition.Status)
				s.Require().Equal(tss.GroupID(1), s.keeper.GetCurrentGroup(s.ctx).GroupID)
			},
		},
		{
			name:  "no current group id",
			input: 1,
			preProcess: func(s *KeeperTestSuite) {
				s.tssKeeper.EXPECT().MustGetGroup(gomock.Any(), tss.GroupID(1)).
					Return(tsstypes.Group{
						ID:          1,
						ModuleOwner: types.ModuleName,
						Status:      tsstypes.GROUP_STATUS_ACTIVE,
						PubKey:      []byte("pubkey"),
					})

				s.keeper.SetGroupTransition(s.ctx, types.GroupTransition{
					Status:          types.TRANSITION_STATUS_CREATING_GROUP,
					IncomingGroupID: tss.GroupID(1),
					ExecTime:        s.ctx.BlockTime().Add(10 * time.Minute),
				})
				s.tssKeeper.EXPECT().MustGetMembers(gomock.Any(), tss.GroupID(1)).Return(members)
			},
			postCheck: func(s *KeeperTestSuite) {
				transition, found := s.keeper.GetGroupTransition(s.ctx)
				s.Require().True(found)
				s.Require().Equal(types.TRANSITION_STATUS_WAITING_EXECUTION, transition.Status)
				s.Require().Equal(tss.GroupID(1), transition.IncomingGroupID)
				s.Require().Equal(tss.GroupID(0), transition.CurrentGroupID)
				s.Require().Equal(tss.Point(nil), transition.CurrentGroupPubKey)
				s.Require().Equal(tss.Point([]byte("pubkey")), transition.IncomingGroupPubKey)

				s.Require().Equal(tss.GroupID(0), s.keeper.GetCurrentGroup(s.ctx).GroupID)

				for _, member := range members {
					m, err := s.keeper.GetMember(s.ctx, sdk.MustAccAddressFromBech32(member.Address), tss.GroupID(1))
					s.Require().NoError(err)
					s.Require().True(m.IsActive)
				}
			},
		},
		{
			name:  "existing current group id",
			input: 2,
			preProcess: func(s *KeeperTestSuite) {
				s.tssKeeper.EXPECT().MustGetGroup(gomock.Any(), tss.GroupID(2)).
					Return(tsstypes.Group{
						ID:          2,
						ModuleOwner: types.ModuleName,
						Status:      tsstypes.GROUP_STATUS_ACTIVE,
						PubKey:      []byte("pubkey-2"),
					})
				s.keeper.SetGroupTransition(s.ctx, types.GroupTransition{
					SigningID:          tss.SigningID(1),
					Status:             types.TRANSITION_STATUS_CREATING_GROUP,
					CurrentGroupID:     tss.GroupID(1),
					CurrentGroupPubKey: tss.Point([]byte("pubkey")),
					IncomingGroupID:    tss.GroupID(2),
					ExecTime:           s.ctx.BlockTime().Add(10 * time.Minute),
				})
				s.keeper.SetCurrentGroup(s.ctx, types.NewCurrentGroup(1, s.ctx.BlockTime()))
				s.accountKeeper.EXPECT().GetModuleAccount(gomock.Any(), types.ModuleName).Return(
					s.moduleAcc,
				)
				s.tssKeeper.EXPECT().RequestSigning(
					gomock.Any(),
					tss.GroupID(1),
					gomock.Any(),
					types.NewGroupTransitionSignatureOrder(
						tss.Point([]byte("pubkey-2")),
						s.ctx.BlockTime().Add(10*time.Minute),
					),
				).Return(tss.SigningID(1), nil)
			},
			postCheck: func(s *KeeperTestSuite) {
				transition, found := s.keeper.GetGroupTransition(s.ctx)
				s.Require().True(found)
				s.Require().Equal(types.TRANSITION_STATUS_WAITING_SIGN, transition.Status)
				s.Require().Equal(tss.GroupID(2), transition.IncomingGroupID)
				s.Require().Equal(tss.GroupID(1), transition.CurrentGroupID)
				s.Require().Equal(tss.Point([]byte("pubkey")), transition.CurrentGroupPubKey)
				s.Require().Equal(tss.SigningID(1), transition.SigningID)
				s.Require().Equal(tss.Point([]byte("pubkey-2")), transition.IncomingGroupPubKey)

				s.Require().Equal(tss.GroupID(1), s.keeper.GetCurrentGroup(s.ctx).GroupID)

				for _, member := range members {
					ok := s.keeper.HasMember(s.ctx, sdk.MustAccAddressFromBech32(member.Address), tss.GroupID(2))
					s.Require().False(ok)
				}
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			if tc.preProcess != nil {
				tc.preProcess(s)
			}

			s.tssCallback.OnGroupCreationCompleted(s.ctx, tc.input)

			if tc.postCheck != nil {
				tc.postCheck(s)
			}
		})
	}
}
