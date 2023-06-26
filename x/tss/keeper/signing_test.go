package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/tss/types"
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
				PubD:     hexDecode("02234d901b8d6404b509e9926407d1a2749f456d18b159af647a65f3e907d61ef1"),
				PubE:     hexDecode("028a1f3e214831b2f2d6e27384817132ddaa222928b05e9372472aa2735cf1f797"),
				PubNonce: hexDecode("03cbb6a27c62baa195dff6c75eae7b6b7713f978732a671855f7d7b86b06e6ac67"),
			},
		},
		Message:       []byte("data"),
		GroupPubNonce: hexDecode("03fae45376abb0d60c3ae2b5caee749118125ec3d73725f3ad03b0b6e686d0f31a"),
		Commitment:    []byte("commitment"),
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
				PubD:     hexDecode("02234d901b8d6404b509e9926407d1a2749f456d18b159af647a65f3e907d61ef1"),
				PubE:     hexDecode("028a1f3e214831b2f2d6e27384817132ddaa222928b05e9372472aa2735cf1f797"),
				PubNonce: hexDecode("03cbb6a27c62baa195dff6c75eae7b6b7713f978732a671855f7d7b86b06e6ac67"),
			},
		},
		Message:       []byte("data"),
		GroupPubNonce: hexDecode("03fae45376abb0d60c3ae2b5caee749118125ec3d73725f3ad03b0b6e686d0f31a"),
		Commitment:    []byte("commitment"),
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
				PubD:     hexDecode("02234d901b8d6404b509e9926407d1a2749f456d18b159af647a65f3e907d61ef1"),
				PubE:     hexDecode("028a1f3e214831b2f2d6e27384817132ddaa222928b05e9372472aa2735cf1f797"),
				PubNonce: hexDecode("03cbb6a27c62baa195dff6c75eae7b6b7713f978732a671855f7d7b86b06e6ac67"),
			},
		},
		Message:       []byte("data"),
		GroupPubNonce: hexDecode("03fae45376abb0d60c3ae2b5caee749118125ec3d73725f3ad03b0b6e686d0f31a"),
		Commitment:    []byte("commitment"),
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

func (s *KeeperTestSuite) TestGetSetPendingSign() {
	ctx, k := s.ctx, s.app.TSSKeeper
	address, _ := sdk.AccAddressFromBech32("band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs")
	signingID := tss.SigningID(1)

	// Set PendingSign
	k.SetPendingSign(ctx, address, signingID)

	// Get PendingSign
	got := k.GetPendingSign(ctx, address, signingID)

	s.Require().True(got)
}

func (s *KeeperTestSuite) TestDeletePendingSign() {
	ctx, k := s.ctx, s.app.TSSKeeper
	address, _ := sdk.AccAddressFromBech32("band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs")
	signingID := tss.SigningID(1)

	// Set PendingSign
	k.SetPendingSign(ctx, address, signingID)

	// Confirm PendingSign was set
	got := k.GetPendingSign(ctx, address, signingID)
	s.Require().True(got)

	// Delete PendingSign
	k.DeletePendingSign(ctx, address, signingID)

	// Confirm PendingSign was deleted
	got = k.GetPendingSign(ctx, address, signingID)
	s.Require().False(got)
}

func (s *KeeperTestSuite) TestGetPendingSignIDs() {
	ctx, k := s.ctx, s.app.TSSKeeper
	address, _ := sdk.AccAddressFromBech32("band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs")
	signingIDs := []tss.SigningID{1, 2, 3}

	// Set PendingSign for multiple SigningIDs
	for _, id := range signingIDs {
		k.SetPendingSign(ctx, address, id)
	}

	// Get all PendingSignIDs
	got := k.GetPendingSignIDs(ctx, address)

	// Convert got (which is []uint64) to []tss.SigningID for comparison
	var gotConverted []tss.SigningID
	for _, id := range got {
		gotConverted = append(gotConverted, tss.SigningID(id))
	}

	// Check if the returned IDs are equal to the ones we set
	s.Require().Equal(signingIDs, gotConverted)
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
	s.Require().ElementsMatch(expected, got)
}

func (s *KeeperTestSuite) TestGetSetRollingSeed() {
	ctx, k := s.ctx, s.app.TSSKeeper
	rollingSeed := []byte("sample-rolling-seed")

	// Set RollingSeed
	k.SetRollingSeed(ctx, rollingSeed)

	// Get and check RollingSeed
	gotSeed := k.GetRollingSeed(ctx)
	s.Require().Equal(rollingSeed, gotSeed)
}

func (s *KeeperTestSuite) TestGetRandomAssigningParticipants() {
	ctx, k := s.ctx, s.app.TSSKeeper
	signingID := uint64(1)
	size := uint64(10)
	t := uint64(5)

	// Set RollingSeed
	k.SetRollingSeed(ctx, []byte("sample-rolling-seed"))

	// Generate random participants
	participants, err := k.GetRandomAssigningParticipants(ctx, signingID, size, t)
	s.Require().NoError(err)

	// Check that the number of participants is correct
	s.Require().Len(participants, int(t))

	// Check that there are no duplicate participants
	participantSet := make(map[tss.MemberID]struct{})
	for _, participant := range participants {
		_, exists := participantSet[participant]
		s.Require().False(exists)
		participantSet[participant] = struct{}{}
	}

	// Check that if use same block and rolling seed will got same answer
	s.Require().Equal([]tss.MemberID{2, 4, 5, 6, 8}, participants)

	// Test that it returns an error if t > size
	_, err = k.GetRandomAssigningParticipants(ctx, signingID, t-1, t)
	s.Require().Error(err)
}
