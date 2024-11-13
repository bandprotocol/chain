package keeper_test

import (
	"github.com/bandprotocol/chain/v3/pkg/tss"
	"github.com/bandprotocol/chain/v3/x/tss/types"
)

func (s *KeeperTestSuite) TestHandleProcessGroup() {
	ctx, k := s.ctx, s.keeper
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

func (s *KeeperTestSuite) TestProcessExpiredGroups() {
	ctx, k := s.ctx, s.keeper

	// handle expired group with no group is requested
	k.HandleExpiredGroups(ctx)
	lastExpiredGroupID := k.GetLastExpiredGroupID(ctx)
	s.Require().Equal(tss.GroupID(0), lastExpiredGroupID)

	// Create group
	groupID := k.AddGroup(ctx, 3, 2, "test")
	k.SetMember(ctx, types.Member{
		ID:          1,
		GroupID:     groupID,
		Address:     "band1p40yh3zkmhcv0ecqp3mcazy83sa57rgjp07dun",
		PubKey:      nil,
		IsMalicious: false,
	})

	// Set the current block height
	ctx = ctx.WithBlockHeight(30001)

	// Handle expired groups
	k.HandleExpiredGroups(ctx)

	// Assert that the last expired group ID is updated correctly
	lastExpiredGroupID = k.GetLastExpiredGroupID(ctx)
	s.Require().Equal(groupID, lastExpiredGroupID)

	// Set to the next block height
	blockHeight := int64(30002)
	ctx = ctx.WithBlockHeight(blockHeight)

	// Handle expired groups
	k.HandleExpiredGroups(ctx)

	// Assert that the last expired group ID is updated correctly
	lastExpiredGroupID = k.GetLastExpiredGroupID(ctx)
	s.Require().Equal(groupID, lastExpiredGroupID)
}

func (s *KeeperTestSuite) TestProcessExpiredOnlyFirstGroups() {
	ctx, k := s.ctx, s.keeper

	// Create a group
	_ = k.AddGroup(ctx, 3, 2, "test")

	// skip to some height and create a new group
	ctx = ctx.WithBlockHeight(int64(types.DefaultCreationPeriod) / 2)
	_ = k.AddGroup(ctx, 3, 2, "test")

	// Set the current block height at which the first group is expired
	ctx = ctx.WithBlockHeight(int64(types.DefaultCreationPeriod) + 1)

	// Handle expired groups
	k.HandleExpiredGroups(ctx)

	// Assert that the last expired group ID is updated correctly
	lastExpiredGroupID := k.GetLastExpiredGroupID(ctx)
	s.Require().Equal(tss.GroupID(1), lastExpiredGroupID)

	count := k.GetGroupCount(ctx)
	s.Require().Equal(uint64(2), count)
}

func (s *KeeperTestSuite) TestSetLastExpiredGroupID() {
	ctx, k := s.ctx, s.keeper
	groupID := tss.GroupID(1)
	k.SetLastExpiredGroupID(ctx, groupID)

	got := k.GetLastExpiredGroupID(ctx)
	s.Require().Equal(groupID, got)
}

func (s *KeeperTestSuite) TestGetSetLastExpiredGroupID() {
	ctx, k := s.ctx, s.keeper

	// Set the last expired group ID
	groupID := tss.GroupID(98765)
	k.SetLastExpiredGroupID(ctx, groupID)

	// Get the last expired group ID
	got := k.GetLastExpiredGroupID(ctx)

	// Assert equality
	s.Require().Equal(groupID, got)
}

func (s *KeeperTestSuite) TestGetSetPendingProcessGroups() {
	ctx, k := s.ctx, s.keeper
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

func (s *KeeperTestSuite) TestDeleteAllDKGInterimData() {
	ctx, k := s.ctx, s.keeper
	groupID := tss.GroupID(1)
	groupSize := uint64(5)
	groupThreshold := uint64(3)
	dkgContext := []byte("dkg-context")

	// Assuming you have corresponding Set methods for each Delete method
	// Setting up initial state
	k.SetDKGContext(ctx, groupID, dkgContext)

	for i := uint64(1); i <= groupSize; i++ {
		memberID := tss.MemberID(i)
		round1Info := types.Round1Info{
			MemberID: memberID,
			CoefficientCommits: tss.Points{
				[]byte("point1"),
				[]byte("point2"),
			},
			OneTimePubKey:    []byte("OneTimePubKeySimple"),
			A0Signature:      []byte("A0SignatureSimple"),
			OneTimeSignature: []byte("OneTimeSignatureSimple"),
		}
		round2Info := types.Round2Info{
			MemberID: memberID,
			EncryptedSecretShares: tss.EncSecretShares{
				[]byte("e_12"),
				[]byte("e_13"),
				[]byte("e_14"),
			},
		}
		complainWithStatus := types.ComplaintsWithStatus{
			MemberID: memberID,
			ComplaintsWithStatus: []types.ComplaintWithStatus{
				{
					Complaint: types.Complaint{
						Complainant: 1,
						Respondent:  2,
						KeySym:      []byte("key_sym"),
						Signature:   []byte("signature"),
					},
					ComplaintStatus: types.COMPLAINT_STATUS_SUCCESS,
				},
			},
		}
		confirm := types.Confirm{
			MemberID:     memberID,
			OwnPubKeySig: []byte("own_pub_key_sig"),
		}

		k.AddRound1Info(ctx, groupID, round1Info)
		k.AddRound2Info(ctx, groupID, round2Info)
		k.AddComplaintsWithStatus(ctx, groupID, complainWithStatus)
		k.AddConfirm(ctx, groupID, confirm)
	}

	for i := uint64(0); i < groupThreshold; i++ {
		k.SetAccumulatedCommit(ctx, groupID, i, []byte("point1"))
	}

	// Delete all interim data
	k.DeleteAllDKGInterimData(ctx, groupID)

	// Check if all data is deleted
	s.Require().Nil(k.GetDKGContext(ctx, groupID))

	for i := uint64(1); i <= groupSize; i++ {
		memberID := tss.MemberID(i)

		gotRound1Info, err := k.GetRound1Info(ctx, groupID, memberID)
		s.Require().ErrorIs(types.ErrRound1InfoNotFound, err)
		s.Require().Empty(types.Round1Info{}, gotRound1Info)

		gotRound2Info, err := k.GetRound2Info(ctx, groupID, memberID)
		s.Require().ErrorIs(types.ErrRound2InfoNotFound, err)
		s.Require().Empty(types.Round2Info{}, gotRound2Info)

		gotComplaintWithStatus, err := k.GetComplaintsWithStatus(ctx, groupID, memberID)
		s.Require().ErrorIs(types.ErrComplaintsWithStatusNotFound, err)
		s.Require().Empty(types.ComplaintWithStatus{}, gotComplaintWithStatus)
	}

	for i := uint64(0); i < groupThreshold; i++ {
		s.Require().Empty(tss.Point{}, k.GetAccumulatedCommit(ctx, groupID, i))
	}

	s.Require().Equal(uint64(0), k.GetRound1InfoCount(ctx, groupID))
	s.Require().Equal(uint64(0), k.GetRound2InfoCount(ctx, groupID))
	s.Require().Equal(uint64(0), k.GetConfirmComplainCount(ctx, groupID))
}
