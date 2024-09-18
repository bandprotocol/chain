package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

func TestGetSetPartialSignatureCount(t *testing.T) {
	s := NewKeeperTestSuite(t)
	ctx, k := s.Ctx, s.Keeper

	k.SetPartialSignatureCount(ctx, 1, 1, 1)

	got := k.GetPartialSignatureCount(ctx, 1, 1)
	require.Equal(t, uint64(1), got)

	// not found should return 0
	got = k.GetPartialSignatureCount(ctx, 1, 2)
	require.Equal(t, uint64(0), got)
}

func TestAddPartialSignatureCount(t *testing.T) {
	s := NewKeeperTestSuite(t)
	ctx, k := s.Ctx, s.Keeper

	k.AddPartialSignatureCount(ctx, 1, 1)
	got := k.GetPartialSignatureCount(ctx, 1, 1)
	require.Equal(t, uint64(1), got)

	k.AddPartialSignatureCount(ctx, 1, 1)
	got = k.GetPartialSignatureCount(ctx, 1, 1)
	require.Equal(t, uint64(2), got)
}

func TestDeletePartialSignatureCounts(t *testing.T) {
	s := NewKeeperTestSuite(t)
	ctx, k := s.Ctx, s.Keeper

	k.AddPartialSignatureCount(ctx, 1, 1)
	k.AddPartialSignatureCount(ctx, 1, 1)
	k.AddPartialSignatureCount(ctx, 1, 2)
	k.AddPartialSignatureCount(ctx, 2, 1)

	// delete all partial signature counts of signing ID 1
	k.DeletePartialSignatureCount(ctx, 1, 1)

	// check result
	got := k.GetPartialSignatureCount(ctx, 1, 1)
	require.Equal(t, uint64(0), got)
	got = k.GetPartialSignatureCount(ctx, 1, 2)
	require.Equal(t, uint64(1), got)
	got = k.GetPartialSignatureCount(ctx, 2, 1)
	require.Equal(t, uint64(1), got)
}

func TestGetSetPartialSignature(t *testing.T) {
	s := NewKeeperTestSuite(t)
	ctx, k := s.Ctx, s.Keeper

	sig := tss.Signature([]byte("test"))
	k.SetPartialSignature(ctx, 1, 1, 1, sig)

	// Get partial signature
	got, err := k.GetPartialSignature(ctx, 1, 1, 1)
	require.NoError(t, err)
	require.Equal(t, sig, got)

	// Get partial signature not found error
	_, err = k.GetPartialSignature(ctx, 1, 1, 2)
	require.ErrorIs(t, err, types.ErrPartialSignatureNotFound)
}

func TestHasPartialSignature(t *testing.T) {
	s := NewKeeperTestSuite(t)
	ctx, k := s.Ctx, s.Keeper

	sig := tss.Signature([]byte("test"))
	k.SetPartialSignature(ctx, 1, 1, 1, sig)

	require.True(t, k.HasPartialSignature(ctx, 1, 1, 1))
	require.False(t, k.HasPartialSignature(ctx, 1, 1, 2))
}

func TestAddPartialSignature(t *testing.T) {
	s := NewKeeperTestSuite(t)
	ctx, k := s.Ctx, s.Keeper

	sig := tss.Signature([]byte("test"))
	k.AddPartialSignature(ctx, 1, 1, 1, sig)

	// Get partial signature
	got, err := k.GetPartialSignature(ctx, 1, 1, 1)
	require.NoError(t, err)
	require.Equal(t, sig, got)
	count := k.GetPartialSignatureCount(ctx, 1, 1)
	require.Equal(t, uint64(1), count)

	// add new signature from new memberID
	sig2 := tss.Signature([]byte("test"))
	k.AddPartialSignature(ctx, 1, 1, 2, sig2)

	// check result
	got, err = k.GetPartialSignature(ctx, 1, 1, 2)
	require.NoError(t, err)
	require.Equal(t, sig2, got)
	count = k.GetPartialSignatureCount(ctx, 1, 1)
	require.Equal(t, uint64(2), count)
}

func TestGetPartialSignatures(t *testing.T) {
	s := NewKeeperTestSuite(t)
	ctx, k := s.Ctx, s.Keeper

	sig := tss.Signature([]byte("test"))
	k.AddPartialSignature(ctx, 1, 1, 1, sig)
	sig2 := tss.Signature([]byte("test2"))
	k.AddPartialSignature(ctx, 1, 1, 2, sig2)

	// check result
	sigs := k.GetPartialSignatures(ctx, 1, 1)
	require.Equal(t, tss.Signatures([]tss.Signature{sig, sig2}), sigs)

	// get empty list
	sigs = k.GetPartialSignatures(ctx, 1, 2)
	require.Equal(t, tss.Signatures(nil), sigs)
}

func TestGetPartialSignaturesWithKey(t *testing.T) {
	s := NewKeeperTestSuite(t)
	ctx, k := s.Ctx, s.Keeper

	sig := tss.Signature([]byte("test"))
	k.AddPartialSignature(ctx, 1, 1, 1, sig)
	sig2 := tss.Signature([]byte("test2"))
	k.AddPartialSignature(ctx, 1, 1, 2, sig2)

	got := k.GetPartialSignaturesWithKey(ctx, 1, 1)
	require.Equal(t, []types.PartialSignature{
		{SigningID: 1, SigningAttempt: 1, MemberID: 1, Signature: sig},
		{SigningID: 1, SigningAttempt: 1, MemberID: 2, Signature: sig2},
	}, got)

	// get empty list
	got = k.GetPartialSignaturesWithKey(ctx, 1, 2)
	require.Equal(t, []types.PartialSignature(nil), got)
}

func TestGetMemberNotSubmitSignature(t *testing.T) {
	s := NewKeeperTestSuite(t)
	ctx, k := s.Ctx, s.Keeper

	sa := GetExampleSigningAttempt()
	k.SetSigningAttempt(ctx, sa)

	sig := tss.Signature([]byte("test"))
	k.AddPartialSignature(ctx, 1, 1, 1, sig)

	got := k.GetMembersNotSubmitSignature(ctx, sa.SigningID, sa.Attempt)
	require.Equal(t, []sdk.AccAddress{sdk.MustAccAddressFromBech32(sa.AssignedMembers[1].Address)}, got)

	// get empty list
	sig2 := tss.Signature([]byte("test2"))
	k.AddPartialSignature(ctx, 1, 1, 2, sig2)
	got = k.GetMembersNotSubmitSignature(ctx, sa.SigningID, sa.Attempt)
	require.Equal(t, []sdk.AccAddress(nil), got)
}

func TestDeletePartialSignatures(t *testing.T) {
	s := NewKeeperTestSuite(t)
	ctx, k := s.Ctx, s.Keeper

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
	require.ErrorIs(t, err, types.ErrPartialSignatureNotFound)
	_, err = k.GetPartialSignature(ctx, 1, 1, 2)
	require.ErrorIs(t, err, types.ErrPartialSignatureNotFound)

	_, err = k.GetPartialSignature(ctx, 2, 1, 1)
	require.NoError(t, err)
	_, err = k.GetPartialSignature(ctx, 1, 2, 1)
	require.NoError(t, err)
}
