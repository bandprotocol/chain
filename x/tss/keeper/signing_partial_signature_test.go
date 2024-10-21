package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/pkg/tss"
	"github.com/bandprotocol/chain/v3/x/tss/types"
)

func (s *KeeperTestSuite) TestGetSetPartialSignatureCount() {
	ctx, k := s.ctx, s.keeper

	k.SetPartialSignatureCount(ctx, 1, 1, 1)

	got := k.GetPartialSignatureCount(ctx, 1, 1)
	s.Require().Equal(uint64(1), got)

	// not found should return 0
	got = k.GetPartialSignatureCount(ctx, 1, 2)
	s.Require().Equal(uint64(0), got)
}

func (s *KeeperTestSuite) TestAddPartialSignatureCount() {
	ctx, k := s.ctx, s.keeper

	k.AddPartialSignatureCount(ctx, 1, 1)
	got := k.GetPartialSignatureCount(ctx, 1, 1)
	s.Require().Equal(uint64(1), got)

	k.AddPartialSignatureCount(ctx, 1, 1)
	got = k.GetPartialSignatureCount(ctx, 1, 1)
	s.Require().Equal(uint64(2), got)
}

func (s *KeeperTestSuite) TestDeletePartialSignatureCounts() {
	ctx, k := s.ctx, s.keeper

	k.AddPartialSignatureCount(ctx, 1, 1)
	k.AddPartialSignatureCount(ctx, 1, 1)
	k.AddPartialSignatureCount(ctx, 1, 2)
	k.AddPartialSignatureCount(ctx, 2, 1)

	// delete all partial signature counts of signing ID 1
	k.DeletePartialSignatureCount(ctx, 1, 1)

	// check result
	got := k.GetPartialSignatureCount(ctx, 1, 1)
	s.Require().Equal(uint64(0), got)
	got = k.GetPartialSignatureCount(ctx, 1, 2)
	s.Require().Equal(uint64(1), got)
	got = k.GetPartialSignatureCount(ctx, 2, 1)
	s.Require().Equal(uint64(1), got)
}

func (s *KeeperTestSuite) TestGetSetPartialSignature() {
	ctx, k := s.ctx, s.keeper

	sig := tss.Signature([]byte("test"))
	k.SetPartialSignature(ctx, 1, 1, 1, sig)

	// Get partial signature
	got, err := k.GetPartialSignature(ctx, 1, 1, 1)
	s.Require().NoError(err)
	s.Require().Equal(sig, got)

	// Get partial signature not found error
	_, err = k.GetPartialSignature(ctx, 1, 1, 2)
	s.Require().ErrorIs(err, types.ErrPartialSignatureNotFound)
}

func (s *KeeperTestSuite) TestHasPartialSignature() {
	ctx, k := s.ctx, s.keeper

	sig := tss.Signature([]byte("test"))
	k.SetPartialSignature(ctx, 1, 1, 1, sig)

	s.Require().True(k.HasPartialSignature(ctx, 1, 1, 1))
	s.Require().False(k.HasPartialSignature(ctx, 1, 1, 2))
}

func (s *KeeperTestSuite) TestAddPartialSignature() {
	ctx, k := s.ctx, s.keeper

	sig := tss.Signature([]byte("test"))
	k.AddPartialSignature(ctx, 1, 1, 1, sig)

	// Get partial signature
	got, err := k.GetPartialSignature(ctx, 1, 1, 1)
	s.Require().NoError(err)
	s.Require().Equal(sig, got)
	count := k.GetPartialSignatureCount(ctx, 1, 1)
	s.Require().Equal(uint64(1), count)

	// add new signature from new memberID
	sig2 := tss.Signature([]byte("test"))
	k.AddPartialSignature(ctx, 1, 1, 2, sig2)

	// check result
	got, err = k.GetPartialSignature(ctx, 1, 1, 2)
	s.Require().NoError(err)
	s.Require().Equal(sig2, got)
	count = k.GetPartialSignatureCount(ctx, 1, 1)
	s.Require().Equal(uint64(2), count)
}

func (s *KeeperTestSuite) TestGetPartialSignatures() {
	ctx, k := s.ctx, s.keeper

	sig := tss.Signature([]byte("test"))
	k.AddPartialSignature(ctx, 1, 1, 1, sig)
	sig2 := tss.Signature([]byte("test2"))
	k.AddPartialSignature(ctx, 1, 1, 2, sig2)

	// check result
	sigs := k.GetPartialSignatures(ctx, 1, 1)
	s.Require().Equal(tss.Signatures([]tss.Signature{sig, sig2}), sigs)

	// get empty list
	sigs = k.GetPartialSignatures(ctx, 1, 2)
	s.Require().Equal(tss.Signatures(nil), sigs)
}

func (s *KeeperTestSuite) TestGetPartialSignaturesWithKey() {
	ctx, k := s.ctx, s.keeper

	sig := tss.Signature([]byte("test"))
	k.AddPartialSignature(ctx, 1, 1, 1, sig)
	sig2 := tss.Signature([]byte("test2"))
	k.AddPartialSignature(ctx, 1, 1, 2, sig2)

	got := k.GetPartialSignaturesWithKey(ctx, 1, 1)
	s.Require().Equal([]types.PartialSignature{
		{SigningID: 1, SigningAttempt: 1, MemberID: 1, Signature: sig},
		{SigningID: 1, SigningAttempt: 1, MemberID: 2, Signature: sig2},
	}, got)

	// get empty list
	got = k.GetPartialSignaturesWithKey(ctx, 1, 2)
	s.Require().Equal([]types.PartialSignature(nil), got)
}

func (s *KeeperTestSuite) TestGetMemberNotSubmitSignature() {
	ctx, k := s.ctx, s.keeper

	sa := GetExampleSigningAttempt()
	k.SetSigningAttempt(ctx, sa)

	sig := tss.Signature([]byte("test"))
	k.AddPartialSignature(ctx, 1, 1, 1, sig)

	got := k.GetMembersNotSubmitSignature(ctx, sa.SigningID, sa.Attempt)
	s.Require().Equal([]sdk.AccAddress{sdk.MustAccAddressFromBech32(sa.AssignedMembers[1].Address)}, got)

	// get empty list
	sig2 := tss.Signature([]byte("test2"))
	k.AddPartialSignature(ctx, 1, 1, 2, sig2)
	got = k.GetMembersNotSubmitSignature(ctx, sa.SigningID, sa.Attempt)
	s.Require().Equal([]sdk.AccAddress(nil), got)
}

func (s *KeeperTestSuite) TestDeletePartialSignatures() {
	ctx, k := s.ctx, s.keeper

	sig := tss.Signature([]byte("test"))
	k.AddPartialSignature(ctx, 1, 1, 1, sig)
	sig2 := tss.Signature([]byte("test2"))
	k.AddPartialSignature(ctx, 1, 1, 2, sig2)
	sig3 := tss.Signature([]byte("test3"))
	k.AddPartialSignature(ctx, 1, 2, 1, sig3)
	sig4 := tss.Signature([]byte("test4"))
	k.AddPartialSignature(ctx, 2, 1, 1, sig4)

	k.DeletePartialSignatures(ctx, 1, 1)

	// check partial signature
	_, err := k.GetPartialSignature(ctx, 1, 1, 1)
	s.Require().ErrorIs(err, types.ErrPartialSignatureNotFound)
	_, err = k.GetPartialSignature(ctx, 1, 1, 2)
	s.Require().ErrorIs(err, types.ErrPartialSignatureNotFound)

	_, err = k.GetPartialSignature(ctx, 2, 1, 1)
	s.Require().NoError(err)
	_, err = k.GetPartialSignature(ctx, 1, 2, 1)
	s.Require().NoError(err)
}
