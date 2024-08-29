package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/bandtss/types"
	tsstypes "github.com/bandprotocol/chain/v2/x/tss/types"
)

func TestCallbackOnSignFailed(t *testing.T) {
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
				s.Keeper.SetSigningIDMapping(s.Ctx, 1, 1)
				s.Keeper.SetSigning(s.Ctx, types.Signing{
					ID:                     1,
					FeePerSigner:           sdk.NewCoins(),
					Requester:              requestor.String(),
					CurrentGroupSigningID:  1,
					IncomingGroupSigningID: 0,
				})
				s.Keeper.SetGroupTransition(s.Ctx, types.GroupTransition{
					SigningID: tss.SigningID(3),
					Status:    types.TRANSITION_STATUS_WAITING_SIGN,
				})
			},
			postCheck: func(s *KeeperTestSuite) {
				require.Zero(s.T(), s.Keeper.GetSigningIDMapping(s.Ctx, 1))

				transition, found := s.Keeper.GetGroupTransition(s.Ctx)
				require.True(s.T(), found)
				require.Equal(s.T(), types.TRANSITION_STATUS_WAITING_SIGN, transition.Status)
			},
		},
		{
			name:  "signing incomingGroup",
			input: 2,
			preProcess: func(s *KeeperTestSuite) {
				s.Keeper.SetSigningIDMapping(s.Ctx, 2, 1)
				s.Keeper.SetSigningIDMapping(s.Ctx, 1, 1)
				s.Keeper.SetSigning(s.Ctx, types.Signing{
					ID:                     1,
					FeePerSigner:           sdk.NewCoins(),
					Requester:              requestor.String(),
					CurrentGroupSigningID:  1,
					IncomingGroupSigningID: 2,
				})
				s.Keeper.SetGroupTransition(s.Ctx, types.GroupTransition{
					SigningID: tss.SigningID(3),
					Status:    types.TRANSITION_STATUS_WAITING_SIGN,
				})
			},
			postCheck: func(s *KeeperTestSuite) {
				require.Zero(s.T(), s.Keeper.GetSigningIDMapping(s.Ctx, 2))
				require.Equal(s.T(), types.SigningID(1), s.Keeper.GetSigningIDMapping(s.Ctx, 1))

				transition, found := s.Keeper.GetGroupTransition(s.Ctx)
				require.True(s.T(), found)
				require.Equal(s.T(), types.TRANSITION_STATUS_WAITING_SIGN, transition.Status)
			},
		},
		{
			name:  "no signingID mapping",
			input: 4,
			preProcess: func(s *KeeperTestSuite) {
				s.Keeper.SetSigningIDMapping(s.Ctx, 2, 1)
				s.Keeper.SetSigningIDMapping(s.Ctx, 1, 1)
				s.Keeper.SetSigning(s.Ctx, types.Signing{
					ID:                     1,
					FeePerSigner:           sdk.NewCoins(),
					Requester:              requestor.String(),
					CurrentGroupSigningID:  1,
					IncomingGroupSigningID: 2,
				})
				s.Keeper.SetGroupTransition(s.Ctx, types.GroupTransition{
					SigningID: tss.SigningID(3),
					Status:    types.TRANSITION_STATUS_WAITING_SIGN,
				})
			},
			postCheck: func(s *KeeperTestSuite) {
				require.Equal(s.T(), types.SigningID(1), s.Keeper.GetSigningIDMapping(s.Ctx, 2))
				require.Equal(s.T(), types.SigningID(1), s.Keeper.GetSigningIDMapping(s.Ctx, 1))

				transition, found := s.Keeper.GetGroupTransition(s.Ctx)
				require.True(s.T(), found)
				require.Equal(s.T(), types.TRANSITION_STATUS_WAITING_SIGN, transition.Status)
			},
		},
		{
			name:  "signing on group transition message",
			input: 3,
			preProcess: func(s *KeeperTestSuite) {
				s.Keeper.SetSigningIDMapping(s.Ctx, 2, 1)
				s.Keeper.SetSigningIDMapping(s.Ctx, 1, 1)
				s.Keeper.SetSigning(s.Ctx, types.Signing{
					ID:                     1,
					FeePerSigner:           sdk.NewCoins(),
					Requester:              requestor.String(),
					CurrentGroupSigningID:  1,
					IncomingGroupSigningID: 2,
				})
				s.Keeper.SetGroupTransition(s.Ctx, types.GroupTransition{
					SigningID:      tss.SigningID(3),
					CurrentGroupID: tss.GroupID(1),
					Status:         types.TRANSITION_STATUS_WAITING_SIGN,
				})
			},
			postCheck: func(s *KeeperTestSuite) {
				require.Equal(s.T(), types.SigningID(1), s.Keeper.GetSigningIDMapping(s.Ctx, 2))
				require.Equal(s.T(), types.SigningID(1), s.Keeper.GetSigningIDMapping(s.Ctx, 1))

				_, found := s.Keeper.GetGroupTransition(s.Ctx)
				require.False(s.T(), found)
			},
		},
		{
			name:  "signing on group transition message; but transition undefined",
			input: 3,
			preProcess: func(s *KeeperTestSuite) {
				s.Keeper.SetSigningIDMapping(s.Ctx, 2, 1)
				s.Keeper.SetSigningIDMapping(s.Ctx, 1, 1)
				s.Keeper.SetSigning(s.Ctx, types.Signing{
					ID:                     1,
					FeePerSigner:           sdk.NewCoins(),
					Requester:              requestor.String(),
					CurrentGroupSigningID:  1,
					IncomingGroupSigningID: 2,
				})
			},
			postCheck: func(s *KeeperTestSuite) {
				require.Equal(s.T(), types.SigningID(1), s.Keeper.GetSigningIDMapping(s.Ctx, 2))
				require.Equal(s.T(), types.SigningID(1), s.Keeper.GetSigningIDMapping(s.Ctx, 1))

				_, found := s.Keeper.GetGroupTransition(s.Ctx)
				require.False(s.T(), found)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewKeeperTestSuite(t)

			if tc.preProcess != nil {
				tc.preProcess(&s)
			}

			s.TssCallback.OnSigningFailed(s.Ctx, tc.input)

			if tc.postCheck != nil {
				tc.postCheck(&s)
			}
		})
	}
}

func TestCallbackOnSignTimeout(t *testing.T) {
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
				s.MockTSSKeeper.EXPECT().MustGetSigning(gomock.Any(), tss.SigningID(1)).
					Return(tsstypes.Signing{
						ID:      1,
						GroupID: 1,
					})
				s.Keeper.SetCurrentGroupID(s.Ctx, 1)
				s.Keeper.SetSigningIDMapping(s.Ctx, 1, 1)
				s.Keeper.SetMember(s.Ctx, types.Member{
					Address:    penalizedMembers[0].String(),
					GroupID:    tss.GroupID(1),
					IsActive:   true,
					Since:      s.Ctx.BlockTime(),
					LastActive: s.Ctx.BlockTime(),
				})

				s.Keeper.SetSigning(s.Ctx, types.Signing{
					ID:                     1,
					FeePerSigner:           sdk.NewCoins(sdk.NewInt64Coin("uband", 10)),
					Requester:              requestor.String(),
					CurrentGroupSigningID:  1,
					IncomingGroupSigningID: 0,
				})
				s.Keeper.SetGroupTransition(s.Ctx, types.GroupTransition{
					SigningID:       tss.SigningID(3),
					Status:          types.TRANSITION_STATUS_WAITING_SIGN,
					CurrentGroupID:  tss.GroupID(1),
					IncomingGroupID: tss.GroupID(0),
				})
				s.MockTSSKeeper.EXPECT().
					DeactivateMember(gomock.Any(), tss.GroupID(1), penalizedMembers[0]).
					Return(nil)
			},
			postCheck: func(s *KeeperTestSuite) {
				require.Equal(s.T(), types.SigningID(1), s.Keeper.GetSigningIDMapping(s.Ctx, 1))

				member, err := s.Keeper.GetMember(s.Ctx, penalizedMembers[0], tss.GroupID(1))
				require.NoError(s.T(), err)
				require.False(s.T(), member.IsActive)

				transition, found := s.Keeper.GetGroupTransition(s.Ctx)
				require.True(s.T(), found)
				require.Equal(s.T(), types.TRANSITION_STATUS_WAITING_SIGN, transition.Status)
			},
		},
		{
			name:  "signing incomingGroup",
			input: input{2, penalizedMembers},
			preProcess: func(s *KeeperTestSuite) {
				s.MockTSSKeeper.EXPECT().MustGetSigning(gomock.Any(), tss.SigningID(2)).
					Return(tsstypes.Signing{
						ID:      2,
						GroupID: 2,
					})
				s.Keeper.SetCurrentGroupID(s.Ctx, 1)
				s.Keeper.SetSigningIDMapping(s.Ctx, 1, 1)
				s.Keeper.SetSigningIDMapping(s.Ctx, 2, 1)
				s.Keeper.SetMember(s.Ctx, types.Member{
					Address:    penalizedMembers[0].String(),
					GroupID:    tss.GroupID(2),
					IsActive:   true,
					Since:      s.Ctx.BlockTime(),
					LastActive: s.Ctx.BlockTime(),
				})
				s.Keeper.SetMember(s.Ctx, types.Member{
					Address:    penalizedMembers[0].String(),
					GroupID:    tss.GroupID(1),
					IsActive:   true,
					Since:      s.Ctx.BlockTime(),
					LastActive: s.Ctx.BlockTime(),
				})
				s.Keeper.SetGroupTransition(s.Ctx, types.GroupTransition{
					SigningID:       tss.SigningID(3),
					Status:          types.TRANSITION_STATUS_WAITING_SIGN,
					CurrentGroupID:  tss.GroupID(1),
					IncomingGroupID: tss.GroupID(2),
				})

				s.Keeper.SetSigning(s.Ctx, types.Signing{
					ID:                     1,
					FeePerSigner:           sdk.NewCoins(sdk.NewInt64Coin("uband", 10)),
					Requester:              requestor.String(),
					CurrentGroupSigningID:  1,
					IncomingGroupSigningID: 2,
				})

				s.MockTSSKeeper.EXPECT().
					DeactivateMember(gomock.Any(), tss.GroupID(2), penalizedMembers[0]).
					Return(nil)
			},
			postCheck: func(s *KeeperTestSuite) {
				require.Equal(s.T(), types.SigningID(1), s.Keeper.GetSigningIDMapping(s.Ctx, 2))
				require.Equal(s.T(), types.SigningID(1), s.Keeper.GetSigningIDMapping(s.Ctx, 1))

				member, err := s.Keeper.GetMember(s.Ctx, penalizedMembers[0], tss.GroupID(1))
				require.NoError(s.T(), err)
				require.True(s.T(), member.IsActive)

				member, err = s.Keeper.GetMember(s.Ctx, penalizedMembers[0], tss.GroupID(2))
				require.NoError(s.T(), err)
				require.False(s.T(), member.IsActive)

				transition, found := s.Keeper.GetGroupTransition(s.Ctx)
				require.True(s.T(), found)
				require.Equal(s.T(), types.TRANSITION_STATUS_WAITING_SIGN, transition.Status)
			},
		},
		{
			name:  "no signingID mapping",
			input: input{4, penalizedMembers},
			preProcess: func(s *KeeperTestSuite) {
				s.MockTSSKeeper.EXPECT().MustGetSigning(gomock.Any(), tss.SigningID(4)).
					Return(
						tsstypes.Signing{
							ID:      4,
							GroupID: 3,
						})
				s.Keeper.SetSigningIDMapping(s.Ctx, 2, 1)
				s.Keeper.SetSigningIDMapping(s.Ctx, 1, 1)
				s.Keeper.SetSigning(s.Ctx, types.Signing{
					ID:                     1,
					FeePerSigner:           sdk.NewCoins(),
					Requester:              requestor.String(),
					CurrentGroupSigningID:  1,
					IncomingGroupSigningID: 2,
				})
				s.Keeper.SetGroupTransition(s.Ctx, types.GroupTransition{
					SigningID: tss.SigningID(3),
					Status:    types.TRANSITION_STATUS_WAITING_SIGN,
				})
				s.Keeper.SetMember(s.Ctx, types.Member{
					Address:    penalizedMembers[0].String(),
					GroupID:    tss.GroupID(2),
					IsActive:   true,
					Since:      s.Ctx.BlockTime(),
					LastActive: s.Ctx.BlockTime(),
				})
				s.Keeper.SetMember(s.Ctx, types.Member{
					Address:    penalizedMembers[0].String(),
					GroupID:    tss.GroupID(1),
					IsActive:   true,
					Since:      s.Ctx.BlockTime(),
					LastActive: s.Ctx.BlockTime(),
				})
			},
			postCheck: func(s *KeeperTestSuite) {
				require.Equal(s.T(), types.SigningID(1), s.Keeper.GetSigningIDMapping(s.Ctx, 2))
				require.Equal(s.T(), types.SigningID(1), s.Keeper.GetSigningIDMapping(s.Ctx, 1))

				transition, found := s.Keeper.GetGroupTransition(s.Ctx)
				require.True(s.T(), found)
				require.Equal(s.T(), types.TRANSITION_STATUS_WAITING_SIGN, transition.Status)

				member, err := s.Keeper.GetMember(s.Ctx, penalizedMembers[0], tss.GroupID(1))
				require.NoError(s.T(), err)
				require.True(s.T(), member.IsActive)

				member, err = s.Keeper.GetMember(s.Ctx, penalizedMembers[0], tss.GroupID(2))
				require.NoError(s.T(), err)
				require.True(s.T(), member.IsActive)
			},
		},
		{
			name:  "signing on group transition message",
			input: input{3, penalizedMembers},
			preProcess: func(s *KeeperTestSuite) {
				s.MockTSSKeeper.EXPECT().MustGetSigning(gomock.Any(), tss.SigningID(3)).
					Return(tsstypes.Signing{
						ID:      3,
						GroupID: 1,
					})
				s.Keeper.SetSigningIDMapping(s.Ctx, 2, 1)
				s.Keeper.SetSigningIDMapping(s.Ctx, 1, 1)
				s.Keeper.SetSigning(s.Ctx, types.Signing{
					ID:                     1,
					FeePerSigner:           sdk.NewCoins(),
					Requester:              requestor.String(),
					CurrentGroupSigningID:  1,
					IncomingGroupSigningID: 2,
				})
				s.Keeper.SetMember(s.Ctx, types.Member{
					Address:    penalizedMembers[0].String(),
					GroupID:    tss.GroupID(2),
					IsActive:   true,
					Since:      s.Ctx.BlockTime(),
					LastActive: s.Ctx.BlockTime(),
				})
				s.Keeper.SetMember(s.Ctx, types.Member{
					Address:    penalizedMembers[0].String(),
					GroupID:    tss.GroupID(1),
					IsActive:   true,
					Since:      s.Ctx.BlockTime(),
					LastActive: s.Ctx.BlockTime(),
				})

				s.Keeper.SetGroupTransition(s.Ctx, types.GroupTransition{
					SigningID: tss.SigningID(3),
					Status:    types.TRANSITION_STATUS_WAITING_SIGN,
				})
				s.MockTSSKeeper.EXPECT().
					DeactivateMember(gomock.Any(), tss.GroupID(1), penalizedMembers[0]).
					Return(nil)
			},
			postCheck: func(s *KeeperTestSuite) {
				require.Equal(s.T(), types.SigningID(1), s.Keeper.GetSigningIDMapping(s.Ctx, 2))
				require.Equal(s.T(), types.SigningID(1), s.Keeper.GetSigningIDMapping(s.Ctx, 1))

				transition, found := s.Keeper.GetGroupTransition(s.Ctx)
				require.True(s.T(), found)
				require.Equal(s.T(), types.TRANSITION_STATUS_WAITING_SIGN, transition.Status)

				member, err := s.Keeper.GetMember(s.Ctx, penalizedMembers[0], tss.GroupID(1))
				require.NoError(s.T(), err)
				require.False(s.T(), member.IsActive)

				member, err = s.Keeper.GetMember(s.Ctx, penalizedMembers[0], tss.GroupID(2))
				require.NoError(s.T(), err)
				require.True(s.T(), member.IsActive)
			},
		},
		{
			name:  "signing on group transition message; but transition already expired",
			input: input{3, penalizedMembers},
			preProcess: func(s *KeeperTestSuite) {
				s.MockTSSKeeper.EXPECT().MustGetSigning(gomock.Any(), tss.SigningID(3)).
					Return(
						tsstypes.Signing{
							ID:      3,
							GroupID: 1,
						})
				s.Keeper.SetSigningIDMapping(s.Ctx, 2, 1)
				s.Keeper.SetSigningIDMapping(s.Ctx, 1, 1)
				s.Keeper.SetSigning(s.Ctx, types.Signing{
					ID:                     1,
					FeePerSigner:           sdk.NewCoins(),
					Requester:              requestor.String(),
					CurrentGroupSigningID:  1,
					IncomingGroupSigningID: 2,
				})
				s.Keeper.SetMember(s.Ctx, types.Member{
					Address:    penalizedMembers[0].String(),
					GroupID:    tss.GroupID(2),
					IsActive:   true,
					Since:      s.Ctx.BlockTime(),
					LastActive: s.Ctx.BlockTime(),
				})
				s.Keeper.SetMember(s.Ctx, types.Member{
					Address:    penalizedMembers[0].String(),
					GroupID:    tss.GroupID(1),
					IsActive:   true,
					Since:      s.Ctx.BlockTime(),
					LastActive: s.Ctx.BlockTime(),
				})
				s.MockTSSKeeper.EXPECT().
					DeactivateMember(gomock.Any(), tss.GroupID(1), penalizedMembers[0]).
					Return(nil)
			},
			postCheck: func(s *KeeperTestSuite) {
				require.Equal(s.T(), types.SigningID(1), s.Keeper.GetSigningIDMapping(s.Ctx, 2))
				require.Equal(s.T(), types.SigningID(1), s.Keeper.GetSigningIDMapping(s.Ctx, 1))

				_, found := s.Keeper.GetGroupTransition(s.Ctx)
				require.False(s.T(), found)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewKeeperTestSuite(t)

			if tc.preProcess != nil {
				tc.preProcess(&s)
			}

			s.TssCallback.OnSigningTimeout(s.Ctx, tc.input.signingID, tc.input.idleMembers)

			if tc.postCheck != nil {
				tc.postCheck(&s)
			}
		})
	}
}

func TestCallbackOnSignCompleted(t *testing.T) {
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
				s.Keeper.SetSigningIDMapping(s.Ctx, 1, 1)
				s.Keeper.SetSigning(s.Ctx, types.Signing{
					ID:                     1,
					FeePerSigner:           sdk.NewCoins(sdk.NewInt64Coin("uband", 10)),
					Requester:              requestor.String(),
					CurrentGroupSigningID:  1,
					IncomingGroupSigningID: 0,
				})
				s.Keeper.SetCurrentGroupID(s.Ctx, 1)
				s.Keeper.SetGroupTransition(s.Ctx, types.GroupTransition{
					SigningID:       tss.SigningID(3),
					Status:          types.TRANSITION_STATUS_WAITING_SIGN,
					IncomingGroupID: tss.GroupID(2),
					ExecTime:        s.Ctx.BlockTime().Add(10 * time.Minute),
				})

				s.MockBankKeeper.EXPECT().SendCoinsFromModuleToAccount(
					gomock.Any(),
					types.ModuleName,
					sdk.MustAccAddressFromBech32("band1t5x8hrmht463eq4m0xhfgz95h62dyvkq049eek"),
					sdk.NewCoins(sdk.NewInt64Coin("uband", 10)),
				)
				s.MockBankKeeper.EXPECT().SendCoinsFromModuleToAccount(
					gomock.Any(),
					types.ModuleName,
					sdk.MustAccAddressFromBech32("band1a22hgwm4tz8gj82y6zad3de2dcg5dpymtj20m5"),
					sdk.NewCoins(sdk.NewInt64Coin("uband", 10)),
				)
			},
			postCheck: func(s *KeeperTestSuite) {
				require.Zero(s.T(), s.Keeper.GetSigningIDMapping(s.Ctx, 1))
				transition, found := s.Keeper.GetGroupTransition(s.Ctx)
				require.True(s.T(), found)
				require.Equal(s.T(), types.TRANSITION_STATUS_WAITING_SIGN, transition.Status)
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
				s.Keeper.SetSigningIDMapping(s.Ctx, 1, 1)
				s.Keeper.SetSigning(s.Ctx, types.Signing{
					ID:                     1,
					FeePerSigner:           sdk.NewCoins(),
					Requester:              requestor.String(),
					CurrentGroupSigningID:  1,
					IncomingGroupSigningID: 0,
				})
				s.Keeper.SetCurrentGroupID(s.Ctx, 1)
				s.Keeper.SetGroupTransition(s.Ctx, types.GroupTransition{
					SigningID:       tss.SigningID(3),
					Status:          types.TRANSITION_STATUS_WAITING_SIGN,
					IncomingGroupID: tss.GroupID(2),
					ExecTime:        s.Ctx.BlockTime().Add(10 * time.Minute),
				})
			},
			postCheck: func(s *KeeperTestSuite) {
				require.Zero(s.T(), s.Keeper.GetSigningIDMapping(s.Ctx, 1))
				transition, found := s.Keeper.GetGroupTransition(s.Ctx)
				require.True(s.T(), found)
				require.Equal(s.T(), types.TRANSITION_STATUS_WAITING_SIGN, transition.Status)
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
				s.Keeper.SetSigningIDMapping(s.Ctx, 1, 1)
				s.Keeper.SetSigning(s.Ctx, types.Signing{
					ID:                     1,
					FeePerSigner:           sdk.NewCoins(sdk.NewInt64Coin("uband", 10)),
					Requester:              requestor.String(),
					CurrentGroupSigningID:  1,
					IncomingGroupSigningID: 0,
				})
				s.Keeper.SetCurrentGroupID(s.Ctx, 2)

				s.MockBankKeeper.EXPECT().SendCoinsFromModuleToAccount(
					gomock.Any(),
					types.ModuleName,
					sdk.MustAccAddressFromBech32("band1t5x8hrmht463eq4m0xhfgz95h62dyvkq049eek"),
					sdk.NewCoins(sdk.NewInt64Coin("uband", 10)),
				)
				s.MockBankKeeper.EXPECT().SendCoinsFromModuleToAccount(
					gomock.Any(),
					types.ModuleName,
					sdk.MustAccAddressFromBech32("band1a22hgwm4tz8gj82y6zad3de2dcg5dpymtj20m5"),
					sdk.NewCoins(sdk.NewInt64Coin("uband", 10)),
				)
			},
			postCheck: func(s *KeeperTestSuite) {
				require.Zero(s.T(), s.Keeper.GetSigningIDMapping(s.Ctx, 1))
				_, found := s.Keeper.GetGroupTransition(s.Ctx)
				require.False(s.T(), found)
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
				s.Keeper.SetSigningIDMapping(s.Ctx, 2, 1)
				s.Keeper.SetSigningIDMapping(s.Ctx, 1, 1)
				s.Keeper.SetSigning(s.Ctx, types.Signing{
					ID:                     1,
					FeePerSigner:           sdk.NewCoins(sdk.NewInt64Coin("uband", 10)),
					Requester:              requestor.String(),
					CurrentGroupSigningID:  1,
					IncomingGroupSigningID: 2,
				})
				s.Keeper.SetGroupTransition(s.Ctx, types.GroupTransition{
					SigningID:       tss.SigningID(3),
					Status:          types.TRANSITION_STATUS_WAITING_SIGN,
					IncomingGroupID: tss.GroupID(2),
					ExecTime:        s.Ctx.BlockTime().Add(10 * time.Minute),
				})
			},
			postCheck: func(s *KeeperTestSuite) {
				require.Zero(s.T(), s.Keeper.GetSigningIDMapping(s.Ctx, 2))
				require.Equal(s.T(), types.SigningID(1), s.Keeper.GetSigningIDMapping(s.Ctx, 1))
				transition, found := s.Keeper.GetGroupTransition(s.Ctx)
				require.True(s.T(), found)
				require.Equal(s.T(), types.TRANSITION_STATUS_WAITING_SIGN, transition.Status)
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
				s.Keeper.SetSigningIDMapping(s.Ctx, 2, 1)
				s.Keeper.SetSigningIDMapping(s.Ctx, 1, 1)
				s.Keeper.SetSigning(s.Ctx, types.Signing{
					ID:                     1,
					CurrentGroupSigningID:  1,
					IncomingGroupSigningID: 2,
				})
				s.Keeper.SetGroupTransition(s.Ctx, types.GroupTransition{
					SigningID:           tss.SigningID(3),
					Status:              types.TRANSITION_STATUS_WAITING_SIGN,
					CurrentGroupID:      tss.GroupID(1),
					CurrentGroupPubKey:  tss.Point([]byte("pubkey-1")),
					IncomingGroupID:     tss.GroupID(2),
					IncomingGroupPubKey: tss.Point([]byte("pubkey-2")),
					ExecTime:            s.Ctx.BlockTime().Add(10 * time.Minute),
				})

				s.MockTSSKeeper.EXPECT().MustGetMembers(gomock.Any(), tss.GroupID(2)).Return(group2Members)
				s.MockTSSKeeper.EXPECT().GetSigningResult(gomock.Any(), tss.SigningID(3)).Return(
					&tsstypes.SigningResult{
						EVMSignature: &tsstypes.EVMSignature{
							RAddress:  []byte("raddress"),
							Signature: []byte("sig"),
						},
					}, nil,
				)
			},

			postCheck: func(s *KeeperTestSuite) {
				require.Equal(s.T(), types.SigningID(1), s.Keeper.GetSigningIDMapping(s.Ctx, 2))
				require.Equal(s.T(), types.SigningID(1), s.Keeper.GetSigningIDMapping(s.Ctx, 1))
				transition, found := s.Keeper.GetGroupTransition(s.Ctx)
				require.True(s.T(), found)
				require.Equal(s.T(), types.TRANSITION_STATUS_WAITING_EXECUTION, transition.Status)
				for _, m := range group2Members {
					member, err := s.Keeper.GetMember(s.Ctx, sdk.MustAccAddressFromBech32(m.Address), tss.GroupID(2))
					require.NoError(s.T(), err)
					require.True(s.T(), member.IsActive)
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewKeeperTestSuite(t)

			if tc.preProcess != nil {
				tc.preProcess(&s)
			}

			s.TssCallback.OnSigningCompleted(s.Ctx, tc.input.signingID, tc.input.assignedMembers)

			if tc.postCheck != nil {
				tc.postCheck(&s)
			}
		})
	}
}

func TestCallbackOnGroupCreationComplete(t *testing.T) {
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
				s.Keeper.SetCurrentGroupID(s.Ctx, 1)
			},
			postCheck: func(s *KeeperTestSuite) {
				_, found := s.Keeper.GetGroupTransition(s.Ctx)
				require.False(s.T(), found)
				require.Equal(s.T(), tss.GroupID(1), s.Keeper.GetCurrentGroupID(s.Ctx))
			},
		},
		{
			name:  "transition exec time is already expired",
			input: 2,
			preProcess: func(s *KeeperTestSuite) {
				s.Keeper.SetGroupTransition(s.Ctx, types.GroupTransition{
					SigningID:       tss.SigningID(1),
					Status:          types.TRANSITION_STATUS_CREATING_GROUP,
					IncomingGroupID: tss.GroupID(2),
					ExecTime:        s.Ctx.BlockTime().Add(-10 * time.Minute),
				})
				s.Keeper.SetCurrentGroupID(s.Ctx, 1)
			},
			postCheck: func(s *KeeperTestSuite) {
				transition, found := s.Keeper.GetGroupTransition(s.Ctx)
				require.True(s.T(), found)
				require.Equal(s.T(), types.TRANSITION_STATUS_CREATING_GROUP, transition.Status)
				require.Equal(s.T(), tss.GroupID(1), s.Keeper.GetCurrentGroupID(s.Ctx))
			},
		},
		{
			name:  "transition group ID does not match",
			input: 2,
			preProcess: func(s *KeeperTestSuite) {
				s.Keeper.SetGroupTransition(s.Ctx, types.GroupTransition{
					SigningID:       tss.SigningID(1),
					Status:          types.TRANSITION_STATUS_CREATING_GROUP,
					IncomingGroupID: tss.GroupID(3),
					ExecTime:        s.Ctx.BlockTime().Add(10 * time.Minute),
				})
				s.Keeper.SetCurrentGroupID(s.Ctx, 1)
			},
			postCheck: func(s *KeeperTestSuite) {
				transition, found := s.Keeper.GetGroupTransition(s.Ctx)
				require.True(s.T(), found)
				require.Equal(s.T(), types.TRANSITION_STATUS_CREATING_GROUP, transition.Status)
				require.Equal(s.T(), tss.GroupID(1), s.Keeper.GetCurrentGroupID(s.Ctx))
			},
		},
		{
			name:  "no current group id",
			input: 1,
			preProcess: func(s *KeeperTestSuite) {
				s.MockTSSKeeper.EXPECT().MustGetGroup(gomock.Any(), tss.GroupID(1)).
					Return(tsstypes.Group{
						ID:          1,
						ModuleOwner: types.ModuleName,
						Status:      tsstypes.GROUP_STATUS_ACTIVE,
						PubKey:      []byte("pubkey"),
					})

				s.Keeper.SetGroupTransition(s.Ctx, types.GroupTransition{
					Status:          types.TRANSITION_STATUS_CREATING_GROUP,
					IncomingGroupID: tss.GroupID(1),
					ExecTime:        s.Ctx.BlockTime().Add(10 * time.Minute),
				})
				s.MockTSSKeeper.EXPECT().MustGetMembers(gomock.Any(), tss.GroupID(1)).Return(members)
			},
			postCheck: func(s *KeeperTestSuite) {
				transition, found := s.Keeper.GetGroupTransition(s.Ctx)
				require.True(s.T(), found)
				require.Equal(s.T(), types.TRANSITION_STATUS_WAITING_EXECUTION, transition.Status)
				require.Equal(s.T(), tss.GroupID(1), transition.IncomingGroupID)
				require.Equal(s.T(), tss.GroupID(0), transition.CurrentGroupID)
				require.Equal(s.T(), tss.Point(nil), transition.CurrentGroupPubKey)
				require.Equal(s.T(), tss.Point([]byte("pubkey")), transition.IncomingGroupPubKey)

				require.Equal(s.T(), tss.GroupID(0), s.Keeper.GetCurrentGroupID(s.Ctx))

				for _, member := range members {
					m, err := s.Keeper.GetMember(s.Ctx, sdk.MustAccAddressFromBech32(member.Address), tss.GroupID(1))
					require.NoError(s.T(), err)
					require.True(s.T(), m.IsActive)
				}
			},
		},
		{
			name:  "existing current group id",
			input: 2,
			preProcess: func(s *KeeperTestSuite) {
				s.MockTSSKeeper.EXPECT().MustGetGroup(gomock.Any(), tss.GroupID(2)).
					Return(tsstypes.Group{
						ID:          2,
						ModuleOwner: types.ModuleName,
						Status:      tsstypes.GROUP_STATUS_ACTIVE,
						PubKey:      []byte("pubkey-2"),
					})
				s.Keeper.SetGroupTransition(s.Ctx, types.GroupTransition{
					SigningID:          tss.SigningID(1),
					Status:             types.TRANSITION_STATUS_CREATING_GROUP,
					CurrentGroupID:     tss.GroupID(1),
					CurrentGroupPubKey: tss.Point([]byte("pubkey")),
					IncomingGroupID:    tss.GroupID(2),
					ExecTime:           s.Ctx.BlockTime().Add(10 * time.Minute),
				})
				s.Keeper.SetCurrentGroupID(s.Ctx, 1)
				s.MockAccountKeeper.EXPECT().GetModuleAccount(gomock.Any(), types.ModuleName).Return(
					s.ModuleAcc,
				)
				s.MockTSSKeeper.EXPECT().RequestSigning(
					gomock.Any(),
					tss.GroupID(1),
					gomock.Any(),
					types.NewGroupTransitionSignatureOrder(tss.Point([]byte("pubkey-2"))),
				).Return(tss.SigningID(1), nil)
			},
			postCheck: func(s *KeeperTestSuite) {
				transition, found := s.Keeper.GetGroupTransition(s.Ctx)
				require.True(s.T(), found)
				require.Equal(s.T(), types.TRANSITION_STATUS_WAITING_SIGN, transition.Status)
				require.Equal(s.T(), tss.GroupID(2), transition.IncomingGroupID)
				require.Equal(s.T(), tss.GroupID(1), transition.CurrentGroupID)
				require.Equal(s.T(), tss.Point([]byte("pubkey")), transition.CurrentGroupPubKey)
				require.Equal(s.T(), tss.SigningID(1), transition.SigningID)
				require.Equal(s.T(), tss.Point([]byte("pubkey-2")), transition.IncomingGroupPubKey)

				require.Equal(s.T(), tss.GroupID(1), s.Keeper.GetCurrentGroupID(s.Ctx))

				for _, member := range members {
					ok := s.Keeper.HasMember(s.Ctx, sdk.MustAccAddressFromBech32(member.Address), tss.GroupID(2))
					require.False(s.T(), ok)
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewKeeperTestSuite(t)

			if tc.preProcess != nil {
				tc.preProcess(&s)
			}

			s.TssCallback.OnGroupCreationCompleted(s.Ctx, tc.input)

			if tc.postCheck != nil {
				tc.postCheck(&s)
			}
		})
	}
}
