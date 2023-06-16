package tss_test

import (
	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/pkg/tss/testutil"
)

func (suite *TSSTestSuite) TestGenerateRound1Info() {
	mid := tss.MemberID(1)
	dkgContext := []byte("DKGContext")
	threshold := uint64(2)

	data, err := tss.GenerateRound1Info(mid, threshold, dkgContext)
	suite.Require().NoError(err)
	err = tss.VerifyOneTimeSig(mid, dkgContext, data.OneTimeSig, data.OneTimePubKey)
	suite.Require().NoError(err)

	err = tss.VerifyA0Sig(mid, dkgContext, data.A0Sig, data.A0PubKey)
	suite.Require().NoError(err)

	for i, coeff := range data.Coefficients {
		commit, err := tss.PrivateKey(coeff).PublicKey()
		suite.Require().NoError(err)
		suite.Require().Equal(tss.PublicKey(data.CoefficientsCommit[i]), commit)
	}
}

func (suite *TSSTestSuite) TestSignOneTime() {
	suite.RunOnMember(suite.testCases, func(tc testutil.TestCase, member testutil.Member) {
		sig, err := tss.SignOneTime(
			member.ID,
			tc.Group.DKGContext,
			member.OneTimePubKey(),
			member.OneTimePrivKey,
		)
		suite.Require().NoError(err)
		suite.Require().Equal(member.OneTimeSig, sig)
	})
}

func (suite *TSSTestSuite) TestVerifyOneTimeSig() {
	suite.RunOnMember(suite.testCases, func(tc testutil.TestCase, member testutil.Member) {
		// Success case
		err := tss.VerifyOneTimeSig(member.ID, tc.Group.DKGContext, member.OneTimeSig, member.OneTimePubKey())
		suite.Require().NoError(err)

		// Wrong ID case
		err = tss.VerifyOneTimeSig(0, tc.Group.DKGContext, member.OneTimeSig, member.OneTimePubKey())
		suite.Require().ErrorIs(err, tss.ErrInvalidSignature)

		// Wrong DKGContext case
		err = tss.VerifyOneTimeSig(member.ID, []byte("fake DKGContext"), member.OneTimeSig, member.OneTimePubKey())
		suite.Require().ErrorIs(err, tss.ErrInvalidSignature)

		// Wrong signature case
		err = tss.VerifyOneTimeSig(member.ID, tc.Group.DKGContext, testutil.FakeSig, member.OneTimePubKey())
		suite.Require().ErrorIs(err, tss.ErrInvalidSignature)

		// Wrong public key case
		err = tss.VerifyOneTimeSig(member.ID, tc.Group.DKGContext, member.OneTimeSig, testutil.FakePubKey)
		suite.Require().ErrorIs(err, tss.ErrInvalidSignature)
	})
}

func (suite *TSSTestSuite) TestSignA0() {
	suite.RunOnMember(suite.testCases, func(tc testutil.TestCase, member testutil.Member) {
		sig, err := tss.SignA0(
			member.ID,
			tc.Group.DKGContext,
			member.A0PubKey(),
			member.A0PrivKey,
		)
		suite.Require().NoError(err)
		suite.Require().Equal(member.A0Sig, sig)
	})
}

func (suite *TSSTestSuite) TestVerifyA0Sig() {
	suite.RunOnMember(suite.testCases, func(tc testutil.TestCase, member testutil.Member) {
		// Success case
		err := tss.VerifyA0Sig(member.ID, tc.Group.DKGContext, member.A0Sig, member.A0PubKey())
		suite.Require().NoError(err)

		// Wrong ID case
		err = tss.VerifyA0Sig(0, tc.Group.DKGContext, member.A0Sig, member.A0PubKey())
		suite.Require().ErrorIs(err, tss.ErrInvalidSignature)

		// Wrong DKGContext case
		err = tss.VerifyA0Sig(member.ID, []byte("fake DKGContext"), member.A0Sig, member.A0PubKey())
		suite.Require().ErrorIs(err, tss.ErrInvalidSignature)

		// Wrong signature case
		err = tss.VerifyA0Sig(member.ID, tc.Group.DKGContext, testutil.FakeSig, member.A0PubKey())
		suite.Require().ErrorIs(err, tss.ErrInvalidSignature)

		// Wrong public key case
		err = tss.VerifyA0Sig(member.ID, tc.Group.DKGContext, member.A0Sig, testutil.FakePubKey)
		suite.Require().ErrorIs(err, tss.ErrInvalidSignature)
	})
}
