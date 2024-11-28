package keeper_test

import (
	"go.uber.org/mock/gomock"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/pkg/tss"
	bandtesting "github.com/bandprotocol/chain/v3/testing"
	"github.com/bandprotocol/chain/v3/x/bandtss/types"
	tsstypes "github.com/bandprotocol/chain/v3/x/tss/types"
)

func (s *KeeperTestSuite) TestCreateDirectSigningRequest() {
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
				s.keeper.SetCurrentGroup(s.ctx, types.NewCurrentGroup(currentGroupID, s.ctx.BlockTime()))

				s.tssKeeper.EXPECT().GetGroup(gomock.Any(), currentGroupID).
					Return(currentGroup, nil).
					AnyTimes()
				s.tssKeeper.EXPECT().RequestSigning(gomock.Any(), currentGroupID, gomock.Any(), content).
					Return(tss.SigningID(1), nil)
				s.bankKeeper.EXPECT().SendCoinsFromAccountToModule(
					gomock.Any(),
					bandtesting.Alice.Address,
					types.ModuleName,
					sdk.NewCoins(sdk.NewInt64Coin("uband", 20)),
				).Return(nil)
			},
			postCheck: func(s *KeeperTestSuite) {
				// check mapping of tss signingID -> bandtss signingID
				actualMappedSigningID := s.keeper.GetSigningIDMapping(s.ctx, tss.SigningID(1))
				s.Require().Equal(types.SigningID(1), actualMappedSigningID)

				// check bandtssSigning
				bandtssSigning, err := s.keeper.GetSigning(s.ctx, types.SigningID(1))
				s.Require().NoError(err)
				s.Require().Equal(types.Signing{
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
			name: "test failed insufficient member in current group even normal incoming group",
			preProcess: func(s *KeeperTestSuite) {
				incomingGroupID := tss.GroupID(2)
				transition := types.GroupTransition{
					SigningID:       tss.SigningID(1),
					Status:          types.TRANSITION_STATUS_WAITING_EXECUTION,
					CurrentGroupID:  currentGroupID,
					IncomingGroupID: incomingGroupID,
				}
				s.keeper.SetGroupTransition(s.ctx, transition)
				s.keeper.SetCurrentGroup(s.ctx, types.NewCurrentGroup(currentGroupID, s.ctx.BlockTime()))

				s.tssKeeper.EXPECT().GetGroup(gomock.Any(), currentGroupID).
					Return(currentGroup, nil).
					AnyTimes()

				s.tssKeeper.EXPECT().RequestSigning(gomock.Any(), currentGroupID, gomock.Any(), content).
					DoAndReturn(func(
						ctx sdk.Context,
						groupID tss.GroupID,
						originator tsstypes.Originator,
						content tsstypes.Content,
					) (tss.SigningID, error) {
						ctx.KVStore(s.key).Set([]byte{0xff, 0xfe}, []byte("test"))
						return tss.SigningID(0), tsstypes.ErrInsufficientSigners
					})

				s.bankKeeper.EXPECT().SendCoinsFromAccountToModule(
					gomock.Any(),
					bandtesting.Alice.Address,
					types.ModuleName,
					sdk.NewCoins(sdk.NewInt64Coin("uband", 20)),
				).Return(nil)
			},
			postCheck: func(s *KeeperTestSuite) {
				s.Require().Equal([]byte("test"), s.ctx.KVStore(s.key).Get([]byte{0xff, 0xfe}))
			},
			input: input{
				sender:   bandtesting.Alice.Address,
				feeLimit: sdk.NewCoins(sdk.NewInt64Coin("uband", 100)),
			},
			expectErr: tsstypes.ErrInsufficientSigners,
		},
		{
			name: "test success with only current group; insufficient member on incoming group",
			preProcess: func(s *KeeperTestSuite) {
				incomingGroupID := tss.GroupID(2)
				transition := types.GroupTransition{
					SigningID:       tss.SigningID(1),
					Status:          types.TRANSITION_STATUS_WAITING_EXECUTION,
					CurrentGroupID:  currentGroupID,
					IncomingGroupID: incomingGroupID,
				}
				s.keeper.SetGroupTransition(s.ctx, transition)
				s.keeper.SetCurrentGroup(s.ctx, types.NewCurrentGroup(currentGroupID, s.ctx.BlockTime()))

				s.tssKeeper.EXPECT().GetGroup(gomock.Any(), currentGroupID).
					Return(currentGroup, nil).
					AnyTimes()

				s.tssKeeper.EXPECT().RequestSigning(gomock.Any(), currentGroupID, gomock.Any(), content).
					DoAndReturn(func(
						ctx sdk.Context,
						groupID tss.GroupID,
						originator tsstypes.Originator,
						content tsstypes.Content,
					) (tss.SigningID, error) {
						ctx.KVStore(s.key).Set([]byte{0xff, 0xfe}, []byte("test"))
						return tss.SigningID(1), nil
					})

				s.tssKeeper.EXPECT().RequestSigning(gomock.Any(), incomingGroupID, gomock.Any(), content).
					DoAndReturn(func(
						ctx sdk.Context,
						groupID tss.GroupID,
						originator tsstypes.Originator,
						content tsstypes.Content,
					) (tss.SigningID, error) {
						ctx.KVStore(s.key).Set([]byte{0xff, 0xff}, []byte("test"))
						return tss.SigningID(0), tsstypes.ErrInsufficientSigners
					})

				s.bankKeeper.EXPECT().SendCoinsFromAccountToModule(
					gomock.Any(),
					bandtesting.Alice.Address,
					types.ModuleName,
					sdk.NewCoins(sdk.NewInt64Coin("uband", 20)),
				).Return(nil)
			},
			postCheck: func(s *KeeperTestSuite) {
				// check mapping of tss signingID -> bandtss signingID
				actualMappedSigningID := s.keeper.GetSigningIDMapping(s.ctx, tss.SigningID(1))
				s.Require().Equal(types.SigningID(1), actualMappedSigningID)

				// check bandtssSigning
				bandtssSigning, err := s.keeper.GetSigning(s.ctx, types.SigningID(1))
				s.Require().NoError(err)
				s.Require().Equal(types.Signing{
					ID:                     types.SigningID(1),
					FeePerSigner:           sdk.NewCoins(sdk.NewInt64Coin("uband", 10)),
					Requester:              bandtesting.Alice.Address.String(),
					CurrentGroupSigningID:  tss.SigningID(1),
					IncomingGroupSigningID: tss.SigningID(0),
				}, bandtssSigning)

				s.Require().Equal([]byte("test"), s.ctx.KVStore(s.key).Get([]byte{0xff, 0xfe}))
				s.Require().Nil(s.ctx.KVStore(s.key).Get([]byte{0xff, 0xff}))
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
				s.keeper.SetGroupTransition(s.ctx, transition)
				s.keeper.SetCurrentGroup(s.ctx, types.NewCurrentGroup(currentGroupID, s.ctx.BlockTime()))

				s.tssKeeper.EXPECT().GetGroup(gomock.Any(), currentGroupID).
					Return(currentGroup, nil).
					AnyTimes()
				s.tssKeeper.EXPECT().RequestSigning(gomock.Any(), currentGroupID, gomock.Any(), content).
					Return(tss.SigningID(2), nil)
				s.tssKeeper.EXPECT().RequestSigning(gomock.Any(), incomingGroupID, gomock.Any(), content).
					Return(tss.SigningID(3), nil)
				s.bankKeeper.EXPECT().SendCoinsFromAccountToModule(
					gomock.Any(),
					bandtesting.Alice.Address,
					types.ModuleName,
					sdk.NewCoins(sdk.NewInt64Coin("uband", 20)),
				).Return(nil)
			},
			postCheck: func(s *KeeperTestSuite) {
				// check mapping of tss signingID -> bandtss signingID
				bandtssSignignID := s.keeper.GetSigningIDMapping(s.ctx, tss.SigningID(2))
				s.Require().Equal(types.SigningID(1), bandtssSignignID)
				bandtssSignignID = s.keeper.GetSigningIDMapping(s.ctx, tss.SigningID(3))
				s.Require().Equal(types.SigningID(1), bandtssSignignID)

				// check bandtssSigning
				bandtssSigning, err := s.keeper.GetSigning(s.ctx, types.SigningID(1))
				s.Require().NoError(err)
				s.Require().Equal(types.Signing{
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
				s.keeper.SetGroupTransition(s.ctx, transition)
				s.keeper.SetCurrentGroup(s.ctx, types.NewCurrentGroup(currentGroupID, s.ctx.BlockTime()))

				s.tssKeeper.EXPECT().GetGroup(gomock.Any(), currentGroupID).
					Return(currentGroup, nil).
					AnyTimes()
				s.tssKeeper.EXPECT().RequestSigning(gomock.Any(), currentGroupID, gomock.Any(), content).
					Return(tss.SigningID(4), nil)
				s.bankKeeper.EXPECT().SendCoinsFromAccountToModule(
					gomock.Any(),
					bandtesting.Alice.Address,
					types.ModuleName,
					sdk.NewCoins(sdk.NewInt64Coin("uband", 20)),
				).Return(nil)
			},
			postCheck: func(s *KeeperTestSuite) {
				// check mapping of tss signingID -> bandtss signingID
				bandtssSignignID := s.keeper.GetSigningIDMapping(s.ctx, tss.SigningID(4))
				s.Require().Equal(types.SigningID(1), bandtssSignignID)

				// check bandtssSigning
				bandtssSigning, err := s.keeper.GetSigning(s.ctx, types.SigningID(1))
				s.Require().NoError(err)
				s.Require().Equal(types.Signing{
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
				s.keeper.SetGroupTransition(s.ctx, transition)

				s.tssKeeper.EXPECT().RequestSigning(gomock.Any(), incomingGroupID, gomock.Any(), content).
					Return(tss.SigningID(1), nil)
			},
			postCheck: func(s *KeeperTestSuite) {
				// check mapping of tss signingID -> bandtss signingID
				bandtssSignignID := s.keeper.GetSigningIDMapping(s.ctx, tss.SigningID(1))
				s.Require().Equal(types.SigningID(1), bandtssSignignID)

				// check bandtssSigning
				bandtssSigning, err := s.keeper.GetSigning(s.ctx, types.SigningID(1))
				s.Require().NoError(err)
				s.Require().Equal(types.Signing{
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
				params := s.keeper.GetParams(s.ctx)
				params.FeePerSigner = sdk.NewCoins(sdk.NewInt64Coin("uband", 100))
				err := s.keeper.SetParams(s.ctx, params)
				s.Require().NoError(err)
				s.keeper.SetCurrentGroup(s.ctx, types.NewCurrentGroup(currentGroupID, s.ctx.BlockTime()))

				s.tssKeeper.EXPECT().GetGroup(gomock.Any(), currentGroupID).
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
		s.Run(tc.name, func() {
			if tc.preProcess != nil {
				tc.preProcess(s)
			}

			_, err := s.keeper.CreateDirectSigningRequest(s.ctx, content, "", tc.input.sender, tc.input.feeLimit)
			if tc.expectErr != nil {
				s.Require().ErrorIs(err, tc.expectErr)
			} else {
				s.Require().NoError(err)
			}

			if tc.postCheck != nil {
				tc.postCheck(s)
			}
		})
	}
}

func (s *KeeperTestSuite) TestCreateDirectSigningRequestWithAuthority() {
	currentGID := tss.GroupID(1)
	content := &tsstypes.TextSignatureOrder{Message: []byte("test")}

	s.keeper.SetCurrentGroup(s.ctx, types.NewCurrentGroup(currentGID, s.ctx.BlockTime()))
	s.tssKeeper.EXPECT().RequestSigning(
		gomock.Any(),
		currentGID,
		gomock.Any(),
		content,
	).Return(tss.SigningID(1), nil)

	feeLimit := sdk.NewCoins(sdk.NewInt64Coin("uband", 100))
	_, err := s.keeper.CreateDirectSigningRequest(s.ctx, content, "", s.authority, feeLimit)
	s.Require().NoError(err)

	actualMappedSigningID := s.keeper.GetSigningIDMapping(s.ctx, tss.SigningID(1))
	s.Require().Equal(types.SigningID(1), actualMappedSigningID)

	// check bandtssSigning
	bandtssSigning, err := s.keeper.GetSigning(s.ctx, types.SigningID(1))
	s.Require().NoError(err)
	s.Require().Equal(types.Signing{
		ID:                     types.SigningID(1),
		FeePerSigner:           nil,
		Requester:              s.authority.String(),
		CurrentGroupSigningID:  tss.SigningID(1),
		IncomingGroupSigningID: tss.SigningID(0),
	}, bandtssSigning)
}
