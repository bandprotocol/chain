package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	bandtesting "github.com/bandprotocol/chain/v2/testing"
	"github.com/bandprotocol/chain/v2/x/bandtss/testutil"
	"github.com/bandprotocol/chain/v2/x/bandtss/types"
	tsstypes "github.com/bandprotocol/chain/v2/x/tss/types"
)

func TestHandleCreateSigning(t *testing.T) {
	s := testutil.NewTestSuite(t)
	ctx, k := s.Ctx, s.Keeper

	currentGroupID := tss.GroupID(1)
	currentGroup := tsstypes.Group{
		ID:        currentGroupID,
		Status:    tsstypes.GROUP_STATUS_ACTIVE,
		Size_:     3,
		Threshold: 2,
	}
	content := &tsstypes.TextSignatureOrder{Message: []byte("test")}
	params := types.DefaultParams()
	params.Fee = sdk.NewCoins(sdk.NewInt64Coin("uband", 10))

	k.SetCurrentGroupID(ctx, currentGroupID)
	err := k.SetParams(ctx, params)
	require.NoError(t, err)

	s.MockTSSKeeper.EXPECT().HandleSigningContent(ctx, content).Return([]byte("test"), nil).AnyTimes()
	s.MockTSSKeeper.EXPECT().GetGroup(ctx, currentGroupID).Return(currentGroup, nil).AnyTimes()

	type input struct {
		sender   sdk.AccAddress
		feeLimit sdk.Coins
	}

	testCases := []struct {
		name        string
		preProcess  func()
		input       input
		expectErr   error
		postProcess func()
		postCheck   func()
	}{
		{
			name: "test success with only current group",
			preProcess: func() {
				expectCurrentGroupSigning := &tsstypes.Signing{
					ID:      tss.SigningID(1),
					GroupID: currentGroupID,
					Status:  tsstypes.SIGNING_STATUS_WAITING,
				}
				s.MockTSSKeeper.EXPECT().CreateSigning(
					ctx,
					currentGroup,
					[]byte("test"),
				).Return(expectCurrentGroupSigning, nil)
				s.MockBankKeeper.EXPECT().SendCoinsFromAccountToModule(
					ctx,
					bandtesting.Alice.Address,
					types.ModuleName,
					sdk.NewCoins(sdk.NewInt64Coin("uband", 20)),
				).Return(nil)
			},
			postCheck: func() {
				// check mapping of tss signingID -> bandtss signingID
				actualMappedSigningID := k.GetSigningIDMapping(ctx, tss.SigningID(1))
				require.Equal(t, types.SigningID(1), actualMappedSigningID)

				// check bandtssSigning
				bandtssSigning, err := k.GetSigning(ctx, types.SigningID(1))
				require.NoError(t, err)
				require.Equal(t, types.Signing{
					ID:                      types.SigningID(1),
					Fee:                     sdk.NewCoins(sdk.NewInt64Coin("uband", 10)),
					Requester:               bandtesting.Alice.Address.String(),
					CurrentGroupSigningID:   tss.SigningID(1),
					ReplacingGroupSigningID: tss.SigningID(0),
				}, bandtssSigning)
			},
			input: input{sender: bandtesting.Alice.Address, feeLimit: sdk.NewCoins(sdk.NewInt64Coin("uband", 100))},
			postProcess: func() {
				k.SetSigningCount(ctx, 0)
				k.DeleteSigningIDMapping(ctx, tss.SigningID(1))
			},
			expectErr: nil,
		},
		{
			name: "test success with creator is authority",
			preProcess: func() {
				expectCurrentGroupSigning := &tsstypes.Signing{
					ID:      tss.SigningID(1),
					GroupID: currentGroupID,
					Status:  tsstypes.SIGNING_STATUS_WAITING,
				}
				s.MockTSSKeeper.EXPECT().CreateSigning(
					ctx,
					currentGroup,
					[]byte("test"),
				).Return(expectCurrentGroupSigning, nil)
			},
			postCheck: func() {
				// check mapping of tss signingID -> bandtss signingID
				actualMappedSigningID := k.GetSigningIDMapping(ctx, tss.SigningID(1))
				require.Equal(t, types.SigningID(1), actualMappedSigningID)

				// check bandtssSigning
				bandtssSigning, err := k.GetSigning(ctx, types.SigningID(1))
				require.NoError(t, err)
				require.Equal(t, types.Signing{
					ID:                      types.SigningID(1),
					Fee:                     nil,
					Requester:               s.Authority.String(),
					CurrentGroupSigningID:   tss.SigningID(1),
					ReplacingGroupSigningID: tss.SigningID(0),
				}, bandtssSigning)
			},
			input:     input{sender: s.Authority, feeLimit: sdk.NewCoins(sdk.NewInt64Coin("uband", 100))},
			expectErr: nil,
			postProcess: func() {
				k.SetSigningCount(ctx, 0)
				k.DeleteSigningIDMapping(ctx, tss.SigningID(1))
			},
		},
		{
			name: "test success with both current and replacing group",
			preProcess: func() {
				replaceGroupID := tss.GroupID(2)
				replaceGroup := tsstypes.Group{
					ID:        currentGroupID,
					Status:    tsstypes.GROUP_STATUS_ACTIVE,
					Size_:     3,
					Threshold: 3,
				}
				replacement := types.Replacement{
					SigningID:      tss.SigningID(1),
					Status:         types.REPLACEMENT_STATUS_WAITING_REPLACE,
					CurrentGroupID: currentGroupID,
					NewGroupID:     replaceGroupID,
				}
				expectCurrentGroupSigning := &tsstypes.Signing{
					ID:      tss.SigningID(2),
					GroupID: currentGroupID,
					Status:  tsstypes.SIGNING_STATUS_WAITING,
				}
				expectReplaceGroupSigning := &tsstypes.Signing{
					ID:      tss.SigningID(3),
					GroupID: currentGroupID,
					Status:  tsstypes.SIGNING_STATUS_WAITING,
				}
				k.SetReplacement(ctx, replacement)

				s.MockTSSKeeper.EXPECT().GetGroup(
					ctx,
					replaceGroupID,
				).Return(replaceGroup, nil)
				s.MockTSSKeeper.EXPECT().CreateSigning(
					ctx,
					currentGroup,
					[]byte("test"),
				).Return(expectCurrentGroupSigning, nil)
				s.MockTSSKeeper.EXPECT().CreateSigning(
					ctx,
					replaceGroup,
					[]byte("test"),
				).Return(expectReplaceGroupSigning, nil)
				s.MockBankKeeper.EXPECT().SendCoinsFromAccountToModule(
					ctx,
					bandtesting.Alice.Address,
					types.ModuleName,
					sdk.NewCoins(sdk.NewInt64Coin("uband", 20)),
				).Return(nil)
			},
			postCheck: func() {
				// check mapping of tss signingID -> bandtss signingID
				bandtssSignignID := k.GetSigningIDMapping(ctx, tss.SigningID(2))
				require.Equal(t, types.SigningID(1), bandtssSignignID)
				bandtssSignignID = k.GetSigningIDMapping(ctx, tss.SigningID(3))
				require.Equal(t, types.SigningID(1), bandtssSignignID)

				// check bandtssSigning
				bandtssSigning, err := k.GetSigning(ctx, types.SigningID(1))
				require.NoError(t, err)
				require.Equal(t, types.Signing{
					ID:                      types.SigningID(1),
					Fee:                     sdk.NewCoins(sdk.NewInt64Coin("uband", 10)),
					Requester:               bandtesting.Alice.Address.String(),
					CurrentGroupSigningID:   tss.SigningID(2),
					ReplacingGroupSigningID: tss.SigningID(3),
				}, bandtssSigning)
			},
			input: input{sender: bandtesting.Alice.Address, feeLimit: sdk.NewCoins(sdk.NewInt64Coin("uband", 100))},
			postProcess: func() {
				k.SetReplacement(ctx, types.Replacement{})
				k.SetSigningCount(ctx, 0)
				k.DeleteSigningIDMapping(ctx, tss.SigningID(1))
			},
			expectErr: nil,
		},
		{
			name: "test success with current group. Replacing group request is not signed",
			preProcess: func() {
				replaceGroupID := tss.GroupID(2)
				replacement := types.Replacement{
					SigningID:      tss.SigningID(1),
					Status:         types.REPLACEMENT_STATUS_WAITING_SIGN,
					CurrentGroupID: currentGroupID,
					NewGroupID:     replaceGroupID,
				}
				expectCurrentGroupSigning := &tsstypes.Signing{
					ID:      tss.SigningID(4),
					GroupID: currentGroupID,
					Status:  tsstypes.SIGNING_STATUS_WAITING,
				}
				k.SetReplacement(ctx, replacement)

				s.MockTSSKeeper.EXPECT().CreateSigning(
					ctx,
					currentGroup,
					[]byte("test"),
				).Return(expectCurrentGroupSigning, nil)
				s.MockBankKeeper.EXPECT().SendCoinsFromAccountToModule(
					ctx,
					bandtesting.Alice.Address,
					types.ModuleName,
					sdk.NewCoins(sdk.NewInt64Coin("uband", 20)),
				).Return(nil)
			},
			postCheck: func() {
				// check mapping of tss signingID -> bandtss signingID
				bandtssSignignID := k.GetSigningIDMapping(ctx, tss.SigningID(4))
				require.Equal(t, types.SigningID(1), bandtssSignignID)

				// check bandtssSigning
				bandtssSigning, err := k.GetSigning(ctx, types.SigningID(1))
				require.NoError(t, err)
				require.Equal(t, types.Signing{
					ID:                      types.SigningID(1),
					Fee:                     sdk.NewCoins(sdk.NewInt64Coin("uband", 10)),
					Requester:               bandtesting.Alice.Address.String(),
					CurrentGroupSigningID:   tss.SigningID(4),
					ReplacingGroupSigningID: tss.SigningID(0),
				}, bandtssSigning)
			},
			input: input{sender: bandtesting.Alice.Address, feeLimit: sdk.NewCoins(sdk.NewInt64Coin("uband", 100))},
			postProcess: func() {
				k.SetReplacement(ctx, types.Replacement{})
				k.SetSigningCount(ctx, 0)
				k.DeleteSigningIDMapping(ctx, tss.SigningID(1))
			},
			expectErr: nil,
		},
		{
			name:        "error no active group",
			preProcess:  func() { k.SetCurrentGroupID(ctx, 0) },
			postProcess: func() { k.SetCurrentGroupID(ctx, currentGroupID) },
			input:       input{sender: bandtesting.Alice.Address, feeLimit: sdk.NewCoins(sdk.NewInt64Coin("uband", 100))},
			expectErr:   types.ErrNoActiveGroup,
		},
		{
			name: "fee more than limit",
			preProcess: func() {
				params.Fee = sdk.NewCoins(sdk.NewInt64Coin("uband", 100))
				err := k.SetParams(ctx, params)
				require.NoError(t, err)
			},
			postProcess: func() {
				params.Fee = sdk.NewCoins(sdk.NewInt64Coin("uband", 10))
				err := k.SetParams(ctx, params)
				require.NoError(t, err)
			},
			input:     input{sender: bandtesting.Alice.Address, feeLimit: sdk.NewCoins(sdk.NewInt64Coin("uband", 100))},
			expectErr: types.ErrFeeExceedsLimit,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.preProcess != nil {
				tc.preProcess()
			}

			_, err := k.HandleCreateSigning(ctx, content, tc.input.sender, tc.input.feeLimit)
			if tc.expectErr != nil {
				require.ErrorIs(t, err, tc.expectErr)
			} else {
				require.NoError(t, err)
			}

			if tc.postCheck != nil {
				tc.postCheck()
			}

			if tc.postProcess != nil {
				tc.postProcess()
			}
		})
	}
}
