package keeper_test

import (
	"fmt"

	"github.com/bandprotocol/chain/v3/pkg/tss"
	"github.com/bandprotocol/chain/v3/pkg/tss/testutil"
	"github.com/bandprotocol/chain/v3/x/tss/types"
)

func (s *KeeperTestSuite) TestHandleVerifyComplain() {
	ctx, k := s.ctx, s.keeper

	for _, tc := range testutil.TestCases {
		s.Run(fmt.Sprintf("Case %s", tc.Name), func() {
			for _, m := range tc.Group.Members {
				// Set member
				k.SetMember(ctx, types.Member{
					ID:          m.ID,
					GroupID:     tc.Group.ID,
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
			err := k.VerifyComplaint(ctx, tc.Group.ID, types.Complaint{
				Complainant: complainant.ID,
				Respondent:  respondent.ID,
				KeySym:      complainant.KeySyms[complainantSlot],
				Signature:   complainant.ComplaintSignatures[complainantSlot],
			})
			s.Require().Error(err)

			// Get respondent round 2 info
			respondentRound2, err := k.GetRound2Info(ctx, tc.Group.ID, respondent.ID)
			s.Require().NoError(err)

			// Set fake encrypted secret shares
			respondentRound2.EncryptedSecretShares[respondentSlot] = testutil.FalseEncSecretShare
			k.AddRound2Info(ctx, tc.Group.ID, respondentRound2)

			// Success case - wrong encrypted secret share
			err = k.VerifyComplaint(ctx, tc.Group.ID, types.Complaint{
				Complainant: complainant.ID,
				Respondent:  respondent.ID,
				KeySym:      complainant.KeySyms[complainantSlot],
				Signature:   complainant.ComplaintSignatures[complainantSlot],
			})
			s.Require().NoError(err)
		})
	}
}

func (s *KeeperTestSuite) TestHandleVerifyOwnPubKeySig() {
	ctx, k := s.ctx, s.keeper

	for _, tc := range testutil.TestCases {
		// Set dkg context
		k.SetDKGContext(ctx, tc.Group.ID, tc.Group.DKGContext)

		for _, m := range tc.Group.Members {
			// Set member
			k.SetMember(ctx, types.Member{
				ID:          m.ID,
				GroupID:     tc.Group.ID,
				Address:     "member_address",
				PubKey:      m.PubKey(),
				IsMalicious: false,
			})

			// Sign
			sig, err := tss.SignOwnPubKey(m.ID, tc.Group.DKGContext, m.PubKey(), m.PrivKey)
			s.Require().NoError(err)

			// Verify own public key signature
			err = k.VerifyOwnPubKeySignature(ctx, tc.Group.ID, m.ID, sig)
			s.Require().NoError(err)
		}
	}
}

func (s *KeeperTestSuite) TestGetSetComplaintsWithStatus() {
	ctx, k := s.ctx, s.keeper
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
	ctx, k := s.ctx, s.keeper
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

func (s *KeeperTestSuite) TestGetAllComplainsWithStatus() {
	ctx, k := s.ctx, s.keeper
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
	ctx, k := s.ctx, s.keeper
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
	ctx, k := s.ctx, s.keeper
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

func (s *KeeperTestSuite) TestDeleteRound3Infos() {
	ctx, k := s.ctx, s.keeper
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

	confirm := types.Confirm{
		MemberID:     memberID,
		OwnPubKeySig: []byte("own_pub_key_sig"),
	}
	k.AddConfirm(ctx, groupID, confirm)

	// Delete complaints with status
	k.DeleteConfirmComplains(ctx, groupID)

	_, err := k.GetComplaintsWithStatus(ctx, groupID, memberID)
	s.Require().Error(err)

	_, err = k.GetConfirm(ctx, groupID, memberID)
	s.Require().Error(err)

	cnt := k.GetConfirmComplainCount(ctx, groupID)
	s.Require().Equal(uint64(0), cnt)
}

func (s *KeeperTestSuite) TestGetConfirms() {
	ctx, k := s.ctx, s.keeper
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
	ctx, k := s.ctx, s.keeper
	groupID := tss.GroupID(1)

	// Get confirm complain count before assign
	got1 := k.GetConfirmComplainCount(ctx, groupID)
	s.Require().Equal(uint64(0), got1)

	k.AddConfirm(ctx, groupID, types.Confirm{MemberID: tss.MemberID(1)})
	k.AddComplaintsWithStatus(ctx, groupID, types.ComplaintsWithStatus{MemberID: tss.MemberID(2)})

	// Get confirm complain count
	got2 := k.GetConfirmComplainCount(ctx, groupID)
	s.Require().Equal(uint64(2), got2)
}
