package tss_test

import (
	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/pkg/tss/testutil"
)

func (suite *TSSTestSuite) TestComputeOwnPublicKey() {
	suite.RunOnMember(suite.testCases, func(tc testutil.TestCase, member testutil.Member) {
		accCommits, err := tc.Group.GetAccumulatedCommits()
		suite.Require().NoError(err)

		pubKey, err := tss.ComputeOwnPublicKey(
			accCommits,
			member.ID,
		)
		suite.Require().NoError(err)
		suite.Require().Equal(member.PubKey(), pubKey)
	})
}

func (suite *TSSTestSuite) TestComputeGroupPublicKey() {
	for _, tc := range suite.testCases {
		suite.Run(tc.Name, func() {
			pubKey, err := tss.ComputeGroupPublicKey(tc.Group.GetCommits(0)...)
			suite.Require().NoError(err)
			suite.Require().Equal(tc.Group.PubKey, pubKey)
		})
	}
}

func (suite *TSSTestSuite) TestComputeOwnPrivateKey() {
	suite.RunOnMember(suite.testCases, func(tc testutil.TestCase, member testutil.Member) {
		var allSecrets []tss.Scalar
		ownSecret, err := tss.ComputeSecretShare(member.Coefficients, member.ID)
		allSecrets = append(allSecrets, ownSecret)

		for _, m := range tc.Group.Members {
			if m.ID == member.ID {
				continue
			}
			allSecrets = append(allSecrets, m.SecretShares[testutil.GetSlot(m.ID, member.ID)])
		}

		privKey, err := tss.ComputeOwnPrivateKey(allSecrets...)
		suite.Require().NoError(err)
		suite.Require().Equal(member.PrivKey, privKey)
	})
}

func (suite *TSSTestSuite) TestVerifySecretShare() {
	suite.RunOnPairMembers(
		suite.testCases,
		func(tc testutil.TestCase, memberI testutil.Member, memberJ testutil.Member) {
			err := tss.VerifySecretShare(
				memberJ.ID,
				memberI.SecretShares[testutil.GetSlot(memberI.ID, memberJ.ID)],
				memberI.CoefficientsCommit,
			)
			suite.Require().NoError(err)
		},
	)
}

func (suite *TSSTestSuite) TestComputeSecretShareCommit() {
	suite.RunOnPairMembers(
		suite.testCases,
		func(tc testutil.TestCase, memberI testutil.Member, memberJ testutil.Member) {
			secretCommit, err := tss.ComputeSecretShareCommit(memberI.CoefficientsCommit, memberJ.ID)
			suite.Require().NoError(err)

			expSecretCommit := memberI.SecretShares[testutil.GetSlot(memberI.ID, memberJ.ID)].Point()
			suite.Require().Equal(tss.Point(expSecretCommit), secretCommit)
		},
	)
}

func (suite *TSSTestSuite) TestDecryptSecretShares() {
	suite.RunOnMember(suite.testCases, func(tc testutil.TestCase, member testutil.Member) {
		secretShares, err := tss.DecryptSecretShares(member.EncSecretShares, member.KeySyms)
		suite.Require().NoError(err)
		suite.Require().Equal(member.SecretShares, secretShares)
	})
}

func (suite *TSSTestSuite) TestDecryptSecretShare() {
	suite.RunOnMember(suite.testCases, func(tc testutil.TestCase, member testutil.Member) {
		for i, encSecretShare := range member.EncSecretShares {
			secretShare, err := tss.DecryptSecretShare(encSecretShare, member.KeySyms[i])
			suite.Require().NoError(err)
			suite.Require().Equal(member.SecretShares[i], secretShare)
		}
	})
}

func (suite *TSSTestSuite) TestSignOwnPubKey() {
	suite.RunOnMember(suite.testCases, func(tc testutil.TestCase, member testutil.Member) {
		sig, err := tss.SignOwnPubkey(
			member.ID,
			tc.Group.DKGContext,
			member.PubKey(),
			member.PrivKey,
		)
		suite.Require().NoError(err)

		err = tss.VerifyOwnPubKeySig(member.ID, tc.Group.DKGContext, sig, member.PubKey())
		suite.Require().NoError(err)
	})
}

func (suite *TSSTestSuite) TestVerifyOwnPubKeySig() {
	suite.RunOnMember(suite.testCases, func(tc testutil.TestCase, member testutil.Member) {
		// Success case
		err := tss.VerifyOwnPubKeySig(member.ID, tc.Group.DKGContext, member.PubKeySig, member.PubKey())
		suite.Require().NoError(err)

		// Wrong ID case
		err = tss.VerifyOwnPubKeySig(0, tc.Group.DKGContext, member.PubKeySig, member.PubKey())
		suite.Require().ErrorIs(err, tss.ErrInvalidSignature)

		// Wrong DKGContext case
		err = tss.VerifyOwnPubKeySig(
			member.ID,
			[]byte("fake DKGContext"),
			member.PubKeySig,
			member.PubKey(),
		)
		suite.Require().ErrorIs(err, tss.ErrInvalidSignature)

		// Wrong signature case
		err = tss.VerifyOwnPubKeySig(member.ID, tc.Group.DKGContext, testutil.FakeSig, member.PubKey())
		suite.Require().ErrorIs(err, tss.ErrInvalidSignature)

		// Wrong public key case
		err = tss.VerifyOwnPubKeySig(member.ID, tc.Group.DKGContext, member.PubKeySig, testutil.FakePubKey)
		suite.Require().ErrorIs(err, tss.ErrInvalidSignature)
	})
}

func (suite *TSSTestSuite) TestSignComplaint() {
	suite.RunOnPairMembers(
		suite.testCases,
		func(tc testutil.TestCase, memberI testutil.Member, memberJ testutil.Member) {
			sig, keySym, err := tss.SignComplaint(
				memberI.OneTimePubKey(),
				memberJ.OneTimePubKey(),
				memberI.OneTimePrivKey,
			)
			suite.Require().NoError(err)

			err = tss.VerifyComplaintSig(memberI.OneTimePubKey(), memberJ.OneTimePubKey(), keySym, sig)
			suite.Require().NoError(err)

			suite.Require().
				Equal(memberI.KeySyms[testutil.GetSlot(memberI.ID, memberJ.ID)], keySym)
		})
}

func (suite *TSSTestSuite) TestVerifyComplaint() {
	suite.RunOnPairMembers(
		suite.testCases,
		func(tc testutil.TestCase, memberI testutil.Member, memberJ testutil.Member) {
			iSlot := testutil.GetSlot(memberI.ID, memberJ.ID)
			jSlot := testutil.GetSlot(memberJ.ID, memberI.ID)
			// Success case - wrong encrypted secret share
			err := tss.VerifyComplaint(
				memberI.OneTimePubKey(),
				memberJ.OneTimePubKey(),
				memberI.KeySyms[iSlot],
				memberI.ComplaintSigs[iSlot],
				testutil.FakePrivKey,
				memberI.ID,
				memberJ.CoefficientsCommit,
			)
			suite.Require().NoError(err)

			// Failed case - correct encrypted secret share
			err = tss.VerifyComplaint(
				memberI.OneTimePubKey(),
				memberJ.OneTimePubKey(),
				memberI.KeySyms[iSlot],
				memberI.ComplaintSigs[iSlot],
				memberJ.EncSecretShares[jSlot],
				memberI.ID,
				memberJ.CoefficientsCommit,
			)
			suite.Require().ErrorIs(err, tss.ErrValidSecretShare)
		})
}

func (suite *TSSTestSuite) TestVerifyComplaintSig() {
	suite.RunOnPairMembers(
		suite.testCases,
		func(tc testutil.TestCase, memberI testutil.Member, memberJ testutil.Member) {
			slot := testutil.GetSlot(memberI.ID, memberJ.ID)
			// Success case
			err := tss.VerifyComplaintSig(
				memberI.OneTimePubKey(),
				memberJ.OneTimePubKey(),
				memberI.KeySyms[slot],
				memberI.ComplaintSigs[slot],
			)
			suite.Require().NoError(err)

			// Wrong public key I case
			err = tss.VerifyComplaintSig(
				testutil.FakePubKey,
				memberJ.OneTimePubKey(),
				memberI.KeySyms[slot],
				memberI.ComplaintSigs[slot],
			)
			suite.Require().ErrorIs(err, tss.ErrInvalidSignature)

			// Wrong public key J case
			err = tss.VerifyComplaintSig(
				memberI.OneTimePubKey(),
				testutil.FakePubKey,
				memberI.KeySyms[slot],
				memberI.ComplaintSigs[slot],
			)
			suite.Require().ErrorIs(err, tss.ErrInvalidSignature)

			// Wrong key sym case
			err = tss.VerifyComplaintSig(
				memberI.OneTimePubKey(),
				memberJ.OneTimePubKey(),
				testutil.FakePubKey,
				memberI.ComplaintSigs[slot],
			)
			suite.Require().ErrorIs(err, tss.ErrInvalidSignature)

			// Wrong signature case
			err = tss.VerifyComplaintSig(
				memberI.OneTimePubKey(),
				memberJ.OneTimePubKey(),
				memberI.KeySyms[slot],
				testutil.FakeComplaintSig,
			)
			suite.Require().ErrorIs(err, tss.ErrInvalidSignature)
		})
}
