package keeper_test

import (
	"testing"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/stretchr/testify/suite"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/pkg/tss/testutil"
	"github.com/bandprotocol/chain/v2/testing/testapp"
	"github.com/bandprotocol/chain/v2/x/tss/keeper"
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

type KeeperTestSuite struct {
	suite.Suite

	app         *testapp.TestingApp
	ctx         sdk.Context
	queryClient types.QueryClient
	msgSrvr     types.MsgServer
	requester   sdk.AccAddress
	authority   sdk.AccAddress
}

var (
	PrivD = testutil.HexDecode("de6aedbe8ba688dd6d342881eb1e67c3476e825106477360148e2858a5eb565c")
	PrivE = testutil.HexDecode("3ff4fb2beac0cee0ab230829a5ae0881310046282e79c978ca22f44897ea434a")
	PubD  = tss.Scalar(PrivD).Point()
	PubE  = tss.Scalar(PrivE).Point()
)

func (s *KeeperTestSuite) SetupTest() {
	app, ctx, _ := testapp.CreateTestInput(true)
	s.app = app
	s.ctx = ctx

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, keeper.NewQueryServer(&app.TSSKeeper))
	queryClient := types.NewQueryClient(queryHelper)

	s.queryClient = queryClient
	s.msgSrvr = keeper.NewMsgServerImpl(&app.TSSKeeper)
	s.authority = authtypes.NewModuleAddress(govtypes.ModuleName)
}

func (s *KeeperTestSuite) setupCreateGroup() {
	ctx, msgSrvr := s.ctx, s.msgSrvr

	// Create group from testutil
	for _, tc := range testutil.TestCases {
		// Initialize members
		var members []string
		for _, m := range tc.Group.Members {
			address := sdk.AccAddress(m.PubKey())
			members = append(members, address.String())

			s.app.TSSKeeper.SetMemberStatus(ctx, types.Status{
				Address: address.String(),
				Status:  types.MEMBER_STATUS_ACTIVE,
				Since:   ctx.BlockTime(),
			})
		}

		// Create group
		_, err := msgSrvr.CreateGroup(ctx, &types.MsgCreateGroup{
			Members:   members,
			Threshold: tc.Group.Threshold,
			Fee:       sdk.NewCoins(sdk.NewInt64Coin("uband", 10)),
			Authority: s.authority.String(),
		})
		s.Require().NoError(err)

		// Set DKG context
		s.app.TSSKeeper.SetDKGContext(ctx, tc.Group.ID, tc.Group.DKGContext)
	}
}

func (s *KeeperTestSuite) setupRound1() {
	s.setupCreateGroup()

	ctx, app, msgSrvr := s.ctx, s.app, s.msgSrvr
	for _, tc := range testutil.TestCases {
		for _, m := range tc.Group.Members {
			// Submit Round 1 information for each member
			_, err := msgSrvr.SubmitDKGRound1(ctx, &types.MsgSubmitDKGRound1{
				GroupID: tc.Group.ID,
				Round1Info: types.Round1Info{
					MemberID:           m.ID,
					CoefficientCommits: m.CoefficientCommits,
					OneTimePubKey:      m.OneTimePubKey(),
					A0Signature:        m.A0Signature,
					OneTimeSignature:   m.OneTimeSignature,
				},
				Address: sdk.AccAddress(m.PubKey()).String(),
			})
			s.Require().NoError(err)
		}
	}

	// Execute the EndBlocker to process groups
	app.EndBlocker(ctx, abci.RequestEndBlock{Height: ctx.BlockHeight() + 1})
}

func (s *KeeperTestSuite) setupRound2() {
	s.setupRound1()

	ctx, app, msgSrvr := s.ctx, s.app, s.msgSrvr
	for _, tc := range testutil.TestCases {
		for _, m := range tc.Group.Members {
			// Submit Round 2 information for each member
			_, err := msgSrvr.SubmitDKGRound2(ctx, &types.MsgSubmitDKGRound2{
				GroupID: tc.Group.ID,
				Round2Info: types.Round2Info{
					MemberID:              m.ID,
					EncryptedSecretShares: m.EncSecretShares,
				},
				Address: sdk.AccAddress(m.PubKey()).String(),
			})
			s.Require().NoError(err)
		}
	}

	// Execute the EndBlocker to process groups
	app.EndBlocker(ctx, abci.RequestEndBlock{Height: ctx.BlockHeight() + 1})
}

func (s *KeeperTestSuite) setupConfirm() {
	s.setupRound2()

	ctx, app, msgSrvr := s.ctx, s.app, s.msgSrvr
	for _, tc := range testutil.TestCases {
		for _, m := range tc.Group.Members {
			// Confirm the group participation for each member
			_, err := msgSrvr.Confirm(ctx, &types.MsgConfirm{
				GroupID:      tc.Group.ID,
				MemberID:     m.ID,
				OwnPubKeySig: m.PubKeySignature,
				Address:      sdk.AccAddress(m.PubKey()).String(),
			})
			s.Require().NoError(err)
		}
	}

	// Execute the EndBlocker to process groups
	app.EndBlocker(ctx, abci.RequestEndBlock{Height: ctx.BlockHeight() + 1})
}

func (s *KeeperTestSuite) setupDE() {
	ctx, msgSrvr := s.ctx, s.msgSrvr

	for _, tc := range testutil.TestCases {
		for _, m := range tc.Group.Members {
			// Submit DEs for each member
			_, err := msgSrvr.SubmitDEs(ctx, &types.MsgSubmitDEs{
				DEs: []types.DE{
					{PubD: PubD, PubE: PubE},
					{PubD: PubD, PubE: PubE},
					{PubD: PubD, PubE: PubE},
					{PubD: PubD, PubE: PubE},
					{PubD: PubD, PubE: PubE},
				},
				Address: sdk.AccAddress(m.PubKey()).String(),
			})
			s.Require().NoError(err)
		}
	}
}

func (s *KeeperTestSuite) SetupGroup(groupStatus types.GroupStatus) {
	switch groupStatus {
	case types.GROUP_STATUS_ROUND_1:
		s.setupCreateGroup()
	case types.GROUP_STATUS_ROUND_2:
		s.setupRound1()
	case types.GROUP_STATUS_ROUND_3:
		s.setupRound2()
	case types.GROUP_STATUS_ACTIVE:
		s.setupConfirm()
		s.setupDE()
	}
}

func (s *KeeperTestSuite) TestGetSetGroupCount() {
	ctx, k := s.ctx, s.app.TSSKeeper
	k.SetGroupCount(ctx, 1)

	groupCount := k.GetGroupCount(ctx)
	s.Require().Equal(uint64(1), groupCount)
}

func (s *KeeperTestSuite) TestGetNextGroupID() {
	ctx, k := s.ctx, s.app.TSSKeeper

	// Initial group count
	k.SetGroupCount(ctx, 0)

	groupID1 := k.GetNextGroupID(ctx)
	s.Require().Equal(tss.GroupID(1), groupID1)
	groupID2 := k.GetNextGroupID(ctx)
	s.Require().Equal(tss.GroupID(2), groupID2)
}

func (s *KeeperTestSuite) TestIsGrantee() {
	ctx, k := s.ctx, s.app.TSSKeeper
	expTime := time.Unix(0, 0)

	// Init grantee address
	grantee, _ := sdk.AccAddressFromBech32("band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs")

	// Init granter address
	granter, _ := sdk.AccAddressFromBech32("band1p40yh3zkmhcv0ecqp3mcazy83sa57rgjp07dun")

	// Save grant msgs to grantee
	for _, m := range types.GetTSSGrantMsgTypes() {
		s.app.AuthzKeeper.SaveGrant(ctx, grantee, granter, authz.NewGenericAuthorization(m), &expTime)
	}

	isGrantee := k.CheckIsGrantee(ctx, granter, grantee)
	s.Require().True(isGrantee)
}

func (s *KeeperTestSuite) TestCreateNewGroup() {
	ctx, k := s.ctx, s.app.TSSKeeper

	group := types.Group{
		Size_:     5,
		Threshold: 3,
		PubKey:    nil,
		Status:    types.GROUP_STATUS_ROUND_1,
	}

	// Create new group
	groupID := k.CreateNewGroup(ctx, group)

	// init group ID
	group.ID = groupID

	// Get group by id
	got, err := k.GetGroup(ctx, groupID)
	s.Require().NoError(err)
	s.Require().Equal(group, got)
}

func (s *KeeperTestSuite) TestSetGroup() {
	ctx, k := s.ctx, s.app.TSSKeeper
	group := types.Group{
		Size_:     5,
		Threshold: 3,
		PubKey:    nil,
		Status:    types.GROUP_STATUS_ROUND_1,
	}

	// Set new group
	groupID := k.CreateNewGroup(ctx, group)

	// Update group size value
	group.Size_ = 6

	// Add group ID
	group.ID = groupID

	k.SetGroup(ctx, group)

	// Get group from chain state
	got, err := k.GetGroup(ctx, groupID)

	// Validate group size value
	s.Require().NoError(err)
	s.Require().Equal(group.Size_, got.Size_)
}

func (s *KeeperTestSuite) TestGetGroups() {
	ctx, k := s.ctx, s.app.TSSKeeper
	group := types.Group{
		ID:        1,
		Size_:     5,
		Threshold: 3,
		PubKey:    nil,
		Status:    types.GROUP_STATUS_ROUND_1,
	}

	// Set new group
	k.SetGroup(ctx, group)

	// Get group from chain state
	got := k.GetGroups(ctx)
	s.Require().Equal([]types.Group{group}, got)
}

func (s *KeeperTestSuite) TestDeleteGroup() {
	ctx, k := s.ctx, s.app.TSSKeeper

	// Create a sample group ID
	groupID := tss.GroupID(123)

	// Set up a sample group in the store
	group := types.Group{
		ID: groupID,
		// Set other fields as needed
	}
	k.SetGroup(ctx, group)

	// Delete the group
	k.DeleteGroup(ctx, groupID)

	// Verify that the group is deleted
	_, err := k.GetGroup(ctx, groupID)
	s.Require().Error(err)
}

func (s *KeeperTestSuite) TestGetSetDKGContext() {
	ctx, k := s.ctx, s.app.TSSKeeper

	dkgContext := []byte("dkg-context sample")
	k.SetDKGContext(ctx, 1, dkgContext)

	got, err := k.GetDKGContext(ctx, 1)
	s.Require().NoError(err)
	s.Require().Equal(dkgContext, got)
}

func (s *KeeperTestSuite) TestGetSetMember() {
	ctx, k := s.ctx, s.app.TSSKeeper
	groupID, memberID := tss.GroupID(1), tss.MemberID(1)
	member := types.Member{
		ID:          1,
		GroupID:     groupID,
		Address:     "band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs",
		PubKey:      nil,
		IsMalicious: false,
	}
	k.SetMember(ctx, member)

	got, err := k.GetMember(ctx, groupID, memberID)
	s.Require().NoError(err)
	s.Require().Equal(member, got)
}

func (s *KeeperTestSuite) TestGetMembers() {
	ctx, k := s.ctx, s.app.TSSKeeper
	groupID := tss.GroupID(1)
	members := []types.Member{
		{
			ID:          1,
			GroupID:     groupID,
			Address:     "band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs",
			PubKey:      nil,
			IsMalicious: false,
		},
		{
			ID:          2,
			GroupID:     groupID,
			Address:     "band1p40yh3zkmhcv0ecqp3mcazy83sa57rgjp07dun",
			PubKey:      nil,
			IsMalicious: false,
		},
	}

	// Set members
	for _, m := range members {
		k.SetMember(ctx, m)
	}

	got, err := k.GetGroupMembers(ctx, groupID)
	s.Require().NoError(err)
	s.Require().Equal(members, got)
}

func (s *KeeperTestSuite) TestSetLastExpiredGroupID() {
	ctx, k := s.ctx, s.app.TSSKeeper
	groupID := tss.GroupID(1)
	k.SetLastExpiredGroupID(ctx, groupID)

	got := k.GetLastExpiredGroupID(ctx)
	s.Require().Equal(groupID, got)
}

func (s *KeeperTestSuite) TestGetSetLastExpiredGroupID() {
	ctx, k := s.ctx, s.app.TSSKeeper

	// Set the last expired group ID
	groupID := tss.GroupID(98765)
	k.SetLastExpiredGroupID(ctx, groupID)

	// Get the last expired group ID
	got := k.GetLastExpiredGroupID(ctx)

	// Assert equality
	s.Require().Equal(groupID, got)
}

func (s *KeeperTestSuite) TestProcessExpiredGroups() {
	ctx, k := s.ctx, s.app.TSSKeeper

	// Create group
	groupID := k.CreateNewGroup(ctx, types.Group{})
	k.SetMember(ctx, types.Member{
		ID:          1,
		GroupID:     groupID,
		Address:     "band1p40yh3zkmhcv0ecqp3mcazy83sa57rgjp07dun",
		PubKey:      nil,
		IsMalicious: false,
	})

	// Set the current block height
	blockHeight := int64(30001)
	ctx = ctx.WithBlockHeight(blockHeight)

	// Handle expired groups
	k.HandleExpiredGroups(ctx)

	// Assert that the last expired group ID is updated correctly
	lastExpiredGroupID := k.GetLastExpiredGroupID(ctx)
	s.Require().Equal(groupID, lastExpiredGroupID)
}

func (s *KeeperTestSuite) TestGetSetPendingProcessGroups() {
	ctx, k := s.ctx, s.app.TSSKeeper
	groupID := tss.GroupID(1)

	// Set the pending process group in the store
	k.SetPendingProcessGroups(ctx, types.PendingProcessGroups{
		GroupIDs: []tss.GroupID{groupID},
	})

	got := k.GetPendingProcessGroups(ctx)

	// Check if the retrieved pending process groups match the original sample
	s.Require().Len(got, 1)
	s.Require().Equal(groupID, got[0])
}

func (s *KeeperTestSuite) TestHandleProcessGroup() {
	ctx, k := s.ctx, s.app.TSSKeeper
	groupID, memberID := tss.GroupID(1), tss.MemberID(1)
	member := types.Member{
		ID:          memberID,
		GroupID:     groupID,
		IsMalicious: false,
	}

	k.SetMember(ctx, member)

	k.SetGroup(ctx, types.Group{
		ID:     groupID,
		Status: types.GROUP_STATUS_ROUND_1,
	})
	k.HandleProcessGroup(ctx, groupID)
	group := k.MustGetGroup(ctx, groupID)
	s.Require().Equal(types.GROUP_STATUS_ROUND_2, group.Status)

	k.SetGroup(ctx, types.Group{
		ID:     groupID,
		Status: types.GROUP_STATUS_ROUND_2,
	})
	k.HandleProcessGroup(ctx, groupID)
	group = k.MustGetGroup(ctx, groupID)
	s.Require().Equal(types.GROUP_STATUS_ROUND_3, group.Status)

	k.SetGroup(ctx, types.Group{
		ID:     groupID,
		Status: types.GROUP_STATUS_FALLEN,
	})
	k.HandleProcessGroup(ctx, groupID)
	group = k.MustGetGroup(ctx, groupID)
	s.Require().Equal(types.GROUP_STATUS_FALLEN, group.Status)

	k.SetGroup(ctx, types.Group{
		ID:     groupID,
		Status: types.GROUP_STATUS_ROUND_3,
	})
	k.HandleProcessGroup(ctx, groupID)
	group = k.MustGetGroup(ctx, groupID)
	s.Require().Equal(types.GROUP_STATUS_ACTIVE, group.Status)

	// if member is malicious
	k.SetGroup(ctx, types.Group{
		ID:     groupID,
		Status: types.GROUP_STATUS_ROUND_3,
	})
	member.IsMalicious = true
	k.SetMember(ctx, member)
	k.HandleProcessGroup(ctx, groupID)
	group = k.MustGetGroup(ctx, groupID)
	s.Require().Equal(types.GROUP_STATUS_FALLEN, group.Status)
}

func (s *KeeperTestSuite) TestGetSetReplacementCount() {
	ctx, k := s.ctx, s.app.TSSKeeper
	k.SetReplacementCount(ctx, 1)

	replacementCount := k.GetReplacementCount(ctx)
	s.Require().Equal(uint64(1), replacementCount)
}

func (s *KeeperTestSuite) TestGetNextReplacementID() {
	ctx, k := s.ctx, s.app.TSSKeeper

	// Initial replacement count
	k.SetReplacementCount(ctx, 1)

	replacementCount1 := k.GetNextReplacementCount(ctx)
	s.Require().Equal(uint64(2), replacementCount1)
	replacementCount2 := k.GetNextReplacementCount(ctx)
	s.Require().Equal(uint64(3), replacementCount2)
}

func (s *KeeperTestSuite) TestGetSetReplacement() {
	ctx, k := s.ctx, s.app.TSSKeeper

	// Create a replacement
	replacement := types.Replacement{
		ID:          1,
		SigningID:   1,
		FromGroupID: 2,
		ToGroupID:   1,
		FromPubKey:  []byte("test_pub_key"),
		ToPubKey:    []byte("test_pub_key"),
		Status:      types.REPLACEMENT_STATUS_WAITING,
		ExecTime:    time.Now().UTC(),
	}

	// Set the replacement using SetReplacement
	k.SetReplacement(ctx, replacement)

	// Get the stored replacement using GetReplacement
	got, err := k.GetReplacement(ctx, replacement.ID)
	s.Require().NoError(err)
	s.Require().Equal(replacement, got)
}

func (s *KeeperTestSuite) TestReplacementQueues() {
	ctx, msgSrvr, k := s.ctx, s.msgSrvr, s.app.TSSKeeper

	replacementID := uint64(1)

	s.SetupGroup(types.GROUP_STATUS_ACTIVE)

	now := time.Now()

	_, err := msgSrvr.ReplaceGroup(ctx, &types.MsgReplaceGroup{
		FromGroupID: 2,
		ToGroupID:   1,
		ExecTime:    now,
		Authority:   s.authority.String(),
	})
	s.Require().NoError(err)

	replacement, err := k.GetReplacement(ctx, replacementID)
	s.Require().NoError(err)

	replacementIterator := k.ReplacementQueueIterator(ctx, now)
	s.Require().True(replacementIterator.Valid())

	gotReplacementID, _ := types.SplitReplacementQueueKey(replacementIterator.Key())
	s.Require().Equal(replacement.ID, gotReplacementID)

	replacementIterator.Close()
}

func (s *KeeperTestSuite) TestSuccessHandleReplaceGroup() {
	ctx, k := s.ctx, s.app.TSSKeeper
	signingID := tss.SigningID(1)
	fromGroupID := tss.GroupID(1)
	toGroupID := tss.GroupID(2)
	replacementID := uint64(1)

	// Set up initial state for testing
	initialFromGroup := types.Group{
		ID:            fromGroupID,
		Size_:         7,
		Threshold:     4,
		PubKey:        testutil.HexDecode("02a37461c1621d12f2c436b98ffe95d6ff0fedc102e8b5b35a08c96b889cb448fd"),
		Status:        types.GROUP_STATUS_ACTIVE,
		Fee:           sdk.NewCoins(sdk.NewInt64Coin("uband", 15)),
		CreatedHeight: 2,
	}
	initialToGroup := types.Group{
		ID:            toGroupID,
		Size_:         5,
		Threshold:     3,
		PubKey:        testutil.HexDecode("0260aa1c85288f77aeaba5d02e984d987b16dd7f6722544574a03d175b48d8b83b"),
		Status:        types.GROUP_STATUS_ACTIVE,
		Fee:           sdk.NewCoins(sdk.NewInt64Coin("uband", 10)),
		CreatedHeight: 1,
	}
	initialSigning := types.Signing{
		ID:     signingID,
		Status: types.SIGNING_STATUS_SUCCESS,
		// ... other fields ...
	}
	initialReplacement := types.Replacement{
		ID:          replacementID,
		SigningID:   signingID,
		FromGroupID: initialFromGroup.ID,
		FromPubKey:  initialFromGroup.PubKey,
		ToGroupID:   initialToGroup.ID,
		ToPubKey:    initialToGroup.PubKey,
		Status:      types.REPLACEMENT_STATUS_WAITING,
		ExecTime:    time.Now(),
	}
	k.SetGroup(ctx, initialFromGroup)
	k.SetGroup(ctx, initialToGroup)
	k.SetSigning(ctx, initialSigning)

	k.SetReplacement(ctx, types.Replacement{})

	// Call HandleReplaceGroup to process the pending replace group
	k.HandleReplaceGroup(ctx, initialReplacement)

	// Verify that the fromGroup was replaced with the toGroup's data
	updatedGroup := k.MustGetGroup(ctx, toGroupID)
	// Verify unchanged data
	s.Require().Equal(toGroupID, updatedGroup.ID)
	s.Require().Equal(initialToGroup.CreatedHeight, updatedGroup.CreatedHeight)
	s.Require().Equal(initialToGroup.LatestReplacementID, updatedGroup.LatestReplacementID)
	// Verify changed data
	s.Require().Equal(initialFromGroup.Size_, updatedGroup.Size_)
	s.Require().Equal(initialFromGroup.Threshold, updatedGroup.Threshold)
	s.Require().Equal(initialFromGroup.PubKey, updatedGroup.PubKey)
	s.Require().Equal(initialFromGroup.Status, updatedGroup.Status)
	s.Require().Equal(initialFromGroup.Fee, updatedGroup.Fee)
}

func (s *KeeperTestSuite) TestFailedHandleReplaceGroup() {
	ctx, k := s.ctx, s.app.TSSKeeper
	signingID := tss.SigningID(1)
	fromGroupID := tss.GroupID(1)
	toGroupID := tss.GroupID(2)
	replacementID := uint64(1)

	// Set up initial state for testing
	initialFromGroup := types.Group{
		ID:            fromGroupID,
		Size_:         7,
		Threshold:     4,
		PubKey:        testutil.HexDecode("02a37461c1621d12f2c436b98ffe95d6ff0fedc102e8b5b35a08c96b889cb448fd"),
		Status:        types.GROUP_STATUS_ACTIVE,
		Fee:           sdk.NewCoins(sdk.NewInt64Coin("uband", 15)),
		CreatedHeight: 2,
	}
	initialToGroup := types.Group{
		ID:            toGroupID,
		Size_:         5,
		Threshold:     3,
		PubKey:        testutil.HexDecode("0260aa1c85288f77aeaba5d02e984d987b16dd7f6722544574a03d175b48d8b83b"),
		Status:        types.GROUP_STATUS_ACTIVE,
		Fee:           sdk.NewCoins(sdk.NewInt64Coin("uband", 10)),
		CreatedHeight: 1,
	}
	initialSigning := types.Signing{
		ID:     signingID,
		Status: types.SIGNING_STATUS_FALLEN,
		// ... other fields ...
	}
	initialReplacement := types.Replacement{
		ID:          replacementID,
		SigningID:   signingID,
		FromGroupID: initialFromGroup.ID,
		FromPubKey:  initialFromGroup.PubKey,
		ToGroupID:   initialToGroup.ID,
		ToPubKey:    initialToGroup.PubKey,
		Status:      types.REPLACEMENT_STATUS_WAITING,
		ExecTime:    time.Now(),
	}
	k.SetGroup(ctx, initialFromGroup)
	k.SetGroup(ctx, initialToGroup)
	k.SetSigning(ctx, initialSigning)

	// Call HandleReplaceGroup to process the pending replace group
	k.HandleReplaceGroup(ctx, initialReplacement)

	// Verify that the fromGroup is not replaced by the toGroup's
	updatedGroup := k.MustGetGroup(ctx, toGroupID)
	s.Require().Equal(initialToGroup, updatedGroup)
}

func (s *KeeperTestSuite) TestParams() {
	k := s.app.TSSKeeper

	testCases := []struct {
		name         string
		input        types.Params
		expectErr    bool
		expectErrStr string
	}{
		{
			name: "set invalid params",
			input: types.Params{
				MaxGroupSize:            0,
				MaxDESize:               0,
				CreatingPeriod:          1,
				SigningPeriod:           1,
				ActiveDuration:          time.Duration(0),
				InactivePenaltyDuration: time.Duration(0),
				JailPenaltyDuration:     time.Duration(0),
				RewardPercentage:        0,
			},
			expectErr:    true,
			expectErrStr: "must be positive:",
		},
		{
			name: "set full valid params",
			input: types.Params{
				MaxGroupSize:            types.DefaultMaxGroupSize,
				MaxDESize:               types.DefaultMaxDESize,
				CreatingPeriod:          types.DefaultCreatingPeriod,
				SigningPeriod:           types.DefaultSigningPeriod,
				ActiveDuration:          types.DefaultActiveDuration,
				InactivePenaltyDuration: types.DefaultInactivePenaltyDuration,
				JailPenaltyDuration:     types.DefaultJailPenaltyDuration,
				RewardPercentage:        types.DefaultRewardPercentage,
			},
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			expected := k.GetParams(s.ctx)
			err := k.SetParams(s.ctx, tc.input)
			if tc.expectErr {
				s.Require().Error(err)
				s.Require().ErrorContains(err, tc.expectErrStr)
			} else {
				expected = tc.input
				s.Require().NoError(err)
			}

			p := k.GetParams(s.ctx)
			s.Require().Equal(expected, p)
		})
	}
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
