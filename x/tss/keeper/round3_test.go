package keeper_test

import (
	"fmt"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/pkg/tss/testutil"
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

func (s *KeeperTestSuite) TestHandleVerifyComplain() {
	ctx, k := s.ctx, s.app.TSSKeeper

	for _, tc := range testutil.TestCases {
		s.Run(fmt.Sprintf("Case %s", tc.Name), func() {
			for _, m := range tc.Group.Members {
				// Set member
				k.SetMember(ctx, tc.Group.ID, m.ID, types.Member{
					Address:     "member_address",
					PubKey:      m.PubKey(),
					IsMalicious: false,
				})

				// Set round 1 info
				k.SetRound1Info(ctx, tc.Group.ID, types.Round1Info{
					MemberID:           m.ID,
					CoefficientsCommit: m.CoefficientsCommit,
					OneTimePubKey:      m.OneTimePubKey(),
					A0Sig:              m.A0Sig,
					OneTimeSig:         m.OneTimeSig,
				})

				// Set round 2 info
				k.SetRound2Info(ctx, tc.Group.ID, types.Round2Info{
					MemberID:              m.ID,
					EncryptedSecretShares: m.EncSecretShares,
				})
			}

			memberI := tc.Group.Members[0]
			memberJ := tc.Group.Members[1]
			iSlot := types.FindMemberSlot(memberI.ID, memberJ.ID)
			jSlot := types.FindMemberSlot(memberJ.ID, memberI.ID)

			// Failed case - correct encrypted secret share
			err := k.HandleVerifyComplaint(ctx, tc.Group.ID, types.Complaint{
				Complainer:  memberI.ID,
				Complainant: memberJ.ID,
				KeySym:      memberI.KeySyms[iSlot],
				Signature:   memberI.ComplaintSigs[iSlot],
			})
			s.Require().Error(err)

			// Get round 2 info Complainant
			round2J, err := k.GetRound2Info(ctx, tc.Group.ID, memberJ.ID)
			s.Require().NoError(err)

			// Set fake encrypted secret shares
			round2J.EncryptedSecretShares[jSlot] = testutil.FakePrivKey
			k.SetRound2Info(ctx, tc.Group.ID, round2J)

			// Success case - wrong encrypted secret share
			err = k.HandleVerifyComplaint(ctx, tc.Group.ID, types.Complaint{
				Complainer:  memberI.ID,
				Complainant: memberJ.ID,
				KeySym:      memberI.KeySyms[iSlot],
				Signature:   memberI.ComplaintSigs[iSlot],
			})
			s.Require().NoError(err)
		})
	}
}

func (s *KeeperTestSuite) TestHandleVerifyOwnPubKeySig() {
	ctx, k := s.ctx, s.app.TSSKeeper

	for _, tc := range testutil.TestCases {
		// Set dkg context
		k.SetDKGContext(ctx, tc.Group.ID, tc.Group.DKGContext)

		for _, m := range tc.Group.Members {
			// Set member
			k.SetMember(ctx, tc.Group.ID, m.ID, types.Member{
				Address:     "member_address",
				PubKey:      m.PubKey(),
				IsMalicious: false,
			})

			// Sign
			sig, err := tss.SignOwnPubkey(m.ID, tc.Group.DKGContext, m.PubKey(), m.PrivKey)
			s.Require().NoError(err)

			// Verify own public key signature
			err = k.HandleVerifyOwnPubKeySig(ctx, tc.Group.ID, m.ID, sig)
			s.Require().NoError(err)
		}
	}
}

func (s *KeeperTestSuite) TestGetSetComplaintsWithStatus() {
	ctx, k := s.ctx, s.app.TSSKeeper
	groupID := tss.GroupID(1)
	memberID := tss.MemberID(1)
	complaintWithStatus := types.ComplaintsWithStatus{
		MemberID: memberID,
		ComplaintsWithStatus: []types.ComplaintWithStatus{
			{
				Complaint: types.Complaint{
					Complainer:  1,
					Complainant: 2,
					KeySym:      []byte("key_sym"),
					Signature:   []byte("signature"),
				},
				ComplaintStatus: types.COMPLAINT_STATUS_SUCCESS,
			},
		},
	}

	// Set complaints with status
	k.SetComplaintsWithStatus(ctx, groupID, complaintWithStatus)

	got, err := k.GetComplaintsWithStatus(ctx, groupID, memberID)
	s.Require().NoError(err)
	s.Require().Equal(complaintWithStatus, got)
}

func (s *KeeperTestSuite) TestDeleteComplainsWithStatus() {
	ctx, k := s.ctx, s.app.TSSKeeper
	groupID := tss.GroupID(1)
	memberID := tss.MemberID(1)
	complainWithStatus := types.ComplaintsWithStatus{
		MemberID: memberID,
		ComplaintsWithStatus: []types.ComplaintWithStatus{
			{
				Complaint: types.Complaint{
					Complainer:  1,
					Complainant: 2,
					KeySym:      []byte("key_sym"),
					Signature:   []byte("signature"),
				},
				ComplaintStatus: types.COMPLAINT_STATUS_SUCCESS,
			},
		},
	}

	// Set complaints with status
	k.SetComplaintsWithStatus(ctx, groupID, complainWithStatus)
	// Delete complaints with status
	k.DeleteComplainsWithStatus(ctx, groupID, memberID)

	_, err := k.GetComplaintsWithStatus(ctx, groupID, memberID)
	s.Require().Error(err)
}

func (s *KeeperTestSuite) TestGetAllComplainsWithStatus() {
	ctx, k := s.ctx, s.app.TSSKeeper
	groupID := tss.GroupID(1)
	memberID1 := tss.MemberID(1)
	memberID2 := tss.MemberID(2)
	complainWithStatus1 := types.ComplaintsWithStatus{
		MemberID: memberID1,
		ComplaintsWithStatus: []types.ComplaintWithStatus{
			{
				Complaint: types.Complaint{
					Complainer:  1,
					Complainant: 2,
					KeySym:      []byte("key_sym"),
					Signature:   []byte("signature"),
				},
				ComplaintStatus: types.COMPLAINT_STATUS_SUCCESS,
			},
		},
	}
	complainWithStatus2 := types.ComplaintsWithStatus{
		MemberID: memberID2,
		ComplaintsWithStatus: []types.ComplaintWithStatus{
			{
				Complaint: types.Complaint{
					Complainer:  1,
					Complainant: 2,
					KeySym:      []byte("key_sym"),
					Signature:   []byte("signature"),
				},
				ComplaintStatus: types.COMPLAINT_STATUS_SUCCESS,
			},
		},
	}

	// Set complaints with status
	k.SetComplaintsWithStatus(ctx, groupID, complainWithStatus1)
	k.SetComplaintsWithStatus(ctx, groupID, complainWithStatus2)

	got := k.GetAllComplainsWithStatus(ctx, groupID)
	s.Require().Equal([]types.ComplaintsWithStatus{complainWithStatus1, complainWithStatus2}, got)
}

func (s *KeeperTestSuite) TestGetSetConfirm() {
	ctx, k := s.ctx, s.app.TSSKeeper
	groupID := tss.GroupID(1)
	memberID := tss.MemberID(1)
	confirm := types.Confirm{
		MemberID:     memberID,
		OwnPubKeySig: []byte("own_pub_key_sig"),
	}

	// Set confirm
	k.SetConfirm(ctx, groupID, confirm)

	got, err := k.GetConfirm(ctx, groupID, memberID)
	s.Require().NoError(err)
	s.Require().Equal(confirm, got)

	// Get confirm or complain count
	count := k.GetConfirmComplainCount(ctx, groupID)
	s.Require().Equal(uint64(1), count)
}

func (s *KeeperTestSuite) TestDeleteConfirm() {
	ctx, k := s.ctx, s.app.TSSKeeper
	groupID := tss.GroupID(1)
	memberID := tss.MemberID(1)
	confirm := types.Confirm{
		MemberID:     memberID,
		OwnPubKeySig: []byte("own_pub_key_sig"),
	}

	// Set confirm
	k.SetConfirm(ctx, groupID, confirm)

	// Delete confirm
	k.DeleteConfirm(ctx, groupID, memberID)

	_, err := k.GetConfirm(ctx, groupID, memberID)
	s.Require().Error(err)
}

func (s *KeeperTestSuite) TestGetConfirms() {
	ctx, k := s.ctx, s.app.TSSKeeper
	groupID := tss.GroupID(1)
	memberID1 := tss.MemberID(1)
	memberID2 := tss.MemberID(2)
	confirm1 := types.Confirm{
		MemberID:     memberID1,
		OwnPubKeySig: []byte("own_pub_key_sig"),
	}
	confirm2 := types.Confirm{
		MemberID:     memberID2,
		OwnPubKeySig: []byte("own_pub_key_sig"),
	}

	// Set confirm
	k.SetConfirm(ctx, groupID, confirm1)
	k.SetConfirm(ctx, groupID, confirm2)

	got := k.GetConfirms(ctx, groupID)
	s.Require().Equal([]types.Confirm{confirm1, confirm2}, got)
}

func (s *KeeperTestSuite) TestGetSetConfirmComplainCount() {
	ctx, k := s.ctx, s.app.TSSKeeper
	groupID := tss.GroupID(1)
	count := uint64(1)

	// Get confirm complain count before assign
	got1 := k.GetConfirmComplainCount(ctx, groupID)
	s.Require().Equal(uint64(0), got1)

	// Set confirm complain count
	k.SetConfirmComplainCount(ctx, groupID, count)

	// Get confirm complain count
	got2 := k.GetConfirmComplainCount(ctx, groupID)
	s.Require().Equal(count, got2)
}

func (s *KeeperTestSuite) TestDeleteConfirmComplainCount() {
	ctx, k := s.ctx, s.app.TSSKeeper
	groupID := tss.GroupID(1)
	count := uint64(5)

	// Set confirm complain count
	k.SetConfirmComplainCount(ctx, groupID, count)

	// Delete confirm complain count
	k.DeleteConfirmComplainCount(ctx, groupID)

	got := k.GetConfirmComplainCount(ctx, groupID)
	s.Require().Empty(got)
}

func (s *KeeperTestSuite) TestMarkMalicious() {
	ctx, k := s.ctx, s.app.TSSKeeper
	groupID := tss.GroupID(1)
	memberID := tss.MemberID(1)

	// Set member
	k.SetMember(ctx, groupID, memberID, types.Member{
		Address:     "member_address",
		PubKey:      []byte("pub_key"),
		IsMalicious: false,
	})

	// Mark malicious
	err := k.MarkMalicious(ctx, groupID, memberID)
	s.Require().NoError(err)

	// Get members
	members, err := k.GetMembers(ctx, groupID)
	s.Require().NoError(err)

	got := types.HaveMalicious(members)
	s.Require().Equal(got, true)
}

func (s *KeeperTestSuite) TestDeleteAllDKGInterimData() {
	ctx, k := s.ctx, s.app.TSSKeeper
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
			CoefficientsCommit: tss.Points{
				[]byte("point1"),
				[]byte("point2"),
			},
			OneTimePubKey: []byte("OneTimePubKeySimple"),
			A0Sig:         []byte("A0SigSimple"),
			OneTimeSig:    []byte("OneTimeSigSimple"),
		}
		round2Info := types.Round2Info{
			MemberID: memberID,
			EncryptedSecretShares: tss.Scalars{
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
						Complainer:  1,
						Complainant: 2,
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

		k.SetRound1Info(ctx, groupID, round1Info)
		k.SetRound2Info(ctx, groupID, round2Info)
		k.SetComplaintsWithStatus(ctx, groupID, complainWithStatus)
		k.SetConfirm(ctx, groupID, confirm)
	}

	for i := uint64(0); i < groupThreshold; i++ {
		k.SetAccumulatedCommit(ctx, groupID, i, []byte("point1"))
	}

	k.SetRound1InfoCount(ctx, groupID, 1)
	k.SetRound2InfoCount(ctx, groupID, 1)
	k.SetConfirmComplainCount(ctx, groupID, 1)

	// Delete all interim data
	k.DeleteAllDKGInterimData(ctx, groupID, groupSize, groupThreshold)

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
		s.Require().ErrorIs(types.ErrComplainsWithStatusNotFound, err)
		s.Require().Empty(types.ComplaintWithStatus{}, gotComplaintWithStatus)
	}

	for i := uint64(0); i < groupThreshold; i++ {
		s.Require().Empty(tss.Point{}, k.GetAccumulatedCommit(ctx, groupID, i))
	}

	s.Require().Equal(uint64(0), k.GetRound1InfoCount(ctx, groupID))
	s.Require().Equal(uint64(0), k.GetRound2InfoCount(ctx, groupID))
	s.Require().Equal(uint64(0), k.GetConfirmComplainCount(ctx, groupID))
}
