package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	bandtesting "github.com/bandprotocol/chain/v2/testing"
	"github.com/bandprotocol/chain/v2/x/bandtss/types"
	tsstypes "github.com/bandprotocol/chain/v2/x/tss/types"
)

func TestCreateDirectSigningRequest(t *testing.T) {
	currentGroupID := tss.GroupID(1)
	currentGroup := tsstypes.Group{
		ID:        currentGroupID,
		Status:    tsstypes.GROUP_STATUS_ACTIVE,
		Size_:     3,
		Threshold: 2,
	}
	content := &tsstypes.TextSignatureOrder{Message: []byte("test")}

	type input struct {
		sender   sdk.AccAddress
		feeLimit sdk.Coins
	}
	testCases := []struct {
		name       string
		preProcess func(s *KeeperTestSuite)
		input      input
		expectErr  error
		postCheck  func(s *KeeperTestSuite)
	}{
		{
			name: "test success with only current group",
			preProcess: func(s *KeeperTestSuite) {
				s.Keeper.SetCurrentGroupID(s.Ctx, currentGroupID)

				s.MockTSSKeeper.EXPECT().GetGroup(gomock.Any(), currentGroupID).
					Return(currentGroup, nil).
					AnyTimes()
				s.MockTSSKeeper.EXPECT().RequestSigning(gomock.Any(), currentGroupID, gomock.Any(), content).
					Return(tss.SigningID(1), nil)
				s.MockBankKeeper.EXPECT().SendCoinsFromAccountToModule(
					gomock.Any(),
					bandtesting.Alice.Address,
					types.ModuleName,
					sdk.NewCoins(sdk.NewInt64Coin("uband", 20)),
				).Return(nil)
			},
			postCheck: func(s *KeeperTestSuite) {
				// check mapping of tss signingID -> bandtss signingID
				actualMappedSigningID := s.Keeper.GetSigningIDMapping(s.Ctx, tss.SigningID(1))
				require.Equal(t, types.SigningID(1), actualMappedSigningID)

				// check bandtssSigning
				bandtssSigning, err := s.Keeper.GetSigning(s.Ctx, types.SigningID(1))
				require.NoError(t, err)
				require.Equal(t, types.Signing{
					ID:                     types.SigningID(1),
					FeePerSigner:           sdk.NewCoins(sdk.NewInt64Coin("uband", 10)),
					Requester:              bandtesting.Alice.Address.String(),
					CurrentGroupSigningID:  tss.SigningID(1),
					IncomingGroupSigningID: tss.SigningID(0),
				}, bandtssSigning)
			},
			input: input{
				sender:   bandtesting.Alice.Address,
				feeLimit: sdk.NewCoins(sdk.NewInt64Coin("uband", 100)),
			},
			expectErr: nil,
		},
		{
			name: "test success with both current and incoming group",
			preProcess: func(s *KeeperTestSuite) {
				incomingGroupID := tss.GroupID(2)
				transition := types.GroupTransition{
					SigningID:       tss.SigningID(1),
					Status:          types.TRANSITION_STATUS_WAITING_EXECUTION,
					CurrentGroupID:  currentGroupID,
					IncomingGroupID: incomingGroupID,
				}
				s.Keeper.SetGroupTransition(s.Ctx, transition)
				s.Keeper.SetCurrentGroupID(s.Ctx, currentGroupID)

				s.MockTSSKeeper.EXPECT().GetGroup(gomock.Any(), currentGroupID).
					Return(currentGroup, nil).
					AnyTimes()
				s.MockTSSKeeper.EXPECT().RequestSigning(gomock.Any(), currentGroupID, gomock.Any(), content).
					Return(tss.SigningID(2), nil)
				s.MockTSSKeeper.EXPECT().RequestSigning(gomock.Any(), incomingGroupID, gomock.Any(), content).
					Return(tss.SigningID(3), nil)
				s.MockBankKeeper.EXPECT().SendCoinsFromAccountToModule(
					gomock.Any(),
					bandtesting.Alice.Address,
					types.ModuleName,
					sdk.NewCoins(sdk.NewInt64Coin("uband", 20)),
				).Return(nil)
			},
			postCheck: func(s *KeeperTestSuite) {
				// check mapping of tss signingID -> bandtss signingID
				bandtssSignignID := s.Keeper.GetSigningIDMapping(s.Ctx, tss.SigningID(2))
				require.Equal(t, types.SigningID(1), bandtssSignignID)
				bandtssSignignID = s.Keeper.GetSigningIDMapping(s.Ctx, tss.SigningID(3))
				require.Equal(t, types.SigningID(1), bandtssSignignID)

				// check bandtssSigning
				bandtssSigning, err := s.Keeper.GetSigning(s.Ctx, types.SigningID(1))
				require.NoError(t, err)
				require.Equal(t, types.Signing{
					ID:                     types.SigningID(1),
					FeePerSigner:           sdk.NewCoins(sdk.NewInt64Coin("uband", 10)),
					Requester:              bandtesting.Alice.Address.String(),
					CurrentGroupSigningID:  tss.SigningID(2),
					IncomingGroupSigningID: tss.SigningID(3),
				}, bandtssSigning)
			},
			input: input{
				sender:   bandtesting.Alice.Address,
				feeLimit: sdk.NewCoins(sdk.NewInt64Coin("uband", 100)),
			},
			expectErr: nil,
		},
		{
			name: "request only current group; transition message is not signed",
			preProcess: func(s *KeeperTestSuite) {
				incomingGroupID := tss.GroupID(2)
				transition := types.GroupTransition{
					SigningID:       tss.SigningID(1),
					Status:          types.TRANSITION_STATUS_WAITING_SIGN,
					CurrentGroupID:  currentGroupID,
					IncomingGroupID: incomingGroupID,
				}
				s.Keeper.SetGroupTransition(s.Ctx, transition)
				s.Keeper.SetCurrentGroupID(s.Ctx, currentGroupID)

				s.MockTSSKeeper.EXPECT().GetGroup(gomock.Any(), currentGroupID).
					Return(currentGroup, nil).
					AnyTimes()
				s.MockTSSKeeper.EXPECT().RequestSigning(gomock.Any(), currentGroupID, gomock.Any(), content).
					Return(tss.SigningID(4), nil)
				s.MockBankKeeper.EXPECT().SendCoinsFromAccountToModule(
					gomock.Any(),
					bandtesting.Alice.Address,
					types.ModuleName,
					sdk.NewCoins(sdk.NewInt64Coin("uband", 20)),
				).Return(nil)
			},
			postCheck: func(s *KeeperTestSuite) {
				// check mapping of tss signingID -> bandtss signingID
				bandtssSignignID := s.Keeper.GetSigningIDMapping(s.Ctx, tss.SigningID(4))
				require.Equal(t, types.SigningID(1), bandtssSignignID)

				// check bandtssSigning
				bandtssSigning, err := s.Keeper.GetSigning(s.Ctx, types.SigningID(1))
				require.NoError(t, err)
				require.Equal(t, types.Signing{
					ID:                     types.SigningID(1),
					FeePerSigner:           sdk.NewCoins(sdk.NewInt64Coin("uband", 10)),
					Requester:              bandtesting.Alice.Address.String(),
					CurrentGroupSigningID:  tss.SigningID(4),
					IncomingGroupSigningID: tss.SigningID(0),
				}, bandtssSigning)
			},
			input: input{
				sender:   bandtesting.Alice.Address,
				feeLimit: sdk.NewCoins(sdk.NewInt64Coin("uband", 100)),
			},
			expectErr: nil,
		},
		{
			name: "request only incoming group; no current group",
			preProcess: func(s *KeeperTestSuite) {
				incomingGroupID := tss.GroupID(2)
				transition := types.GroupTransition{
					SigningID:       tss.SigningID(0),
					Status:          types.TRANSITION_STATUS_WAITING_EXECUTION,
					CurrentGroupID:  tss.GroupID(0),
					IncomingGroupID: incomingGroupID,
				}
				s.Keeper.SetGroupTransition(s.Ctx, transition)

				s.MockTSSKeeper.EXPECT().RequestSigning(gomock.Any(), incomingGroupID, gomock.Any(), content).
					Return(tss.SigningID(1), nil)
			},
			postCheck: func(s *KeeperTestSuite) {
				// check mapping of tss signingID -> bandtss signingID
				bandtssSignignID := s.Keeper.GetSigningIDMapping(s.Ctx, tss.SigningID(1))
				require.Equal(t, types.SigningID(1), bandtssSignignID)

				// check bandtssSigning
				bandtssSigning, err := s.Keeper.GetSigning(s.Ctx, types.SigningID(1))
				require.NoError(t, err)
				require.Equal(t, types.Signing{
					ID:                     types.SigningID(1),
					FeePerSigner:           nil,
					Requester:              bandtesting.Alice.Address.String(),
					CurrentGroupSigningID:  tss.SigningID(0),
					IncomingGroupSigningID: tss.SigningID(1),
				}, bandtssSigning)
			},
			input: input{
				sender:   bandtesting.Alice.Address,
				feeLimit: sdk.NewCoins(sdk.NewInt64Coin("uband", 100)),
			},
			expectErr: nil,
		},
		{
			name:       "error no current group; no incoming group",
			preProcess: func(s *KeeperTestSuite) {},
			input: input{
				sender:   bandtesting.Alice.Address,
				feeLimit: sdk.NewCoins(sdk.NewInt64Coin("uband", 100)),
			},
			expectErr: types.ErrNoActiveGroup,
		},
		{
			name: "error: fee more than limit",
			preProcess: func(s *KeeperTestSuite) {
				params := s.Keeper.GetParams(s.Ctx)
				params.Fee = sdk.NewCoins(sdk.NewInt64Coin("uband", 100))
				err := s.Keeper.SetParams(s.Ctx, params)
				require.NoError(t, err)
				s.Keeper.SetCurrentGroupID(s.Ctx, currentGroupID)

				s.MockTSSKeeper.EXPECT().GetGroup(gomock.Any(), currentGroupID).
					Return(currentGroup, nil).
					AnyTimes()
			},
			input: input{
				sender:   bandtesting.Alice.Address,
				feeLimit: sdk.NewCoins(sdk.NewInt64Coin("uband", 100)),
			},
			expectErr: types.ErrFeeExceedsLimit,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewKeeperTestSuite(t)

			if tc.preProcess != nil {
				tc.preProcess(&s)
			}

			_, err := s.Keeper.CreateDirectSigningRequest(s.Ctx, content, "", tc.input.sender, tc.input.feeLimit)
			if tc.expectErr != nil {
				require.ErrorIs(t, err, tc.expectErr)
			} else {
				require.NoError(t, err)
			}

			if tc.postCheck != nil {
				tc.postCheck(&s)
			}
		})
	}
}

func TestCreateDirectSigningRequestWithAuthority(t *testing.T) {
	s := NewKeeperTestSuite(t)

	currentGID := tss.GroupID(1)
	content := &tsstypes.TextSignatureOrder{Message: []byte("test")}

	s.Keeper.SetCurrentGroupID(s.Ctx, currentGID)
	s.MockTSSKeeper.EXPECT().RequestSigning(
		gomock.Any(),
		currentGID,
		gomock.Any(),
		content,
	).Return(tss.SigningID(1), nil)

	feeLimit := sdk.NewCoins(sdk.NewInt64Coin("uband", 100))
	_, err := s.Keeper.CreateDirectSigningRequest(s.Ctx, content, "", s.Authority, feeLimit)
	require.NoError(t, err)

	actualMappedSigningID := s.Keeper.GetSigningIDMapping(s.Ctx, tss.SigningID(1))
	require.Equal(t, types.SigningID(1), actualMappedSigningID)

	// check bandtssSigning
	bandtssSigning, err := s.Keeper.GetSigning(s.Ctx, types.SigningID(1))
	require.NoError(t, err)
	require.Equal(t, types.Signing{
		ID:                     types.SigningID(1),
		FeePerSigner:           nil,
		Requester:              s.Authority.String(),
		CurrentGroupSigningID:  tss.SigningID(1),
		IncomingGroupSigningID: tss.SigningID(0),
	}, bandtssSigning)
}
