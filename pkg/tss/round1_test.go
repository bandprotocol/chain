package tss_test

import (
	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/pkg/tss/testutil"
)

func (suite *TSSTestSuite) TestGenerateRound1Info() {
	mid := tss.NewMemberID(1)
	dkgContext := []byte("DKGContext")
	threshold := uint64(2)

	data, err := tss.GenerateRound1Info(mid, threshold, dkgContext)
	suite.Require().NoError(err)
	err = tss.VerifyOneTimeSignature(mid, dkgContext, data.OneTimeSignature, data.OneTimePubKey)
	suite.Require().NoError(err)

	err = tss.VerifyA0Signature(mid, dkgContext, data.A0Signature, data.A0PubKey)
	suite.Require().NoError(err)

	for i, coeff := range data.Coefficients {
		commit := coeff.Point()
		suite.Require().NoError(err)
		suite.Require().Equal(data.CoefficientCommits[i], commit)
	}
}

func (suite *TSSTestSuite) TestSignOneTime() {
	suite.RunOnMember(suite.testCases, func(tc testutil.TestCase, member testutil.Member) {
		signature, err := tss.SignOneTime(
			member.ID,
			tc.Group.DKGContext,
			member.OneTimePubKey(),
			member.OneTimePrivKey,
		)
		suite.Require().NoError(err)

		err = tss.VerifyOneTimeSignature(member.ID, tc.Group.DKGContext, signature, member.OneTimePubKey())
		suite.Require().NoError(err)
	})
}

func (suite *TSSTestSuite) TestVerifyOneTimeSignature() {
	suite.RunOnMember(suite.testCases, func(tc testutil.TestCase, member testutil.Member) {
		// Success case
		err := tss.VerifyOneTimeSignature(
			member.ID,
			tc.Group.DKGContext,
			member.OneTimeSignature,
			member.OneTimePubKey(),
		)
		suite.Require().NoError(err)

		// Wrong ID case
		err = tss.VerifyOneTimeSignature(0, tc.Group.DKGContext, member.OneTimeSignature, member.OneTimePubKey())
		suite.Require().ErrorIs(err, tss.ErrInvalidSignature)

		// Wrong DKGContext case
		err = tss.VerifyOneTimeSignature(
			member.ID,
			[]byte("fake DKGContext"),
			member.OneTimeSignature,
			member.OneTimePubKey(),
		)
		suite.Require().ErrorIs(err, tss.ErrInvalidSignature)

		// Wrong signature case
		err = tss.VerifyOneTimeSignature(
			member.ID,
			tc.Group.DKGContext,
			testutil.FakeSignature,
			member.OneTimePubKey(),
		)
		suite.Require().ErrorIs(err, tss.ErrInvalidSignature)

		// Wrong public key case
		err = tss.VerifyOneTimeSignature(member.ID, tc.Group.DKGContext, member.OneTimeSignature, testutil.FakePubKey)
		suite.Require().ErrorIs(err, tss.ErrInvalidSignature)
	})
}

func (suite *TSSTestSuite) TestSignA0() {
	suite.RunOnMember(suite.testCases, func(tc testutil.TestCase, member testutil.Member) {
		signature, err := tss.SignA0(
			member.ID,
			tc.Group.DKGContext,
			member.A0PubKey(),
			member.A0PrivKey,
		)
		suite.Require().NoError(err)

		err = tss.VerifyA0Signature(member.ID, tc.Group.DKGContext, signature, member.A0PubKey())
		suite.Require().NoError(err)
	})
}

func (suite *TSSTestSuite) TestVerifyA0Signature() {
	suite.RunOnMember(suite.testCases, func(tc testutil.TestCase, member testutil.Member) {
		// Success case
		err := tss.VerifyA0Signature(member.ID, tc.Group.DKGContext, member.A0Signature, member.A0PubKey())
		suite.Require().NoError(err)

		// Wrong ID case
		err = tss.VerifyA0Signature(0, tc.Group.DKGContext, member.A0Signature, member.A0PubKey())
		suite.Require().ErrorIs(err, tss.ErrInvalidSignature)

		// Wrong DKGContext case
		err = tss.VerifyA0Signature(member.ID, []byte("fake DKGContext"), member.A0Signature, member.A0PubKey())
		suite.Require().ErrorIs(err, tss.ErrInvalidSignature)

		// Wrong signature case
		err = tss.VerifyA0Signature(member.ID, tc.Group.DKGContext, testutil.FakeSignature, member.A0PubKey())
		suite.Require().ErrorIs(err, tss.ErrInvalidSignature)

		// Wrong public key case
		err = tss.VerifyA0Signature(member.ID, tc.Group.DKGContext, member.A0Signature, testutil.FakePubKey)
		suite.Require().ErrorIs(err, tss.ErrInvalidSignature)
	})
}
