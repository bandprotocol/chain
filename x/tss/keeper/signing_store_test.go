package keeper_test

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

func TestGetSetSigningCount(t *testing.T) {
	s := NewKeeperTestSuite(t)
	ctx, k := s.Ctx, s.Keeper

	k.SetSigningCount(ctx, 1)

	got := k.GetSigningCount(ctx)
	require.Equal(t, uint64(1), got)
}

func TestGetSetSigning(t *testing.T) {
	s := NewKeeperTestSuite(t)
	ctx, k := s.Ctx, s.Keeper

	signing := GetExampleSigning()
	k.SetSigning(ctx, signing)

	// Get signing
	got, err := k.GetSigning(ctx, signing.ID)
	require.NoError(t, err)
	require.Equal(t, signing, got)

	// Get Signing not found error
	_, err = k.GetSigning(ctx, 2)
	require.ErrorIs(t, err, types.ErrSigningNotFound)
}

func TestHasSigning(t *testing.T) {
	s := NewKeeperTestSuite(t)
	ctx, k := s.Ctx, s.Keeper

	signing := GetExampleSigning()
	k.SetSigning(ctx, signing)

	require.True(t, k.HasSigning(ctx, 1))
	require.False(t, k.HasSigning(ctx, 2))
}

func TestMustGetSigning(t *testing.T) {
	s := NewKeeperTestSuite(t)
	ctx, k := s.Ctx, s.Keeper

	signing := GetExampleSigning()
	k.SetSigning(ctx, signing)

	// Get signing
	got := k.MustGetSigning(ctx, signing.ID)
	require.Equal(t, signing, got)

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("the code below should panic")
		}
	}()
	_ = k.MustGetSigning(ctx, 2)
}

func TestCreateSigningSuccess(t *testing.T) {
	s := NewKeeperTestSuite(t)
	ctx, k := s.Ctx, s.Keeper

	group := GetExampleGroup()
	k.SetGroup(ctx, group)

	// Create a sample signing object
	signingID, err := k.CreateSigning(ctx, 1, []byte("originator"), []byte("message"))
	require.NoError(t, err)
	require.Equal(t, tss.SigningID(1), signingID)

	signingMsg := k.GetSigningMessage(ctx, 1, []byte("originator"), []byte("message"))
	expectSigning := types.Signing{
		ID:               1,
		CurrentAttempt:   0,
		GroupID:          1,
		GroupPubKey:      group.PubKey,
		Originator:       []byte("originator"),
		Message:          signingMsg,
		CreatedHeight:    uint64(ctx.BlockHeight()),
		CreatedTimestamp: ctx.BlockTime(),
		Status:           types.SIGNING_STATUS_WAITING,
	}

	got, err := k.GetSigning(ctx, signingID)
	require.NoError(t, err)
	require.Equal(t, expectSigning, got)
}

func TestCreateSigningFailGroupStatusNotReady(t *testing.T) {
	s := NewKeeperTestSuite(t)
	ctx, k := s.Ctx, s.Keeper

	group := GetExampleGroup()
	group.Status = types.GROUP_STATUS_ROUND_2
	k.SetGroup(ctx, group)

	// Create a sample signing object
	_, err := k.CreateSigning(ctx, 1, []byte("originator"), []byte("message"))
	require.ErrorIs(t, err, types.ErrGroupIsNotActive)
}

func TestGetSigningMessage(t *testing.T) {
	s := NewKeeperTestSuite(t)
	ctx, k := s.Ctx, s.Keeper

	got := k.GetSigningMessage(ctx, 1, []byte("originator"), []byte("message"))
	strHex := hex.EncodeToString(got)
	expected := "" +
		"c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470" +
		"bac0e8e27c59b287045fc0a3df1b9bc08bca23b9c7d4e8d21f6c311f67a7ef4b" +
		"000000005e0be100" +
		"0000000000000001" +
		"6d657373616765"

	require.Equal(t, expected, strHex)
}

func TestGetSetSigningAttempt(t *testing.T) {
	s := NewKeeperTestSuite(t)
	ctx, k := s.Ctx, s.Keeper

	sa := GetExampleSigningAttempt()
	k.SetSigningAttempt(ctx, sa)

	// Get SigningAttempt
	got, err := k.GetSigningAttempt(ctx, sa.SigningID, sa.Attempt)
	require.NoError(t, err)
	require.Equal(t, sa, got)

	_, err = k.GetSigningAttempt(ctx, sa.SigningID, 10)
	require.ErrorIs(t, err, types.ErrSigningAttemptNotFound)

	_, err = k.GetSigningAttempt(ctx, 3, sa.Attempt)
	require.ErrorIs(t, err, types.ErrSigningAttemptNotFound)
}

func TestMustGetSigningAttempt(t *testing.T) {
	s := NewKeeperTestSuite(t)
	ctx, k := s.Ctx, s.Keeper

	sa := GetExampleSigningAttempt()
	k.SetSigningAttempt(ctx, sa)

	// Get SigningAttempt
	got := k.MustGetSigningAttempt(ctx, sa.SigningID, sa.Attempt)
	require.Equal(t, sa, got)

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("the code below should panic")
		}
	}()
	_ = k.MustGetSigningAttempt(ctx, 3, sa.Attempt)
}

func TestDeleteSigningAttempts(t *testing.T) {
	s := NewKeeperTestSuite(t)
	ctx, k := s.Ctx, s.Keeper

	sa1 := GetExampleSigningAttempt()
	k.SetSigningAttempt(ctx, sa1)

	sa2 := GetExampleSigningAttempt()
	sa2.SigningID = tss.SigningID(2)
	k.SetSigningAttempt(ctx, sa2)

	sa3 := GetExampleSigningAttempt()
	sa3.Attempt = 2
	k.SetSigningAttempt(ctx, sa3)

	// get signing attempt normally
	for _, sa := range []types.SigningAttempt{sa1, sa2, sa3} {
		got, err := k.GetSigningAttempt(ctx, sa.SigningID, sa.Attempt)
		require.NoError(t, err)
		require.Equal(t, sa, got)
	}

	k.DeleteSigningAttempt(ctx, sa1.SigningID, sa1.Attempt)

	// check remaining signing Attempt

	for _, sa := range []types.SigningAttempt{sa2, sa3} {
		got, err := k.GetSigningAttempt(ctx, sa.SigningID, sa.Attempt)
		require.NoError(t, err)
		require.Equal(t, sa, got)
	}

	_, err := k.GetSigningAttempt(ctx, sa1.SigningID, sa1.Attempt)
	require.ErrorIs(t, err, types.ErrSigningAttemptNotFound)
}
