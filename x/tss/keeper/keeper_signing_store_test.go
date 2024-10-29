package keeper_test

import (
	"encoding/hex"

	"github.com/bandprotocol/chain/v3/pkg/tss"
	"github.com/bandprotocol/chain/v3/x/tss/types"
)

func (s *KeeperTestSuite) TestGetSetSigningCount() {
	ctx, k := s.ctx, s.keeper

	k.SetSigningCount(ctx, 1)

	got := k.GetSigningCount(ctx)
	s.Require().Equal(uint64(1), got)
}

func (s *KeeperTestSuite) TestGetSetSigning() {
	ctx, k := s.ctx, s.keeper

	signing := GetExampleSigning()
	k.SetSigning(ctx, signing)

	// Get signing
	got, err := k.GetSigning(ctx, signing.ID)
	s.Require().NoError(err)
	s.Require().Equal(signing, got)

	// Get Signing not found error
	_, err = k.GetSigning(ctx, 2)
	s.Require().ErrorIs(err, types.ErrSigningNotFound)
}

func (s *KeeperTestSuite) TestHasSigning() {
	ctx, k := s.ctx, s.keeper

	signing := GetExampleSigning()
	k.SetSigning(ctx, signing)

	s.Require().True(k.HasSigning(ctx, 1))
	s.Require().False(k.HasSigning(ctx, 2))
}

func (s *KeeperTestSuite) TestMustGetSigning() {
	ctx, k := s.ctx, s.keeper

	signing := GetExampleSigning()
	k.SetSigning(ctx, signing)

	// Get signing
	got := k.MustGetSigning(ctx, signing.ID)
	s.Require().Equal(signing, got)

	// Get Signing not found, should panic.
	s.Require().Panics(func() {
		_ = k.MustGetSigning(ctx, 2)
	})
}

func (s *KeeperTestSuite) TestCreateSigningSuccess() {
	ctx, k := s.ctx, s.keeper

	group := GetExampleGroup()
	k.SetGroup(ctx, group)

	// Create a sample signing object
	signingID, err := k.CreateSigning(ctx, 1, []byte("originator"), []byte("message"))
	s.Require().NoError(err)
	s.Require().Equal(tss.SigningID(1), signingID)

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
	s.Require().NoError(err)
	s.Require().Equal(expectSigning, got)
}

func (s *KeeperTestSuite) TestCreateSigningFailGroupStatusNotReady() {
	ctx, k := s.ctx, s.keeper

	group := GetExampleGroup()
	group.Status = types.GROUP_STATUS_ROUND_2
	k.SetGroup(ctx, group)

	// Create a sample signing object
	_, err := k.CreateSigning(ctx, 1, []byte("originator"), []byte("message"))
	s.Require().ErrorIs(err, types.ErrGroupIsNotActive)
}

func (s *KeeperTestSuite) TestGetSigningMessage() {
	ctx, k := s.ctx, s.keeper

	got := k.GetSigningMessage(ctx, 1, []byte("originator"), []byte("message"))
	strHex := hex.EncodeToString(got)
	expected := "" +
		"c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470" +
		"bac0e8e27c59b287045fc0a3df1b9bc08bca23b9c7d4e8d21f6c311f67a7ef4b" +
		"000000005e0be100" +
		"0000000000000001" +
		"6d657373616765"

	s.Require().Equal(expected, strHex)
}

func (s *KeeperTestSuite) TestGetSetSigningAttempt() {
	ctx, k := s.ctx, s.keeper

	sa := GetExampleSigningAttempt()
	k.SetSigningAttempt(ctx, sa)

	got, err := k.GetSigningAttempt(ctx, sa.SigningID, sa.Attempt)
	s.Require().NoError(err)
	s.Require().Equal(sa, got)
}

func (s *KeeperTestSuite) TestGetSigningAttemptIncorrectAttempt() {
	ctx, k := s.ctx, s.keeper

	sa := GetExampleSigningAttempt()
	k.SetSigningAttempt(ctx, sa)

	_, err := k.GetSigningAttempt(ctx, sa.SigningID, 10)
	s.Require().ErrorIs(err, types.ErrSigningAttemptNotFound)
}

func (s *KeeperTestSuite) TestGetSigningAttemptIncorrectSigningID() {
	ctx, k := s.ctx, s.keeper

	sa := GetExampleSigningAttempt()
	k.SetSigningAttempt(ctx, sa)

	_, err := k.GetSigningAttempt(ctx, 3, sa.Attempt)
	s.Require().ErrorIs(err, types.ErrSigningAttemptNotFound)
}

func (s *KeeperTestSuite) TestMustGetSigningAttempt() {
	ctx, k := s.ctx, s.keeper

	sa := GetExampleSigningAttempt()
	k.SetSigningAttempt(ctx, sa)

	// Get SigningAttempt
	got := k.MustGetSigningAttempt(ctx, sa.SigningID, sa.Attempt)
	s.Require().Equal(sa, got)

	s.Require().Panics(func() {
		_ = k.MustGetSigningAttempt(ctx, 3, sa.Attempt)
	})
}

func (s *KeeperTestSuite) TestDeleteSigningAttempts() {
	ctx, k := s.ctx, s.keeper

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
		s.Require().NoError(err)
		s.Require().Equal(sa, got)
	}

	k.DeleteSigningAttempt(ctx, sa1.SigningID, sa1.Attempt)

	// check remaining signing Attempt

	for _, sa := range []types.SigningAttempt{sa2, sa3} {
		got, err := k.GetSigningAttempt(ctx, sa.SigningID, sa.Attempt)
		s.Require().NoError(err)
		s.Require().Equal(sa, got)
	}

	_, err := k.GetSigningAttempt(ctx, sa1.SigningID, sa1.Attempt)
	s.Require().ErrorIs(err, types.ErrSigningAttemptNotFound)
}
