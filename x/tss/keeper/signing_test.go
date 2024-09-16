package keeper_test

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	tsstestutil "github.com/bandprotocol/chain/v2/x/tss/testutil"
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

func TestGetRandomMembersSuccess(t *testing.T) {
	s := NewKeeperTestSuite(t)
	ctx, k := s.Ctx, s.Keeper

	// create a group context and mock the result.
	groupCtx, err := tsstestutil.CompleteGroupCreation(ctx, k, 4, 2)
	require.NoError(t, err)
	s.RollingseedKeeper.EXPECT().GetRollingSeed(gomock.Any()).
		Return([]byte("RandomStringThatShouldBeLongEnough"))

	// call GetRandomMembers and check results
	ams, err := k.GetRandomMembers(ctx, groupCtx.GroupID, []byte("test_nonce"))
	require.NoError(t, err)

	memberIDs := []tss.MemberID{}
	for _, am := range ams {
		memberIDs = append(memberIDs, am.ID)
	}
	require.Equal(t, memberIDs, []tss.MemberID{3, 4})
}

func TestGetRandomMembersInsufficientActiveMembers(t *testing.T) {
	s := NewKeeperTestSuite(t)
	ctx, k := s.Ctx, s.Keeper

	// create a group context.
	groupCtx, err := tsstestutil.CompleteGroupCreation(ctx, k, 4, 2)
	require.NoError(t, err)

	// set every member to inactive.
	members, err := k.GetGroupMembers(ctx, groupCtx.GroupID)
	require.NoError(t, err)
	for _, member := range members[:len(members)-1] {
		member.IsActive = false
		k.SetMember(ctx, member)
	}

	_, err = k.GetRandomMembers(ctx, tss.GroupID(1), []byte("test_nonce"))
	require.ErrorIs(t, err, types.ErrInsufficientActiveMembers)
}

func TestGetRandomMembersNoActiveMember(t *testing.T) {
	s := NewKeeperTestSuite(t)
	ctx, k := s.Ctx, s.Keeper

	// create a group context.
	groupCtx, err := tsstestutil.CompleteGroupCreation(ctx, k, 4, 2)
	require.NoError(t, err)

	// set every member to inactive.
	members, err := k.GetGroupMembers(ctx, groupCtx.GroupID)
	require.NoError(t, err)
	for _, member := range members {
		member.IsActive = false
		k.SetMember(ctx, member)
	}

	_, err = k.GetRandomMembers(ctx, tss.GroupID(1), []byte("test_nonce"))
	require.ErrorIs(t, err, types.ErrNoActiveMember)
}

func TestAssignedMembers(t *testing.T) {
	s := NewKeeperTestSuite(t)
	ctx, k := s.Ctx, s.Keeper

	// create a group context and mock the result.
	groupCtx, err := tsstestutil.CompleteGroupCreation(ctx, k, 4, 2)
	require.NoError(t, err)
	s.RollingseedKeeper.EXPECT().GetRollingSeed(gomock.Any()).
		Return([]byte("RandomStringThatShouldBeLongEnough"))

	DEQueuesBeforePolling := []types.DEQueue{}
	for _, acc := range groupCtx.Accounts {
		deQueue := k.GetDEQueue(ctx, acc.Address)
		DEQueuesBeforePolling = append(DEQueuesBeforePolling, deQueue)
	}

	// check result if no error and check selected MemberIDs
	assignedMembers, err := k.AssignMembers(ctx, groupCtx.GroupID, []byte("message"), []byte("test_nonce"))
	require.NoError(t, err)
	assignedMemberIDs := assignedMembers.MemberIDs()
	require.Equal(t, assignedMemberIDs, []tss.MemberID{3, 4})

	// check if DE is correctly polled; check DEQueue Head is shifted by 1.
	for id, acc := range groupCtx.Accounts {
		deQueue := k.GetDEQueue(ctx, acc.Address)
		expectedHead := DEQueuesBeforePolling[id].Head
		if slices.Contains(assignedMemberIDs, tss.MemberID(id+1)) {
			expectedHead += 1
		}
		require.Equal(t, expectedHead, deQueue.Head)
	}
}

func TestInitiateNewSigningRoundForFirstRound(t *testing.T) {
	s := NewKeeperTestSuite(t)
	ctx, k := s.Ctx, s.Keeper

	// create a group context and mock the result.
	groupCtx, err := tsstestutil.CompleteGroupCreation(ctx, k, 4, 2)
	require.NoError(t, err)
	s.RollingseedKeeper.EXPECT().GetRollingSeed(gomock.Any()).
		Return([]byte("RandomStringThatShouldBeLongEnough"))

	signingID, err := k.CreateSigning(ctx, groupCtx.GroupID, []byte("originator"), []byte("message"))
	require.NoError(t, err)
	require.Equal(t, tss.SigningID(1), signingID)

	err = k.InitiateNewSigningRound(ctx, signingID)
	require.NoError(t, err)

	// check signing
	signing, err := k.GetSigning(ctx, signingID)
	require.NoError(t, err)
	require.Equal(t, uint64(1), signing.CurrentAttempt)
	require.NotNil(t, signing.GroupPubNonce)

	// check signing expiration
	signingExpirations := k.GetSigningExpirations(ctx)
	require.Len(t, signingExpirations, 1)
	require.Equal(
		t,
		types.SigningExpiration{SigningID: signingID, SigningAttempt: signing.CurrentAttempt},
		signingExpirations[0],
	)

	// check signingAttempt
	sa, err := k.GetSigningAttempt(ctx, signingID, signing.CurrentAttempt)
	require.NoError(t, err)
	require.Equal(t, uint64(ctx.BlockHeight()+int64(k.GetParams(ctx).SigningPeriod)), sa.ExpiredHeight)

	assignedMembers := types.AssignedMembers(sa.AssignedMembers)
	require.Equal(t, []tss.MemberID{1, 2}, assignedMembers.MemberIDs())
}

func TestInitiateNewSigningRoundForThirdRound(t *testing.T) {
	s := NewKeeperTestSuite(t)
	ctx, k := s.Ctx, s.Keeper

	// create a group context and mock the result.
	groupCtx, err := tsstestutil.CompleteGroupCreation(ctx, k, 4, 2)
	require.NoError(t, err)
	s.RollingseedKeeper.EXPECT().GetRollingSeed(gomock.Any()).
		Return([]byte("RandomStringThatShouldBeLongEnough"))

	signingID, err := k.CreateSigning(ctx, groupCtx.GroupID, []byte("originator"), []byte("message"))
	require.NoError(t, err)
	require.Equal(t, tss.SigningID(1), signingID)

	// update signing attempt
	signing, err := k.GetSigning(ctx, signingID)
	require.NoError(t, err)
	signing.CurrentAttempt = 2
	k.SetSigning(ctx, signing)

	err = k.InitiateNewSigningRound(ctx, signingID)
	require.NoError(t, err)

	// check signing
	signing, err = k.GetSigning(ctx, signingID)
	require.NoError(t, err)
	require.Equal(t, uint64(3), signing.CurrentAttempt)
	require.NotNil(t, signing.GroupPubNonce)

	// check signing expiration
	signingExpirations := k.GetSigningExpirations(ctx)
	require.Len(t, signingExpirations, 1)
	require.Equal(
		t,
		types.SigningExpiration{SigningID: signingID, SigningAttempt: signing.CurrentAttempt},
		signingExpirations[0],
	)

	// check signingAttempt
	sa, err := k.GetSigningAttempt(ctx, signingID, signing.CurrentAttempt)
	require.NoError(t, err)
	require.Equal(t, uint64(ctx.BlockHeight()+int64(k.GetParams(ctx).SigningPeriod)), sa.ExpiredHeight)

	assignedMembers := types.AssignedMembers(sa.AssignedMembers)
	require.Equal(t, []tss.MemberID{2, 4}, assignedMembers.MemberIDs())
}

func TestInitiateNewSigningRoundOverMaxAttempt(t *testing.T) {
	s := NewKeeperTestSuite(t)
	ctx, k := s.Ctx, s.Keeper

	// create a group context and mock the result.
	groupCtx, err := tsstestutil.CompleteGroupCreation(ctx, k, 4, 2)
	require.NoError(t, err)

	signingID, err := k.CreateSigning(ctx, groupCtx.GroupID, []byte("originator"), []byte("message"))
	require.NoError(t, err)
	require.Equal(t, tss.SigningID(1), signingID)

	// update signing attempt
	signing, err := k.GetSigning(ctx, signingID)
	require.NoError(t, err)
	signing.CurrentAttempt = k.GetParams(ctx).MaxSigningAttempt
	k.SetSigning(ctx, signing)

	err = k.InitiateNewSigningRound(ctx, signingID)
	require.ErrorIs(t, err, types.ErrMaxSigningAttemptReached)
}

func TestRequestSigning(t *testing.T) {
	s := NewKeeperTestSuite(t)
	ctx, k := s.Ctx, s.Keeper

	groupCtx, err := tsstestutil.CompleteGroupCreation(ctx, k, 4, 2)
	groupID := groupCtx.GroupID
	require.NoError(t, err)

	s.RollingseedKeeper.EXPECT().GetRollingSeed(gomock.Any()).
		Return([]byte("RandomStringThatShouldBeLongEnough"))

	// Create a new request for the request signature
	content := types.NewTextSignatureOrder([]byte("example"))
	signingID, err := k.RequestSigning(ctx, groupID, types.DirectOriginator{}, content)
	require.NoError(t, err)

	signing, err := k.GetSigning(ctx, signingID)
	require.NoError(t, err)
	require.Equal(t, groupID, signing.GroupID)
	require.Equal(t, types.SIGNING_STATUS_WAITING, signing.Status)
}

func TestMustGetCurrentAssignedMembers(t *testing.T) {
	s := NewKeeperTestSuite(t)
	ctx, k := s.Ctx, s.Keeper

	signing := GetExampleSigning()
	k.SetSigning(ctx, signing)
	sa := GetExampleSigningAttempt()
	k.SetSigningAttempt(ctx, sa)

	// Get current assigned members
	assignedMembers := k.MustGetCurrentAssignedMembers(ctx, signing.ID)
	require.Len(t, assignedMembers, 2)
	require.Equal(t, "band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs", assignedMembers[0].String())

	// Get incorrect signingID
	require.PanicsWithError(t, "failed to get Signing with ID: 999: signing not found", func() {
		k.MustGetCurrentAssignedMembers(ctx, 999)
	})
}
