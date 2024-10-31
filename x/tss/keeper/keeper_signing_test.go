package keeper_test

import (
	"slices"

	"go.uber.org/mock/gomock"

	"github.com/bandprotocol/chain/v3/pkg/tss"
	tsstestutil "github.com/bandprotocol/chain/v3/x/tss/testutil"
	"github.com/bandprotocol/chain/v3/x/tss/types"
)

func (s *KeeperTestSuite) TestAssignedMembers() {
	ctx, k := s.ctx, s.keeper

	// create a group context and mock the result.
	groupCtx, err := tsstestutil.CompleteGroupCreation(ctx, k, 4, 2)
	s.Require().NoError(err)
	s.rollingseedKeeper.EXPECT().GetRollingSeed(gomock.Any()).
		Return([]byte("RandomStringThatShouldBeLongEnough"))

	DEQueuesBeforePolling := []types.DEQueue{}
	for _, acc := range groupCtx.Accounts {
		deQueue := k.GetDEQueue(ctx, acc.Address)
		DEQueuesBeforePolling = append(DEQueuesBeforePolling, deQueue)
	}

	// check result if no error and check selected MemberIDs
	assignedMembers, err := k.AssignMembersForSigning(ctx, groupCtx.GroupID, []byte("message"), []byte("test_nonce"))
	s.Require().NoError(err)
	assignedMemberIDs := assignedMembers.MemberIDs()
	s.Require().Equal(assignedMemberIDs, []tss.MemberID{3, 4})

	// check if DE is correctly polled; check DEQueue Head is shifted by 1.
	for id, acc := range groupCtx.Accounts {
		deQueue := k.GetDEQueue(ctx, acc.Address)
		expectedHead := DEQueuesBeforePolling[id].Head
		if slices.Contains(assignedMemberIDs, tss.MemberID(id+1)) {
			expectedHead += 1
		}
		s.Require().Equal(expectedHead, deQueue.Head)
	}
}

func (s *KeeperTestSuite) TestInitiateNewSigningRoundForFirstRound() {
	ctx, k := s.ctx, s.keeper

	// create a group context and mock the result.
	groupCtx, err := tsstestutil.CompleteGroupCreation(ctx, k, 4, 2)
	s.Require().NoError(err)
	s.rollingseedKeeper.EXPECT().GetRollingSeed(gomock.Any()).
		Return([]byte("RandomStringThatShouldBeLongEnough"))

	signingID, err := k.CreateSigning(ctx, groupCtx.GroupID, []byte("originator"), []byte("message"))
	s.Require().NoError(err)
	s.Require().Equal(tss.SigningID(1), signingID)

	err = k.InitiateNewSigningRound(ctx, signingID)
	s.Require().NoError(err)

	// check signing
	signing, err := k.GetSigning(ctx, signingID)
	s.Require().NoError(err)
	s.Require().Equal(uint64(1), signing.CurrentAttempt)
	s.Require().NotNil(signing.GroupPubNonce)

	// check signing expiration
	signingExpirations := k.GetSigningExpirations(ctx)
	s.Require().Len(signingExpirations, 1)
	s.Require().Equal(
		types.SigningExpiration{SigningID: signingID, SigningAttempt: signing.CurrentAttempt},
		signingExpirations[0],
	)

	// check signingAttempt
	sa, err := k.GetSigningAttempt(ctx, signingID, signing.CurrentAttempt)
	s.Require().NoError(err)
	s.Require().Equal(uint64(ctx.BlockHeight()+int64(k.GetParams(ctx).SigningPeriod)), sa.ExpiredHeight)

	assignedMembers := types.AssignedMembers(sa.AssignedMembers)
	s.Require().Equal([]tss.MemberID{1, 2}, assignedMembers.MemberIDs())
}

func (s *KeeperTestSuite) TestInitiateNewSigningRoundForThirdRound() {
	ctx, k := s.ctx, s.keeper

	// create a group context and mock the result.
	groupCtx, err := tsstestutil.CompleteGroupCreation(ctx, k, 4, 2)
	s.Require().NoError(err)
	s.rollingseedKeeper.EXPECT().GetRollingSeed(gomock.Any()).
		Return([]byte("RandomStringThatShouldBeLongEnough"))

	signingID, err := k.CreateSigning(ctx, groupCtx.GroupID, []byte("originator"), []byte("message"))
	s.Require().NoError(err)
	s.Require().Equal(tss.SigningID(1), signingID)

	// update signing attempt
	signing, err := k.GetSigning(ctx, signingID)
	s.Require().NoError(err)
	signing.CurrentAttempt = 2
	k.SetSigning(ctx, signing)

	err = k.InitiateNewSigningRound(ctx, signingID)
	s.Require().NoError(err)

	// check signing
	signing, err = k.GetSigning(ctx, signingID)
	s.Require().NoError(err)
	s.Require().Equal(uint64(3), signing.CurrentAttempt)
	s.Require().NotNil(signing.GroupPubNonce)

	// check signing expiration
	signingExpirations := k.GetSigningExpirations(ctx)
	s.Require().Len(signingExpirations, 1)
	s.Require().Equal(
		types.SigningExpiration{SigningID: signingID, SigningAttempt: signing.CurrentAttempt},
		signingExpirations[0],
	)

	// check signingAttempt
	sa, err := k.GetSigningAttempt(ctx, signingID, signing.CurrentAttempt)
	s.Require().NoError(err)
	s.Require().Equal(uint64(ctx.BlockHeight()+int64(k.GetParams(ctx).SigningPeriod)), sa.ExpiredHeight)

	assignedMembers := types.AssignedMembers(sa.AssignedMembers)
	s.Require().Equal([]tss.MemberID{2, 4}, assignedMembers.MemberIDs())
}

func (s *KeeperTestSuite) TestInitiateNewSigningRoundOverMaxAttempt() {
	ctx, k := s.ctx, s.keeper

	// create a group context and mock the result.
	groupCtx, err := tsstestutil.CompleteGroupCreation(ctx, k, 4, 2)
	s.Require().NoError(err)

	signingID, err := k.CreateSigning(ctx, groupCtx.GroupID, []byte("originator"), []byte("message"))
	s.Require().NoError(err)
	s.Require().Equal(tss.SigningID(1), signingID)

	// update signing attempt
	signing, err := k.GetSigning(ctx, signingID)
	s.Require().NoError(err)
	signing.CurrentAttempt = k.GetParams(ctx).MaxSigningAttempt
	k.SetSigning(ctx, signing)

	err = k.InitiateNewSigningRound(ctx, signingID)
	s.Require().ErrorIs(err, types.ErrMaxSigningAttemptReached)
}

func (s *KeeperTestSuite) TestRequestSigning() {
	ctx, k := s.ctx, s.keeper

	groupCtx, err := tsstestutil.CompleteGroupCreation(ctx, k, 4, 2)
	groupID := groupCtx.GroupID
	s.Require().NoError(err)

	s.rollingseedKeeper.EXPECT().GetRollingSeed(gomock.Any()).
		Return([]byte("RandomStringThatShouldBeLongEnough"))

	// Create a new request for the request signature
	content := types.NewTextSignatureOrder([]byte("example"))
	signingID, err := k.RequestSigning(ctx, groupID, types.DirectOriginator{}, content)
	s.Require().NoError(err)

	signing, err := k.GetSigning(ctx, signingID)
	s.Require().NoError(err)
	s.Require().Equal(groupID, signing.GroupID)
	s.Require().Equal(types.SIGNING_STATUS_WAITING, signing.Status)
}

func (s *KeeperTestSuite) TestMustGetCurrentAssignedMembers() {
	ctx, k := s.ctx, s.keeper

	signing := GetExampleSigning()
	k.SetSigning(ctx, signing)
	sa := GetExampleSigningAttempt()
	k.SetSigningAttempt(ctx, sa)

	// Get current assigned members
	assignedMembers := k.MustGetCurrentAssignedMembers(ctx, signing.ID)
	s.Require().Len(assignedMembers, 2)
	s.Require().Equal("band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs", assignedMembers[0].String())

	// Get incorrect signingID
	s.Require().PanicsWithError("failed to get Signing with ID: 999: signing not found", func() {
		k.MustGetCurrentAssignedMembers(ctx, 999)
	})
}

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

	signingMsg := types.EncodeSigning(ctx, 1, []byte("originator"), []byte("message"))
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
