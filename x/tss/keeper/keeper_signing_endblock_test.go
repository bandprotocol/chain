package keeper_test

import (
	"go.uber.org/mock/gomock"

	"github.com/bandprotocol/chain/v3/pkg/tss"
	tsstestutil "github.com/bandprotocol/chain/v3/x/tss/testutil"
	"github.com/bandprotocol/chain/v3/x/tss/types"
)

func (s *KeeperTestSuite) TestSetGetPendingProcessSigning() {
	ctx, k := s.ctx, s.keeper

	signingIDs := []tss.SigningID{1, 2}
	k.SetPendingProcessSignings(ctx, types.PendingProcessSignings{SigningIDs: signingIDs})

	got := k.GetPendingProcessSignings(ctx)
	s.Require().Equal(signingIDs, got)
}

func (s *KeeperTestSuite) TestAddPendingProcessSigning() {
	ctx, k := s.ctx, s.keeper

	k.AddPendingProcessSigning(ctx, 1)
	k.AddPendingProcessSigning(ctx, 2)

	got := k.GetPendingProcessSignings(ctx)
	s.Require().Equal([]tss.SigningID{1, 2}, got)
}

func (s *KeeperTestSuite) TestAggregatePartialSignatures() {
	ctx, k := s.ctx, s.keeper

	groupCtx, err := tsstestutil.CompleteGroupCreation(ctx, k, 4, 2)
	s.Require().NoError(err)

	s.rollingseedKeeper.EXPECT().GetRollingSeed(gomock.Any()).
		Return([]byte("RandomStringThatShouldBeLongEnough"))
	signingID, err := k.RequestSigning(
		ctx,
		groupCtx.GroupID,
		types.DirectOriginator{},
		&types.TextSignatureOrder{Message: []byte("test")},
	)
	s.Require().NoError(err)

	signing, err := k.GetSigning(ctx, signingID)
	s.Require().NoError(err)
	s.Require().Equal(types.SIGNING_STATUS_WAITING, signing.Status)

	err = groupCtx.SubmitSignature(ctx, k, s.msgServer, signing.ID)
	s.Require().NoError(err)

	s.Require().Equal(types.SIGNING_STATUS_WAITING, signing.Status)

	err = k.AggregatePartialSignatures(ctx, signing.ID)
	s.Require().NoError(err)

	signing, err = k.GetSigning(ctx, signing.ID)
	s.Require().NoError(err)
	s.Require().Equal(types.SIGNING_STATUS_SUCCESS, signing.Status)
	s.Require().NotNil(signing.Signature)
}

func (s *KeeperTestSuite) TestHandleFailedSigning() {
	ctx, k := s.ctx, s.keeper

	groupCtx, err := tsstestutil.CompleteGroupCreation(ctx, k, 4, 2)
	s.Require().NoError(err)

	s.rollingseedKeeper.EXPECT().GetRollingSeed(gomock.Any()).
		Return([]byte("RandomStringThatShouldBeLongEnough"))
	signingID, err := k.RequestSigning(
		ctx,
		groupCtx.GroupID,
		types.DirectOriginator{},
		&types.TextSignatureOrder{Message: []byte("test")},
	)
	s.Require().NoError(err)

	signing, err := k.GetSigning(ctx, signingID)
	s.Require().NoError(err)
	s.Require().Equal(types.SIGNING_STATUS_WAITING, signing.Status)

	k.HandleFailedSigning(ctx, signing.ID, "test")

	signing, err = k.GetSigning(ctx, signing.ID)
	s.Require().NoError(err)
	s.Require().Equal(types.SIGNING_STATUS_FALLEN, signing.Status)
}

func (s *KeeperTestSuite) TestGetSetSigningExpirations() {
	ctx, k := s.ctx, s.keeper

	signingExpirations := []types.SigningExpiration{{SigningID: 1, SigningAttempt: 1}}
	k.SetSigningExpirations(ctx, types.SigningExpirations{
		SigningExpirations: signingExpirations,
	})

	got := k.GetSigningExpirations(ctx)
	s.Require().Equal(signingExpirations, got)
}

func (s *KeeperTestSuite) TestAddSigningExpirations() {
	ctx, k := s.ctx, s.keeper

	k.AddSigningExpiration(ctx, 1, 1)
	k.AddSigningExpiration(ctx, 1, 2)

	got := k.GetSigningExpirations(ctx)
	s.Require().Equal(
		[]types.SigningExpiration{
			{SigningID: 1, SigningAttempt: 1},
			{SigningID: 1, SigningAttempt: 2},
		},
		got,
	)
}

func (s *KeeperTestSuite) TestHandleSigningEndblockMembersSubmitSignature() {
	ctx, k := s.ctx, s.keeper

	groupCtx, err := tsstestutil.CompleteGroupCreation(ctx, k, 4, 2)
	s.Require().NoError(err)

	s.rollingseedKeeper.EXPECT().GetRollingSeed(gomock.Any()).
		Return([]byte("RandomStringThatShouldBeLongEnough")).AnyTimes()

	signingID, err := k.RequestSigning(
		ctx,
		groupCtx.GroupID,
		types.DirectOriginator{},
		&types.TextSignatureOrder{Message: []byte("test")},
	)
	s.Require().NoError(err)
	err = groupCtx.SubmitSignature(ctx, k, s.msgServer, signingID)
	s.Require().NoError(err)

	k.HandleSigningEndBlock(ctx)

	// check signingID status
	signing, err := k.GetSigning(ctx, signingID)
	s.Require().NoError(err)
	s.Require().Equal(types.SIGNING_STATUS_SUCCESS, signing.Status)
	s.Require().NotNil(signing.Signature)

	// check signingID interim data
	sa, err := k.GetSigningAttempt(ctx, signing.ID, signing.CurrentAttempt)
	s.Require().NoError(err)

	signatureCount := k.GetPartialSignatureCount(ctx, signing.ID, signing.CurrentAttempt)
	s.Require().Equal(signatureCount, uint64(len(sa.AssignedMembers)))

	partialSigs := k.GetPartialSignatures(ctx, signing.ID, signing.CurrentAttempt)
	s.Require().Len(partialSigs, len(sa.AssignedMembers))

	// check pendingProcessSignings
	pendingProcessSignings := k.GetPendingProcessSignings(ctx)
	s.Require().Len(pendingProcessSignings, 0)

	// check signingExpirations
	signingExpirations := k.GetSigningExpirations(ctx)
	s.Require().Len(signingExpirations, 1)

	// skip time to handle signing expiration
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + int64(k.GetParams(ctx).SigningPeriod))
	k.HandleSigningEndBlock(ctx)

	// check signing data
	_, err = k.GetSigningAttempt(ctx, signing.ID, signing.CurrentAttempt)
	s.Require().ErrorIs(err, types.ErrSigningAttemptNotFound)

	signatureCount = k.GetPartialSignatureCount(ctx, signing.ID, signing.CurrentAttempt)
	s.Require().Zero(signatureCount)

	partialSigs = k.GetPartialSignatures(ctx, signing.ID, signing.CurrentAttempt)
	s.Require().Len(partialSigs, 0)

	// check pendingProcessSignings
	pendingProcessSignings = k.GetPendingProcessSignings(ctx)
	s.Require().Len(pendingProcessSignings, 0)

	// check signingExpirations
	signingExpirations = k.GetSigningExpirations(ctx)
	s.Require().Len(signingExpirations, 0)
}

func (s *KeeperTestSuite) TestHandleSigningEndblockTimeoutSigning() {
	ctx, k := s.ctx, s.keeper

	groupCtx, err := tsstestutil.CompleteGroupCreation(ctx, k, 4, 2)
	s.Require().NoError(err)

	s.rollingseedKeeper.EXPECT().GetRollingSeed(gomock.Any()).
		Return([]byte("RandomStringThatShouldBeLongEnough")).AnyTimes()

	signingID, err := k.RequestSigning(
		ctx,
		groupCtx.GroupID,
		types.DirectOriginator{},
		&types.TextSignatureOrder{Message: []byte("test")},
	)
	s.Require().NoError(err)

	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + int64(k.GetParams(ctx).SigningPeriod))

	k.HandleSigningEndBlock(ctx)

	// check signingID status
	signing, err := k.GetSigning(ctx, signingID)
	s.Require().NoError(err)
	s.Require().Equal(types.SIGNING_STATUS_WAITING, signing.Status)
	s.Require().Nil(signing.Signature)
	s.Require().Equal(uint64(2), signing.CurrentAttempt)

	// check signingID interim data
	sa, err := k.GetSigningAttempt(ctx, signing.ID, signing.CurrentAttempt)
	s.Require().NoError(err)
	s.Require().Equal(uint64(ctx.BlockHeight()+int64(k.GetParams(ctx).SigningPeriod)), sa.ExpiredHeight)

	// check pendingProcessSignings
	pendingProcessSignings := k.GetPendingProcessSignings(ctx)
	s.Require().Len(pendingProcessSignings, 0)

	// check signingExpirations
	signingExpirations := k.GetSigningExpirations(ctx)
	s.Require().Len(signingExpirations, 1)
	s.Require().Equal(
		types.SigningExpiration{SigningID: signing.ID, SigningAttempt: signing.CurrentAttempt},
		signingExpirations[0],
	)
}

func (s *KeeperTestSuite) TestHandleSigningEndblockFailAggregate() {
	ctx, k := s.ctx, s.keeper

	groupCtx, err := tsstestutil.CompleteGroupCreation(ctx, k, 4, 2)
	s.Require().NoError(err)

	s.rollingseedKeeper.EXPECT().GetRollingSeed(gomock.Any()).
		Return([]byte("RandomStringThatShouldBeLongEnough")).AnyTimes()

	signingID, err := k.RequestSigning(
		ctx,
		groupCtx.GroupID,
		types.DirectOriginator{},
		&types.TextSignatureOrder{Message: []byte("test")},
	)
	s.Require().NoError(err)

	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
	err = groupCtx.SubmitSignature(ctx, k, s.msgServer, signingID)
	s.Require().NoError(err)

	// change raw message, so that signature aggregation will fail.
	signing, err := k.GetSigning(ctx, signingID)
	s.Require().NoError(err)
	signing.Message = []byte("message")
	k.SetSigning(ctx, signing)

	k.HandleSigningEndBlock(ctx)

	// check signingID status
	signing, err = k.GetSigning(ctx, signingID)
	s.Require().NoError(err)
	s.Require().Equal(types.SIGNING_STATUS_WAITING, signing.Status)
	s.Require().Nil(signing.Signature)
	s.Require().Equal(uint64(2), signing.CurrentAttempt)

	// check signingID interim data
	sa, err := k.GetSigningAttempt(ctx, signing.ID, signing.CurrentAttempt)
	s.Require().NoError(err)
	s.Require().Equal(uint64(ctx.BlockHeight()+int64(k.GetParams(ctx).SigningPeriod)), sa.ExpiredHeight)

	// check pendingProcessSignings
	pendingProcessSignings := k.GetPendingProcessSignings(ctx)
	s.Require().Len(pendingProcessSignings, 0)

	// check signingExpirations
	signingExpirations := k.GetSigningExpirations(ctx)
	s.Require().Len(signingExpirations, 2)
	s.Require().Equal(
		types.SigningExpiration{SigningID: signing.ID, SigningAttempt: signing.CurrentAttempt},
		signingExpirations[1],
	)

	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + int64(k.GetParams(ctx).SigningPeriod) - 1)
	k.HandleSigningEndBlock(ctx)

	// check signingID status
	signing, err = k.GetSigning(ctx, signingID)
	s.Require().NoError(err)
	s.Require().Equal(types.SIGNING_STATUS_WAITING, signing.Status)
	s.Require().Nil(signing.Signature)
	s.Require().Equal(uint64(2), signing.CurrentAttempt)

	// check signingExpirations
	signingExpirations = k.GetSigningExpirations(ctx)
	s.Require().Len(signingExpirations, 1)
	s.Require().Equal(
		types.SigningExpiration{SigningID: signing.ID, SigningAttempt: signing.CurrentAttempt},
		signingExpirations[0],
	)
}

func (s *KeeperTestSuite) TestHandleSigningEndblockFailAggregateAndExpired() {
	ctx, k := s.ctx, s.keeper

	groupCtx, err := tsstestutil.CompleteGroupCreation(ctx, k, 4, 2)
	s.Require().NoError(err)

	s.rollingseedKeeper.EXPECT().GetRollingSeed(gomock.Any()).
		Return([]byte("RandomStringThatShouldBeLongEnough")).AnyTimes()

	signingID, err := k.RequestSigning(
		ctx,
		groupCtx.GroupID,
		types.DirectOriginator{},
		&types.TextSignatureOrder{Message: []byte("test")},
	)
	s.Require().NoError(err)

	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + int64(k.GetParams(ctx).SigningPeriod))
	err = groupCtx.SubmitSignature(ctx, k, s.msgServer, signingID)
	s.Require().NoError(err)

	// change raw message, so that signature aggregation will fail.
	signing, err := k.GetSigning(ctx, signingID)
	s.Require().NoError(err)
	signing.Message = []byte("message")
	k.SetSigning(ctx, signing)

	k.HandleSigningEndBlock(ctx)

	// check signingID status
	signing, err = k.GetSigning(ctx, signingID)
	s.Require().NoError(err)
	s.Require().Equal(types.SIGNING_STATUS_WAITING, signing.Status)
	s.Require().Nil(signing.Signature)
	s.Require().Equal(uint64(2), signing.CurrentAttempt)

	// check signingID interim data
	sa, err := k.GetSigningAttempt(ctx, signing.ID, signing.CurrentAttempt)
	s.Require().NoError(err)
	s.Require().Equal(uint64(ctx.BlockHeight()+int64(k.GetParams(ctx).SigningPeriod)), sa.ExpiredHeight)

	// check pendingProcessSignings
	pendingProcessSignings := k.GetPendingProcessSignings(ctx)
	s.Require().Len(pendingProcessSignings, 0)

	// check signingExpirations
	signingExpirations := k.GetSigningExpirations(ctx)
	s.Require().Len(signingExpirations, 1)
	s.Require().Equal(
		types.SigningExpiration{SigningID: signing.ID, SigningAttempt: signing.CurrentAttempt},
		signingExpirations[0],
	)
}
