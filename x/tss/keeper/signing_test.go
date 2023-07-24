package keeper_test

import (
	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/pkg/tss/testutil"
	"github.com/bandprotocol/chain/v2/x/tss/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (s *KeeperTestSuite) TestGetSetSigningCount() {
	ctx, k := s.ctx, s.app.TSSKeeper

	// Set signing count
	count := uint64(42)
	k.SetSigningCount(ctx, count)

	// Get signing count
	got := k.GetSigningCount(ctx)

	// Assert equality
	s.Require().Equal(count, got)
}

func (s *KeeperTestSuite) TestGetNextSigningID() {
	ctx, k := s.ctx, s.app.TSSKeeper

	// Get initial signing count
	initialCount := k.GetSigningCount(ctx)

	// Get next signing ID
	signingID := k.GetNextSigningID(ctx)

	// Get updated signing count
	updatedCount := k.GetSigningCount(ctx)

	// Assert that the signing ID is incremented and the signing count is updated
	s.Require().Equal(tss.SigningID(initialCount+1), signingID)
	s.Require().Equal(initialCount+1, updatedCount)
}

func (s *KeeperTestSuite) TestGetSetSigning() {
	ctx, k := s.ctx, s.app.TSSKeeper

	// Create a sample signing object
	signingID := tss.SigningID(1)
	groupID := tss.GroupID(1)
	member1 := tss.MemberID(1)
	signing := types.Signing{
		SigningID: signingID,
		GroupID:   groupID,
		AssignedMembers: []types.AssignedMember{
			{
				MemberID: member1,
				Member:   "band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs",
				PubD:     testutil.HexDecode("02234d901b8d6404b509e9926407d1a2749f456d18b159af647a65f3e907d61ef1"),
				PubE:     testutil.HexDecode("028a1f3e214831b2f2d6e27384817132ddaa222928b05e9372472aa2735cf1f797"),
				PubNonce: testutil.HexDecode("03cbb6a27c62baa195dff6c75eae7b6b7713f978732a671855f7d7b86b06e6ac67"),
			},
		},
		Message:       []byte("data"),
		GroupPubNonce: testutil.HexDecode("03fae45376abb0d60c3ae2b5caee749118125ec3d73725f3ad03b0b6e686d0f31a"),
		Signature:     nil,
	}

	// Set signing
	k.SetSigning(ctx, signing)

	// Get signing
	got, err := k.GetSigning(ctx, signingID)

	// Assert no error and equality
	s.Require().NoError(err)
	s.Require().Equal(signing, got)
}

func (s *KeeperTestSuite) TestAddSigning() {
	ctx, k := s.ctx, s.app.TSSKeeper

	// Create a sample signing object
	groupID := tss.GroupID(1)
	member1 := tss.MemberID(1)
	signing := types.Signing{
		GroupID: groupID,
		AssignedMembers: []types.AssignedMember{
			{
				MemberID: member1,
				Member:   "band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs",
				PubD:     testutil.HexDecode("02234d901b8d6404b509e9926407d1a2749f456d18b159af647a65f3e907d61ef1"),
				PubE:     testutil.HexDecode("028a1f3e214831b2f2d6e27384817132ddaa222928b05e9372472aa2735cf1f797"),
				PubNonce: testutil.HexDecode("03cbb6a27c62baa195dff6c75eae7b6b7713f978732a671855f7d7b86b06e6ac67"),
			},
		},
		Message:       []byte("data"),
		GroupPubNonce: testutil.HexDecode("03fae45376abb0d60c3ae2b5caee749118125ec3d73725f3ad03b0b6e686d0f31a"),
		Signature:     nil,
	}

	// Add signing
	signingID := k.AddSigning(ctx, signing)

	// Get added signing
	got, err := k.GetSigning(ctx, signingID)

	// Assert no error and equality
	s.Require().NoError(err)
	s.Require().Equal(signingID, got.SigningID)
}

func (s *KeeperTestSuite) TestDeleteSigning() {
	ctx, k := s.ctx, s.app.TSSKeeper

	// Create a sample signing object
	signingID := tss.SigningID(1)
	groupID := tss.GroupID(1)
	member1 := tss.MemberID(1)
	signing := types.Signing{
		SigningID: signingID,
		GroupID:   groupID,
		AssignedMembers: []types.AssignedMember{
			{
				MemberID: member1,
				Member:   "band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs",
				PubD:     testutil.HexDecode("02234d901b8d6404b509e9926407d1a2749f456d18b159af647a65f3e907d61ef1"),
				PubE:     testutil.HexDecode("028a1f3e214831b2f2d6e27384817132ddaa222928b05e9372472aa2735cf1f797"),
				PubNonce: testutil.HexDecode("03cbb6a27c62baa195dff6c75eae7b6b7713f978732a671855f7d7b86b06e6ac67"),
			},
		},
		Message:       []byte("data"),
		GroupPubNonce: testutil.HexDecode("03fae45376abb0d60c3ae2b5caee749118125ec3d73725f3ad03b0b6e686d0f31a"),
		Signature:     nil,
	}

	// Set signing
	k.SetSigning(ctx, signing)

	// Delete the signing
	k.DeleteSigning(ctx, signingID)

	// Verify that the signing data is deleted
	_, err := k.GetSigning(ctx, signingID)
	s.Require().Error(err)
}

func (s *KeeperTestSuite) TestGetPendingSigns() {
	ctx, k := s.ctx, s.app.TSSKeeper
	memberID := tss.MemberID(1)
	address := "band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs"

	signing := types.Signing{
		AssignedMembers: []types.AssignedMember{
			{
				MemberID: memberID,
				Member:   address,
				PubD:     testutil.HexDecode("02234d901b8d6404b509e9926407d1a2749f456d18b159af647a65f3e907d61ef1"),
				PubE:     testutil.HexDecode("028a1f3e214831b2f2d6e27384817132ddaa222928b05e9372472aa2735cf1f797"),
				PubNonce: testutil.HexDecode("03cbb6a27c62baa195dff6c75eae7b6b7713f978732a671855f7d7b86b06e6ac67"),
			},
		},
	}

	// Set signing
	signingID := k.AddSigning(ctx, signing)

	// Get all PendingSignIDs
	got := k.GetPendingSigns(ctx, address)

	// Update signing id
	signing.SigningID = signingID

	// Check if the returned signings are equal to the ones we set
	s.Require().Equal(signing, got[0])
}

func (s *KeeperTestSuite) TestSetGetSigCount() {
	ctx, k := s.ctx, s.app.TSSKeeper
	signingID := tss.SigningID(1)

	// Set initial SigCount
	initialCount := uint64(5)
	k.SetSigCount(ctx, signingID, initialCount)

	// Get and check SigCount
	gotCount := k.GetSigCount(ctx, signingID)
	s.Require().Equal(initialCount, gotCount)
}

func (s *KeeperTestSuite) TestAddSigCount() {
	ctx, k := s.ctx, s.app.TSSKeeper
	signingID := tss.SigningID(1)

	// Set initial SigCount
	initialCount := uint64(5)
	k.SetSigCount(ctx, signingID, initialCount)

	// Add to SigCount
	k.AddSigCount(ctx, signingID)

	// Get and check incremented SigCount
	gotCount := k.GetSigCount(ctx, signingID)
	s.Require().Equal(initialCount+1, gotCount)
}

func (s *KeeperTestSuite) TestDeleteSigCount() {
	ctx, k := s.ctx, s.app.TSSKeeper
	signingID := tss.SigningID(1)

	// Set initial SigCount
	initialCount := uint64(5)
	k.SetSigCount(ctx, signingID, initialCount)

	// Delete SigCount
	k.DeleteSigCount(ctx, signingID)

	// Get and check SigCount after deletion
	gotCount := k.GetSigCount(ctx, signingID)
	s.Require().Equal(uint64(0), gotCount) // usually, Get on a non-existing key will return the zero value of the type
}

func (s *KeeperTestSuite) TestGetSetPartialSig() {
	ctx, k := s.ctx, s.app.TSSKeeper
	signingID := tss.SigningID(1)
	memberID := tss.MemberID(1)
	sig := tss.Signature("sample-signature")

	// Set PartialSignature
	k.SetPartialSig(ctx, signingID, memberID, sig)

	// Get and check PartialSignature
	gotSig, err := k.GetPartialSig(ctx, signingID, memberID)
	s.Require().NoError(err)
	s.Require().Equal(sig, gotSig)
}

func (s *KeeperTestSuite) TestDeletePartialSig() {
	ctx, k := s.ctx, s.app.TSSKeeper
	signingID := tss.SigningID(1)
	memberID := tss.MemberID(1)
	sig := tss.Signature("sample-signature")

	// Set PartialSignature
	k.SetPartialSig(ctx, signingID, memberID, sig)

	// Delete PartialSignature
	k.DeletePartialSig(ctx, signingID, memberID)

	// Try to get the deleted PartialSignature, expecting an error
	_, err := k.GetPartialSig(ctx, signingID, memberID)
	s.Require().Error(err)
}

func (s *KeeperTestSuite) TestGetPartialSigs() {
	ctx, k := s.ctx, s.app.TSSKeeper
	signingID := tss.SigningID(1)
	memberIDs := []tss.MemberID{1, 2, 3}
	sigs := tss.Signatures{
		tss.Signature("sample-signature-1"),
		tss.Signature("sample-signature-2"),
		tss.Signature("sample-signature-3"),
	}

	// Set PartialSigs
	for i, memberID := range memberIDs {
		k.SetPartialSig(ctx, signingID, memberID, sigs[i])
	}

	// Get all PartialSigs
	got := k.GetPartialSigs(ctx, signingID)

	// Check if the returned signatures are equal to the ones we set
	s.Require().ElementsMatch(sigs, got)
}

func (s *KeeperTestSuite) TestGetPartialSigsWithKey() {
	ctx, k := s.ctx, s.app.TSSKeeper
	signingID := tss.SigningID(1)
	memberIDs := []tss.MemberID{1, 2, 3}
	sigs := tss.Signatures{
		tss.Signature("sample-signature-1"),
		tss.Signature("sample-signature-2"),
		tss.Signature("sample-signature-3"),
	}

	// Set PartialSigs
	for i, memberID := range memberIDs {
		k.SetPartialSig(ctx, signingID, memberID, sigs[i])
	}

	// Get all PartialSigs with keys
	got := k.GetPartialSigsWithKey(ctx, signingID)

	// Construct expected result
	expected := []types.PartialSignature{}
	for i, memberID := range memberIDs {
		expected = append(expected, types.PartialSignature{
			MemberID:  memberID,
			Signature: sigs[i],
		})
	}

	// Check if the returned signatures with keys are equal to the ones we set
	s.Require().Equal(expected, got)
}

func (s *KeeperTestSuite) TestGetRandomAssigningParticipants() {
	ctx, k := s.ctx, s.app.TSSKeeper
	signingID := uint64(1)
	members := []types.Member{
		{
			MemberID:    1,
			Address:     "band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs",
			PubKey:      nil,
			IsMalicious: false,
		},
		{
			MemberID:    2,
			Address:     "band1p40yh3zkmhcv0ecqp3mcazy83sa57rgjp07dun",
			PubKey:      nil,
			IsMalicious: false,
		},
	}
	t := uint64(1)

	// Generate random participants
	participants, err := k.GetRandomAssigningParticipants(ctx, signingID, members, t)
	s.Require().NoError(err)

	// Check that the number of participants is correct
	s.Require().Len(participants, int(t))

	// Check that there are no duplicate participants
	participantSet := make(map[tss.MemberID]struct{})
	for _, participant := range participants {
		_, exists := participantSet[participant.MemberID]
		s.Require().False(exists)
		participantSet[participant.MemberID] = struct{}{}
	}

	// Check that if use same block and rolling seed will got same answer
	s.Require().Equal([]types.Member{members[1]}, participants)

	// Test that it returns an error if t > size
	_, err = k.GetRandomAssigningParticipants(ctx, signingID, members, uint64(len(members)+1))
	s.Require().Error(err)
}

func (s *KeeperTestSuite) TestGetSetLastExpiredSigningID() {
	ctx, k := s.ctx, s.app.TSSKeeper

	// Set the last expired signing ID
	signingID := tss.SigningID(12345)
	k.SetLastExpiredSigningID(ctx, signingID)

	// Get the last expired signing ID
	got := k.GetLastExpiredSigningID(ctx)

	// Assert equality
	s.Require().Equal(signingID, got)
}

func (s *KeeperTestSuite) TestGetSetPendingSignings() {
	ctx, k := s.ctx, s.app.TSSKeeper

	// Create signingIDs
	signingIDs := []tss.SigningID{1, 2}

	// Set the pending process signings in the store
	k.SetPendingSignings(ctx, types.PendingProcessSignings{
		SigningIDs: signingIDs,
	})

	// Retrieve the pending process signings from the store
	got := k.GetPendingSignings(ctx)

	// Check if the retrieved signing IDs match the original sample
	s.Require().Len(got, len(signingIDs))

	// Check each individual signing ID from the retrieved list against the original sample
	for i, sid := range signingIDs {
		s.Require().Equal(signingIDs[i], sid)
	}
}

func (s *KeeperTestSuite) TestProcessExpiredSignings() {
	ctx, k := s.ctx, s.app.TSSKeeper
	groupID, memberID := tss.GroupID(1), tss.MemberID(1)
	addressStr := "band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs"
	accAddress := sdk.MustAccAddressFromBech32(addressStr)

	// Set member
	k.SetMember(ctx, groupID, types.Member{
		MemberID: memberID,
		Address:  addressStr,
	})

	// Set status
	k.SetStatus(ctx, accAddress, types.Status{
		MemberID: memberID,
		GroupID:  groupID,
		IsActive: true,
		Since:    ctx.BlockTime(),
	})

	// Create signing
	signingID := k.AddSigning(ctx, types.Signing{
		GroupID: 1,
		AssignedMembers: []types.AssignedMember{
			{
				MemberID: memberID,
			},
		},
		Status: types.SIGNING_STATUS_WAITING,
	})

	// Set the current block height
	blockHeight := int64(101)
	ctx = ctx.WithBlockHeight(blockHeight)

	// Handle expired signings
	k.HandleExpiredSignings(ctx)

	// Assert that the last expired signing is updated correctly
	gotSigning, err := k.GetSigning(ctx, signingID)
	s.Require().NoError(err)
	s.Require().Equal(types.SIGNING_STATUS_EXPIRED, gotSigning.Status)
	s.Require().Nil(gotSigning.AssignedMembers)
	gotStatus, err := k.GetStatus(ctx, accAddress, groupID)
	s.Require().NoError(err)
	s.Require().False(gotStatus.IsActive)
	s.Require().False(gotStatus.IsActive)
	gotLastExpiredSigningID := k.GetLastExpiredSigningID(ctx)
	s.Require().Equal(signingID, gotLastExpiredSigningID)
	gotPZs := k.GetPartialSigs(ctx, signingID)
	s.Require().Empty(gotPZs)
}
