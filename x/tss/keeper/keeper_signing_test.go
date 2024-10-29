package keeper_test

import (
	"slices"

	"go.uber.org/mock/gomock"

	"github.com/bandprotocol/chain/v3/pkg/tss"
	tsstestutil "github.com/bandprotocol/chain/v3/x/tss/testutil"
	"github.com/bandprotocol/chain/v3/x/tss/types"
)

func (s *KeeperTestSuite) TestGetRandomMembersSuccess() {
	ctx, k := s.ctx, s.keeper

	// create a group context and mock the result.
	groupCtx, err := tsstestutil.CompleteGroupCreation(ctx, k, 4, 2)
	s.Require().NoError(err)
	s.rollingseedKeeper.EXPECT().GetRollingSeed(gomock.Any()).
		Return([]byte("RandomStringThatShouldBeLongEnough"))

	// call GetRandomMembers and check results
	ams, err := k.GetRandomMembers(ctx, groupCtx.GroupID, []byte("test_nonce"))
	s.Require().NoError(err)

	memberIDs := []tss.MemberID{}
	for _, am := range ams {
		memberIDs = append(memberIDs, am.ID)
	}
	s.Require().Equal(memberIDs, []tss.MemberID{3, 4})
}

func (s *KeeperTestSuite) TestGetRandomMembersInsufficientActiveMembers() {
	ctx, k := s.ctx, s.keeper

	// create a group context.
	groupCtx, err := tsstestutil.CompleteGroupCreation(ctx, k, 4, 2)
	s.Require().NoError(err)

	// set every member to inactive.
	members, err := k.GetGroupMembers(ctx, groupCtx.GroupID)
	s.Require().NoError(err)
	for _, member := range members[:len(members)-1] {
		member.IsActive = false
		k.SetMember(ctx, member)
	}

	_, err = k.GetRandomMembers(ctx, tss.GroupID(1), []byte("test_nonce"))
	s.Require().ErrorIs(err, types.ErrInsufficientActiveMembers)
}

func (s *KeeperTestSuite) TestGetRandomMembersNoActiveMember() {
	ctx, k := s.ctx, s.keeper

	// create a group context.
	groupCtx, err := tsstestutil.CompleteGroupCreation(ctx, k, 4, 2)
	s.Require().NoError(err)

	// set every member to inactive.
	members, err := k.GetGroupMembers(ctx, groupCtx.GroupID)
	s.Require().NoError(err)
	for _, member := range members {
		member.IsActive = false
		k.SetMember(ctx, member)
	}

	_, err = k.GetRandomMembers(ctx, tss.GroupID(1), []byte("test_nonce"))
	s.Require().ErrorIs(err, types.ErrNoActiveMember)
}

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
	assignedMembers, err := k.AssignMembers(ctx, groupCtx.GroupID, []byte("message"), []byte("test_nonce"))
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
