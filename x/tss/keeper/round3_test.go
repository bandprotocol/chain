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
				k.SetMember(ctx, tc.Group.ID, types.Member{
					MemberID:    m.ID,
					Address:     "member_address",
					PubKey:      m.PubKey(),
					IsMalicious: false,
				})

				// Add round 1 info
				k.AddRound1Info(ctx, tc.Group.ID, types.Round1Info{
					MemberID:           m.ID,
					CoefficientCommits: m.CoefficientCommits,
					OneTimePubKey:      m.OneTimePubKey(),
					A0Signature:        m.A0Signature,
					OneTimeSignature:   m.OneTimeSignature,
				})

				// Set round 2 info
				k.AddRound2Info(ctx, tc.Group.ID, types.Round2Info{
					MemberID:              m.ID,
					EncryptedSecretShares: m.EncSecretShares,
				})
			}

			complainant := tc.Group.Members[0]
			respondent := tc.Group.Members[1]
			complainantSlot := types.FindMemberSlot(complainant.ID, respondent.ID)
			respondentSlot := types.FindMemberSlot(respondent.ID, complainant.ID)

			// Failed case - correct encrypted secret share
			err := k.HandleVerifyComplaint(ctx, tc.Group.ID, types.Complaint{
				Complainant: complainant.ID,
				Respondent:  respondent.ID,
				KeySym:      complainant.KeySyms[complainantSlot],
				Signature:   complainant.ComplaintSigs[complainantSlot],
			})
			s.Require().Error(err)

			// Get respondent round 2 info
			respondentRound2, err := k.GetRound2Info(ctx, tc.Group.ID, respondent.ID)
			s.Require().NoError(err)

			// Set fake encrypted secret shares
			respondentRound2.EncryptedSecretShares[respondentSlot] = testutil.FakePrivKey
			k.AddRound2Info(ctx, tc.Group.ID, respondentRound2)

			// Success case - wrong encrypted secret share
			err = k.HandleVerifyComplaint(ctx, tc.Group.ID, types.Complaint{
				Complainant: complainant.ID,
				Respondent:  respondent.ID,
				KeySym:      complainant.KeySyms[complainantSlot],
				Signature:   complainant.ComplaintSigs[complainantSlot],
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
			k.SetMember(ctx, tc.Group.ID, types.Member{
				MemberID:    m.ID,
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
					Complainant: 1,
					Respondent:  2,
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

func (s *KeeperTestSuite) TestAddComplaintsWithStatus() {
	ctx, k := s.ctx, s.app.TSSKeeper
	groupID := tss.GroupID(1)
	memberID := tss.MemberID(1)
	complaintWithStatus := types.ComplaintsWithStatus{
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

	// Add complaints with status
	k.AddComplaintsWithStatus(ctx, groupID, complaintWithStatus)

	gotComplaintsWithStatus, err := k.GetComplaintsWithStatus(ctx, groupID, memberID)
	s.Require().NoError(err)
	s.Require().Equal(complaintWithStatus, gotComplaintsWithStatus)
	gotCount := k.GetConfirmComplainCount(ctx, groupID)
	s.Require().Equal(uint64(1), gotCount)
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
					Complainant: 1,
					Respondent:  2,
					KeySym:      []byte("key_sym"),
					Signature:   []byte("signature"),
				},
				ComplaintStatus: types.COMPLAINT_STATUS_SUCCESS,
			},
		},
	}

	// Add complaints with status
	k.AddComplaintsWithStatus(ctx, groupID, complainWithStatus)
	// Delete complaints with status
	k.DeleteComplainsWithStatus(ctx, groupID, memberID)

	_, err := k.GetComplaintsWithStatus(ctx, groupID, memberID)
	s.Require().Error(err)
}

func (s *KeeperTestSuite) TestDeleteAllComplainsWithStatus() {
	ctx, k := s.ctx, s.app.TSSKeeper
	groupID := tss.GroupID(1)
	memberID := tss.MemberID(1)
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

	// Add complaints with status
	k.AddComplaintsWithStatus(ctx, groupID, complainWithStatus)
	// Delete complaints with status
	k.DeleteAllComplainsWithStatus(ctx, groupID)

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
					Complainant: 1,
					Respondent:  2,
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
					Complainant: 1,
					Respondent:  2,
					KeySym:      []byte("key_sym"),
					Signature:   []byte("signature"),
				},
				ComplaintStatus: types.COMPLAINT_STATUS_SUCCESS,
			},
		},
	}

	// Set complaints with status
	k.AddComplaintsWithStatus(ctx, groupID, complainWithStatus1)
	k.AddComplaintsWithStatus(ctx, groupID, complainWithStatus2)

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
}

func (s *KeeperTestSuite) TestAddConfirm() {
	ctx, k := s.ctx, s.app.TSSKeeper
	groupID := tss.GroupID(1)
	memberID := tss.MemberID(1)
	confirm := types.Confirm{
		MemberID:     memberID,
		OwnPubKeySig: []byte("own_pub_key_sig"),
	}

	// Add confirm
	k.AddConfirm(ctx, groupID, confirm)

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

func (s *KeeperTestSuite) TestDeleteConfirms() {
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
	k.DeleteConfirms(ctx, groupID)

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

	// Add confirm
	k.AddConfirm(ctx, groupID, confirm1)
	k.AddConfirm(ctx, groupID, confirm2)

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
	k.SetMember(ctx, groupID, types.Member{
		MemberID:    memberID,
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

	got := types.Members(members).HaveMalicious()
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

	k.SetRound1InfoCount(ctx, groupID, 1)
	k.SetRound2InfoCount(ctx, groupID, 1)
	k.SetConfirmComplainCount(ctx, groupID, 1)

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
