package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	tsstestutil "github.com/bandprotocol/chain/v2/x/tss/testutil"
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

func TestSetGetPendingProcessSigning(t *testing.T) {
	s := NewKeeperTestSuite(t)
	ctx, k := s.Ctx, s.Keeper

	signingIDs := []tss.SigningID{1, 2}
	k.SetPendingProcessSignings(ctx, types.PendingProcessSignings{SigningIDs: signingIDs})

	got := k.GetPendingProcessSignings(ctx)
	require.Equal(t, signingIDs, got)
}

func TestAddPendingProcessSigning(t *testing.T) {
	s := NewKeeperTestSuite(t)
	ctx, k := s.Ctx, s.Keeper

	k.AddPendingProcessSigning(ctx, 1)
	k.AddPendingProcessSigning(ctx, 2)

	got := k.GetPendingProcessSignings(ctx)
	require.Equal(t, []tss.SigningID{1, 2}, got)
}

func TestAggregatePartialSignatures(t *testing.T) {
	s := NewKeeperTestSuite(t)
	ctx, k := s.Ctx, s.Keeper

	groupCtx, err := tsstestutil.CompleteGroupCreation(ctx, k, 4, 2)
	require.NoError(t, err)

	s.RollingseedKeeper.EXPECT().GetRollingSeed(gomock.Any()).
		Return([]byte("RandomStringThatShouldBeLongEnough"))
	signingID, err := k.RequestSigning(
		ctx,
		groupCtx.GroupID,
		types.DirectOriginator{},
		&types.TextSignatureOrder{Message: []byte("test")},
	)
	require.NoError(t, err)

	signing, err := k.GetSigning(ctx, signingID)
	require.NoError(t, err)
	require.Equal(t, types.SIGNING_STATUS_WAITING, signing.Status)

	err = groupCtx.SubmitSignature(ctx, k, s.MsgServer, signing.ID)
	require.NoError(t, err)

	require.Equal(t, types.SIGNING_STATUS_WAITING, signing.Status)

	err = k.AggregatePartialSignatures(ctx, signing.ID)
	require.NoError(t, err)

	signing, err = k.GetSigning(ctx, signing.ID)
	require.NoError(t, err)
	require.Equal(t, types.SIGNING_STATUS_SUCCESS, signing.Status)
	require.NotNil(t, signing.Signature)
}

func TestHandleFailedSigning(t *testing.T) {
	s := NewKeeperTestSuite(t)
	ctx, k := s.Ctx, s.Keeper

	groupCtx, err := tsstestutil.CompleteGroupCreation(ctx, k, 4, 2)
	require.NoError(t, err)

	s.RollingseedKeeper.EXPECT().GetRollingSeed(gomock.Any()).
		Return([]byte("RandomStringThatShouldBeLongEnough"))
	signingID, err := k.RequestSigning(
		ctx,
		groupCtx.GroupID,
		types.DirectOriginator{},
		&types.TextSignatureOrder{Message: []byte("test")},
	)
	require.NoError(t, err)

	signing, err := k.GetSigning(ctx, signingID)
	require.NoError(t, err)
	require.Equal(t, types.SIGNING_STATUS_WAITING, signing.Status)

	k.HandleFailedSigning(ctx, signing.ID, "test")

	signing, err = k.GetSigning(ctx, signing.ID)
	require.NoError(t, err)
	require.Equal(t, types.SIGNING_STATUS_FALLEN, signing.Status)
}

func TestGetSetSigningExpirations(t *testing.T) {
	s := NewKeeperTestSuite(t)
	ctx, k := s.Ctx, s.Keeper

	signingExpirations := []types.SigningExpiration{{SigningID: 1, SigningAttempt: 1}}
	k.SetSigningExpirations(ctx, types.SigningExpirations{
		SigningExpirations: signingExpirations,
	})

	got := k.GetSigningExpirations(ctx)
	require.Equal(t, signingExpirations, got)
}

func TestAddSigningExpirations(t *testing.T) {
	s := NewKeeperTestSuite(t)
	ctx, k := s.Ctx, s.Keeper

	k.AddSigningExpiration(ctx, 1, 1)
	k.AddSigningExpiration(ctx, 1, 2)

	got := k.GetSigningExpirations(ctx)
	require.Equal(
		t,
		[]types.SigningExpiration{
			{SigningID: 1, SigningAttempt: 1},
			{SigningID: 1, SigningAttempt: 2},
		},
		got,
	)
}

func TestHandleSigningEndblockMembersSubmitSignature(t *testing.T) {
	s := NewKeeperTestSuite(t)
	ctx, k := s.Ctx, s.Keeper

	groupCtx, err := tsstestutil.CompleteGroupCreation(ctx, k, 4, 2)
	require.NoError(t, err)

	s.RollingseedKeeper.EXPECT().GetRollingSeed(gomock.Any()).
		Return([]byte("RandomStringThatShouldBeLongEnough")).AnyTimes()

	signingID, err := k.RequestSigning(
		ctx,
		groupCtx.GroupID,
		types.DirectOriginator{},
		&types.TextSignatureOrder{Message: []byte("test")},
	)
	require.NoError(t, err)
	err = groupCtx.SubmitSignature(ctx, k, s.MsgServer, signingID)
	require.NoError(t, err)

	k.HandleSigningEndBlock(ctx)

	// check signingID status
	signing, err := k.GetSigning(ctx, signingID)
	require.NoError(t, err)
	require.Equal(t, types.SIGNING_STATUS_SUCCESS, signing.Status)
	require.NotNil(t, signing.Signature)

	// check signingID interim data
	sa, err := k.GetSigningAttempt(ctx, signing.ID, signing.CurrentAttempt)
	require.NoError(t, err)

	signatureCount := k.GetPartialSignatureCount(ctx, signing.ID, signing.CurrentAttempt)
	require.Equal(t, signatureCount, uint64(len(sa.AssignedMembers)))

	partialSigs := k.GetPartialSignatures(ctx, signing.ID, signing.CurrentAttempt)
	require.Len(t, partialSigs, len(sa.AssignedMembers))

	// check pendingProcessSignings
	pendingProcessSignings := k.GetPendingProcessSignings(ctx)
	require.Len(t, pendingProcessSignings, 0)

	// check signingExpirations
	signingExpirations := k.GetSigningExpirations(ctx)
	require.Len(t, signingExpirations, 1)

	// skip time to handle signing expiration
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + int64(k.GetParams(ctx).SigningPeriod))
	k.HandleSigningEndBlock(ctx)

	// check signing data
	_, err = k.GetSigningAttempt(ctx, signing.ID, signing.CurrentAttempt)
	require.ErrorIs(t, err, types.ErrSigningAttemptNotFound)

	signatureCount = k.GetPartialSignatureCount(ctx, signing.ID, signing.CurrentAttempt)
	require.Zero(t, signatureCount)

	partialSigs = k.GetPartialSignatures(ctx, signing.ID, signing.CurrentAttempt)
	require.Len(t, partialSigs, 0)

	// check pendingProcessSignings
	pendingProcessSignings = k.GetPendingProcessSignings(ctx)
	require.Len(t, pendingProcessSignings, 0)

	// check signingExpirations
	signingExpirations = k.GetSigningExpirations(ctx)
	require.Len(t, signingExpirations, 0)
}

func TestHandleSigningEndblockTimeoutSigning(t *testing.T) {
	s := NewKeeperTestSuite(t)
	ctx, k := s.Ctx, s.Keeper

	groupCtx, err := tsstestutil.CompleteGroupCreation(ctx, k, 4, 2)
	require.NoError(t, err)

	s.RollingseedKeeper.EXPECT().GetRollingSeed(gomock.Any()).
		Return([]byte("RandomStringThatShouldBeLongEnough")).AnyTimes()

	signingID, err := k.RequestSigning(
		ctx,
		groupCtx.GroupID,
		types.DirectOriginator{},
		&types.TextSignatureOrder{Message: []byte("test")},
	)
	require.NoError(t, err)

	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + int64(k.GetParams(ctx).SigningPeriod))

	k.HandleSigningEndBlock(ctx)

	// check signingID status
	signing, err := k.GetSigning(ctx, signingID)
	require.NoError(t, err)
	require.Equal(t, types.SIGNING_STATUS_WAITING, signing.Status)
	require.Nil(t, signing.Signature)
	require.Equal(t, uint64(2), signing.CurrentAttempt)

	// check signingID interim data
	sa, err := k.GetSigningAttempt(ctx, signing.ID, signing.CurrentAttempt)
	require.NoError(t, err)
	require.Equal(t, uint64(ctx.BlockHeight()+int64(k.GetParams(ctx).SigningPeriod)), sa.ExpiredHeight)

	// check pendingProcessSignings
	pendingProcessSignings := k.GetPendingProcessSignings(ctx)
	require.Len(t, pendingProcessSignings, 0)

	// check signingExpirations
	signingExpirations := k.GetSigningExpirations(ctx)
	require.Len(t, signingExpirations, 1)
	require.Equal(
		t,
		types.SigningExpiration{SigningID: signing.ID, SigningAttempt: signing.CurrentAttempt},
		signingExpirations[0],
	)
}

func TestHandleSigningEndblockFailAggregate(t *testing.T) {
	s := NewKeeperTestSuite(t)
	ctx, k := s.Ctx, s.Keeper

	groupCtx, err := tsstestutil.CompleteGroupCreation(ctx, k, 4, 2)
	require.NoError(t, err)

	s.RollingseedKeeper.EXPECT().GetRollingSeed(gomock.Any()).
		Return([]byte("RandomStringThatShouldBeLongEnough")).AnyTimes()

	signingID, err := k.RequestSigning(
		ctx,
		groupCtx.GroupID,
		types.DirectOriginator{},
		&types.TextSignatureOrder{Message: []byte("test")},
	)
	require.NoError(t, err)

	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
	err = groupCtx.SubmitSignature(ctx, k, s.MsgServer, signingID)
	require.NoError(t, err)

	// change raw message, so that signature aggregation will fail.
	signing, err := k.GetSigning(ctx, signingID)
	require.NoError(t, err)
	signing.Message = []byte("message")
	k.SetSigning(ctx, signing)

	k.HandleSigningEndBlock(ctx)

	// check signingID status
	signing, err = k.GetSigning(ctx, signingID)
	require.NoError(t, err)
	require.Equal(t, types.SIGNING_STATUS_WAITING, signing.Status)
	require.Nil(t, signing.Signature)
	require.Equal(t, uint64(2), signing.CurrentAttempt)

	// check signingID interim data
	sa, err := k.GetSigningAttempt(ctx, signing.ID, signing.CurrentAttempt)
	require.NoError(t, err)
	require.Equal(t, uint64(ctx.BlockHeight()+int64(k.GetParams(ctx).SigningPeriod)), sa.ExpiredHeight)

	// check pendingProcessSignings
	pendingProcessSignings := k.GetPendingProcessSignings(ctx)
	require.Len(t, pendingProcessSignings, 0)

	// check signingExpirations
	signingExpirations := k.GetSigningExpirations(ctx)
	require.Len(t, signingExpirations, 2)
	require.Equal(
		t,
		types.SigningExpiration{SigningID: signing.ID, SigningAttempt: signing.CurrentAttempt},
		signingExpirations[1],
	)

	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + int64(k.GetParams(ctx).SigningPeriod) - 1)
	k.HandleSigningEndBlock(ctx)

	// check signingID status
	signing, err = k.GetSigning(ctx, signingID)
	require.NoError(t, err)
	require.Equal(t, types.SIGNING_STATUS_WAITING, signing.Status)
	require.Nil(t, signing.Signature)
	require.Equal(t, uint64(2), signing.CurrentAttempt)

	// check signingExpirations
	signingExpirations = k.GetSigningExpirations(ctx)
	require.Len(t, signingExpirations, 1)
	require.Equal(
		t,
		types.SigningExpiration{SigningID: signing.ID, SigningAttempt: signing.CurrentAttempt},
		signingExpirations[0],
	)
}

func TestHandleSigningEndblockFailAggregateAndExpired(t *testing.T) {
	s := NewKeeperTestSuite(t)
	ctx, k := s.Ctx, s.Keeper

	groupCtx, err := tsstestutil.CompleteGroupCreation(ctx, k, 4, 2)
	require.NoError(t, err)

	s.RollingseedKeeper.EXPECT().GetRollingSeed(gomock.Any()).
		Return([]byte("RandomStringThatShouldBeLongEnough")).AnyTimes()

	signingID, err := k.RequestSigning(
		ctx,
		groupCtx.GroupID,
		types.DirectOriginator{},
		&types.TextSignatureOrder{Message: []byte("test")},
	)
	require.NoError(t, err)

	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + int64(k.GetParams(ctx).SigningPeriod))
	err = groupCtx.SubmitSignature(ctx, k, s.MsgServer, signingID)
	require.NoError(t, err)

	// change raw message, so that signature aggregation will fail.
	signing, err := k.GetSigning(ctx, signingID)
	require.NoError(t, err)
	signing.Message = []byte("message")
	k.SetSigning(ctx, signing)

	k.HandleSigningEndBlock(ctx)

	// check signingID status
	signing, err = k.GetSigning(ctx, signingID)
	require.NoError(t, err)
	require.Equal(t, types.SIGNING_STATUS_WAITING, signing.Status)
	require.Nil(t, signing.Signature)
	require.Equal(t, uint64(2), signing.CurrentAttempt)

	// check signingID interim data
	sa, err := k.GetSigningAttempt(ctx, signing.ID, signing.CurrentAttempt)
	require.NoError(t, err)
	require.Equal(t, uint64(ctx.BlockHeight()+int64(k.GetParams(ctx).SigningPeriod)), sa.ExpiredHeight)

	// check pendingProcessSignings
	pendingProcessSignings := k.GetPendingProcessSignings(ctx)
	require.Len(t, pendingProcessSignings, 0)

	// check signingExpirations
	signingExpirations := k.GetSigningExpirations(ctx)
	require.Len(t, signingExpirations, 1)
	require.Equal(
		t,
		types.SigningExpiration{SigningID: signing.ID, SigningAttempt: signing.CurrentAttempt},
		signingExpirations[0],
	)
}
